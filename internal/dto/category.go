package dto

import (
	"time"

	"go_expense_service/internal/model"
)

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Color       string `json:"color" validate:"omitempty,max=30"`
}

// UpdateCategoryRequest — частичное обновление: nil означает «поле не трогаем»,
// поэтому все поля указатели.
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" validate:"omitempty,max=30"`
}

type CategoryResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func MapCategoryResponse(c *model.Category) *CategoryResponse {
	if c == nil {
		return nil
	}
	return &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Color:       c.Color,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

func MapCategoriesResponse(categories []model.Category) []*CategoryResponse {
	resp := make([]*CategoryResponse, 0, len(categories))
	for i := range categories {
		resp = append(resp, MapCategoryResponse(&categories[i]))
	}
	return resp
}
