package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateAccountInput struct {
	Name     string
	Type     string
	Currency string
	Balance  decimal.Decimal
}

type UpdateAccountInput struct {
	Name string
	Type string
}

type Account interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateAccountInput) (*domain.Account, error)
	List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error)
	Update(ctx context.Context, userID, id uuid.UUID, in UpdateAccountInput) (*domain.Account, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type AccountService struct {
	repo repository.Account
}

func NewAccountService(repo repository.Account) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Create(ctx context.Context, userID uuid.UUID, in CreateAccountInput) (*domain.Account, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name required", domain.ErrInvalidInput)
	}
	if !isValidAccountType(in.Type) {
		return nil, fmt.Errorf("%w: invalid account type", domain.ErrInvalidInput)
	}
	if len(in.Currency) != 3 {
		return nil, fmt.Errorf("%w: currency must be 3 letters", domain.ErrInvalidInput)
	}

	account := &domain.Account{
		UserID:   userID,
		Name:     in.Name,
		Type:     in.Type,
		Currency: strings.ToUpper(in.Currency),
		Balance:  in.Balance,
	}
	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	return s.repo.List(ctx, userID)
}

func (s *AccountService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *AccountService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateAccountInput) (*domain.Account, error) {
	if in.Type != "" && !isValidAccountType(in.Type) {
		return nil, fmt.Errorf("%w: invalid account type", domain.ErrInvalidInput)
	}
	if err := s.repo.Update(ctx, userID, id, in.Name, in.Type); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

func (s *AccountService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}

func isValidAccountType(t string) bool {
	switch t {
	case domain.AccountTypeCash, domain.AccountTypeCard, domain.AccountTypeSavings, domain.AccountTypeOther:
		return true
	}
	return false
}
