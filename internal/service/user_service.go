package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/config"
	"go_expense_service/internal/model"
	"go_expense_service/internal/repository"
)

const (
	minLoginLen    = 3
	maxLoginLen    = 50
	minPasswordLen = 4
)

// UserService — управление пользователями. Используется только консольной
// утилитой useradm: регистрации через API в сервисе нет.
type UserService struct {
	repo repository.UserRepositoryInterface
	cfg  *config.Config
}

func NewUserService(repo repository.UserRepositoryInterface, cfg *config.Config) *UserService {
	return &UserService{repo: repo, cfg: cfg}
}

func (s *UserService) Create(ctx context.Context, login, password, name string) (*model.User, error) {
	login = strings.TrimSpace(login)
	if err := validateLogin(login); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetByLogin(ctx, login); err == nil {
		return nil, apperr.New(apperr.CodeConflict, "пользователь с таким логином уже существует")
	} else if !errors.Is(err, apperr.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return nil, err
	}

	now := s.cfg.Clock.Now()
	return s.repo.Create(ctx, &model.User{
		Login:        login,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(name),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

// ChangePassword меняет пароль. Ранее выданные JWT при этом остаются
// действительными до истечения срока — отзыва токенов в сервисе нет.
// Чтобы погасить все токены разом, меняется JWT_SECRET и сервис перезапускается.
func (s *UserService) ChangePassword(ctx context.Context, login, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(ctx, user.ID, string(hash))
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Unlock(ctx context.Context, login string) error {
	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return err
	}
	return s.repo.ResetLoginFailures(ctx, user.ID)
}

func validateLogin(login string) error {
	if len(login) < minLoginLen || len(login) > maxLoginLen {
		return fmt.Errorf("%w: логин должен быть длиной от %d до %d символов",
			apperr.ErrValidation, minLoginLen, maxLoginLen)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLen {
		return fmt.Errorf("%w: пароль должен быть не короче %d символов",
			apperr.ErrValidation, minPasswordLen)
	}
	return nil
}
