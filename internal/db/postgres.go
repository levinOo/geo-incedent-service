package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/levinOo/geo-incedent-service/config"
)

const (
	connectTimeout = 5 * time.Second
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(cfg *config.PostgresConfig) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	pgConfig, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	pgPool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := pgPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &Postgres{
		Pool: pgPool,
	}, nil
}

func (p *Postgres) Close() error {
	p.Pool.Close()
	slog.Info("postgres connection closed")
	return nil
}
