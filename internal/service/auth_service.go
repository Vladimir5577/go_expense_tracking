package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/config"
	"go_expense_service/internal/dto"
	"go_expense_service/internal/model"
	"go_expense_service/internal/repository"
)

type AuthServiceInterface interface {
	Login(ctx context.Context, req dto.LoginRequest) (string, time.Time, *model.User, error)
	Me(ctx context.Context, userID int64) (*model.User, error)
}

type AuthService struct {
	repo repository.UserRepositoryInterface
	cfg  *config.Config
	// dummyHash — заглушка для сравнения пароля при несуществующем логине.
	// Без неё несуществующий логин отвечает за миллисекунду, а существующий —
	// за время bcrypt, и по разнице времени перебираются валидные логины.
	dummyHash []byte
}

func NewAuthService(repo repository.UserRepositoryInterface, cfg *config.Config) *AuthService {
	// Считаем заглушку тем же cost, что и рабочие хеши, иначе она не выровняет
	// время ответа. Разово при старте — это нормально.
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), cfg.BcryptCost)
	if err != nil {
		// Ошибка здесь означает некорректный cost — сервис всё равно нерабочий.
		panic("не удалось подготовить bcrypt-заглушку: " + err.Error())
	}
	return &AuthService{repo: repo, cfg: cfg, dummyHash: hash}
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (string, time.Time, *model.User, error) {
	now := s.cfg.Clock.Now()

	user, err := s.repo.GetByLogin(ctx, req.Login)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			// Тратим то же время, что и на реальную проверку.
			_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Password))
			return "", time.Time{}, nil, invalidCredentials()
		}
		return "", time.Time{}, nil, err
	}

	if user.IsLocked(now) {
		return "", time.Time{}, nil, apperr.New(apperr.CodeTooManyAttempts, "too many attempts")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		if err := s.registerFailure(ctx, user, now); err != nil {
			return "", time.Time{}, nil, err
		}
		return "", time.Time{}, nil, invalidCredentials()
	}

	if user.FailedAttempts > 0 || user.LockedUntil != nil {
		if err := s.repo.ResetLoginFailures(ctx, user.ID); err != nil {
			return "", time.Time{}, nil, err
		}
	}

	token, expiresAt, err := s.issueToken(user, now)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	return token, expiresAt, user, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64) (*model.User, error) {
	return s.repo.GetByID(ctx, userID)
}

// registerFailure увеличивает счётчик неудачных попыток и при достижении лимита
// ставит блокировку. Счётчик при этом обнуляется, чтобы после снятия блокировки
// у пользователя снова был полный набор попыток, а не одна.
func (s *AuthService) registerFailure(ctx context.Context, user *model.User, now time.Time) error {
	attempts := user.FailedAttempts + 1
	var lockedUntil *time.Time

	if attempts >= s.cfg.LoginMaxAttempts {
		until := now.Add(s.cfg.LoginLockDuration)
		lockedUntil = &until
		attempts = 0
	}

	return s.repo.SetLoginFailure(ctx, user.ID, attempts, lockedUntil)
}

func (s *AuthService) issueToken(user *model.User, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(s.cfg.JWTTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   strconv.FormatInt(user.ID, 10),
		"login": user.Login,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	})

	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// invalidCredentials — единый ответ и на несуществующий логин, и на неверный
// пароль. Различать их нельзя: по разнице ответов перебираются валидные логины.
func invalidCredentials() error {
	return apperr.New(apperr.CodeInvalidCredentials, "invalid credentials")
}
