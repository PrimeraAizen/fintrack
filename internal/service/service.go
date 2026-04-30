package service

import (
	"context"

	"github.com/diyas/fintrack/config"
	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	postgres "github.com/diyas/fintrack/pkg/adapter"
)

type Health interface {
	Ping(ctx context.Context) error
}

type Service struct {
	Health      Health
	Auth        Auth
	Account     Account
	Category    Category
	Currency    Currency
	Transaction Transaction
	Budget      Budget
	Transfer    Transfer
	SavingGoal  SavingGoal
	Report      Report
	CSV         CSV
}

type Deps struct {
	Repos  *repository.Repository
	Config *config.Config
	Redis  *postgres.RedisClient
}

func NewServices(deps Deps) *Service {
	tokenManager := NewTokenManager(&deps.Config.JWT, deps.Redis)
	auth := NewAuthService(deps.Repos.User, tokenManager)
	categories := NewCategoryService(deps.Repos.Category)
	currency := NewCurrencyService(&deps.Config.Currency, deps.Redis, deps.Repos.ExchangeRate)
	transactions := NewTransactionService(
		deps.Repos.Transaction,
		deps.Repos.Account,
		deps.Repos.Category,
		deps.Repos.User,
		currency,
	)
	budgets := NewBudgetService(deps.Repos.Budget)
	transactions.SetBudgetUpdater(budgets)

	auth.RegisterHook(func(ctx context.Context, user *domain.User) error {
		return categories.SeedDefaults(ctx, user.ID)
	})

	return &Service{
		Health:      NewHealthService(deps.Repos.Health),
		Auth:        auth,
		Account:     NewAccountService(deps.Repos.Account),
		Category:    categories,
		Currency:    currency,
		Transaction: transactions,
		Budget:      budgets,
		Transfer:    NewTransferService(deps.Repos.Transfer, deps.Repos.Account, currency),
		SavingGoal:  NewSavingGoalService(deps.Repos.SavingGoal),
		Report:      NewReportService(deps.Repos.Report),
		CSV:         NewCSVService(transactions, deps.Repos.Account, deps.Repos.Category, deps.Repos.Transaction),
	}
}

type HealthService struct {
	repo repository.Health
}

func NewHealthService(repo repository.Health) *HealthService {
	return &HealthService{repo: repo}
}

func (s *HealthService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
