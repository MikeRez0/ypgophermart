package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/MikeRez0/ypgophermart/internal/adapter/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
	dsn          string
	QueryBuilder *squirrel.StatementBuilderType
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func NewDBStorage(ctx context.Context, config *config.Database) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	dbs := DB{
		Pool:         pool,
		dsn:          config.DSN,
		QueryBuilder: &psql,
	}

	return &dbs, nil
}

func (db *DB) RunMigrations() error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, db.dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
