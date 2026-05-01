package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type CreateTransactionInput struct {
	AccountID       uuid.UUID
	CategoryID      uuid.UUID
	Amount          decimal.Decimal
	Currency        string
	Note            string
	TransactionDate *time.Time
}

type TransactionResult struct {
	Transaction     *domain.Transaction
	Category        *domain.Category
	BudgetExceeded  bool
}

// BudgetUpdater is a hook-style interface so the transaction service can update
// budgets without depending on the budget service directly. Wired in service/service.go.
type BudgetUpdater interface {
	OnTransactionCreatedTx(ctx context.Context, tx pgx.Tx, userID, categoryID uuid.UUID, convertedAmount decimal.Decimal, categoryType string) (exceeded bool, err error)
	OnTransactionDeletedTx(ctx context.Context, tx pgx.Tx, userID, categoryID uuid.UUID, convertedAmount decimal.Decimal, categoryType string) error
}

type UpdateTransactionInput struct {
	CategoryID      *uuid.UUID
	Amount          *decimal.Decimal
	Note            *string
	TransactionDate *time.Time
}

type Transaction interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateTransactionInput) (*TransactionResult, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error)
	List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int64, error)
	Update(ctx context.Context, userID, id uuid.UUID, in UpdateTransactionInput) (*domain.Transaction, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	SetBudgetUpdater(b BudgetUpdater)
}

type TransactionService struct {
	transactions repository.Transaction
	accounts     repository.Account
	categories   repository.Category
	users        repository.User
	currency     Currency
	budgets      BudgetUpdater
}

func NewTransactionService(
	transactions repository.Transaction,
	accounts repository.Account,
	categories repository.Category,
	users repository.User,
	currency Currency,
) *TransactionService {
	return &TransactionService{
		transactions: transactions,
		accounts:     accounts,
		categories:   categories,
		users:        users,
		currency:     currency,
	}
}

func (s *TransactionService) SetBudgetUpdater(b BudgetUpdater) {
	s.budgets = b
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, in CreateTransactionInput) (*TransactionResult, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	account, err := s.accounts.GetByID(ctx, userID, in.AccountID)
	if err != nil {
		return nil, err
	}

	category, err := s.categories.GetByID(ctx, userID, in.CategoryID)
	if err != nil {
		return nil, err
	}

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = account.Currency
	}

	convertedAmount := in.Amount
	if currency != user.BaseCurrency {
		converted, err := s.currency.Convert(ctx, in.Amount, currency, user.BaseCurrency)
		if err != nil {
			return nil, fmt.Errorf("convert currency: %w", err)
		}
		convertedAmount = converted
	}

	txDate := time.Now()
	if in.TransactionDate != nil {
		txDate = *in.TransactionDate
	}

	t := &domain.Transaction{
		AccountID:       in.AccountID,
		CategoryID:      in.CategoryID,
		Amount:          in.Amount,
		Currency:        currency,
		ConvertedAmount: convertedAmount,
		Note:            in.Note,
		TransactionDate: txDate,
	}

	delta := signedAmount(in.Amount, category.Type)

	dbTx, err := s.transactions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, dbTx)

	if err := s.transactions.CreateTx(ctx, dbTx, t); err != nil {
		return nil, err
	}
	if err := s.accounts.UpdateBalance(ctx, dbTx, account.ID, delta); err != nil {
		return nil, err
	}

	exceeded := false
	if s.budgets != nil && category.Type == domain.CategoryTypeExpense {
		ex, err := s.budgets.OnTransactionCreatedTx(ctx, dbTx, userID, category.ID, convertedAmount, category.Type)
		if err != nil {
			return nil, err
		}
		exceeded = ex
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &TransactionResult{
		Transaction:    t,
		Category:       category,
		BudgetExceeded: exceeded,
	}, nil
}

func (s *TransactionService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	return s.transactions.GetForUser(ctx, userID, id)
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	return s.transactions.List(ctx, userID, filter)
}

func (s *TransactionService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateTransactionInput) (*domain.Transaction, error) {
	dbTx, err := s.transactions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, dbTx)

	t, err := s.transactions.GetForUserTx(ctx, dbTx, userID, id)
	if err != nil {
		return nil, err
	}

	oldCategory, err := s.categories.GetByID(ctx, userID, t.CategoryID)
	if err != nil {
		return nil, err
	}

	// Apply partial updates to the local struct.
	if in.CategoryID != nil {
		newCat, err := s.categories.GetByID(ctx, userID, *in.CategoryID)
		if err != nil {
			return nil, err
		}
		t.CategoryID = newCat.ID
		oldCategory = newCat // category type may have changed
	}
	if in.Note != nil {
		t.Note = *in.Note
	}
	if in.TransactionDate != nil {
		t.TransactionDate = *in.TransactionDate
	}

	// If amount changed, recompute converted_amount and update account balance.
	if in.Amount != nil && !in.Amount.Equal(t.Amount) {
		if in.Amount.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
		}
		user, err := s.users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		newConverted := *in.Amount
		if t.Currency != user.BaseCurrency {
			converted, err := s.currency.Convert(ctx, *in.Amount, t.Currency, user.BaseCurrency)
			if err != nil {
				return nil, fmt.Errorf("convert currency: %w", err)
			}
			newConverted = converted
		}
		// Reverse old balance effect, apply new balance effect.
		oldDelta := signedAmount(t.Amount, oldCategory.Type)
		newDelta := signedAmount(*in.Amount, oldCategory.Type)
		balanceDelta := newDelta.Sub(oldDelta)
		if err := s.accounts.UpdateBalance(ctx, dbTx, t.AccountID, balanceDelta); err != nil {
			return nil, err
		}
		t.Amount = *in.Amount
		t.ConvertedAmount = newConverted
	}

	if err := s.transactions.UpdateTx(ctx, dbTx, t); err != nil {
		return nil, err
	}
	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update transaction: %w", err)
	}
	return t, nil
}

func (s *TransactionService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	dbTx, err := s.transactions.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer rollback(ctx, dbTx)

	t, err := s.transactions.GetForUserTx(ctx, dbTx, userID, id)
	if err != nil {
		return err
	}

	category, err := s.categories.GetByID(ctx, userID, t.CategoryID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	delta := decimal.Zero
	if category != nil {
		delta = signedAmount(t.Amount, category.Type).Neg()
	}

	if err := s.accounts.UpdateBalance(ctx, dbTx, t.AccountID, delta); err != nil {
		return err
	}
	if err := s.transactions.DeleteTx(ctx, dbTx, t.ID); err != nil {
		return err
	}

	if s.budgets != nil && category != nil && category.Type == domain.CategoryTypeExpense {
		if err := s.budgets.OnTransactionDeletedTx(ctx, dbTx, userID, category.ID, t.ConvertedAmount, category.Type); err != nil {
			return err
		}
	}

	return dbTx.Commit(ctx)
}

func signedAmount(amount decimal.Decimal, categoryType string) decimal.Decimal {
	if categoryType == domain.CategoryTypeExpense {
		return amount.Neg()
	}
	return amount
}

func rollback(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}
