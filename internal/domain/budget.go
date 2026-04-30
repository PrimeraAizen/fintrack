package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	BudgetPeriodWeekly  = "weekly"
	BudgetPeriodMonthly = "monthly"
)

type Budget struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	CategoryID    uuid.UUID       `json:"category_id"`
	SpendingLimit decimal.Decimal `json:"spending_limit"`
	Period        string          `json:"period"`
	Spent         decimal.Decimal `json:"spent"`
	PeriodStart   time.Time       `json:"period_start"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (b *Budget) Remaining() decimal.Decimal {
	return b.SpendingLimit.Sub(b.Spent)
}

func (b *Budget) Exceeded() bool {
	return b.Spent.GreaterThan(b.SpendingLimit)
}

func (b *Budget) NextPeriodStart() time.Time {
	if b.Period == BudgetPeriodWeekly {
		return b.PeriodStart.AddDate(0, 0, 7)
	}
	return b.PeriodStart.AddDate(0, 1, 0)
}
