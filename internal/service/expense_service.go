package service

import (
	"context"
	"errors"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/config"
	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/model"
	"go_expense_service/internal/repository"
)

// ExpenseList — страница списка трат вместе с параметрами пагинации.
type ExpenseList struct {
	Items []model.Expense
	Page  int
	Limit int
	// TotalItems — сколько записей подходит под фильтр, а не сумма денег.
	TotalItems int64
}

type ExpenseServiceInterface interface {
	List(ctx context.Context, userID int64, q ExpenseQuery) (*ExpenseList, error)
	Get(ctx context.Context, userID, id int64) (*model.Expense, error)
	Create(ctx context.Context, userID int64, req dto.CreateExpenseRequest) (*model.Expense, error)
	Update(ctx context.Context, userID, id int64, req dto.UpdateExpenseRequest) (*model.Expense, error)
	Delete(ctx context.Context, userID, id int64) error
}

type ExpenseService struct {
	repo         repository.ExpenseRepositoryInterface
	categoryRepo repository.CategoryRepositoryInterface
	cfg          *config.Config
}

func NewExpenseService(
	repo repository.ExpenseRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface,
	cfg *config.Config,
) *ExpenseService {
	return &ExpenseService{repo: repo, categoryRepo: categoryRepo, cfg: cfg}
}

func (s *ExpenseService) List(ctx context.Context, userID int64, q ExpenseQuery) (*ExpenseList, error) {
	filter, page, limit, err := s.buildFilter(userID, q)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ExpenseList{Items: items, Page: page, Limit: limit, TotalItems: total}, nil
}

func (s *ExpenseService) Get(ctx context.Context, userID, id int64) (*model.Expense, error) {
	e, err := s.repo.GetByID(ctx, userID, id)
	return e, mapExpenseNotFound(err)
}

func (s *ExpenseService) Create(ctx context.Context, userID int64, req dto.CreateExpenseRequest) (*model.Expense, error) {
	// Категория обязана принадлежать тому же пользователю — иначе трату можно
	// было бы записать в чужую категорию.
	category, err := s.categoryRepo.GetByID(ctx, userID, req.CategoryID)
	if err != nil {
		return nil, mapCategoryNotFound(err)
	}

	spentAt, err := s.resolveSpentAt(req.SpentAt)
	if err != nil {
		return nil, err
	}

	now := s.cfg.Clock.Now()
	expense := &model.Expense{
		UserID:       userID,
		CategoryID:   category.ID,
		CategoryName: category.Name,
		Amount:       helper.RoundMoney(req.Amount),
		Description:  req.Description,
		SpentAt:      spentAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return s.repo.Create(ctx, expense)
}

func (s *ExpenseService) Update(ctx context.Context, userID, id int64, req dto.UpdateExpenseRequest) (*model.Expense, error) {
	expense, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, mapExpenseNotFound(err)
	}

	if req.CategoryID != nil && *req.CategoryID != expense.CategoryID {
		category, err := s.categoryRepo.GetByID(ctx, userID, *req.CategoryID)
		if err != nil {
			return nil, mapCategoryNotFound(err)
		}
		expense.CategoryID = category.ID
		expense.CategoryName = category.Name
	}
	if req.Amount != nil {
		expense.Amount = helper.RoundMoney(*req.Amount)
	}
	if req.Description != nil {
		expense.Description = *req.Description
	}
	if req.SpentAt != nil {
		spentAt, err := s.resolveSpentAt(*req.SpentAt)
		if err != nil {
			return nil, err
		}
		expense.SpentAt = spentAt
	}

	expense.UpdatedAt = s.cfg.Clock.Now()

	if err := s.repo.Update(ctx, expense); err != nil {
		return nil, mapExpenseNotFound(err)
	}
	return expense, nil
}

func (s *ExpenseService) Delete(ctx context.Context, userID, id int64) error {
	return mapExpenseNotFound(s.repo.Delete(ctx, userID, id))
}

// resolveSpentAt: пустое значение означает «сегодня» в таймзоне сервиса.
func (s *ExpenseService) resolveSpentAt(raw string) (string, error) {
	if raw == "" {
		return helper.Today(s.cfg.Timezone).Format(helper.DateLayout), nil
	}
	d, err := helper.ParseDate(raw)
	if err != nil {
		return "", err
	}
	return d.Format(helper.DateLayout), nil
}

// buildFilter собирает фильтр репозитория из параметров запроса.
func (s *ExpenseService) buildFilter(userID int64, q ExpenseQuery) (model.ExpenseFilter, int, int, error) {
	dates, err := resolveRange(q, s.cfg.Timezone)
	if err != nil {
		return model.ExpenseFilter{}, 0, 0, err
	}

	filter := model.ExpenseFilter{
		UserID:      userID,
		From:        dates.From,
		To:          dates.To,
		CategoryIDs: q.CategoryIDs,
	}

	limit, offset, page, err := normalizePaging(q)
	if err != nil {
		return model.ExpenseFilter{}, 0, 0, err
	}
	filter.Limit = limit
	filter.Offset = offset

	return filter, page, limit, nil
}

func mapExpenseNotFound(err error) error {
	if errors.Is(err, apperr.ErrNotFound) {
		return apperr.New(apperr.CodeExpenseNotFound, "expense not found")
	}
	return err
}
