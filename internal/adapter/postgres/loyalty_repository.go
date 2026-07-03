package postgres

import (
	"context"
	"errors"
	"fmt"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LoyaltyRepository struct {
	pool *pgxpool.Pool
}

func NewLoyaltyRepository(pool *pgxpool.Pool) *LoyaltyRepository {
	return &LoyaltyRepository{pool: pool}
}

// AwardForReceipt inserts one movement tied to the receipt and atomically
// increments users.points. The unique partial index on
// loyalty_transactions(receipt_id) makes this idempotent: a second call
// with the same receipt_id fails with code 23505 and is translated to
// ports.ErrAlreadyAwarded. Even zero-point daily-limit rows reference the
// receipt so the user's history explains why no points were granted.
func (r *LoyaltyRepository) AwardForReceipt(ctx context.Context, userID, receiptID string, points int, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
INSERT INTO loyalty_transactions (user_id, receipt_id, points, reason, created_at)
VALUES ($1::uuid, $2::uuid, $3, $4, NOW())
`, userID, receiptID, points, reason)
	if err != nil {
		if isUniqueViolation(err) {
			return ports.ErrAlreadyAwarded
		}
		return fmt.Errorf("insert loyalty transaction: %w", err)
	}

	if points != 0 {
		if _, err := tx.Exec(ctx, `
UPDATE users SET points = points + $2, updated_at = NOW() WHERE id::text = $1
`, userID, points); err != nil {
			return fmt.Errorf("update user points: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *LoyaltyRepository) DailyGrantCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
SELECT COUNT(DISTINCT receipt_id)
FROM loyalty_transactions
WHERE user_id::text = $1
  AND points > 0
  AND created_at >= date_trunc('day', NOW())
`, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("daily grant count: %w", err)
	}
	return count, nil
}

func (r *LoyaltyRepository) Balance(ctx context.Context, userID string) (int, error) {
	var points int
	err := r.pool.QueryRow(ctx, `SELECT points FROM users WHERE id::text = $1`, userID).Scan(&points)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ports.ErrUserNotFound
		}
		return 0, fmt.Errorf("balance: %w", err)
	}
	return points, nil
}

func (r *LoyaltyRepository) History(ctx context.Context, userID string, limit int) ([]entities.LoyaltyTransaction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id::text, user_id::text, points, reason, created_at, receipt_id::text
FROM loyalty_transactions
WHERE user_id::text = $1
ORDER BY created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("loyalty history: %w", err)
	}
	defer rows.Close()

	out := make([]entities.LoyaltyTransaction, 0)
	for rows.Next() {
		var tx entities.LoyaltyTransaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Points, &tx.Reason, &tx.CreatedAt, &tx.ReceiptID); err != nil {
			return nil, fmt.Errorf("scan loyalty transaction: %w", err)
		}
		out = append(out, tx)
	}
	return out, rows.Err()
}

func (r *LoyaltyRepository) ContributorStats(ctx context.Context, userID string) (ports.ContributorStats, error) {
	var stats ports.ContributorStats

	err := r.pool.QueryRow(ctx, `
SELECT
  (SELECT COUNT(*) FROM receipts WHERE user_id::text = $1 AND status = 'CONFIRMED') AS receipts_confirmed,
  (SELECT COUNT(*) FROM price_observations po
     JOIN receipts r ON r.id = po.receipt_id
     WHERE r.user_id::text = $1) AS price_observations,
  (SELECT COUNT(DISTINCT r.store_id) FROM receipts r
     WHERE r.user_id::text = $1 AND r.store_id IS NOT NULL) AS unique_stores,
  (SELECT COUNT(DISTINCT ri.product_id) FROM receipt_items ri
     JOIN receipts r ON r.id = ri.receipt_id
     WHERE r.user_id::text = $1 AND ri.product_id IS NOT NULL) AS unique_products
`, userID).Scan(
		&stats.ReceiptsConfirmed,
		&stats.PriceObservations,
		&stats.UniqueStores,
		&stats.UniqueProducts,
	)
	if err != nil {
		return ports.ContributorStats{}, fmt.Errorf("contributor stats: %w", err)
	}
	return stats, nil
}
