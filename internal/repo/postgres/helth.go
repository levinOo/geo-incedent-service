package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepo interface {
	Ping(ctx context.Context) error
}

type HealthRepoImpl struct {
	pool *pgxpool.Pool
}

func NewHealthRepoImpl(pool *pgxpool.Pool) HealthRepo {
	return &HealthRepoImpl{pool: pool}
}

func (r *HealthRepoImpl) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
