package postgres

import (
	"context"
	"strings"

	"ahorrapp/internal/domain/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReceiptRepository struct {
	pool *pgxpool.Pool
}

func NewReceiptRepository(pool *pgxpool.Pool) *ReceiptRepository {
	return &ReceiptRepository{pool: pool}
}

func (r *ReceiptRepository) FindByUserAndImageHash(ctx context.Context, userID, imageHash string) (*entities.Receipt, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id::text, user_id, image_url, image_hash, status, created_at, updated_at
FROM receipts
WHERE user_id = $1 AND image_hash = $2
`, userID, imageHash)

	var out entities.Receipt
	var status string
	err := row.Scan(&out.ID, &out.UserID, &out.ImageURL, &out.ImageHash, &status, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	out.Status = entities.ReceiptStatus(status)
	return &out, nil
}

func (r *ReceiptRepository) CreatePendingReceipt(ctx context.Context, userID, imageURL, imageHash string) (*entities.Receipt, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
INSERT INTO receipts (user_id, image_url, image_hash, status)
VALUES ($1, $2, $3, 'PENDING')
RETURNING id::text, created_at, updated_at
`, userID, imageURL, imageHash)

	var out entities.Receipt
	out.UserID = userID
	out.ImageURL = imageURL
	out.ImageHash = imageHash
	out.Status = entities.ReceiptStatusPending
	if err := row.Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO ocr_jobs (receipt_id, status)
VALUES ($1::uuid, 'QUEUED')
ON CONFLICT (receipt_id) DO UPDATE SET status = 'QUEUED', attempt = 0, last_error = NULL
`, out.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &out, nil
}

func (r *ReceiptRepository) GetByIDForUser(ctx context.Context, receiptID, userID string) (*entities.EditableSummary, error) {
	row := r.pool.QueryRow(ctx, `
SELECT r.id::text, r.status, r.purchase_date::text, r.total,
       COALESCE(s.name, ''), s.branch, s.address
FROM receipts r
LEFT JOIN stores s ON s.id = r.store_id
WHERE r.id = $1::uuid AND r.user_id = $2
`, receiptID, userID)

	var out entities.EditableSummary
	var status string
	var purchaseDate *string
	var total *float64
	if err := row.Scan(&out.ReceiptID, &status, &purchaseDate, &total, &out.Store.Name, &out.Store.Branch, &out.Store.Address); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	out.Status = entities.ReceiptStatus(status)
	out.PurchaseDate = purchaseDate
	out.Total = total

	itemsRows, err := r.pool.Query(ctx, `
SELECT raw_text, quantity, unit_price, currency
FROM receipt_items
WHERE receipt_id = $1::uuid
ORDER BY created_at ASC
`, receiptID)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	out.Items = make([]entities.EditableItem, 0)
	for itemsRows.Next() {
		var item entities.EditableItem
		if err := itemsRows.Scan(&item.RawText, &item.Quantity, &item.UnitPrice, &item.Currency); err != nil {
			return nil, err
		}
		out.Items = append(out.Items, item)
	}

	return &out, itemsRows.Err()
}

func (r *ReceiptRepository) GetByID(ctx context.Context, receiptID string) (*entities.Receipt, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id::text, user_id, image_url, image_hash, status, created_at, updated_at
FROM receipts
WHERE id = $1::uuid
`, receiptID)

	var out entities.Receipt
	var status string
	err := row.Scan(&out.ID, &out.UserID, &out.ImageURL, &out.ImageHash, &status, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	out.Status = entities.ReceiptStatus(status)
	return &out, nil
}

func (r *ReceiptRepository) MarkNeedsReview(ctx context.Context, receiptID string, summary entities.EditableSummary) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	storeID, err := r.resolveOrCreateStoreTx(ctx, tx, summary.Store)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
UPDATE receipts
SET status = 'NEEDS_REVIEW', store_id = $2::uuid, purchase_date = $3::date, total = $4, updated_at = NOW()
WHERE id = $1::uuid
`, receiptID, storeID, summary.PurchaseDate, summary.Total)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM receipt_items WHERE receipt_id = $1::uuid`, receiptID); err != nil {
		return err
	}

	for _, item := range summary.Items {
		_, err := tx.Exec(ctx, `
INSERT INTO receipt_items (receipt_id, raw_text, normalized_name, quantity, unit_price, currency, line_total)
VALUES ($1::uuid, $2, $3, $4, $5, $6, NULL)
`, receiptID, item.RawText, normalizeName(item.RawText), item.Quantity, item.UnitPrice, item.Currency)
		if err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
UPDATE ocr_jobs
SET status = 'DONE', attempt = attempt + 1, processed_at = NOW(), last_error = NULL
WHERE receipt_id = $1::uuid
`, receiptID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ReceiptRepository) ConfirmReceipt(ctx context.Context, receiptID, userID string, payload entities.ConfirmPayload, observations []entities.PriceObservation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	storeID, err := r.resolveOrCreateStoreTx(ctx, tx, payload.Store)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
UPDATE receipts
SET status = 'CONFIRMED', store_id = $3::uuid, purchase_date = $4::date, total = $5, updated_at = NOW()
WHERE id = $1::uuid AND user_id = $2
`, receiptID, userID, storeID, payload.PurchaseDate, payload.Total); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM receipt_items WHERE receipt_id = $1::uuid`, receiptID); err != nil {
		return err
	}

	for _, item := range payload.Items {
		productID, normalizedName, err := r.normalizeProductTx(ctx, tx, item.RawText)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO receipt_items (receipt_id, raw_text, normalized_name, product_id, quantity, unit_price, currency, line_total)
VALUES ($1::uuid, $2, $3, $4::uuid, $5, $6, $7, NULL)
`, receiptID, item.RawText, normalizedName, productID, item.Quantity, item.UnitPrice, item.Currency)
		if err != nil {
			return err
		}
	}

	for _, obs := range observations {
		_, err := tx.Exec(ctx, `
INSERT INTO price_observations (product_id, store_id, unit_price, currency, observed_at, receipt_id)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6::uuid)
`, obs.ProductID, obs.StoreID, obs.UnitPrice, obs.Currency, obs.ObservedAt, obs.ReceiptID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ReceiptRepository) ResolveOrCreateStore(ctx context.Context, store entities.StoreSummary) (string, error) {
	return r.resolveOrCreateStoreTx(ctx, r.pool, store)
}

func (r *ReceiptRepository) NormalizeProduct(ctx context.Context, rawName string) (string, string, error) {
	return r.normalizeProductTx(ctx, r.pool, rawName)
}

type dbtx interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *ReceiptRepository) resolveOrCreateStoreTx(ctx context.Context, q dbtx, store entities.StoreSummary) (string, error) {
	name := strings.TrimSpace(store.Name)
	if name == "" {
		name = "Unknown"
	}
	row := q.QueryRow(ctx, `
SELECT id::text
FROM stores
WHERE name = $1 AND COALESCE(branch, '') = COALESCE($2, '') AND COALESCE(address, '') = COALESCE($3, '')
LIMIT 1
`, name, store.Branch, store.Address)

	var existing string
	err := row.Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if err != pgx.ErrNoRows {
		return "", err
	}

	insert := q.QueryRow(ctx, `
INSERT INTO stores (name, branch, address)
VALUES ($1, $2, $3)
RETURNING id::text
`, name, store.Branch, store.Address)

	var storeID string
	if err := insert.Scan(&storeID); err != nil {
		return "", err
	}
	return storeID, nil
}

func (r *ReceiptRepository) normalizeProductTx(ctx context.Context, q dbtx, rawName string) (string, string, error) {
	canonical := normalizeName(rawName)
	row := q.QueryRow(ctx, `
SELECT id::text, canonical_name
FROM products
WHERE canonical_name = $1
`, canonical)

	var id, name string
	err := row.Scan(&id, &name)
	if err == nil {
		return id, name, nil
	}
	if err != pgx.ErrNoRows {
		return "", "", err
	}

	insert := q.QueryRow(ctx, `
INSERT INTO products (canonical_name)
VALUES ($1)
RETURNING id::text, canonical_name
`, canonical)
	if err := insert.Scan(&id, &name); err != nil {
		return "", "", err
	}
	return id, name, nil
}

func normalizeName(in string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(in)))
	return strings.Join(fields, " ")
}
