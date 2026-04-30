package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/repository"
	"github.com/google/uuid"
)

var defaultCategories = []struct {
	Name string
	Type string
}{
	{"Salary", domain.CategoryTypeIncome},
	{"Freelance", domain.CategoryTypeIncome},
	{"Investment", domain.CategoryTypeIncome},
	{"Gift", domain.CategoryTypeIncome},
	{"Other Income", domain.CategoryTypeIncome},
	{"Food", domain.CategoryTypeExpense},
	{"Transport", domain.CategoryTypeExpense},
	{"Housing", domain.CategoryTypeExpense},
	{"Entertainment", domain.CategoryTypeExpense},
	{"Health", domain.CategoryTypeExpense},
	{"Education", domain.CategoryTypeExpense},
	{"Shopping", domain.CategoryTypeExpense},
	{"Other", domain.CategoryTypeExpense},
}

type CreateCategoryInput struct {
	Name string
	Type string
	Icon string
}

type UpdateCategoryInput struct {
	Name string
	Icon string
}

type Category interface {
	Create(ctx context.Context, userID uuid.UUID, in CreateCategoryInput) (*domain.Category, error)
	List(ctx context.Context, userID uuid.UUID, typeFilter string) ([]domain.Category, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error)
	Update(ctx context.Context, userID, id uuid.UUID, in UpdateCategoryInput) (*domain.Category, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	SeedDefaults(ctx context.Context, userID uuid.UUID) error
}

type CategoryService struct {
	repo repository.Category
}

func NewCategoryService(repo repository.Category) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, in CreateCategoryInput) (*domain.Category, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name required", domain.ErrInvalidInput)
	}
	if in.Type != domain.CategoryTypeIncome && in.Type != domain.CategoryTypeExpense {
		return nil, fmt.Errorf("%w: invalid category type", domain.ErrInvalidInput)
	}
	c := &domain.Category{
		UserID: userID,
		Name:   in.Name,
		Type:   in.Type,
		Icon:   in.Icon,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID, typeFilter string) ([]domain.Category, error) {
	if typeFilter != "" && typeFilter != domain.CategoryTypeIncome && typeFilter != domain.CategoryTypeExpense {
		return nil, fmt.Errorf("%w: invalid type filter", domain.ErrInvalidInput)
	}
	return s.repo.List(ctx, userID, typeFilter)
}

func (s *CategoryService) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Category, error) {
	return s.repo.GetByID(ctx, userID, id)
}

func (s *CategoryService) Update(ctx context.Context, userID, id uuid.UUID, in UpdateCategoryInput) (*domain.Category, error) {
	if err := s.repo.Update(ctx, userID, id, in.Name, in.Icon); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID, id)
}

func (s *CategoryService) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *CategoryService) SeedDefaults(ctx context.Context, userID uuid.UUID) error {
	items := make([]domain.Category, 0, len(defaultCategories))
	for _, def := range defaultCategories {
		items = append(items, domain.Category{
			UserID: userID,
			Name:   def.Name,
			Type:   def.Type,
		})
	}
	return s.repo.CreateMany(ctx, items)
}
