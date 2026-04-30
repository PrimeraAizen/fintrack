package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateBudgetRequest struct {
	CategoryID    uuid.UUID       `json:"category_id" binding:"required"`
	SpendingLimit decimal.Decimal `json:"spending_limit" binding:"required"`
	Period        string          `json:"period" binding:"required,oneof=weekly monthly"`
}

func (r CreateBudgetRequest) ToInput() service.CreateBudgetInput {
	return service.CreateBudgetInput{
		CategoryID:    r.CategoryID,
		SpendingLimit: r.SpendingLimit,
		Period:        r.Period,
	}
}

type UpdateBudgetRequest struct {
	SpendingLimit *decimal.Decimal `json:"spending_limit"`
	Period        string           `json:"period" binding:"omitempty,oneof=weekly monthly"`
}

func (r UpdateBudgetRequest) ToInput() service.UpdateBudgetInput {
	return service.UpdateBudgetInput{
		SpendingLimit: r.SpendingLimit,
		Period:        r.Period,
	}
}

type BudgetResponse struct {
	ID            uuid.UUID       `json:"id"`
	CategoryID    uuid.UUID       `json:"category_id"`
	SpendingLimit decimal.Decimal `json:"spending_limit"`
	Spent         decimal.Decimal `json:"spent"`
	Remaining     decimal.Decimal `json:"remaining"`
	Period        string          `json:"period"`
	PeriodStart   time.Time       `json:"period_start"`
	Exceeded      bool            `json:"exceeded"`
	CreatedAt     time.Time       `json:"created_at"`
}

func BudgetResponseFrom(b *domain.Budget) BudgetResponse {
	return BudgetResponse{
		ID:            b.ID,
		CategoryID:    b.CategoryID,
		SpendingLimit: b.SpendingLimit,
		Spent:         b.Spent,
		Remaining:     b.Remaining(),
		Period:        b.Period,
		PeriodStart:   b.PeriodStart,
		Exceeded:      b.Exceeded(),
		CreatedAt:     b.CreatedAt,
	}
}

func BudgetViewResponseFrom(views []service.BudgetView) []BudgetResponse {
	out := make([]BudgetResponse, len(views))
	for i := range views {
		out[i] = BudgetResponseFrom(&views[i].Budget)
		out[i].Remaining = views[i].Remaining
		out[i].Exceeded = views[i].Exceeded
	}
	return out
}
