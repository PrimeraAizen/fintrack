package dto

import (
	"github.com/diyas/fintrack/internal/domain"
	"github.com/shopspring/decimal"
)

type ReportResponse struct {
	From         string                     `json:"from"`
	To           string                     `json:"to"`
	TotalIncome  decimal.Decimal            `json:"total_income"`
	TotalExpense decimal.Decimal            `json:"total_expense"`
	Net          decimal.Decimal            `json:"net"`
	ByCategory   []domain.CategoryBreakdown `json:"by_category"`
	DailyTrend   []domain.DailyTrendPoint   `json:"daily_trend"`
}

func ReportResponseFrom(r *domain.Report) ReportResponse {
	return ReportResponse{
		From:         r.From.Format("2006-01-02"),
		To:           r.To.Format("2006-01-02"),
		TotalIncome:  r.TotalIncome,
		TotalExpense: r.TotalExpense,
		Net:          r.Net,
		ByCategory:   r.ByCategory,
		DailyTrend:   r.DailyTrend,
	}
}
