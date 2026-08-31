package repository

import (
	"context"
	"database/sql"
	"time"

	"go_expense_service/internal/model"
)

type UserRepositoryInterface interface {
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, u *model.User) (*model.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	List(ctx context.Context) ([]model.User, error)
	SetLoginFailure(ctx context.Context, id int64, attempts int, lockedUntil *time.Time) error
	ResetLoginFailures(ctx context.Context, id int64) error
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `id, login, password_hash, COALESCE(name, ''), failed_attempts, locked_until, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	var (
		u           model.User
		lockedUntil sql.NullString
		createdAt   string
		updatedAt   string
	)
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Name, &u.FailedAttempts, &lockedUntil, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	u.LockedUntil = parseTimePtr(lockedUntil)
	u.CreatedAt = parseTime(createdAt)
	u.UpdatedAt = parseTime(updatedAt)
	return &u, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE login = ?`, login)
	u, err := scanUser(row)
	if err != nil {
		return nil, NormalizeError(err)
	}
	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = ?`, id)
	u, err := scanUser(row)
	if err != nil {
		return nil, NormalizeError(err)
	}
	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (login, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		u.Login, u.PasswordHash, u.Name, u.CreatedAt.Format(time.RFC3339), u.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.ID = id
	return u, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	// Смена пароля заодно снимает блокировку входа: администратор, меняющий
	// пароль, обычно как раз разбирается с заблокированным пользователем.
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, failed_attempts = 0, locked_until = NULL, updated_at = ? WHERE id = ?`,
		passwordHash, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+userColumns+` FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *UserRepository) SetLoginFailure(ctx context.Context, id int64, attempts int, lockedUntil *time.Time) error {
	var locked any
	if lockedUntil != nil {
		locked = lockedUntil.UTC().Format(time.RFC3339)
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET failed_attempts = ?, locked_until = ? WHERE id = ?`,
		attempts, locked, id)
	return err
}

func (r *UserRepository) ResetLoginFailures(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE id = ?`, id)
	return err
}
