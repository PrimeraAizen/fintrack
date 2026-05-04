package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type CreateBudgetInput struct {
	CategoryID    uuid.UUID
	SpendingLimit decimal.Decimal
	Period        string
}

type UpdateBudgetInput struct {
	SpendingLimit *decimal.Decimal
	Period        string
}

type BudgetView struct {
	domain.Budget
	Remaining decimal.Decimal `json:"remaining"`
	Exceeded  bool            `json:"exceeded"`
}

type Budget interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateBudgetInput) (*domain.Budget, error)
	List(ctx context.Context, userID uuid.UUID, period string) ([]BudgetView, error)
	Update(ctx context.Context, userID, id uuid.UUID, in UpdateBudgetInput) (*domain.Budget, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error

	BudgetUpdater
}

type BudgetService struct {
	repo repository.Budget
}

func NewBudgetService(repo repository.Budget) *BudgetService {
	return &BudgetService{repo: repo}
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, in CreateBudgetInput) (*domain.Budget, error) {
	if in.SpendingLimit.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: limit must be positive", domain.ErrInvalidInput)
	}
	if in.Period != domain.BudgetPeriodWeekly && in.Period != domain.BudgetPeriodMonthly {
		return nil, fmt.Errorf("%w: invalid period", domain.ErrInvalidInput)
	}
	b := &domain.Budget{
		UserID:        userID,
		CategoryID:    in.CategoryID,
		SpendingLimit: in.SpendingLimit,
		Period:        in.Period,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID, period string) ([]BudgetView, error) {
	budgets, err := s.repo.List(ctx, userID, period)
	if err != nil {
		return nil, err
	}
	out := make([]BudgetView, 0, len(budgets))
	now := time.Now()
	for i := range budgets {
		b := &budgets[i]
		if !now.Before(b.NextPeriodStart()) {
			next := nextPeriodStart(b.Period, now)
			if err := s.repo.ResetPeriod(ctx, b.ID, next.Format("2006-01-02")); err != nil {
				return nil, err
			}
			b.Spent = decimal.Zero
			b.PeriodStart = next
		}
		out = append(out, BudgetView{
			Budget:    *b,
			Remaining: b.Remaining(),
			Exceeded:  b.Exceeded(),
		})
	}
	return out, nil
}

func (s *BudgetService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateBudgetInput) (*domain.Budget, error) {
	if in.SpendingLimit != nil && in.SpendingLimit.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: limit must be positive", domain.ErrInvalidInput)
	}
	if in.Period != "" && in.Period != domain.BudgetPeriodWeekly && in.Period != domain.BudgetPeriodMonthly {
		return nil, fmt.Errorf("%w: invalid period", domain.ErrInvalidInput)
	}
	if err := s.repo.Update(ctx, userID, id, in.SpendingLimit, in.Period); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

func (s *BudgetService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *BudgetService) OnTransactionCreatedTx(ctx context.Context, tx pgx.Tx, userID, categoryID uuid.UUID, convertedAmount decimal.Decimal, categoryType string) (bool, error) {
	if categoryType != domain.CategoryTypeExpense {
		return false, nil
	}
	budget, err := s.repo.GetByCategoryTx(ctx, tx, userID, categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if err := s.repo.UpdateSpentTx(ctx, tx, budget.ID, convertedAmount); err != nil {
		return false, err
	}
	newSpent := budget.Spent.Add(convertedAmount)
	return newSpent.GreaterThan(budget.SpendingLimit), nil
}

func (s *BudgetService) OnTransactionDeletedTx(ctx context.Context, tx pgx.Tx, userID, categoryID uuid.UUID, convertedAmount decimal.Decimal, categoryType string) error {
	if categoryType != domain.CategoryTypeExpense {
		return nil
	}
	budget, err := s.repo.GetByCategoryTx(ctx, tx, userID, categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.repo.UpdateSpentTx(ctx, tx, budget.ID, convertedAmount.Neg())
}

func nextPeriodStart(period string, now time.Time) time.Time {
	if period == domain.BudgetPeriodWeekly {
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, now.Location())
	}
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}
