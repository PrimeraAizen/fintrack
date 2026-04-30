package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type Currency interface {
	GetRate(ctx context.Context, from, to string) (decimal.Decimal, error)
	Convert(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error)
}

type CurrencyService struct {
	cfg        *config.Currency
	redis      *postgres.RedisClient
	rates      repository.ExchangeRate
	httpClient *http.Client
}

func NewCurrencyService(cfg *config.Currency, rdb *postgres.RedisClient, rates repository.ExchangeRate) *CurrencyService {
	return &CurrencyService{
		cfg:   cfg,
		redis: rdb,
		rates: rates,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *CurrencyService) GetRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)
	if from == to {
		return decimal.NewFromInt(1), nil
	}

	cacheKey := fmt.Sprintf("exchange_rate:%s:%s", from, to)
	if val, err := s.redis.Client.Get(ctx, cacheKey).Result(); err == nil {
		if rate, decErr := decimal.NewFromString(val); decErr == nil {
			return rate, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		// log-only failure on cache read; continue to fetch
	}

	rate, fetchErr := s.fetchRate(ctx, from, to)
	if fetchErr == nil {
		_ = s.redis.Client.Set(ctx, cacheKey, rate.String(), s.cfg.CacheTTL).Err()
		_ = s.rates.Upsert(ctx, from, to, rate)
		return rate, nil
	}

	if er, dbErr := s.rates.Get(ctx, from, to); dbErr == nil {
		return er.Rate, nil
	}

	return decimal.Zero, fmt.Errorf("get rate %s->%s: %w", from, to, fetchErr)
}

func (s *CurrencyService) Convert(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}
	return amount.Mul(rate).Round(2), nil
}

type exchangeRateAPIResponse struct {
	Rates map[string]float64 `json:"rates"`
}

func (s *CurrencyService) fetchRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	endpoint, err := url.Parse(s.cfg.APIURL)
	if err != nil {
		return decimal.Zero, fmt.Errorf("parse currency api url: %w", err)
	}
	q := endpoint.Query()
	q.Set("base", from)
	q.Set("symbols", to)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("build request: %w", err)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("call currency api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("currency api status %d", resp.StatusCode)
	}

	var body exchangeRateAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return decimal.Zero, fmt.Errorf("decode currency api: %w", err)
	}
	rateFloat, ok := body.Rates[to]
	if !ok {
		return decimal.Zero, fmt.Errorf("%w: rate %s->%s missing", domain.ErrNotFound, from, to)
	}
	return decimal.NewFromFloat(rateFloat), nil
}
