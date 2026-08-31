package service

import (
	"context"

	"go_expense_service/internal/config"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/model"
	"go_expense_service/internal/repository"
)

type ReportServiceInterface interface {
	Summary(ctx context.Context, userID int64, q ExpenseQuery) (*model.Summary, error)
}

type ReportService struct {
	repo repository.ExpenseRepositoryInterface
	cfg  *config.Config
}

func NewReportService(repo repository.ExpenseRepositoryInterface, cfg *config.Config) *ReportService {
	return &ReportService{repo: repo, cfg: cfg}
}

// Summary отдаёт сумму за период и разбивку по категориям.
// Обе величины считает SQL: складывать траты циклом в Go нельзя.
func (s *ReportService) Summary(ctx context.Context, userID int64, q ExpenseQuery) (*model.Summary, error) {
	dates, err := resolveRange(q, s.cfg.Timezone)
	if err != nil {
		return nil, err
	}

	filter := model.ExpenseFilter{
		UserID:      userID,
		From:        dates.From,
		To:          dates.To,
		CategoryIDs: q.CategoryIDs,
	}

	total, err := s.repo.Total(ctx, filter)
	if err != nil {
		return nil, err
	}

	byCategory, err := s.repo.SummaryByCategory(ctx, filter)
	if err != nil {
		return nil, err
	}

	if total > 0 {
		for i := range byCategory {
			byCategory[i].Share = helper.RoundMoney(byCategory[i].Total / total * 100)
		}
	}

	return &model.Summary{
		From:       dates.From,
		To:         dates.To,
		Total:      total,
		ByCategory: byCategory,
	}, nil
}
