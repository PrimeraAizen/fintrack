package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateSavingGoalInput struct {
	Name         string
	TargetAmount decimal.Decimal
	Deadline     *time.Time
}

type UpdateSavingGoalInput struct {
	Name         string
	TargetAmount *decimal.Decimal
	Deadline     *string
}

type SavingGoalView struct {
	domain.SavingGoal
	ProgressPercent float64 `json:"progress_percent"`
}

type SavingGoal interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateSavingGoalInput) (*domain.SavingGoal, error)
	List(ctx context.Context, userID uuid.UUID) ([]SavingGoalView, error)
	Update(ctx context.Context, userID, id uuid.UUID, in UpdateSavingGoalInput) (*domain.SavingGoal, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Contribute(ctx context.Context, userID, id uuid.UUID, amount decimal.Decimal) (*SavingGoalView, error)
}

type SavingGoalService struct {
	repo repository.SavingGoal
}

func NewSavingGoalService(repo repository.SavingGoal) *SavingGoalService {
	return &SavingGoalService{repo: repo}
}

func (s *SavingGoalService) Create(ctx context.Context, userID uuid.UUID, in CreateSavingGoalInput) (*domain.SavingGoal, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name required", domain.ErrInvalidInput)
	}
	if in.TargetAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: target must be positive", domain.ErrInvalidInput)
	}
	g := &domain.SavingGoal{
		UserID:       userID,
		Name:         in.Name,
		TargetAmount: in.TargetAmount,
		Deadline:     in.Deadline,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *SavingGoalService) List(ctx context.Context, userID uuid.UUID) ([]SavingGoalView, error) {
	goals, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]SavingGoalView, len(goals))
	for i := range goals {
		out[i] = SavingGoalView{
			SavingGoal:      goals[i],
			ProgressPercent: goals[i].ProgressPercent(),
		}
	}
	return out, nil
}

func (s *SavingGoalService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateSavingGoalInput) (*domain.SavingGoal, error) {
	if in.TargetAmount != nil && in.TargetAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: target must be positive", domain.ErrInvalidInput)
	}
	if err := s.repo.Update(ctx, userID, id, in.Name, in.TargetAmount, in.Deadline); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

func (s *SavingGoalService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *SavingGoalService) Contribute(ctx context.Context, userID, id uuid.UUID, amount decimal.Decimal) (*SavingGoalView, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}
	g, err := s.repo.AddContribution(ctx, userID, id, amount)
	if err != nil {
		return nil, err
	}
	return &SavingGoalView{
		SavingGoal:      *g,
		ProgressPercent: g.ProgressPercent(),
	}, nil
}
