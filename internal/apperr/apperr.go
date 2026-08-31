// Package apperr определяет доменные ошибки, независимые от транспорта (HTTP).
//
// Слои repository и service возвращают эти ошибки (через errors.Is / оборачивание
// в %w), а слой handler мапит их в HTTP-статусы функцией helper.WriteError.
// Так бизнес-логика не знает про HTTP, а handler не угадывает статус по тексту.
package apperr

type ErrorCode string

const (
	CodeInternal    ErrorCode = "internal_error"
	CodeInvalidJSON ErrorCode = "invalid_json"
	CodeValidation  ErrorCode = "validation_failed"

	CodeUnauthorized       ErrorCode = "unauthorized"
	CodeInvalidCredentials ErrorCode = "invalid_credentials"
	CodeTooManyAttempts    ErrorCode = "too_many_attempts"

	CodeNotFound         ErrorCode = "not_found"
	CodeCategoryNotFound ErrorCode = "category_not_found"
	CodeExpenseNotFound  ErrorCode = "expense_not_found"

	CodeConflict            ErrorCode = "conflict"
	CodeCategoryNameExists  ErrorCode = "category_name_exists"
	CodeCategoryHasExpenses ErrorCode = "category_has_expenses"
)

type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Is считает ошибку совпавшей как по точному коду, так и по категории:
// errors.Is(categoryNotFound, ErrNotFound) == true.
func (e *Error) Is(target error) bool {
	targetErr, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == targetErr.Code || categoryCode(e.Code) == targetErr.Code
}

func New(code ErrorCode, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func Wrap(code ErrorCode, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}

func categoryCode(code ErrorCode) ErrorCode {
	switch code {
	case CodeCategoryNotFound, CodeExpenseNotFound:
		return CodeNotFound
	case CodeCategoryNameExists, CodeCategoryHasExpenses:
		return CodeConflict
	case CodeInvalidCredentials:
		return CodeUnauthorized
	case CodeInvalidJSON:
		return CodeValidation
	default:
		return code
	}
}

var (
	// ErrNotFound — запрошенная сущность не существует (или принадлежит другому
	// пользователю — существование чужих данных мы не подтверждаем). → 404
	ErrNotFound = New(CodeNotFound, "not found")

	// ErrUnauthorized — пользователь не аутентифицирован. → 401
	ErrUnauthorized = New(CodeUnauthorized, "unauthorized")

	// ErrConflict — нарушение инварианта/уникальности. → 409
	ErrConflict = New(CodeConflict, "conflict")

	// ErrValidation — некорректные входные данные запроса. → 400
	ErrValidation = New(CodeValidation, "validation failed")
)
