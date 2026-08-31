package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/model"
)

func TestCategoryNameUniquePerUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewCategoryRepository(db)

	alice := mustUser(t, db, "alice")
	bob := mustUser(t, db, "bob")
	now := time.Now().UTC()

	mustCategory(t, db, alice, "Продукты")

	t.Run("повтор у того же пользователя запрещён", func(t *testing.T) {
		_, err := repo.Create(ctx, &model.Category{
			UserID: alice, Name: "Продукты", CreatedAt: now, UpdatedAt: now,
		})
		if err == nil {
			t.Fatal("ожидалось нарушение UNIQUE")
		}
		if !IsUniqueViolation(err) {
			t.Fatalf("IsUniqueViolation не распознал ошибку: %v", err)
		}
	})

	t.Run("то же имя у другого пользователя разрешено", func(t *testing.T) {
		if _, err := repo.Create(ctx, &model.Category{
			UserID: bob, Name: "Продукты", CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			t.Fatalf("Bob должен был завести свою категорию с тем же именем: %v", err)
		}
	})
}

func TestCategoryUserIsolation(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewCategoryRepository(db)

	alice := mustUser(t, db, "alice")
	bob := mustUser(t, db, "bob")
	catID := mustCategory(t, db, alice, "Продукты")

	if _, err := repo.GetByID(ctx, bob, catID); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("чужая категория должна быть not_found, получено: %v", err)
	}
	if err := repo.Delete(ctx, bob, catID); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("удаление чужой категории должно быть not_found, получено: %v", err)
	}

	list, err := repo.List(ctx, bob)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("Bob видит %d чужих категорий", len(list))
	}
}

// Категорию с тратами удалить нельзя. Сервис проверяет это заранее и отдаёт 409,
// а на уровне схемы страхует ON DELETE RESTRICT — проверяем именно его.
func TestCategoryDeleteRestrictedByForeignKey(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewCategoryRepository(db)

	user := mustUser(t, db, "user")
	catID := mustCategory(t, db, user, "Продукты")
	mustExpense(t, db, user, catID, 100, "2026-08-30")

	hasExpenses, err := repo.HasExpenses(ctx, user, catID)
	if err != nil {
		t.Fatalf("HasExpenses: %v", err)
	}
	if !hasExpenses {
		t.Fatal("HasExpenses должен был вернуть true")
	}

	if err := repo.Delete(ctx, user, catID); err == nil {
		t.Fatal("удаление категории с тратами должно было упасть на FK")
	}

	// Пустую категорию удалить можно.
	emptyID := mustCategory(t, db, user, "Развлечения")
	if err := repo.Delete(ctx, user, emptyID); err != nil {
		t.Fatalf("удаление пустой категории: %v", err)
	}
}

func TestCategoryExistsByName(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	repo := NewCategoryRepository(db)

	user := mustUser(t, db, "user")
	catID := mustCategory(t, db, user, "Продукты")

	// Сама редактируемая категория не считается конфликтом.
	taken, err := repo.ExistsByName(ctx, user, "Продукты", catID)
	if err != nil {
		t.Fatalf("ExistsByName: %v", err)
	}
	if taken {
		t.Fatal("категория не должна конфликтовать сама с собой")
	}

	taken, err = repo.ExistsByName(ctx, user, "Продукты", 0)
	if err != nil {
		t.Fatalf("ExistsByName: %v", err)
	}
	if !taken {
		t.Fatal("имя занято, ожидалось true")
	}
}
