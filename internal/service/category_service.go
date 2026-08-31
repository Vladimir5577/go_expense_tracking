package service

import (
	"context"
	"errors"
	"strings"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/config"
	"go_expense_service/internal/dto"
	"go_expense_service/internal/model"
	"go_expense_service/internal/repository"
)

type CategoryServiceInterface interface {
	List(ctx context.Context, userID int64) ([]model.Category, error)
	Get(ctx context.Context, userID, id int64) (*model.Category, error)
	Create(ctx context.Context, userID int64, req dto.CreateCategoryRequest) (*model.Category, error)
	Update(ctx context.Context, userID, id int64, req dto.UpdateCategoryRequest) (*model.Category, error)
	Delete(ctx context.Context, userID, id int64) error
}

type CategoryService struct {
	repo repository.CategoryRepositoryInterface
	cfg  *config.Config
}

func NewCategoryService(repo repository.CategoryRepositoryInterface, cfg *config.Config) *CategoryService {
	return &CategoryService{repo: repo, cfg: cfg}
}

func (s *CategoryService) List(ctx context.Context, userID int64) ([]model.Category, error) {
	return s.repo.List(ctx, userID)
}

func (s *CategoryService) Get(ctx context.Context, userID, id int64) (*model.Category, error) {
	c, err := s.repo.GetByID(ctx, userID, id)
	return c, mapCategoryNotFound(err)
}

func (s *CategoryService) Create(ctx context.Context, userID int64, req dto.CreateCategoryRequest) (*model.Category, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperr.New(apperr.CodeValidation, "name is required")
	}

	taken, err := s.repo.ExistsByName(ctx, userID, name, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, apperr.New(apperr.CodeCategoryNameExists, "category name already exists")
	}

	now := s.cfg.Clock.Now()
	category := &model.Category{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Color:       strings.TrimSpace(req.Color),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.Create(ctx, category)
	if err != nil {
		// Гонка между проверкой и вставкой: UNIQUE(user_id, name) ловит то,
		// что проверка пропустила.
		if repository.IsUniqueViolation(err) {
			return nil, apperr.New(apperr.CodeCategoryNameExists, "category name already exists")
		}
		return nil, err
	}
	return created, nil
}

func (s *CategoryService) Update(ctx context.Context, userID, id int64, req dto.UpdateCategoryRequest) (*model.Category, error) {
	category, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, mapCategoryNotFound(err)
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, apperr.New(apperr.CodeValidation, "name must not be empty")
		}
		if name != category.Name {
			taken, err := s.repo.ExistsByName(ctx, userID, name, id)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, apperr.New(apperr.CodeCategoryNameExists, "category name already exists")
			}
		}
		category.Name = name
	}
	if req.Description != nil {
		category.Description = strings.TrimSpace(*req.Description)
	}
	if req.Color != nil {
		category.Color = strings.TrimSpace(*req.Color)
	}

	category.UpdatedAt = s.cfg.Clock.Now()

	if err := s.repo.Update(ctx, category); err != nil {
		if repository.IsUniqueViolation(err) {
			return nil, apperr.New(apperr.CodeCategoryNameExists, "category name already exists")
		}
		return nil, mapCategoryNotFound(err)
	}
	return category, nil
}

// Delete запрещает удаление категории, на которой висят траты.
// Проверяем явным запросом, чтобы вернуть внятный код, а не ловить сырое
// нарушение внешнего ключа. FK ON DELETE RESTRICT остаётся страховкой.
func (s *CategoryService) Delete(ctx context.Context, userID, id int64) error {
	if _, err := s.repo.GetByID(ctx, userID, id); err != nil {
		return mapCategoryNotFound(err)
	}

	hasExpenses, err := s.repo.HasExpenses(ctx, userID, id)
	if err != nil {
		return err
	}
	if hasExpenses {
		return apperr.New(apperr.CodeCategoryHasExpenses, "category has expenses")
	}

	return mapCategoryNotFound(s.repo.Delete(ctx, userID, id))
}

// mapCategoryNotFound уточняет общий not_found до category_not_found.
func mapCategoryNotFound(err error) error {
	if errors.Is(err, apperr.ErrNotFound) {
		return apperr.New(apperr.CodeCategoryNotFound, "category not found")
	}
	return err
}
