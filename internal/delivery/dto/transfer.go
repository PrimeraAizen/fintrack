package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateTransferRequest struct {
	FromAccountID uuid.UUID       `json:"from_account_id" binding:"required"`
	ToAccountID   uuid.UUID       `json:"to_account_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
}

func (r CreateTransferRequest) ToInput() service.ExecuteTransferInput {
	return service.ExecuteTransferInput{
		FromAccountID: r.FromAccountID,
		ToAccountID:   r.ToAccountID,
		Amount:        r.Amount,
	}
}

type TransferResponse struct {
	ID            uuid.UUID       `json:"id"`
	FromAccountID uuid.UUID       `json:"from_account_id"`
	ToAccountID   uuid.UUID       `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	FromCurrency  string          `json:"from_currency"`
	ToCurrency    string          `json:"to_currency"`
	ExchangeRate  decimal.Decimal `json:"exchange_rate"`
	CreatedAt     time.Time       `json:"created_at"`
}

func TransferResponseFrom(t *domain.Transfer) TransferResponse {
	return TransferResponse{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        t.Amount,
		FromCurrency:  t.FromCurrency,
		ToCurrency:    t.ToCurrency,
		ExchangeRate:  t.ExchangeRate,
		CreatedAt:     t.CreatedAt,
	}
}

func TransferListResponseFrom(items []domain.Transfer) []TransferResponse {
	out := make([]TransferResponse, len(items))
	for i := range items {
		out[i] = TransferResponseFrom(&items[i])
	}
	return out
}
