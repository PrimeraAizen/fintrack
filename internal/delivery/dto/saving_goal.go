package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateSavingGoalRequest struct {
	Name         string          `json:"name" binding:"required,max=100"`
	TargetAmount decimal.Decimal `json:"target_amount" binding:"required"`
	Deadline     *time.Time      `json:"deadline"`
}

func (r CreateSavingGoalRequest) ToInput() service.CreateSavingGoalInput {
	return service.CreateSavingGoalInput{
		Name:         r.Name,
		TargetAmount: r.TargetAmount,
		Deadline:     r.Deadline,
	}
}

type UpdateSavingGoalRequest struct {
	Name         string           `json:"name" binding:"omitempty,max=100"`
	TargetAmount *decimal.Decimal `json:"target_amount"`
	Deadline     *string          `json:"deadline"`
}

func (r UpdateSavingGoalRequest) ToInput() service.UpdateSavingGoalInput {
	return service.UpdateSavingGoalInput{
		Name:         r.Name,
		TargetAmount: r.TargetAmount,
		Deadline:     r.Deadline,
	}
}

type ContributeSavingGoalRequest struct {
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

type SavingGoalResponse struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	TargetAmount    decimal.Decimal `json:"target_amount"`
	CurrentAmount   decimal.Decimal `json:"current_amount"`
	ProgressPercent float64         `json:"progress_percent"`
	Deadline        *time.Time      `json:"deadline,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

func SavingGoalResponseFrom(g *domain.SavingGoal) SavingGoalResponse {
	return SavingGoalResponse{
		ID:              g.ID,
		Name:            g.Name,
		TargetAmount:    g.TargetAmount,
		CurrentAmount:   g.CurrentAmount,
		ProgressPercent: g.ProgressPercent(),
		Deadline:        g.Deadline,
		CreatedAt:       g.CreatedAt,
	}
}

func SavingGoalListResponseFrom(views []service.SavingGoalView) []SavingGoalResponse {
	out := make([]SavingGoalResponse, len(views))
	for i := range views {
		out[i] = SavingGoalResponseFrom(&views[i].SavingGoal)
		out[i].ProgressPercent = views[i].ProgressPercent
	}
	return out
}
