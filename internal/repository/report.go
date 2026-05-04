package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Report interface {
	Totals(ctx context.Context, userID uuid.UUID, from, to time.Time) (income, expense decimal.Decimal, err error)
	ByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.CategoryBreakdown, error)
	DailyTrend(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.DailyTrendPoint, error)
}

type ReportRepository struct {
	pg *postgres.Postgres
}

func NewReportRepository(pg *postgres.Postgres) *ReportRepository {
	return &ReportRepository{pg: pg}
}

func (r *ReportRepository) Totals(ctx context.Context, userID uuid.UUID, from, to time.Time) (decimal.Decimal, decimal.Decimal, error) {
	var income, expense decimal.Decimal
	err := r.pg.Pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.converted_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.converted_amount ELSE 0 END), 0)
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id
		WHERE a.user_id = $1 AND t.transaction_date >= $2 AND t.transaction_date <= $3`,
		userID, from, to,
	).Scan(&income, &expense)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("totals query: %w", err)
	}
	return income, expense, nil
}

func (r *ReportRepository) ByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.CategoryBreakdown, error) {
	rows, err := r.pg.Pool.Query(ctx, `
		SELECT c.id, c.name, c.type, COALESCE(SUM(t.converted_amount), 0)
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id
		WHERE a.user_id = $1 AND t.transaction_date >= $2 AND t.transaction_date <= $3
		GROUP BY c.id, c.name, c.type
		ORDER BY SUM(t.converted_amount) DESC`,
		userID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("by category query: %w", err)
	}
	defer rows.Close()
	out := make([]domain.CategoryBreakdown, 0)
	for rows.Next() {
		var b domain.CategoryBreakdown
		if err := rows.Scan(&b.CategoryID, &b.CategoryName, &b.CategoryType, &b.Total); err != nil {
			return nil, fmt.Errorf("scan breakdown: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *ReportRepository) DailyTrend(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.DailyTrendPoint, error) {
	rows, err := r.pg.Pool.Query(ctx, `
		SELECT
			t.transaction_date,
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.converted_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.converted_amount ELSE 0 END), 0)
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		JOIN categories c ON c.id = t.category_id
		WHERE a.user_id = $1 AND t.transaction_date >= $2 AND t.transaction_date <= $3
		GROUP BY t.transaction_date
		ORDER BY t.transaction_date`,
		userID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("daily trend query: %w", err)
	}
	defer rows.Close()
	out := make([]domain.DailyTrendPoint, 0)
	for rows.Next() {
		var p domain.DailyTrendPoint
		if err := rows.Scan(&p.Date, &p.Income, &p.Expense); err != nil {
			return nil, fmt.Errorf("scan trend: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
