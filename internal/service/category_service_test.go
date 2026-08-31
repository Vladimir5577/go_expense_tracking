package service

import (
	"context"
	"errors"
	"testing"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/dto"
	"go_expense_service/internal/model"
)

const (
	alice = int64(1)
	bob   = int64(2)
)

func requireCode(t *testing.T, err error, want apperr.ErrorCode) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("ожидалась доменная ошибка %s, получено: %v", want, err)
	}
	if appErr.Code != want {
		t.Fatalf("код ошибки = %s, ожидался %s", appErr.Code, want)
	}
}

func TestCategoryDeleteWithExpensesIsRejected(t *testing.T) {
	repo := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	repo.hasExpenses[10] = true
	svc := NewCategoryService(repo, testConfig())

	err := svc.Delete(context.Background(), alice, 10)
	requireCode(t, err, apperr.CodeCategoryHasExpenses)

	if len(repo.deleted) != 0 {
		t.Fatal("категория не должна была удалиться")
	}
}

func TestCategoryDeleteEmptyIsAllowed(t *testing.T) {
	repo := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := NewCategoryService(repo, testConfig())

	if err := svc.Delete(context.Background(), alice, 10); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(repo.deleted) != 1 {
		t.Fatal("категория должна была удалиться")
	}
}

// Чужая категория для пользователя не существует — 404, а не 403.
// Иначе по коду ответа выясняется, какие id заняты у других пользователей.
func TestCategoryForeignAccessIsNotFound(t *testing.T) {
	repo := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := NewCategoryService(repo, testConfig())
	ctx := context.Background()

	_, err := svc.Get(ctx, bob, 10)
	requireCode(t, err, apperr.CodeCategoryNotFound)

	name := "Взломано"
	_, err = svc.Update(ctx, bob, 10, dto.UpdateCategoryRequest{Name: &name})
	requireCode(t, err, apperr.CodeCategoryNotFound)

	err = svc.Delete(ctx, bob, 10)
	requireCode(t, err, apperr.CodeCategoryNotFound)

	if repo.categories[10].Name != "Продукты" {
		t.Fatal("категория Alice не должна была измениться")
	}
}

func TestCategoryDuplicateName(t *testing.T) {
	repo := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := NewCategoryService(repo, testConfig())

	_, err := svc.Create(context.Background(), alice, dto.CreateCategoryRequest{Name: "Продукты"})
	requireCode(t, err, apperr.CodeCategoryNameExists)
}

// Имя обрезается по краям: «  Продукты » и «Продукты» — одна и та же категория.
func TestCategoryNameIsTrimmed(t *testing.T) {
	repo := newFakeCategoryRepo()
	svc := NewCategoryService(repo, testConfig())
	ctx := context.Background()

	created, err := svc.Create(ctx, alice, dto.CreateCategoryRequest{Name: "  Продукты  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Name != "Продукты" {
		t.Fatalf("имя = %q, ожидалось %q", created.Name, "Продукты")
	}

	_, err = svc.Create(ctx, alice, dto.CreateCategoryRequest{Name: " Продукты"})
	requireCode(t, err, apperr.CodeCategoryNameExists)
}

func TestCategoryRenameToExistingName(t *testing.T) {
	repo := newFakeCategoryRepo(
		&model.Category{ID: 10, UserID: alice, Name: "Продукты"},
		&model.Category{ID: 11, UserID: alice, Name: "Транспорт"},
	)
	svc := NewCategoryService(repo, testConfig())

	name := "Продукты"
	_, err := svc.Update(context.Background(), alice, 11, dto.UpdateCategoryRequest{Name: &name})
	requireCode(t, err, apperr.CodeCategoryNameExists)
}

// Переименование в то же самое имя конфликтом не является.
func TestCategoryRenameToOwnName(t *testing.T) {
	repo := newFakeCategoryRepo(&model.Category{ID: 10, UserID: alice, Name: "Продукты"})
	svc := NewCategoryService(repo, testConfig())

	name := "Продукты"
	description := "Еда и бытовая химия"
	updated, err := svc.Update(context.Background(), alice, 10, dto.UpdateCategoryRequest{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Description != description {
		t.Fatalf("описание = %q, ожидалось %q", updated.Description, description)
	}
}
