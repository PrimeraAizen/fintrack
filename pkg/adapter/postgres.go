package postgres

import (
	"context"
	"fmt"

	"github.com/diyas/fintrack/config"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.PG) (*Postgres, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Postgres config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	// if err := RunMigrations(poolConfig.ConnString(), path.Join("migrations", "postgres"), "cdp.goose_db_version"); err != nil {
	// 	return nil, err
	// }

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &Postgres{
		Builder: builder,
		Pool:    pool,
	}, nil
}

func (db *Postgres) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
