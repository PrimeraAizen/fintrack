package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transfer struct {
	ID            uuid.UUID       `json:"id"`
	FromAccountID uuid.UUID       `json:"from_account_id"`
	ToAccountID   uuid.UUID       `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	FromCurrency  string          `json:"from_currency"`
	ToCurrency    string          `json:"to_currency"`
	ExchangeRate  decimal.Decimal `json:"exchange_rate"`
	CreatedAt     time.Time       `json:"created_at"`
}
