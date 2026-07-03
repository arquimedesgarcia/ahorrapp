package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PriceAggregateRepository struct {
	pool *pgxpool.Pool
}

func NewPriceAggregateRepository(pool *pgxpool.Pool) *PriceAggregateRepository {
	return &PriceAggregateRepository{pool: pool}
}

func (r *PriceAggregateRepository) RecomputeAggregate(
	ctx context.Context,
	productID, storeID, currency string,
	ageThresholdDays int,
) error {
	if ageThresholdDays < 1 {
		ageThresholdDays = 90
	}

	query := `
	SELECT
		COALESCE(AVG(unit_price), 0),
		COALESCE(MIN(unit_price), 0),
		COUNT(*),
		COALESCE(MAX(observed_at), NOW())
	FROM price_observations
	WHERE product_id = $1::uuid
	  AND store_id = $2::uuid
	  AND currency = $3
	  AND observed_at >= NOW() - make_interval(days => $4)
	`

	var avg, min float64
	var count int
	var lastObserved time.Time
	err := r.pool.QueryRow(ctx, query, productID, storeID, currency, ageThresholdDays).Scan(&avg, &min, &count, &lastObserved)
	if err != nil {
		return fmt.Errorf("query observations: %w", err)
	}

	// No fresh observations: drop the cached row so the ranking does
	// not display stale data. Upserting a (count=0, avg=0, min=0) row
	// would pollute the table and the ranking query already filters
	// sample_count > 0, but DELETE is the cleaner source of truth.
	if count == 0 {
		_, err = r.pool.Exec(ctx, `
		DELETE FROM price_aggregates
		WHERE product_id = $1::uuid AND store_id = $2::uuid AND currency = $3
		`, productID, storeID, currency)
		if err != nil {
			return fmt.Errorf("delete stale aggregate: %w", err)
		}
		return nil
	}

	upsert := `
	INSERT INTO price_aggregates (product_id, store_id, currency, average_price, min_price, sample_count, last_observed_at, updated_at)
	VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, NOW())
	ON CONFLICT (product_id, store_id, currency)
	DO UPDATE SET
		average_price    = EXCLUDED.average_price,
		min_price        = EXCLUDED.min_price,
		sample_count     = EXCLUDED.sample_count,
		last_observed_at = EXCLUDED.last_observed_at,
		updated_at       = NOW()
	`
	_, err = r.pool.Exec(ctx, upsert, productID, storeID, currency, avg, min, count, lastObserved)
	if err != nil {
		return fmt.Errorf("upsert aggregate: %w", err)
	}
	return nil
}

func (r *PriceAggregateRepository) RecomputeAll(ctx context.Context, ageThresholdDays int) error {
	if ageThresholdDays < 1 {
		ageThresholdDays = 90
	}

	// Two-step recompute so the resulting table contains only rows
	// that have at least one fresh observation. INSERT ... ON CONFLICT
	// cannot represent "delete when the new row would be empty", so we
	// first aggregate, then delete the rows that fell out of the window.
	insertQuery := `
	INSERT INTO price_aggregates (product_id, store_id, currency, average_price, min_price, sample_count, last_observed_at, updated_at)
	SELECT
		product_id,
		store_id,
		currency,
		AVG(unit_price),
		MIN(unit_price),
		COUNT(*),
		MAX(observed_at)
	FROM price_observations
	WHERE observed_at >= NOW() - make_interval(days => $1)
	GROUP BY product_id, store_id, currency
	ON CONFLICT (product_id, store_id, currency)
	DO UPDATE SET
		average_price    = EXCLUDED.average_price,
		min_price        = EXCLUDED.min_price,
		sample_count     = EXCLUDED.sample_count,
		last_observed_at = EXCLUDED.last_observed_at,
		updated_at       = NOW()
	`
	if _, err := r.pool.Exec(ctx, insertQuery, ageThresholdDays); err != nil {
		return fmt.Errorf("recompute all aggregates: %w", err)
	}
	// Drop rows whose only observations have aged out of the window.
	if _, err := r.pool.Exec(ctx, `
	DELETE FROM price_aggregates pa
	WHERE NOT EXISTS (
	  SELECT 1 FROM price_observations po
	  WHERE po.product_id = pa.product_id
	    AND po.store_id = pa.store_id
	    AND po.currency = pa.currency
	    AND po.observed_at >= NOW() - make_interval(days => $1)
	)
	`, ageThresholdDays); err != nil {
		return fmt.Errorf("prune stale aggregates: %w", err)
	}
	return nil
}

func (r *PriceAggregateRepository) GetProductRanking(
	ctx context.Context,
	productID string,
	opts ports.RankingQueryOptions,
) ([]entities.PriceAggregate, error) {
	if opts.HasLocation() {
		return r.getProductRankingWithProximity(ctx, productID, opts)
	}

	return r.getProductRankingDefault(ctx, productID, opts)
}

func (r *PriceAggregateRepository) getProductRankingDefault(
	ctx context.Context,
	productID string,
	opts ports.RankingQueryOptions,
) ([]entities.PriceAggregate, error) {
	_ = opts

	query := `
	SELECT
		pa.store_id::text,
		s.name,
		s.branch,
		pa.currency,
		pa.average_price,
		pa.min_price,
		pa.sample_count,
		pa.last_observed_at,
		NULL::float8 AS distance_km
	FROM price_aggregates pa
	JOIN stores s ON s.id = pa.store_id
	WHERE pa.product_id = $1::uuid
	  AND pa.sample_count > 0
	ORDER BY pa.average_price ASC, s.name ASC
	`

	return r.queryRanking(ctx, query, productID)
}

func (r *PriceAggregateRepository) getProductRankingWithProximity(
	ctx context.Context,
	productID string,
	opts ports.RankingQueryOptions,
) ([]entities.PriceAggregate, error) {
	if opts.Lat == nil || opts.Long == nil {
		return r.getProductRankingDefault(ctx, productID, opts)
	}

	// All numeric inputs are passed as $N parameters; no string
	// interpolation of user data into the SQL. The user point is
	// expressed via ST_MakePoint(long, lat) per the PostGIS convention
	// (longitude first).
	userPoint := "ST_MakePoint($2, $3)::geography"

	var query string
	args := []any{productID, *opts.Long, *opts.Lat}
	if opts.HasRadius() && opts.RadiusKm != nil {
		query = fmt.Sprintf(`
			SELECT
				pa.store_id::text,
				s.name,
				s.branch,
				pa.currency,
				pa.average_price,
				pa.min_price,
				pa.sample_count,
				pa.last_observed_at,
				CASE
					WHEN s.geo IS NOT NULL
						THEN ST_Distance(s.geo, %s) / 1000.0
					ELSE NULL
				END AS distance_km
			FROM price_aggregates pa
			JOIN stores s ON s.id = pa.store_id
			WHERE pa.product_id = $1::uuid
			  AND pa.sample_count > 0
			  AND (s.geo IS NULL OR ST_DWithin(s.geo, %s, $4))
			ORDER BY
				CASE WHEN distance_km IS NULL THEN 1 ELSE 0 END,
				distance_km ASC,
				pa.average_price ASC,
				s.name ASC
			`, userPoint, userPoint)
		args = append(args, *opts.RadiusKm*1000)
	} else {
		query = fmt.Sprintf(`
			SELECT
				pa.store_id::text,
				s.name,
				s.branch,
				pa.currency,
				pa.average_price,
				pa.min_price,
				pa.sample_count,
				pa.last_observed_at,
				CASE
					WHEN s.geo IS NOT NULL
						THEN ST_Distance(s.geo, %s) / 1000.0
					ELSE NULL
				END AS distance_km
			FROM price_aggregates pa
			JOIN stores s ON s.id = pa.store_id
			WHERE pa.product_id = $1::uuid
			  AND pa.sample_count > 0
			ORDER BY
				CASE WHEN distance_km IS NULL THEN 1 ELSE 0 END,
				distance_km ASC,
				pa.average_price ASC,
				s.name ASC
			`, userPoint)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query ranking with proximity: %w", err)
	}
	defer rows.Close()

	var results []entities.PriceAggregate
	for rows.Next() {
		var agg entities.PriceAggregate
		var branch *string
		var distanceKm *float64
		if err := rows.Scan(&agg.StoreID, &agg.StoreName, &branch, &agg.Currency, &agg.AveragePrice, &agg.MinPrice, &agg.SampleCount, &agg.LastObservedAt, &distanceKm); err != nil {
			return nil, fmt.Errorf("scan ranking row: %w", err)
		}
		agg.Branch = branch
		agg.DistanceKm = distanceKm
		results = append(results, agg)
	}
	return results, rows.Err()
}

func (r *PriceAggregateRepository) queryRanking(ctx context.Context, query, productID string) ([]entities.PriceAggregate, error) {
	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("query ranking: %w", err)
	}
	defer rows.Close()

	var results []entities.PriceAggregate
	for rows.Next() {
		var agg entities.PriceAggregate
		var branch *string
		var distanceKm *float64
		if err := rows.Scan(&agg.StoreID, &agg.StoreName, &branch, &agg.Currency, &agg.AveragePrice, &agg.MinPrice, &agg.SampleCount, &agg.LastObservedAt, &distanceKm); err != nil {
			return nil, fmt.Errorf("scan ranking row: %w", err)
		}
		agg.Branch = branch
		agg.DistanceKm = distanceKm
		results = append(results, agg)
	}
	return results, rows.Err()
}

func (r *PriceAggregateRepository) SearchProducts(
	ctx context.Context,
	query string,
) ([]entities.ProductSearchResult, error) {
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if len(normalizedQuery) < 3 {
		return nil, fmt.Errorf("search query must be at least 3 characters")
	}

	productQuery := `
	SELECT id::text, canonical_name
	FROM products
	WHERE unaccent(canonical_name) ILIKE unaccent('%' || $1 || '%')
	ORDER BY canonical_name ASC
	`
	rows, err := r.pool.Query(ctx, productQuery, normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	defer rows.Close()

	type productMatch struct {
		id   string
		name string
	}
	var matches []productMatch
	for rows.Next() {
		var m productMatch
		if err := rows.Scan(&m.id, &m.name); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		matches = append(matches, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return []entities.ProductSearchResult{}, nil
	}

	results := make([]entities.ProductSearchResult, 0, len(matches))
	for _, m := range matches {
		storeRows, err := r.pool.Query(ctx, `
			SELECT
				pa.store_id::text,
				s.name,
				s.branch,
				pa.currency,
				pa.average_price,
				pa.min_price,
				pa.sample_count,
				pa.last_observed_at
			FROM price_aggregates pa
			JOIN stores s ON s.id = pa.store_id
			WHERE pa.product_id = $1::uuid
			  AND pa.sample_count > 0
			ORDER BY pa.average_price ASC, s.name ASC
			`, m.id)
		if err != nil {
			return nil, fmt.Errorf("query best prices for product %s: %w", m.id, err)
		}

		bestPrices := make(map[string]*entities.PriceAggregate)
		for storeRows.Next() {
			var agg entities.PriceAggregate
			var branch *string
			if err := storeRows.Scan(&agg.StoreID, &agg.StoreName, &branch, &agg.Currency, &agg.AveragePrice, &agg.MinPrice, &agg.SampleCount, &agg.LastObservedAt); err != nil {
				storeRows.Close()
				return nil, fmt.Errorf("scan best price: %w", err)
			}
			agg.ProductID = m.id
			agg.Branch = branch
			if existing, ok := bestPrices[agg.Currency]; !ok {
				bestPrices[agg.Currency] = &agg
			} else {
				if agg.AveragePrice < existing.AveragePrice {
					bestPrices[agg.Currency] = &agg
				}
			}
		}
		storeRows.Close()

		results = append(results, entities.ProductSearchResult{
			ProductID:   m.id,
			ProductName: m.name,
			Unit:        nil,
			BestPrices:  bestPrices,
		})
	}

	return results, nil
}

func (r *PriceAggregateRepository) GetProductName(ctx context.Context, productID string) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx, `SELECT canonical_name FROM products WHERE id = $1::uuid`, productID).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ports.ErrProductNotFound
		}
		return "", fmt.Errorf("get product name: %w", err)
	}
	return name, nil
}
