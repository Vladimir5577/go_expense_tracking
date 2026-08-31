package repository

import (
	"context"
	"database/sql"
	"time"

	"go_expense_service/internal/model"
)

type CategoryRepositoryInterface interface {
	List(ctx context.Context, userID int64) ([]model.Category, error)
	GetByID(ctx context.Context, userID, id int64) (*model.Category, error)
	Create(ctx context.Context, c *model.Category) (*model.Category, error)
	Update(ctx context.Context, c *model.Category) error
	Delete(ctx context.Context, userID, id int64) error
	ExistsByName(ctx context.Context, userID int64, name string, excludeID int64) (bool, error)
	HasExpenses(ctx context.Context, userID, categoryID int64) (bool, error)
}

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

const categoryColumns = `id, user_id, name, COALESCE(description, ''), COALESCE(color, ''), created_at, updated_at`

func scanCategory(row interface{ Scan(...any) error }) (*model.Category, error) {
	var (
		c         model.Category
		createdAt string
		updatedAt string
	)
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Description, &c.Color, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	c.CreatedAt = parseTime(createdAt)
	c.UpdatedAt = parseTime(updatedAt)
	return &c, nil
}

func (r *CategoryRepository) List(ctx context.Context, userID int64) ([]model.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+categoryColumns+` FROM categories WHERE user_id = ? ORDER BY name COLLATE NOCASE`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]model.Category, 0)
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *c)
	}
	return categories, rows.Err()
}

// GetByID всегда фильтрует по user_id: чужая категория для пользователя
// не существует, а не «запрещена».
func (r *CategoryRepository) GetByID(ctx context.Context, userID, id int64) (*model.Category, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+categoryColumns+` FROM categories WHERE id = ? AND user_id = ?`, id, userID)
	c, err := scanCategory(row)
	if err != nil {
		return nil, NormalizeError(err)
	}
	return c, nil
}

func (r *CategoryRepository) Create(ctx context.Context, c *model.Category) (*model.Category, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO categories (user_id, name, description, color, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		c.UserID, c.Name, c.Description, c.Color,
		c.CreatedAt.Format(time.RFC3339), c.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	c.ID = id
	return c, nil
}

func (r *CategoryRepository) Update(ctx context.Context, c *model.Category) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name = ?, description = ?, color = ?, updated_at = ?
		 WHERE id = ? AND user_id = ?`,
		c.Name, c.Description, c.Color, c.UpdatedAt.Format(time.RFC3339), c.ID, c.UserID)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func (r *CategoryRepository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// ExistsByName проверяет уникальность имени в рамках пользователя.
// excludeID позволяет не считать конфликтом саму редактируемую категорию.
func (r *CategoryRepository) ExistsByName(ctx context.Context, userID int64, name string, excludeID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM categories WHERE user_id = ? AND name = ? AND id <> ?)`,
		userID, name, excludeID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *CategoryRepository) HasExpenses(ctx context.Context, userID, categoryID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM expenses WHERE user_id = ? AND category_id = ?)`,
		userID, categoryID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// requireAffected превращает «обновлено 0 строк» в ErrNotFound. Ноль строк здесь
// значит, что записи нет или она принадлежит другому пользователю — снаружи
// эти случаи неразличимы намеренно.
func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return NormalizeError(sql.ErrNoRows)
	}
	return nil
}
