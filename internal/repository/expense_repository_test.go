package repository

import (
	"context"
	"errors"
	"testing"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/model"
)

// Изоляция пользователей — единственное правило доступа в сервисе,
// поэтому проверяется на каждой операции репозитория.
func TestExpenseUserIsolation(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewExpenseRepository(db)

	alice := mustUser(t, db, "alice")
	bob := mustUser(t, db, "bob")
	aliceCat := mustCategory(t, db, alice, "Продукты")
	expenseID := mustExpense(t, db, alice, aliceCat, 100.50, "2026-08-30")

	t.Run("чужую трату не прочитать", func(t *testing.T) {
		_, err := repo.GetByID(ctx, bob, expenseID)
		if !errors.Is(err, apperr.ErrNotFound) {
			t.Fatalf("ожидался not_found, получено: %v", err)
		}
	})

	t.Run("чужую трату не удалить", func(t *testing.T) {
		err := repo.Delete(ctx, bob, expenseID)
		if !errors.Is(err, apperr.ErrNotFound) {
			t.Fatalf("ожидался not_found, получено: %v", err)
		}
		if _, err := repo.GetByID(ctx, alice, expenseID); err != nil {
			t.Fatalf("трата Alice должна была остаться: %v", err)
		}
	})

	t.Run("чужую трату не изменить", func(t *testing.T) {
		expense, err := repo.GetByID(ctx, alice, expenseID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		expense.UserID = bob
		expense.Amount = 1
		if err := repo.Update(ctx, expense); !errors.Is(err, apperr.ErrNotFound) {
			t.Fatalf("ожидался not_found, получено: %v", err)
		}
	})

	t.Run("в списке чужих трат нет", func(t *testing.T) {
		items, err := repo.List(ctx, model.ExpenseFilter{UserID: bob, Limit: 10})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("Bob видит %d чужих трат", len(items))
		}
	})

	t.Run("в сумме чужие траты не учитываются", func(t *testing.T) {
		total, err := repo.Total(ctx, model.ExpenseFilter{UserID: bob})
		if err != nil {
			t.Fatalf("Total: %v", err)
		}
		if total != 0 {
			t.Fatalf("сумма Bob = %v, ожидался 0", total)
		}
	})
}

func TestExpenseFilterByCategories(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewExpenseRepository(db)

	user := mustUser(t, db, "user")
	food := mustCategory(t, db, user, "Продукты")
	transport := mustCategory(t, db, user, "Транспорт")
	fun := mustCategory(t, db, user, "Развлечения")

	mustExpense(t, db, user, food, 100, "2026-08-30")
	mustExpense(t, db, user, transport, 200, "2026-08-30")
	mustExpense(t, db, user, fun, 300, "2026-08-30")

	// Несколько категорий — это объединение, а не пересечение.
	items, err := repo.List(ctx, model.ExpenseFilter{
		UserID:      user,
		CategoryIDs: []int64{food, transport},
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("получено %d трат, ожидалось 2", len(items))
	}

	total, err := repo.Total(ctx, model.ExpenseFilter{
		UserID:      user,
		CategoryIDs: []int64{food, transport},
	})
	if err != nil {
		t.Fatalf("Total: %v", err)
	}
	if total != 300 {
		t.Fatalf("сумма = %v, ожидалось 300", total)
	}
}

func TestExpenseFilterByDateRange(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewExpenseRepository(db)

	user := mustUser(t, db, "user")
	cat := mustCategory(t, db, user, "Продукты")

	mustExpense(t, db, user, cat, 10, "2026-08-23")
	mustExpense(t, db, user, cat, 20, "2026-08-24")
	mustExpense(t, db, user, cat, 30, "2026-08-30")
	mustExpense(t, db, user, cat, 40, "2026-08-31")

	// Границы включительные.
	total, err := repo.Total(ctx, model.ExpenseFilter{
		UserID: user,
		From:   "2026-08-24",
		To:     "2026-08-30",
	})
	if err != nil {
		t.Fatalf("Total: %v", err)
	}
	if total != 50 {
		t.Fatalf("сумма за неделю = %v, ожидалось 50", total)
	}

	count, err := repo.Count(ctx, model.ExpenseFilter{UserID: user, From: "2026-08-24", To: "2026-08-30"})
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, ожидалось 2", count)
	}
}

func TestSummaryByCategory(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewExpenseRepository(db)

	user := mustUser(t, db, "user")
	food := mustCategory(t, db, user, "Продукты")
	transport := mustCategory(t, db, user, "Транспорт")

	mustExpense(t, db, user, food, 100.10, "2026-08-30")
	mustExpense(t, db, user, food, 200.20, "2026-08-30")
	mustExpense(t, db, user, transport, 50.05, "2026-08-30")

	summary, err := repo.SummaryByCategory(ctx, model.ExpenseFilter{UserID: user})
	if err != nil {
		t.Fatalf("SummaryByCategory: %v", err)
	}
	if len(summary) != 2 {
		t.Fatalf("получено %d категорий, ожидалось 2", len(summary))
	}

	// Отсортировано по сумме убыванию.
	if summary[0].CategoryID != food {
		t.Fatalf("первой должна идти категория с большей суммой")
	}
	if summary[0].Total != 300.30 {
		t.Fatalf("сумма по продуктам = %v, ожидалось 300.30", summary[0].Total)
	}
	if summary[0].Count != 2 {
		t.Fatalf("count по продуктам = %d, ожидалось 2", summary[0].Count)
	}

	total, err := repo.Total(ctx, model.ExpenseFilter{UserID: user})
	if err != nil {
		t.Fatalf("Total: %v", err)
	}
	if total != 350.35 {
		t.Fatalf("общая сумма = %v, ожидалось 350.35", total)
	}
}

// Сумма с тремя знаками должна отсекаться на уровне схемы —
// сервис округляет её на входе, но CHECK остаётся страховкой.
func TestAmountCheckConstraint(t *testing.T) {
	db := newTestDB(t)
	user := mustUser(t, db, "user")
	cat := mustCategory(t, db, user, "Продукты")

	for _, amount := range []float64{123.456, 0, -10} {
		_, err := db.Exec(
			`INSERT INTO expenses (user_id, category_id, amount, spent_at, created_at, updated_at)
			 VALUES (?, ?, ?, '2026-08-30', '2026-08-30T00:00:00Z', '2026-08-30T00:00:00Z')`,
			user, cat, amount)
		if err == nil {
			t.Errorf("сумма %v должна была быть отклонена CHECK-ограничением", amount)
		}
	}
}

// Формат даты тоже страхуется схемой.
func TestSpentAtCheckConstraint(t *testing.T) {
	db := newTestDB(t)
	user := mustUser(t, db, "user")
	cat := mustCategory(t, db, user, "Продукты")

	for _, date := range []string{"30-08-2026", "2026-8-30", "сегодня", ""} {
		_, err := db.Exec(
			`INSERT INTO expenses (user_id, category_id, amount, spent_at, created_at, updated_at)
			 VALUES (?, ?, 100, ?, '2026-08-30T00:00:00Z', '2026-08-30T00:00:00Z')`,
			user, cat, date)
		if err == nil {
			t.Errorf("дата %q должна была быть отклонена CHECK-ограничением", date)
		}
	}
}
