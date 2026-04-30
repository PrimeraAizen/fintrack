package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CategoryBreakdown struct {
	CategoryID   uuid.UUID       `json:"category_id"`
	CategoryName string          `json:"category_name"`
	CategoryType string          `json:"category_type"`
	Total        decimal.Decimal `json:"total"`
}

type DailyTrendPoint struct {
	Date    time.Time       `json:"date"`
	Income  decimal.Decimal `json:"income"`
	Expense decimal.Decimal `json:"expense"`
}

type Report struct {
	From         time.Time           `json:"from"`
	To           time.Time           `json:"to"`
	TotalIncome  decimal.Decimal     `json:"total_income"`
	TotalExpense decimal.Decimal     `json:"total_expense"`
	Net          decimal.Decimal     `json:"net"`
	ByCategory   []CategoryBreakdown `json:"by_category"`
	DailyTrend   []DailyTrendPoint   `json:"daily_trend"`
}
