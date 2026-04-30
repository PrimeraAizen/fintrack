package service

import (
	"context"
	"fmt"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
)

type Report interface {
	Weekly(ctx context.Context, userID uuid.UUID, anchor time.Time) (*domain.Report, error)
	Monthly(ctx context.Context, userID uuid.UUID, year int, month time.Month) (*domain.Report, error)
}

type ReportService struct {
	repo repository.Report
}

func NewReportService(repo repository.Report) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) Weekly(ctx context.Context, userID uuid.UUID, anchor time.Time) (*domain.Report, error) {
	from := startOfWeek(anchor)
	to := from.AddDate(0, 0, 6)
	return s.build(ctx, userID, from, to)
}

func (s *ReportService) Monthly(ctx context.Context, userID uuid.UUID, year int, month time.Month) (*domain.Report, error) {
	if month < time.January || month > time.December {
		return nil, fmt.Errorf("%w: invalid month", domain.ErrInvalidInput)
	}
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)
	return s.build(ctx, userID, from, to)
}

func (s *ReportService) build(ctx context.Context, userID uuid.UUID, from, to time.Time) (*domain.Report, error) {
	income, expense, err := s.repo.Totals(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	byCategory, err := s.repo.ByCategory(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	trend, err := s.repo.DailyTrend(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}
	return &domain.Report{
		From:         from,
		To:           to,
		TotalIncome:  income,
		TotalExpense: expense,
		Net:          income.Sub(expense),
		ByCategory:   byCategory,
		DailyTrend:   trend,
	}, nil
}

func startOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	day := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}
