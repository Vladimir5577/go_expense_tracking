package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"go_expense_service/internal/model"
)

// newTestDB поднимает базу во временном файле и накатывает боевую миграцию.
// Именно боевую, а не отдельный DDL для тестов: иначе тесты проверяют схему,
// которой нет в проде.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_txlock=immediate")
	if err != nil {
		t.Fatalf("открытие тестовой БД: %v", err)
	}
	db.SetMaxOpenConns(1)

	for _, stmt := range migrationUp(t) {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("миграция: %v\nSQL: %s", err, stmt)
		}
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// migrationUp вытаскивает секцию Up из файла миграции и режет её на выражения.
func migrationUp(t *testing.T) []string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "migrations", "00001_init.sql"))
	if err != nil {
		t.Fatalf("чтение миграции: %v", err)
	}

	body := string(raw)
	body = strings.SplitN(body, "-- +goose Down", 2)[0]
	body = strings.ReplaceAll(body, "-- +goose Up", "")

	statements := make([]string, 0)
	for _, stmt := range strings.Split(body, ";") {
		if strings.TrimSpace(stripComments(stmt)) != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func stripComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			b.WriteString(line)
		}
	}
	return b.String()
}

func mustUser(t *testing.T, db *sql.DB, login string) int64 {
	t.Helper()

	now := time.Now().UTC()
	u, err := NewUserRepository(db).Create(context.Background(), &model.User{
		Login:        login,
		PasswordHash: "hash",
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		t.Fatalf("создание пользователя %s: %v", login, err)
	}
	return u.ID
}

func mustCategory(t *testing.T, db *sql.DB, userID int64, name string) int64 {
	t.Helper()

	now := time.Now().UTC()
	c, err := NewCategoryRepository(db).Create(context.Background(), &model.Category{
		UserID:    userID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("создание категории %s: %v", name, err)
	}
	return c.ID
}

func mustExpense(t *testing.T, db *sql.DB, userID, categoryID int64, amount float64, spentAt string) int64 {
	t.Helper()

	now := time.Now().UTC()
	e, err := NewExpenseRepository(db).Create(context.Background(), &model.Expense{
		UserID:     userID,
		CategoryID: categoryID,
		Amount:     amount,
		SpentAt:    spentAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		t.Fatalf("создание траты: %v", err)
	}
	return e.ID
}
