package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ExchangeRate interface {
	Upsert(ctx context.Context, base, target string, rate decimal.Decimal) error
	Get(ctx context.Context, base, target string) (*domain.ExchangeRate, error)
}

type ExchangeRateRepository struct {
	pg *postgres.Postgres
}

func NewExchangeRateRepository(pg *postgres.Postgres) *ExchangeRateRepository {
	return &ExchangeRateRepository{pg: pg}
}

func (r *ExchangeRateRepository) Upsert(ctx context.Context, base, target string, rate decimal.Decimal) error {
	_, err := r.pg.Pool.Exec(ctx, `
		INSERT INTO exchange_rates (base_currency, target_currency, rate, cached_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (base_currency, target_currency)
		DO UPDATE SET rate = EXCLUDED.rate, cached_at = NOW()`,
		base, target, rate,
	)
	if err != nil {
		return fmt.Errorf("upsert exchange rate: %w", err)
	}
	return nil
}

func (r *ExchangeRateRepository) Get(ctx context.Context, base, target string) (*domain.ExchangeRate, error) {
	var er domain.ExchangeRate
	err := r.pg.Pool.QueryRow(ctx,
		"SELECT base_currency, target_currency, rate, cached_at FROM exchange_rates WHERE base_currency = $1 AND target_currency = $2",
		base, target,
	).Scan(&er.BaseCurrency, &er.TargetCurrency, &er.Rate, &er.CachedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get exchange rate: %w", err)
	}
	return &er, nil
}
