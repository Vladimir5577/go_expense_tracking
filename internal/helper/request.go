package helper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"go_expense_service/internal/apperr"
)

// IDParam читает числовой параметр пути (например {id}) и возвращает его как int64.
func IDParam(r *http.Request, name string) (int64, error) {
	raw := chi.URLParam(r, name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: invalid %s", apperr.ErrValidation, name)
	}
	return id, nil
}

// QueryInt читает целочисленный query-параметр. Пустой параметр даёт fallback.
func QueryInt(r *http.Request, name string, fallback int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s", apperr.ErrValidation, name)
	}
	return v, nil
}

// QueryIDList разбирает список идентификаторов через запятую: "1,3,7".
// Дубликаты схлопываются, порядок сохраняется.
func QueryIDList(r *http.Request, name string, maxItems int) ([]int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	seen := make(map[int64]struct{}, len(parts))
	ids := make([]int64, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("%w: invalid %s", apperr.ErrValidation, name)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) > maxItems {
		return nil, fmt.Errorf("%w: too many %s, max %d", apperr.ErrValidation, name, maxItems)
	}
	return ids, nil
}
