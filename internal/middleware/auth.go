package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/helper"
)

type contextKey string

const UserContextKey contextKey = "user"

// UserClaims — то, что middleware кладёт в контекст запроса.
type UserClaims struct {
	ID    int64
	Login string
}

type AuthMiddleware struct {
	secret []byte
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: []byte(secret)}
}

// Handler проверяет подпись токена и срок действия — и больше ничего.
// В БД не ходит: отзыва отдельных токенов в сервисе нет, аварийный отзыв всех
// токенов делается сменой JWT_SECRET и перезапуском.
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := bearerToken(r)
		if !ok {
			helper.WriteError(w, apperr.ErrUnauthorized)
			return
		}

		// WithExpirationRequired обязателен: без него токен без claim exp
		// считался бы бессрочным.
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return m.secret, nil
		}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || !token.Valid {
			helper.WriteError(w, apperr.ErrUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			helper.WriteError(w, apperr.ErrUnauthorized)
			return
		}

		userClaims, err := parseClaims(claims)
		if err != nil {
			helper.WriteError(w, apperr.ErrUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	return parts[1], true
}

func parseClaims(claims jwt.MapClaims) (UserClaims, error) {
	// sub по спецификации JWT — строка, но чужие библиотеки иногда кладут число.
	// Принимаем оба варианта, чтобы токен не отвергался из-за формата.
	var id int64
	switch v := claims["sub"].(type) {
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return UserClaims{}, errors.New("invalid sub")
		}
		id = parsed
	case float64:
		id = int64(v)
	default:
		return UserClaims{}, errors.New("missing sub")
	}

	if id <= 0 {
		return UserClaims{}, errors.New("invalid sub")
	}

	login, _ := claims["login"].(string)
	return UserClaims{ID: id, Login: login}, nil
}

// GetUser достаёт данные пользователя из контекста запроса.
func GetUser(ctx context.Context) (UserClaims, bool) {
	user, ok := ctx.Value(UserContextKey).(UserClaims)
	return user, ok
}

// UserID — короткая форма для хендлеров: id текущего пользователя либо
// ошибка unauthorized, если middleware по какой-то причине не отработал.
func UserID(ctx context.Context) (int64, error) {
	user, ok := GetUser(ctx)
	if !ok || user.ID <= 0 {
		return 0, apperr.ErrUnauthorized
	}
	return user.ID, nil
}
