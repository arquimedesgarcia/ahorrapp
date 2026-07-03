package postgres

import (
	"context"
	"errors"
	"fmt"

	"ahorrapp/internal/domain/entities"
	"ahorrapp/internal/domain/ports"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash, displayName string) (*entities.User, error) {
	row := r.pool.QueryRow(ctx, `
INSERT INTO users (email, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING id::text, email, password_hash, display_name, points, created_at, updated_at
`, email, passwordHash, displayName)

	var u entities.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Points, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, ports.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id::text, email, password_hash, display_name, points, created_at, updated_at
FROM users WHERE email = $1
`, email)

	var u entities.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Points, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id::text, email, password_hash, display_name, points, created_at, updated_at
FROM users WHERE id::text = $1
`, id)

	var u entities.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Points, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetPoints(ctx context.Context, userID string) (int, error) {
	var points int
	err := r.pool.QueryRow(ctx, `SELECT points FROM users WHERE id::text = $1`, userID).Scan(&points)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ports.ErrUserNotFound
		}
		return 0, fmt.Errorf("get points: %w", err)
	}
	return points, nil
}

func (r *UserRepository) RecentTransactions(ctx context.Context, userID string, limit int) ([]entities.LoyaltyTransaction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id::text, user_id::text, points, reason, created_at, receipt_id::text
FROM loyalty_transactions
WHERE user_id::text = $1
ORDER BY created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recent transactions: %w", err)
	}
	defer rows.Close()

	var out []entities.LoyaltyTransaction
	for rows.Next() {
		var tx entities.LoyaltyTransaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Points, &tx.Reason, &tx.CreatedAt, &tx.ReceiptID); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		out = append(out, tx)
	}
	return out, nil
}

// isUniqueViolation checks if the error is a Postgres unique constraint
// violation (SQLSTATE 23505). Uses the typed *pgconn.PgError so the
// detection is exact and immune to false positives from any error
// message that happens to contain the words "unique" or "duplicate".
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
