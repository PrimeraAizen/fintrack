package service

import (
	"context"
	"fmt"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExecuteTransferInput struct {
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	Amount        decimal.Decimal
}

type Transfer interface {
	Execute(ctx context.Context, userID uuid.UUID, in ExecuteTransferInput) (*domain.Transfer, error)
	List(ctx context.Context, userID uuid.UUID) ([]domain.Transfer, error)
}

type TransferService struct {
	transfers repository.Transfer
	accounts  repository.Account
	currency  Currency
}

func NewTransferService(transfers repository.Transfer, accounts repository.Account, currency Currency) *TransferService {
	return &TransferService{
		transfers: transfers,
		accounts:  accounts,
		currency:  currency,
	}
}

func (s *TransferService) Execute(ctx context.Context, userID uuid.UUID, in ExecuteTransferInput) (*domain.Transfer, error) {
	if in.FromAccountID == in.ToAccountID {
		return nil, fmt.Errorf("%w: source and destination accounts must differ", domain.ErrInvalidInput)
	}
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: amount must be positive", domain.ErrInvalidInput)
	}

	dbTx, err := s.transfers.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx, dbTx)

	from, err := s.accounts.GetByIDTx(ctx, dbTx, userID, in.FromAccountID)
	if err != nil {
		return nil, err
	}
	to, err := s.accounts.GetByIDTx(ctx, dbTx, userID, in.ToAccountID)
	if err != nil {
		return nil, err
	}
	if from.Balance.LessThan(in.Amount) {
		return nil, domain.ErrInsufficientBalance
	}

	rate := decimal.NewFromInt(1)
	convertedAmount := in.Amount
	if from.Currency != to.Currency {
		r, err := s.currency.GetRate(ctx, from.Currency, to.Currency)
		if err != nil {
			return nil, fmt.Errorf("get rate: %w", err)
		}
		rate = r
		convertedAmount = in.Amount.Mul(rate).Round(2)
	}

	if err := s.accounts.UpdateBalance(ctx, dbTx, from.ID, in.Amount.Neg()); err != nil {
		return nil, err
	}
	if err := s.accounts.UpdateBalance(ctx, dbTx, to.ID, convertedAmount); err != nil {
		return nil, err
	}

	transfer := &domain.Transfer{
		FromAccountID: from.ID,
		ToAccountID:   to.ID,
		Amount:        in.Amount,
		FromCurrency:  from.Currency,
		ToCurrency:    to.Currency,
		ExchangeRate:  rate,
	}
	if err := s.transfers.CreateTx(ctx, dbTx, transfer); err != nil {
		return nil, err
	}
	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transfer: %w", err)
	}
	return transfer, nil
}

func (s *TransferService) List(ctx context.Context, userID uuid.UUID) ([]domain.Transfer, error) {
	return s.transfers.List(ctx, userID)
}
