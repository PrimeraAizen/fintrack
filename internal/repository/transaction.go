package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Transaction interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CreateTx(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error
	GetForUser(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error)
	GetForUserTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) (*domain.Transaction, error)
	UpdateTx(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error
	DeleteTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int64, error)
}

type TransactionRepository struct {
	pg *postgres.Postgres
}

func NewTransactionRepository(pg *postgres.Postgres) *TransactionRepository {
	return &TransactionRepository{pg: pg}
}

func (r *TransactionRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pg.Pool.BeginTx(ctx, pgx.TxOptions{})
}

func (r *TransactionRepository) CreateTx(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error {
	query, args, err := r.pg.Builder.
		Insert("transactions").
		Columns("account_id", "category_id", "amount", "currency", "converted_amount", "note", "transaction_date").
		Values(t.AccountID, t.CategoryID, t.Amount, t.Currency, t.ConvertedAmount, nullableString(t.Note), t.TransactionDate).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert transaction: %w", err)
	}
	if err := tx.QueryRow(ctx, query, args...).Scan(&t.ID, &t.CreatedAt); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetForUser(ctx context.Context, userID, id uuid.UUID) (*domain.Transaction, error) {
	t, err := r.queryOne(ctx, r.pg.Pool, userID, id)
	return t, err
}

func (r *TransactionRepository) GetForUserTx(ctx context.Context, tx pgx.Tx, userID, id uuid.UUID) (*domain.Transaction, error) {
	return r.queryOne(ctx, tx, userID, id)
}

type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *TransactionRepository) queryOne(ctx context.Context, q rowQuerier, userID, id uuid.UUID) (*domain.Transaction, error) {
	query, args, err := r.pg.Builder.
		Select(
			"t.id", "t.account_id", "t.category_id", "c.type",
			"t.amount", "t.currency",
			"t.converted_amount", "COALESCE(t.note, '')", "t.transaction_date", "t.created_at",
		).
		From("transactions t").
		Join("accounts a ON a.id = t.account_id").
		Join("categories c ON c.id = t.category_id").
		Where(sq.Eq{"t.id": id, "a.user_id": userID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select transaction: %w", err)
	}
	var t domain.Transaction
	if err := q.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.AccountID, &t.CategoryID, &t.Type,
		&t.Amount, &t.Currency,
		&t.ConvertedAmount, &t.Note, &t.TransactionDate, &t.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select transaction: %w", err)
	}
	return &t, nil
}

func (r *TransactionRepository) UpdateTx(ctx context.Context, tx pgx.Tx, t *domain.Transaction) error {
	query, args, err := r.pg.Builder.
		Update("transactions").
		Set("category_id", t.CategoryID).
		Set("amount", t.Amount).
		Set("currency", t.Currency).
		Set("converted_amount", t.ConvertedAmount).
		Set("note", nullableString(t.Note)).
		Set("transaction_date", t.TransactionDate).
		Where(sq.Eq{"id": t.ID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update transaction: %w", err)
	}
	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepository) DeleteTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := tx.Exec(ctx, "DELETE FROM transactions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepository) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, int64, error) {
	whereClauses := []sq.Sqlizer{sq.Eq{"a.user_id": userID}}
	if filter.AccountID != nil {
		whereClauses = append(whereClauses, sq.Eq{"t.account_id": *filter.AccountID})
	}
	if filter.CategoryID != nil {
		whereClauses = append(whereClauses, sq.Eq{"t.category_id": *filter.CategoryID})
	}
	if filter.FromDate != nil {
		whereClauses = append(whereClauses, sq.GtOrEq{"t.transaction_date": *filter.FromDate})
	}
	if filter.ToDate != nil {
		whereClauses = append(whereClauses, sq.LtOrEq{"t.transaction_date": *filter.ToDate})
	}
	if t := strings.TrimSpace(filter.Type); t != "" {
		whereClauses = append(whereClauses, sq.Eq{"c.type": t})
	}

	countBuilder := r.pg.Builder.
		Select("COUNT(*)").
		From("transactions t").
		Join("accounts a ON a.id = t.account_id").
		Join("categories c ON c.id = t.category_id").
		Where(sq.And(whereClauses))

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count transactions: %w", err)
	}
	var total int64
	if err := r.pg.Pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	page := domain.PageParams{Page: filter.Page, PerPage: filter.PerPage}.Normalize()

	listBuilder := r.pg.Builder.
		Select(
			"t.id", "t.account_id", "t.category_id", "c.type",
			"t.amount", "t.currency",
			"t.converted_amount", "COALESCE(t.note, '')", "t.transaction_date", "t.created_at",
		).
		From("transactions t").
		Join("accounts a ON a.id = t.account_id").
		Join("categories c ON c.id = t.category_id").
		Where(sq.And(whereClauses)).
		OrderBy("t.transaction_date DESC", "t.created_at DESC").
		Limit(uint64(page.PerPage)).
		Offset(uint64(page.Offset()))

	query, args, err := listBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list transactions: %w", err)
	}
	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Transaction, 0)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(
			&t.ID, &t.AccountID, &t.CategoryID, &t.Type,
			&t.Amount, &t.Currency,
			&t.ConvertedAmount, &t.Note, &t.TransactionDate, &t.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}
