package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/shopspring/decimal"
)

type ExchangeRateResponse struct {
	From     string          `json:"from"`
	To       string          `json:"to"`
	Rate     decimal.Decimal `json:"rate"`
	CachedAt time.Time       `json:"cached_at"`
}

func ExchangeRateResponseFrom(er *domain.ExchangeRate) ExchangeRateResponse {
	return ExchangeRateResponse{
		From:     er.BaseCurrency,
		To:       er.TargetCurrency,
		Rate:     er.Rate,
		CachedAt: er.CachedAt,
	}
}
