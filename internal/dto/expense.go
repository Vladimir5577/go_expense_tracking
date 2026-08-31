package dto

import (
	"time"

	"go_expense_service/internal/model"
)

// Потолок суммы (lte=1000000) — защита от опечатки в три нуля,
// которая иначе перекосит все отчёты.
type CreateExpenseRequest struct {
	CategoryID  int64   `json:"categoryId" validate:"required,gt=0"`
	Amount      float64 `json:"amount" validate:"required,gt=0,lte=1000000"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	// Пустое значение означает «сегодня» в таймзоне сервиса.
	SpentAt string `json:"spentAt" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateExpenseRequest struct {
	CategoryID  *int64   `json:"categoryId,omitempty" validate:"omitempty,gt=0"`
	Amount      *float64 `json:"amount,omitempty" validate:"omitempty,gt=0,lte=1000000"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	SpentAt     *string  `json:"spentAt,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

type ExpenseResponse struct {
	ID           int64   `json:"id"`
	CategoryID   int64   `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	SpentAt      string  `json:"spentAt"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type ExpenseListResponse struct {
	Items []*ExpenseResponse `json:"items"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
	// TotalItems — количество записей, подходящих под фильтр (не сумма денег).
	// Сумма денег живёт в отчёте /api/reports/summary под именем total.
	TotalItems int64 `json:"totalItems"`
}

func MapExpenseResponse(e *model.Expense) *ExpenseResponse {
	if e == nil {
		return nil
	}
	return &ExpenseResponse{
		ID:           e.ID,
		CategoryID:   e.CategoryID,
		CategoryName: e.CategoryName,
		Amount:       e.Amount,
		Description:  e.Description,
		SpentAt:      e.SpentAt,
		CreatedAt:    e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    e.UpdatedAt.Format(time.RFC3339),
	}
}

func MapExpensesResponse(expenses []model.Expense) []*ExpenseResponse {
	resp := make([]*ExpenseResponse, 0, len(expenses))
	for i := range expenses {
		resp = append(resp, MapExpenseResponse(&expenses[i]))
	}
	return resp
}
