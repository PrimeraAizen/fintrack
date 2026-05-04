package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/diyas/fintrack/internal/domain"
	postgres "github.com/diyas/fintrack/pkg/adapter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Category interface {
	Create(ctx context.Context, c *domain.Category) error
	CreateMany(ctx context.Context, items []domain.Category) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error)
	List(ctx context.Context, userID uuid.UUID, typeFilter string) ([]domain.Category, error)
	Update(ctx context.Context, userID, id uuid.UUID, name, icon string) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

type CategoryRepository struct {
	pg *postgres.Postgres
}

func NewCategoryRepository(pg *postgres.Postgres) *CategoryRepository {
	return &CategoryRepository{pg: pg}
}

func (r *CategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	query, args, err := r.pg.Builder.
		Insert("categories").
		Columns("user_id", "name", "type", "icon").
		Values(c.UserID, c.Name, c.Type, nullableString(c.Icon)).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert category: %w", err)
	}
	if err := r.pg.Pool.QueryRow(ctx, query, args...).Scan(&c.ID, &c.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (r *CategoryRepository) CreateMany(ctx context.Context, items []domain.Category) error {
	if len(items) == 0 {
		return nil
	}
	builder := r.pg.Builder.Insert("categories").Columns("user_id", "name", "type", "icon")
	for _, item := range items {
		builder = builder.Values(item.UserID, item.Name, item.Type, nullableString(item.Icon))
	}
	builder = builder.Suffix("ON CONFLICT (user_id, name) DO NOTHING")
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build bulk insert categories: %w", err)
	}
	if _, err := r.pg.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("bulk insert categories: %w", err)
	}
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	query, args, err := r.pg.Builder.
		Select("id", "user_id", "name", "type", "COALESCE(icon, '')", "created_at").
		From("categories").
		Where("id = ? AND user_id = ?", id, userID).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select category: %w", err)
	}
	var c domain.Category
	if err := r.pg.Pool.QueryRow(ctx, query, args...).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type, &c.Icon, &c.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select category: %w", err)
	}
	return &c, nil
}

func (r *CategoryRepository) List(ctx context.Context, userID uuid.UUID, typeFilter string) ([]domain.Category, error) {
	builder := r.pg.Builder.
		Select("id", "user_id", "name", "type", "COALESCE(icon, '')", "created_at").
		From("categories").
		Where("user_id = ?", userID).
		OrderBy("name ASC")
	if typeFilter != "" {
		builder = builder.Where("type = ?", typeFilter)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list categories: %w", err)
	}
	rows, err := r.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	var out []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.Icon, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepository) Update(ctx context.Context, userID, id uuid.UUID, name, icon string) error {
	builder := r.pg.Builder.Update("categories").Where("id = ? AND user_id = ?", id, userID)
	if name != "" {
		builder = builder.Set("name", name)
	}
	if icon != "" {
		builder = builder.Set("icon", icon)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build update category: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("update category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	query, args, err := r.pg.Builder.
		Delete("categories").
		Where("id = ? AND user_id = ?", id, userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete category: %w", err)
	}
	tag, err := r.pg.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
