package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Checker struct {
	pool *pgxpool.Pool
}

func NewChecker(pool *pgxpool.Pool) *Checker {
	return &Checker{pool: pool}
}

func (c *Checker) Name() string {
	return "postgres"
}

func (c *Checker) Check(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
