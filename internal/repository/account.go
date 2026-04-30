package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type Account interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error)
	List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	Update(ctx context.Context, userID, id uuid.UUID, name, accType string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	UpdateBalance(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, delta decimal.Decimal) error
	GetByIDTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) (*domain.Account, error)
}

type AccountRepository struct {
	pg *postgres.Postgres
}

func NewAccountRepository(pg *postgres.Postgres) *AccountRepository {
	return &AccountRepository{pg: pg}
}

func (r *AccountRepository) Create(ctx context.Context, a *domain.Account) error {
	query, args, err := r.pg.Builder.
		Insert("accounts").
		Columns("user_id", "name", "type", "currency", "balance").
		Values(a.UserID, a.Name, a.Type, a.Currency, a.Balance).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert account: %w", err)
	}
	if err := r.pg.Pool.QueryRow(ctx, query, args...).Scan(&a.ID, &a.CreatedAt); err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

func (r *AccountRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Account, error) {
	query, args, err := r.pg.Builder.
		Select("id", "user_id", "name", "type", "currency", "balance", "created_at").
		From("accounts").
		Where("id = ? AND user_id = ?", id, userID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select account: %w", err)
	}
	var a domain.Account
	if err := r.pg.Pool.QueryRow(ctx, query, args...).Scan(
		&a.ID, &a.UserID, &a.Name, &a.Type, &a.Currency, &a.Balance, &a.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select account: %w", err)
	}
	return &a, nil
}

func (r *AccountRepository) GetByIDTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) (*domain.Account, error) {
	query, args, err := r.pg.Builder.
		Select("id", "user_id", "name", "type", "currency", "balance", "created_at").
		From("accounts").
		Where("id = ? AND user_id = ?", id, userID).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select account tx: %w", err)
	}
	var a domain.Account
	if err := tx.QueryRow(ctx, query, args...).Scan(
		&a.ID, &a.UserID, &a.Name, &a.Type, &a.Currency, &a.Balance, &a.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select account tx: %w", err)
	}
	return &a, nil
}

func (r *AccountRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	query, args, err := r.pg.Builder.
		Select("id", "user_id", "name", "type", "currency", "balance", "created_at").
		From("accounts").
		Where("user_id = ?", userID).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list accounts: %w", err)
	}
	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query list accounts: %w", err)
	}
	defer rows.Close()

	var out []domain.Account
	for rows.Next() {
		var a domain.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Currency, &a.Balance, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AccountRepository) Update(ctx context.Context, userID, id uuid.UUID, name, accType string) error {
	builder := r.pg.Builder.Update("accounts").Where("id = ? AND user_id = ?", id, userID)
	if name != "" {
		builder = builder.Set("name", name)
	}
	if accType != "" {
		builder = builder.Set("type", accType)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build update account: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	query, args, err := r.pg.Builder.
		Delete("accounts").
		Where("id = ? AND user_id = ?", id, userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete account: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, delta decimal.Decimal) error {
	_, err := tx.Exec(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		delta, accountID,
	)
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}
	return nil
}
