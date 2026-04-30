package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type SavingGoal interface {
	Create(ctx context.Context, g *domain.SavingGoal) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.SavingGoal, error)
	List(ctx context.Context, userID uuid.UUID) ([]domain.SavingGoal, error)
	Update(ctx context.Context, userID, id uuid.UUID, name string, target *decimal.Decimal, deadline *string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	AddContribution(ctx context.Context, userID, id uuid.UUID, amount decimal.Decimal) (*domain.SavingGoal, error)
}

type SavingGoalRepository struct {
	pg *postgres.Postgres
}

func NewSavingGoalRepository(pg *postgres.Postgres) *SavingGoalRepository {
	return &SavingGoalRepository{pg: pg}
}

func (r *SavingGoalRepository) Create(ctx context.Context, g *domain.SavingGoal) error {
	query, args, err := r.pg.Builder.
		Insert("saving_goals").
		Columns("user_id", "name", "target_amount", "current_amount", "deadline").
		Values(g.UserID, g.Name, g.TargetAmount, g.CurrentAmount, nullableDate(g.Deadline)).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert saving goal: %w", err)
	}
	if err := r.pg.Pool.QueryRow(ctx, query, args...).Scan(&g.ID, &g.CreatedAt); err != nil {
		return fmt.Errorf("insert saving goal: %w", err)
	}
	return nil
}

func (r *SavingGoalRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.SavingGoal, error) {
	var g domain.SavingGoal
	err := r.pg.Pool.QueryRow(ctx,
		"SELECT id, user_id, name, target_amount, current_amount, deadline, created_at FROM saving_goals WHERE id = $1 AND user_id = $2",
		id, userID,
	).Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Deadline, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select saving goal: %w", err)
	}
	return &g, nil
}

func (r *SavingGoalRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.SavingGoal, error) {
	rows, err := r.pg.Pool.Query(ctx,
		"SELECT id, user_id, name, target_amount, current_amount, deadline, created_at FROM saving_goals WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query saving goals: %w", err)
	}
	defer rows.Close()
	var out []domain.SavingGoal
	for rows.Next() {
		var g domain.SavingGoal
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Deadline, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan saving goal: %w", err)
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *SavingGoalRepository) Update(ctx context.Context, userID, id uuid.UUID, name string, target *decimal.Decimal, deadline *string) error {
	builder := r.pg.Builder.Update("saving_goals").Where("id = ? AND user_id = ?", id, userID)
	hasUpdate := false
	if name != "" {
		builder = builder.Set("name", name)
		hasUpdate = true
	}
	if target != nil {
		builder = builder.Set("target_amount", *target)
		hasUpdate = true
	}
	if deadline != nil {
		if *deadline == "" {
			builder = builder.Set("deadline", nil)
		} else {
			builder = builder.Set("deadline", *deadline)
		}
		hasUpdate = true
	}
	if !hasUpdate {
		return nil
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build update saving goal: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update saving goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SavingGoalRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := r.pg.Pool.Exec(ctx,
		"DELETE FROM saving_goals WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete saving goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SavingGoalRepository) AddContribution(ctx context.Context, userID, id uuid.UUID, amount decimal.Decimal) (*domain.SavingGoal, error) {
	var g domain.SavingGoal
	err := r.pg.Pool.QueryRow(ctx, `
		UPDATE saving_goals
		SET current_amount = current_amount + $1
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, name, target_amount, current_amount, deadline, created_at`,
		amount, id, userID,
	).Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Deadline, &g.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("contribute to saving goal: %w", err)
	}
	return &g, nil
}

func nullableDate(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
