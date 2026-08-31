package model

import "time"

// Expense — трата. SpentAt — календарная дата в формате YYYY-MM-DD.
type Expense struct {
	ID           int64
	UserID       int64
	CategoryID   int64
	CategoryName string
	Amount       float64
	Description  string
	SpentAt      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ExpenseFilter — параметры выборки трат. Пустые поля означают «без фильтра».
type ExpenseFilter struct {
	UserID      int64
	From        string
	To          string
	CategoryIDs []int64
	Limit       int
	Offset      int
}

// CategoryTotal — строка отчёта: сумма и количество трат по одной категории.
// Share (доля в процентах) заполняет сервис, репозиторий её не считает.
type CategoryTotal struct {
	CategoryID int64
	Name       string
	Total      float64
	Count      int64
	Share      float64
}

// Summary — отчёт по тратам за период в разрезе категорий.
type Summary struct {
	From       string
	To         string
	Total      float64
	ByCategory []CategoryTotal
}
