package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateAccountRequest struct {
	Name     string          `json:"name" binding:"required,max=100"`
	Type     string          `json:"type" binding:"required,oneof=cash card savings other"`
	Currency string          `json:"currency" binding:"required,len=3"`
	Balance  decimal.Decimal `json:"balance"`
}

func (r CreateAccountRequest) ToInput() service.CreateAccountInput {
	return service.CreateAccountInput{
		Name:     r.Name,
		Type:     r.Type,
		Currency: r.Currency,
		Balance:  r.Balance,
	}
}

type UpdateAccountRequest struct {
	Name string `json:"name" binding:"omitempty,max=100"`
	Type string `json:"type" binding:"omitempty,oneof=cash card savings other"`
}

func (r UpdateAccountRequest) ToInput() service.UpdateAccountInput {
	return service.UpdateAccountInput{
		Name: r.Name,
		Type: r.Type,
	}
}

type AccountResponse struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Currency  string          `json:"currency"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
}

func AccountResponseFrom(a *domain.Account) AccountResponse {
	return AccountResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Name:      a.Name,
		Type:      a.Type,
		Currency:  a.Currency,
		Balance:   a.Balance,
		CreatedAt: a.CreatedAt,
	}
}

func AccountListResponseFrom(accounts []domain.Account) []AccountResponse {
	out := make([]AccountResponse, len(accounts))
	for i := range accounts {
		out[i] = AccountResponseFrom(&accounts[i])
	}
	return out
}
