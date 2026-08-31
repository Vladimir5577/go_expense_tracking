package handler

import (
	"net/http"
	"strings"

	"go_expense_service/internal/helper"
	"go_expense_service/internal/service"
)

// maxCategoryFilter — потолок на длину списка categoryIds.
const maxCategoryFilter = 50

// parseExpenseQuery разбирает параметры фильтрации. Общий для списка трат
// и для отчёта — набор параметров у них одинаковый.
func parseExpenseQuery(r *http.Request) (service.ExpenseQuery, error) {
	q := r.URL.Query()

	categoryIDs, err := helper.QueryIDList(r, "categoryIds", maxCategoryFilter)
	if err != nil {
		return service.ExpenseQuery{}, err
	}

	page, err := helper.QueryInt(r, "page", 0)
	if err != nil {
		return service.ExpenseQuery{}, err
	}

	limit, err := helper.QueryInt(r, "limit", 0)
	if err != nil {
		return service.ExpenseQuery{}, err
	}

	return service.ExpenseQuery{
		Period:      strings.TrimSpace(q.Get("period")),
		Date:        strings.TrimSpace(q.Get("date")),
		From:        strings.TrimSpace(q.Get("from")),
		To:          strings.TrimSpace(q.Get("to")),
		CategoryIDs: categoryIDs,
		Page:        page,
		Limit:       limit,
	}, nil
}
