package model

import "time"

// Category — категория затрат. Принадлежит конкретному пользователю,
// имя уникально в рамках пользователя.
type Category struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
