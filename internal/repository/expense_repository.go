package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"go_expense_service/internal/model"
)

type ExpenseRepositoryInterface interface {
	List(ctx context.Context, f model.ExpenseFilter) ([]model.Expense, error)
	Count(ctx context.Context, f model.ExpenseFilter) (int64, error)
	GetByID(ctx context.Context, userID, id int64) (*model.Expense, error)
	Create(ctx context.Context, e *model.Expense) (*model.Expense, error)
	Update(ctx context.Context, e *model.Expense) error
	Delete(ctx context.Context, userID, id int64) error
	Total(ctx context.Context, f model.ExpenseFilter) (float64, error)
	SummaryByCategory(ctx context.Context, f model.ExpenseFilter) ([]model.CategoryTotal, error)
}

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

const expenseColumns = `e.id, e.user_id, e.category_id, c.name, e.amount,
	COALESCE(e.description, ''), e.spent_at, e.created_at, e.updated_at`

func scanExpense(row interface{ Scan(...any) error }) (*model.Expense, error) {
	var (
		e         model.Expense
		createdAt string
		updatedAt string
	)
	err := row.Scan(&e.ID, &e.UserID, &e.CategoryID, &e.CategoryName, &e.Amount,
		&e.Description, &e.SpentAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	e.CreatedAt = parseTime(createdAt)
	e.UpdatedAt = parseTime(updatedAt)
	return &e, nil
}

// buildWhere собирает условие выборки. user_id присутствует всегда — это
// единственное правило доступа в сервисе, и оно не должно зависеть от того,
// передал ли вызывающий остальные фильтры.
func buildWhere(f model.ExpenseFilter) (string, []any) {
	var sb strings.Builder
	args := []any{f.UserID}
	sb.WriteString(" WHERE e.user_id = ?")

	if f.From != "" {
		sb.WriteString(" AND e.spent_at >= ?")
		args = append(args, f.From)
	}
	if f.To != "" {
		sb.WriteString(" AND e.spent_at <= ?")
		args = append(args, f.To)
	}
	if len(f.CategoryIDs) > 0 {
		sb.WriteString(" AND e.category_id IN (")
		sb.WriteString(strings.TrimSuffix(strings.Repeat("?,", len(f.CategoryIDs)), ","))
		sb.WriteString(")")
		for _, id := range f.CategoryIDs {
			args = append(args, id)
		}
	}
	return sb.String(), args
}

func (r *ExpenseRepository) List(ctx context.Context, f model.ExpenseFilter) ([]model.Expense, error) {
	where, args := buildWhere(f)
	query := `SELECT ` + expenseColumns + `
		FROM expenses e
		JOIN categories c ON c.id = e.category_id` + where + `
		ORDER BY e.spent_at DESC, e.id DESC
		LIMIT ? OFFSET ?`
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := make([]model.Expense, 0)
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, *e)
	}
	return expenses, rows.Err()
}

func (r *ExpenseRepository) Count(ctx context.Context, f model.ExpenseFilter) (int64, error) {
	where, args := buildWhere(f)
	var total int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM expenses e`+where, args...).Scan(&total)
	return total, err
}

func (r *ExpenseRepository) GetByID(ctx context.Context, userID, id int64) (*model.Expense, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+expenseColumns+`
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		WHERE e.id = ? AND e.user_id = ?`, id, userID)
	e, err := scanExpense(row)
	if err != nil {
		return nil, NormalizeError(err)
	}
	return e, nil
}

func (r *ExpenseRepository) Create(ctx context.Context, e *model.Expense) (*model.Expense, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO expenses (user_id, category_id, amount, description, spent_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.UserID, e.CategoryID, e.Amount, e.Description, e.SpentAt,
		e.CreatedAt.Format(time.RFC3339), e.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	e.ID = id
	return e, nil
}

func (r *ExpenseRepository) Update(ctx context.Context, e *model.Expense) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE expenses SET category_id = ?, amount = ?, description = ?, spent_at = ?, updated_at = ?
		 WHERE id = ? AND user_id = ?`,
		e.CategoryID, e.Amount, e.Description, e.SpentAt,
		e.UpdatedAt.Format(time.RFC3339), e.ID, e.UserID)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func (r *ExpenseRepository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM expenses WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// Total считает сумму средствами SQL. Складывать траты циклом в Go нельзя:
// SQLite с версии 3.44 суммирует REAL компенсированным алгоритмом
// (Kahan-Babuska) и даёт точный результат, а наивный цикл накапливает погрешность.
func (r *ExpenseRepository) Total(ctx context.Context, f model.ExpenseFilter) (float64, error) {
	where, args := buildWhere(f)
	var total float64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(ROUND(SUM(e.amount), 2), 0) FROM expenses e`+where, args...).Scan(&total)
	return total, err
}

func (r *ExpenseRepository) SummaryByCategory(ctx context.Context, f model.ExpenseFilter) ([]model.CategoryTotal, error) {
	where, args := buildWhere(f)
	query := `SELECT e.category_id, c.name, ROUND(SUM(e.amount), 2), COUNT(*)
		FROM expenses e
		JOIN categories c ON c.id = e.category_id` + where + `
		GROUP BY e.category_id, c.name
		ORDER BY SUM(e.amount) DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totals := make([]model.CategoryTotal, 0)
	for rows.Next() {
		var t model.CategoryTotal
		if err := rows.Scan(&t.CategoryID, &t.Name, &t.Total, &t.Count); err != nil {
			return nil, err
		}
		totals = append(totals, t)
	}
	return totals, rows.Err()
}
