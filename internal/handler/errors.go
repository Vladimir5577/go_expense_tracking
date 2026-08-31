package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	gpvalidator "github.com/go-playground/validator/v10"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/validator"
)

// decodeAndValidate разбирает тело запроса и прогоняет его через валидатор.
// Единая точка: иначе каждый хендлер по-своему отвечает на кривой JSON.
func decodeAndValidate(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return apperr.New(apperr.CodeInvalidJSON, "invalid json")
	}
	if err := validator.Validate.Struct(dst); err != nil {
		return validationError(err)
	}
	return nil
}

// validationError превращает ошибки валидатора в доменную ошибку с перечнем
// полей — клиенту нужно знать, что именно не прошло.
func validationError(err error) error {
	var validationErrors gpvalidator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fields := make([]string, 0, len(validationErrors))
		for _, fieldErr := range validationErrors {
			fields = append(fields, fieldErr.Field()+":"+fieldErr.Tag())
		}
		return apperr.New(apperr.CodeValidation, "validation failed: "+strings.Join(fields, ", "))
	}
	return apperr.New(apperr.CodeValidation, "validation failed")
}
