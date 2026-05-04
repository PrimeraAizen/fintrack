package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateTransactionRequest struct {
	AccountID       uuid.UUID       `json:"account_id" binding:"required"`
	CategoryID      uuid.UUID       `json:"category_id" binding:"required"`
	Amount          decimal.Decimal `json:"amount" binding:"required"`
	Currency        string          `json:"currency" binding:"omitempty,len=3"`
	Note            string          `json:"note"`
	TransactionDate *time.Time      `json:"transaction_date"`
}

func (r CreateTransactionRequest) ToInput() service.CreateTransactionInput {
	return service.CreateTransactionInput{
		AccountID:       r.AccountID,
		CategoryID:      r.CategoryID,
		Amount:          r.Amount,
		Currency:        r.Currency,
		Note:            r.Note,
		TransactionDate: r.TransactionDate,
	}
}

type TransactionResponse struct {
	ID              uuid.UUID       `json:"id"`
	AccountID       uuid.UUID       `json:"account_id"`
	CategoryID      uuid.UUID       `json:"category_id"`
	Type            string          `json:"type"`
	Amount          decimal.Decimal `json:"amount"`
	Currency        string          `json:"currency"`
	ConvertedAmount decimal.Decimal `json:"converted_amount"`
	Note            string          `json:"note,omitempty"`
	TransactionDate time.Time       `json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}

func TransactionResponseFrom(t *domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:              t.ID,
		AccountID:       t.AccountID,
		CategoryID:      t.CategoryID,
		Type:            t.Type,
		Amount:          t.Amount,
		Currency:        t.Currency,
		ConvertedAmount: t.ConvertedAmount,
		Note:            t.Note,
		TransactionDate: t.TransactionDate,
		CreatedAt:       t.CreatedAt,
	}
}

type UpdateTransactionRequest struct {
	CategoryID      *uuid.UUID       `json:"category_id"`
	Amount          *decimal.Decimal `json:"amount"`
	Note            *string          `json:"note"`
	TransactionDate *time.Time       `json:"transaction_date"`
}

func (r UpdateTransactionRequest) ToInput() service.UpdateTransactionInput {
	return service.UpdateTransactionInput{
		CategoryID:      r.CategoryID,
		Amount:          r.Amount,
		Note:            r.Note,
		TransactionDate: r.TransactionDate,
	}
}

func TransactionListResponseFrom(items []domain.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, len(items))
	for i := range items {
		out[i] = TransactionResponseFrom(&items[i])
	}
	return out
}
