package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"go_expense_service/internal/apperr"
)

// NormalizeError переводит ошибки драйвера в доменные.
func NormalizeError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return apperr.ErrNotFound
	}
	return err
}

// IsUniqueViolation сообщает, что запрос нарушил UNIQUE-ограничение.
// Проверяем по тексту ошибки, а не по коду драйвера: так репозиторий не зависит
// от конкретной реализации SQLite.
func IsUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// parseTime разбирает время из TEXT-колонки. Время пишем сами и всегда в
// RFC3339, поэтому ошибка здесь означала бы порчу данных — возвращаем нулевое
// время, а не роняем запрос.
func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseTimePtr(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t := parseTime(ns.String)
	if t.IsZero() {
		return nil
	}
	return &t
}
