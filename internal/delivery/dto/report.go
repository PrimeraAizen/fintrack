package dto

import (
	"github.com/diyas/fintrack/internal/domain"
)

type ReportResponse struct {
	From         string                     `json:"from"`
	To           string                     `json:"to"`
	TotalIncome  string                     `json:"total_income"`
	TotalExpense string                     `json:"total_expense"`
	Net          string                     `json:"net"`
	ByCategory   []domain.CategoryBreakdown `json:"by_category"`
	DailyTrend   []domain.DailyTrendPoint   `json:"daily_trend"`
}

func ReportResponseFrom(r *domain.Report) ReportResponse {
	return ReportResponse{
		From:         r.From.Format("2006-01-02"),
		To:           r.To.Format("2006-01-02"),
		TotalIncome:  r.TotalIncome.String(),
		TotalExpense: r.TotalExpense.String(),
		Net:          r.Net.String(),
		ByCategory:   r.ByCategory,
		DailyTrend:   r.DailyTrend,
	}
}
