package db

import (
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/levinOo/geo-incedent-service/migrations"
	"github.com/pressly/goose/v3"
)

func RunMigrations(pg *Postgres) error {
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	poolConfig := pg.Pool.Config()

	stdDB := stdlib.OpenDB(*poolConfig.ConnConfig)
	defer stdDB.Close()

	if err := goose.Up(stdDB, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
