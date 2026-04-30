package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Transfer interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CreateTx(ctx context.Context, tx pgx.Tx, t *domain.Transfer) error
	List(ctx context.Context, userID uuid.UUID) ([]domain.Transfer, error)
}

type TransferRepository struct {
	pg *postgres.Postgres
}

func NewTransferRepository(pg *postgres.Postgres) *TransferRepository {
	return &TransferRepository{pg: pg}
}

func (r *TransferRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pg.Pool.BeginTx(ctx, pgx.TxOptions{})
}

func (r *TransferRepository) CreateTx(ctx context.Context, tx pgx.Tx, t *domain.Transfer) error {
	query, args, err := r.pg.Builder.
		Insert("transfers").
		Columns("from_account_id", "to_account_id", "amount", "from_currency", "to_currency", "exchange_rate").
		Values(t.FromAccountID, t.ToAccountID, t.Amount, t.FromCurrency, t.ToCurrency, t.ExchangeRate).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert transfer: %w", err)
	}
	if err := tx.QueryRow(ctx, query, args...).Scan(&t.ID, &t.CreatedAt); err != nil {
		return fmt.Errorf("insert transfer: %w", err)
	}
	return nil
}

func (r *TransferRepository) List(ctx context.Context, userID uuid.UUID) ([]domain.Transfer, error) {
	query, args, err := r.pg.Builder.
		Select(
			"t.id", "t.from_account_id", "t.to_account_id", "t.amount",
			"t.from_currency", "t.to_currency", "t.exchange_rate", "t.created_at",
		).
		From("transfers t").
		Join("accounts af ON af.id = t.from_account_id").
		Join("accounts at ON at.id = t.to_account_id").
		Where(sq.Or{sq.Eq{"af.user_id": userID}, sq.Eq{"at.user_id": userID}}).
		OrderBy("t.created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list transfers: %w", err)
	}
	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query transfers: %w", err)
	}
	defer rows.Close()
	var out []domain.Transfer
	for rows.Next() {
		var t domain.Transfer
		if err := rows.Scan(
			&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount,
			&t.FromCurrency, &t.ToCurrency, &t.ExchangeRate, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
