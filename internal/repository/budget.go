package repository

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type Budget interface {
	Create(ctx context.Context, b *domain.Budget) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error)
	GetByCategoryTx(ctx context.Context, tx pgx.Tx, userID, categoryID uuid.UUID) (*domain.Budget, error)
	List(ctx context.Context, userID uuid.UUID, period string) ([]domain.Budget, error)
	Update(ctx context.Context, userID, id uuid.UUID, limit *decimal.Decimal, period string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	UpdateSpentTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta decimal.Decimal) error
	ResetPeriod(ctx context.Context, id uuid.UUID, periodStart string) error
}

type BudgetRepository struct {
	pg *postgres.Postgres
}

func NewBudgetRepository(pg *postgres.Postgres) *BudgetRepository {
	return &BudgetRepository{pg: pg}
}

func (r *BudgetRepository) Create(ctx context.Context, b *domain.Budget) error {
	query, args, err := r.pg.Builder.
		Insert("budgets").
		Columns("user_id", "category_id", "spending_limit", "period").
		Values(b.UserID, b.CategoryID, b.SpendingLimit, b.Period).
		Suffix("RETURNING id, spent, period_start, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert budget: %w", err)
	}
	if err := r.pg.Pool.QueryRow(ctx, query, args...).Scan(&b.ID, &b.Spent, &b.PeriodStart, &b.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert budget: %w", err)
	}
	return nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Budget, error) {
	query, args, err := r.pg.Builder.
		Select("id", "user_id", "category_id", "spending_limit", "period", "spent", "period_start", "created_at").
		From("budgets").
		Where(sq.Eq{"id": id, "user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select budget: %w", err)
	}
	return scanBudget(r.pg.Pool.QueryRow(ctx, query, args...))
}

func (r *BudgetRepository) GetByCategoryTx(ctx context.Context, tx pgx.Tx, userID, categoryID uuid.UUID) (*domain.Budget, error) {
	query, args, err := r.pg.Builder.
		Select("id", "user_id", "category_id", "spending_limit", "period", "spent", "period_start", "created_at").
		From("budgets").
		Where(sq.Eq{"user_id": userID, "category_id": categoryID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select budget tx: %w", err)
	}
	return scanBudget(tx.QueryRow(ctx, query, args...))
}

func (r *BudgetRepository) List(ctx context.Context, userID uuid.UUID, period string) ([]domain.Budget, error) {
	builder := r.pg.Builder.
		Select("id", "user_id", "category_id", "spending_limit", "period", "spent", "period_start", "created_at").
		From("budgets").
		Where("user_id = ?", userID).
		OrderBy("created_at DESC")
	if period != "" {
		builder = builder.Where("period = ?", period)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list budgets: %w", err)
	}
	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query budgets: %w", err)
	}
	defer rows.Close()
	var out []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.CategoryID, &b.SpendingLimit, &b.Period, &b.Spent, &b.PeriodStart, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BudgetRepository) Update(ctx context.Context, userID, id uuid.UUID, limit *decimal.Decimal, period string) error {
	builder := r.pg.Builder.Update("budgets").Where("id = ? AND user_id = ?", id, userID)
	if limit != nil {
		builder = builder.Set("spending_limit", *limit)
	}
	if period != "" {
		builder = builder.Set("period", period)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build update budget: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update budget: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BudgetRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	query, args, err := r.pg.Builder.
		Delete("budgets").
		Where("id = ? AND user_id = ?", id, userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete budget: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BudgetRepository) UpdateSpentTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta decimal.Decimal) error {
	_, err := tx.Exec(ctx,
		"UPDATE budgets SET spent = GREATEST(spent + $1, 0) WHERE id = $2",
		delta, id,
	)
	if err != nil {
		return fmt.Errorf("update budget spent: %w", err)
	}
	return nil
}

func (r *BudgetRepository) ResetPeriod(ctx context.Context, id uuid.UUID, periodStart string) error {
	_, err := r.pg.Pool.Exec(ctx,
		"UPDATE budgets SET spent = 0, period_start = $1::date WHERE id = $2",
		periodStart, id,
	)
	if err != nil {
		return fmt.Errorf("reset budget period: %w", err)
	}
	return nil
}

func scanBudget(row pgx.Row) (*domain.Budget, error) {
	var b domain.Budget
	if err := row.Scan(&b.ID, &b.UserID, &b.CategoryID, &b.SpendingLimit, &b.Period, &b.Spent, &b.PeriodStart, &b.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan budget: %w", err)
	}
	return &b, nil
}
