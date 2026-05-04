package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	AccountTypeCash    = "cash"
	AccountTypeCard    = "card"
	AccountTypeSavings = "savings"
	AccountTypeOther   = "other"
)

type Account struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Currency  string          `json:"currency"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
}
