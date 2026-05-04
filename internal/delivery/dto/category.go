package dto

import (
	"time"

	"github.com/diyas/fintrack/internal/domain"
	"github.com/diyas/fintrack/internal/service"
	"github.com/google/uuid"
)

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required,max=50"`
	Type string `json:"type" binding:"required,oneof=income expense"`
	Icon string `json:"icon" binding:"omitempty,max=50"`
}

func (r CreateCategoryRequest) ToInput() service.CreateCategoryInput {
	return service.CreateCategoryInput{
		Name: r.Name,
		Type: r.Type,
		Icon: r.Icon,
	}
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"omitempty,max=50"`
	Icon string `json:"icon" binding:"omitempty,max=50"`
}

func (r UpdateCategoryRequest) ToInput() service.UpdateCategoryInput {
	return service.UpdateCategoryInput{
		Name: r.Name,
		Icon: r.Icon,
	}
}

type CategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Icon      string    `json:"icon,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func CategoryResponseFrom(c *domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		Icon:      c.Icon,
		CreatedAt: c.CreatedAt,
	}
}

func CategoryListResponseFrom(items []domain.Category) []CategoryResponse {
	out := make([]CategoryResponse, len(items))
	for i := range items {
		out[i] = CategoryResponseFrom(&items[i])
	}
	return out
}
