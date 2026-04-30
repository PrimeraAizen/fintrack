package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	CategoryID      uuid.UUID       `json:"category_id"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	ConvertedAmount decimal.Decimal `json:"converted_amount"`
	Note            string          `json:"note,omitempty"`
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}

type TransactionFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	FromDate   *time.Time
	ToDate     *time.Time
	Type       string
	Page       int
	PerPage    int
}

type ExchangeRate struct {
	BaseCurrency   string          `json:"base_currency"`
	TargetCurrency string          `json:"target_currency"`
	Rate           decimal.Decimal `json:"rate"`
	CachedAt       time.Time       `json:"cached_at"`
}
