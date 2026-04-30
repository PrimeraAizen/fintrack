package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SavingGoal struct {
	ID            uuid.UUID       `json:"id"`
	UserID        uuid.UUID       `json:"user_id"`
	Name          string          `json:"name"`
	TargetAmount  decimal.Decimal `json:"target_amount"`
	CurrentAmount decimal.Decimal `json:"current_amount"`
	Deadline      *time.Time      `json:"deadline,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (g *SavingGoal) ProgressPercent() float64 {
	if g.TargetAmount.IsZero() {
		return 0
	}
	progress := g.CurrentAmount.Div(g.TargetAmount).Mul(decimal.NewFromInt(100))
	if progress.GreaterThan(decimal.NewFromInt(100)) {
		return 100
	}
	f, _ := progress.Float64()
	return f
}
