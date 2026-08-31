// Package helper содержит утилиты HTTP-слоя: единообразную запись JSON-ответов
// и маппинг доменных ошибок (apperr) в HTTP-статусы.
package helper

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go_expense_service/internal/apperr"
)

// WriteJSON пишет data как JSON с указанным статусом.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Заголовки уже отправлены — ответ частично ушёл, писать в него нечего.
		slog.Error("failed to encode response", "error", err)
	}
}

// WriteError переводит доменную ошибку в HTTP-ответ.
func WriteError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := apperr.CodeInternal

	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		code = appErr.Code
		status = statusForErrorCode(appErr.Code)
	}

	// 5xx маскируются в ответе кодом internal_error, поэтому исходную ошибку
	// логируем здесь централизованно — иначе причина пропадёт бесследно.
	if status >= http.StatusInternalServerError {
		slog.Error("request failed", "code", code, "error", err)
		code = apperr.CodeInternal
	}

	WriteJSON(w, status, map[string]string{
		"error": string(code),
		"code":  string(code),
	})
}

func statusForErrorCode(code apperr.ErrorCode) int {
	switch code {
	case apperr.CodeInvalidJSON, apperr.CodeValidation:
		return http.StatusBadRequest
	case apperr.CodeUnauthorized, apperr.CodeInvalidCredentials:
		return http.StatusUnauthorized
	case apperr.CodeNotFound, apperr.CodeCategoryNotFound, apperr.CodeExpenseNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict, apperr.CodeCategoryNameExists, apperr.CodeCategoryHasExpenses:
		return http.StatusConflict
	case apperr.CodeTooManyAttempts:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
