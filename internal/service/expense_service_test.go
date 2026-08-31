package service

import (
	"context"
	"testing"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/model"
)

func newExpenseService(categories *fakeCategoryRepo, expenses *fakeExpenseRepo) *ExpenseService {
	return NewExpenseService(expenses, categories, testConfig())
}

// Трату нельзя записать в чужую категорию — иначе через categoryId можно
// подмешивать свои траты в чужую статистику.
func TestExpenseCreateWithForeignCategory(t *testing.T) {
	categories := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := newExpenseService(categories, newFakeExpenseRepo())

	_, err := svc.Create(context.Background(), bob, dto.CreateExpenseRequest{
		CategoryID: 10,
		Amount:     100,
	})
	requireCode(t, err, apperr.CodeCategoryNotFound)
}

func TestExpenseCreateRoundsAmount(t *testing.T) {
	categories := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := newExpenseService(categories, newFakeExpenseRepo())

	tests := []struct {
		in   float64
		want float64
	}{
		{123.456, 123.46},
		{123.454, 123.45},
		{100, 100},
		{0.005, 0.01},
	}

	for _, tc := range tests {
		created, err := svc.Create(context.Background(), alice, dto.CreateExpenseRequest{
			CategoryID: 10,
			Amount:     tc.in,
		})
		if err != nil {
			t.Fatalf("Create(%v): %v", tc.in, err)
		}
		if created.Amount != tc.want {
			t.Errorf("сумма %v сохранена как %v, ожидалось %v", tc.in, created.Amount, tc.want)
		}
	}
}

// Без spentAt трата попадает на сегодня — в таймзоне сервиса, а не в UTC.
func TestExpenseCreateDefaultsToToday(t *testing.T) {
	categories := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := newExpenseService(categories, newFakeExpenseRepo())

	created, err := svc.Create(context.Background(), alice, dto.CreateExpenseRequest{
		CategoryID: 10,
		Amount:     100,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	want := helper.Today(testConfig().Timezone).Format(helper.DateLayout)
	if created.SpentAt != want {
		t.Fatalf("spentAt = %q, ожидалось %q", created.SpentAt, want)
	}
}

func TestExpenseCreateRejectsBadDate(t *testing.T) {
	categories := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := newExpenseService(categories, newFakeExpenseRepo())

	_, err := svc.Create(context.Background(), alice, dto.CreateExpenseRequest{
		CategoryID: 10,
		Amount:     100,
		SpentAt:    "30-08-2026",
	})
	requireCode(t, err, apperr.CodeValidation)
}

func TestExpenseForeignAccessIsNotFound(t *testing.T) {
	categories := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	expenses := newFakeExpenseRepo(&model.Expense{
		ID: 100, UserID: alice, CategoryID: 10, Amount: 500, SpentAt: "2026-08-30",
	})
	svc := newExpenseService(categories, expenses)
	ctx := context.Background()

	_, err := svc.Get(ctx, bob, 100)
	requireCode(t, err, apperr.CodeExpenseNotFound)

	amount := 1.0
	_, err = svc.Update(ctx, bob, 100, dto.UpdateExpenseRequest{Amount: &amount})
	requireCode(t, err, apperr.CodeExpenseNotFound)

	err = svc.Delete(ctx, bob, 100)
	requireCode(t, err, apperr.CodeExpenseNotFound)

	if expenses.expenses[100].Amount != 500 {
		t.Fatal("трата Alice не должна была измениться")
	}
}

func TestExpenseUpdatePartial(t *testing.T) {
	categories := newFakeCategoryRepo(
		&model.Category{ID: 10, UserID: alice, Name: "Продукты"},
		&model.Category{ID: 11, UserID: alice, Name: "Транспорт"},
	)
	expenses := newFakeExpenseRepo(&model.Expense{
		ID: 100, UserID: alice, CategoryID: 10, CategoryName: "Продукты",
		Amount: 500, Description: "старое", SpentAt: "2026-08-30",
	})
	svc := newExpenseService(categories, expenses)

	// Меняем только категорию — остальное должно остаться как было.
	newCategory := int64(11)
	updated, err := svc.Update(context.Background(), alice, 100, dto.UpdateExpenseRequest{
		CategoryID: &newCategory,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.CategoryID != 11 || updated.CategoryName != "Транспорт" {
		t.Fatalf("категория не обновилась: %d %q", updated.CategoryID, updated.CategoryName)
	}
	if updated.Amount != 500 || updated.Description != "старое" || updated.SpentAt != "2026-08-30" {
		t.Fatalf("непереданные поля изменились: %+v", updated)
	}
}

func TestExpenseUpdateToForeignCategory(t *testing.T) {
	categories := newFakeCategoryRepo(
		&model.Category{ID: 10, UserID: alice, Name: "Продукты"},
		&model.Category{ID: 20, UserID: bob, Name: "Чужая"},
	)
	expenses := newFakeExpenseRepo(&model.Expense{
		ID: 100, UserID: alice, CategoryID: 10, Amount: 500, SpentAt: "2026-08-30",
	})
	svc := newExpenseService(categories, expenses)

	foreign := int64(20)
	_, err := svc.Update(context.Background(), alice, 100, dto.UpdateExpenseRequest{CategoryID: &foreign})
	requireCode(t, err, apperr.CodeCategoryNotFound)
}

func TestExpenseListPaging(t *testing.T) {
	categories := newFakeCategoryRepo()
	expenses := newFakeExpenseRepo()
	svc := newExpenseService(categories, expenses)
	ctx := context.Background()

	t.Run("значения по умолчанию", func(t *testing.T) {
		list, err := svc.List(ctx, alice, ExpenseQuery{})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if list.Page != 1 || list.Limit != defaultLimit {
			t.Fatalf("page=%d limit=%d, ожидалось 1/%d", list.Page, list.Limit, defaultLimit)
		}
	})

	t.Run("offset считается от страницы", func(t *testing.T) {
		if _, err := svc.List(ctx, alice, ExpenseQuery{Page: 3, Limit: 20}); err != nil {
			t.Fatalf("List: %v", err)
		}
		if expenses.lastFilter.Offset != 40 {
			t.Fatalf("offset = %d, ожидалось 40", expenses.lastFilter.Offset)
		}
	})

	t.Run("некорректные значения отклоняются", func(t *testing.T) {
		if _, err := svc.List(ctx, alice, ExpenseQuery{Page: -1}); err == nil {
			t.Error("отрицательная страница должна отклоняться")
		}
		if _, err := svc.List(ctx, alice, ExpenseQuery{Limit: maxLimit + 1}); err == nil {
			t.Error("limit сверх потолка должен отклоняться")
		}
	})
}

// Период разворачивается в диапазон дат, который уходит в репозиторий.
func TestExpenseListResolvesPeriod(t *testing.T) {
	expenses := newFakeExpenseRepo()
	svc := newExpenseService(newFakeCategoryRepo(), expenses)

	if _, err := svc.List(context.Background(), alice, ExpenseQuery{
		Period: helper.PeriodMonth,
		Date:   "2026-08-15",
	}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if expenses.lastFilter.From != "2026-08-01" || expenses.lastFilter.To != "2026-08-31" {
		t.Fatalf("диапазон = [%s; %s], ожидалось [2026-08-01; 2026-08-31]",
			expenses.lastFilter.From, expenses.lastFilter.To)
	}
}

// Явные from/to важнее period.
func TestExplicitRangeBeatsPeriod(t *testing.T) {
	expenses := newFakeExpenseRepo()
	svc := newExpenseService(newFakeCategoryRepo(), expenses)

	if _, err := svc.List(context.Background(), alice, ExpenseQuery{
		Period: helper.PeriodMonth,
		From:   "2026-01-01",
		To:     "2026-01-31",
	}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if expenses.lastFilter.From != "2026-01-01" || expenses.lastFilter.To != "2026-01-31" {
		t.Fatalf("диапазон = [%s; %s], ожидались явные границы",
			expenses.lastFilter.From, expenses.lastFilter.To)
	}
}

func TestRangeValidation(t *testing.T) {
	svc := newExpenseService(newFakeCategoryRepo(), newFakeExpenseRepo())
	ctx := context.Background()

	if _, err := svc.List(ctx, alice, ExpenseQuery{From: "2026-08-31", To: "2026-08-01"}); err == nil {
		t.Error("from позже to должно отклоняться")
	}
	if _, err := svc.List(ctx, alice, ExpenseQuery{From: "31-08-2026"}); err == nil {
		t.Error("некорректный формат даты должен отклоняться")
	}
	if _, err := svc.List(ctx, alice, ExpenseQuery{Period: "year"}); err == nil {
		t.Error("неизвестный период должен отклоняться")
	}
}

func TestSummaryShares(t *testing.T) {
	expenses := newFakeExpenseRepo(
		&model.Expense{ID: 1, UserID: alice, CategoryID: 10, CategoryName: "Продукты", Amount: 750, SpentAt: "2026-08-30"},
		&model.Expense{ID: 2, UserID: alice, CategoryID: 11, CategoryName: "Транспорт", Amount: 250, SpentAt: "2026-08-30"},
	)
	svc := NewReportService(expenses, testConfig())

	summary, err := svc.Summary(context.Background(), alice, ExpenseQuery{})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.Total != 1000 {
		t.Fatalf("итог = %v, ожидалось 1000", summary.Total)
	}

	shares := map[int64]float64{}
	for _, c := range summary.ByCategory {
		shares[c.CategoryID] = c.Share
	}
	if shares[10] != 75 || shares[11] != 25 {
		t.Fatalf("доли = %v, ожидалось 75/25", shares)
	}
}

// На пустом периоде деления на ноль быть не должно.
func TestSummaryEmpty(t *testing.T) {
	svc := NewReportService(newFakeExpenseRepo(), testConfig())

	summary, err := svc.Summary(context.Background(), alice, ExpenseQuery{Period: helper.PeriodWeek})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.Total != 0 || len(summary.ByCategory) != 0 {
		t.Fatalf("ожидался пустой отчёт, получено %+v", summary)
	}
	if summary.From == "" || summary.To == "" {
		t.Fatal("границы периода должны заполняться даже на пустом отчёте")
	}
}
