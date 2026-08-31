package service

import (
	"fmt"
	"time"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/helper"
)

// ExpenseQuery — общие параметры выборки трат. Используются и списком, и отчётом,
// поэтому разбор один на оба эндпоинта.
type ExpenseQuery struct {
	Period      string
	Date        string
	From        string
	To          string
	CategoryIDs []int64
	Page        int
	Limit       int
}

const (
	defaultLimit = 50
	maxLimit     = 200
)

// resolveRange превращает параметры запроса в диапазон дат.
//
// Приоритет: явные from/to важнее period. Если не задано ничего — диапазон
// пустой, то есть выборка за всё время.
func resolveRange(q ExpenseQuery, loc *time.Location) (helper.DateRange, error) {
	if q.From != "" || q.To != "" {
		if q.From != "" {
			if _, err := helper.ParseDate(q.From); err != nil {
				return helper.DateRange{}, err
			}
		}
		if q.To != "" {
			if _, err := helper.ParseDate(q.To); err != nil {
				return helper.DateRange{}, err
			}
		}
		if q.From != "" && q.To != "" && q.From > q.To {
			return helper.DateRange{}, fmt.Errorf("%w: from must not be after to", apperr.ErrValidation)
		}
		return helper.DateRange{From: q.From, To: q.To}, nil
	}

	if q.Period == "" {
		return helper.DateRange{}, nil
	}

	anchor := helper.Today(loc)
	if q.Date != "" {
		parsed, err := helper.ParseDate(q.Date)
		if err != nil {
			return helper.DateRange{}, err
		}
		anchor = parsed
	}

	return helper.ResolvePeriod(q.Period, anchor)
}

func normalizePaging(q ExpenseQuery) (limit, offset, page int, err error) {
	page = q.Page
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return 0, 0, 0, fmt.Errorf("%w: page must be >= 1", apperr.ErrValidation)
	}

	limit = q.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 1 || limit > maxLimit {
		return 0, 0, 0, fmt.Errorf("%w: limit must be between 1 and %d", apperr.ErrValidation, maxLimit)
	}

	return limit, (page - 1) * limit, page, nil
}
