package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type User interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateBaseCurrency(ctx context.Context, id uuid.UUID, baseCurrency string) error
}

type UserRepository struct {
	pg *postgres.Postgres
}

func NewUserRepository(pg *postgres.Postgres) *UserRepository {
	return &UserRepository{pg: pg}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query, args, err := r.pg.Builder.
		Insert("users").
		Columns("email", "password_hash", "base_currency").
		Values(user.Email, user.PasswordHash, user.BaseCurrency).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert user: %w", err)
	}

	row := r.pg.Pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query, args, err := r.pg.Builder.
		Select("id", "email", "password_hash", "base_currency", "created_at", "updated_at").
		From("users").
		Where("email = ?", email).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select user: %w", err)
	}

	var u domain.User
	row := r.pg.Pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.BaseCurrency, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query, args, err := r.pg.Builder.
		Select("id", "email", "password_hash", "base_currency", "created_at", "updated_at").
		From("users").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select user: %w", err)
	}

	var u domain.User
	row := r.pg.Pool.QueryRow(ctx, query, args...)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.BaseCurrency, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) UpdateBaseCurrency(ctx context.Context, id uuid.UUID, baseCurrency string) error {
	query, args, err := r.pg.Builder.
		Update("users").
		Set("base_currency", baseCurrency).
		Set("updated_at", sq.Expr("NOW()")).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update user: %w", err)
	}

	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update user base currency: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
