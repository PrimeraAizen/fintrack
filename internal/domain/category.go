package domain

import "github.com/google/uuid"

const (
	CategoryTypeIncome  = "income"
	CategoryTypeExpense = "expense"
)

type Category struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Type   string    `json:"type"`
	Icon   string    `json:"icon,omitempty"`
}
