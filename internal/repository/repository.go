package repository

import (
	"context"

	postgres "github.com/diyas/fintrack/pkg/adapter"
)

type Health interface {
	Ping(ctx context.Context) error
}

type Repository struct {
	Health       Health
	User         User
	Account      Account
	Category     Category
	Transaction  Transaction
	ExchangeRate ExchangeRate
	Budget       Budget
	Transfer     Transfer
	SavingGoal   SavingGoal
	Report       Report
}

func NewRepositories(pg *postgres.Postgres) *Repository {
	return &Repository{
		Health:       NewHealthRepository(pg),
		User:         NewUserRepository(pg),
		Account:      NewAccountRepository(pg),
		Category:     NewCategoryRepository(pg),
		Transaction:  NewTransactionRepository(pg),
		ExchangeRate: NewExchangeRateRepository(pg),
		Budget:       NewBudgetRepository(pg),
		Transfer:     NewTransferRepository(pg),
		SavingGoal:   NewSavingGoalRepository(pg),
		Report:       NewReportRepository(pg),
	}
}

type HealthRepository struct {
	pg *postgres.Postgres
}

func NewHealthRepository(pg *postgres.Postgres) *HealthRepository {
	return &HealthRepository{pg: pg}
}

func (r *HealthRepository) Ping(ctx context.Context) error {
	return r.pg.Pool.Ping(ctx)
}
