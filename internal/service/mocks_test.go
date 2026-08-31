package service

import (
	"context"
	"time"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/model"
)

// --- пользователи ---

type fakeUserRepo struct {
	users map[int64]*model.User

	setFailureAttempts int
	setFailureLocked   *time.Time
	resetCalled        bool
}

func newFakeUserRepo(users ...*model.User) *fakeUserRepo {
	r := &fakeUserRepo{users: map[int64]*model.User{}}
	for _, u := range users {
		r.users[u.ID] = u
	}
	return r
}

func (r *fakeUserRepo) GetByLogin(_ context.Context, login string) (*model.User, error) {
	for _, u := range r.users {
		if u.Login == login {
			copy := *u
			return &copy, nil
		}
	}
	return nil, apperr.ErrNotFound
}

func (r *fakeUserRepo) GetByID(_ context.Context, id int64) (*model.User, error) {
	if u, ok := r.users[id]; ok {
		copy := *u
		return &copy, nil
	}
	return nil, apperr.ErrNotFound
}

func (r *fakeUserRepo) Create(_ context.Context, u *model.User) (*model.User, error) {
	u.ID = int64(len(r.users) + 1)
	r.users[u.ID] = u
	return u, nil
}

func (r *fakeUserRepo) UpdatePassword(_ context.Context, id int64, hash string) error {
	if u, ok := r.users[id]; ok {
		u.PasswordHash = hash
	}
	return nil
}

func (r *fakeUserRepo) List(context.Context) ([]model.User, error) {
	out := make([]model.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, *u)
	}
	return out, nil
}

func (r *fakeUserRepo) SetLoginFailure(_ context.Context, id int64, attempts int, lockedUntil *time.Time) error {
	r.setFailureAttempts = attempts
	r.setFailureLocked = lockedUntil
	if u, ok := r.users[id]; ok {
		u.FailedAttempts = attempts
		u.LockedUntil = lockedUntil
	}
	return nil
}

func (r *fakeUserRepo) ResetLoginFailures(_ context.Context, id int64) error {
	r.resetCalled = true
	if u, ok := r.users[id]; ok {
		u.FailedAttempts = 0
		u.LockedUntil = nil
	}
	return nil
}

// --- категории ---

type fakeCategoryRepo struct {
	categories  map[int64]*model.Category
	hasExpenses map[int64]bool
	deleted     []int64
}

func newFakeCategoryRepo(categories ...*model.Category) *fakeCategoryRepo {
	r := &fakeCategoryRepo{
		categories:  map[int64]*model.Category{},
		hasExpenses: map[int64]bool{},
	}
	for _, c := range categories {
		r.categories[c.ID] = c
	}
	return r
}

func (r *fakeCategoryRepo) List(_ context.Context, userID int64) ([]model.Category, error) {
	out := make([]model.Category, 0)
	for _, c := range r.categories {
		if c.UserID == userID {
			out = append(out, *c)
		}
	}
	return out, nil
}

// GetByID повторяет боевое поведение: чужая категория не «запрещена», а не найдена.
func (r *fakeCategoryRepo) GetByID(_ context.Context, userID, id int64) (*model.Category, error) {
	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return nil, apperr.ErrNotFound
	}
	copy := *c
	return &copy, nil
}

func (r *fakeCategoryRepo) Create(_ context.Context, c *model.Category) (*model.Category, error) {
	c.ID = int64(len(r.categories) + 1)
	r.categories[c.ID] = c
	return c, nil
}

func (r *fakeCategoryRepo) Update(_ context.Context, c *model.Category) error {
	existing, ok := r.categories[c.ID]
	if !ok || existing.UserID != c.UserID {
		return apperr.ErrNotFound
	}
	r.categories[c.ID] = c
	return nil
}

func (r *fakeCategoryRepo) Delete(_ context.Context, userID, id int64) error {
	c, ok := r.categories[id]
	if !ok || c.UserID != userID {
		return apperr.ErrNotFound
	}
	delete(r.categories, id)
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *fakeCategoryRepo) ExistsByName(_ context.Context, userID int64, name string, excludeID int64) (bool, error) {
	for _, c := range r.categories {
		if c.UserID == userID && c.Name == name && c.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeCategoryRepo) HasExpenses(_ context.Context, _, categoryID int64) (bool, error) {
	return r.hasExpenses[categoryID], nil
}

// --- траты ---

type fakeExpenseRepo struct {
	expenses   map[int64]*model.Expense
	nextID     int64
	lastFilter model.ExpenseFilter
}

func newFakeExpenseRepo(expenses ...*model.Expense) *fakeExpenseRepo {
	r := &fakeExpenseRepo{expenses: map[int64]*model.Expense{}}
	for _, e := range expenses {
		r.expenses[e.ID] = e
		r.nextID = e.ID
	}
	return r
}

func (r *fakeExpenseRepo) List(_ context.Context, f model.ExpenseFilter) ([]model.Expense, error) {
	r.lastFilter = f
	out := make([]model.Expense, 0)
	for _, e := range r.expenses {
		if e.UserID == f.UserID {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (r *fakeExpenseRepo) Count(_ context.Context, f model.ExpenseFilter) (int64, error) {
	r.lastFilter = f
	var n int64
	for _, e := range r.expenses {
		if e.UserID == f.UserID {
			n++
		}
	}
	return n, nil
}

func (r *fakeExpenseRepo) GetByID(_ context.Context, userID, id int64) (*model.Expense, error) {
	e, ok := r.expenses[id]
	if !ok || e.UserID != userID {
		return nil, apperr.ErrNotFound
	}
	copy := *e
	return &copy, nil
}

func (r *fakeExpenseRepo) Create(_ context.Context, e *model.Expense) (*model.Expense, error) {
	r.nextID++
	e.ID = r.nextID
	r.expenses[e.ID] = e
	return e, nil
}

func (r *fakeExpenseRepo) Update(_ context.Context, e *model.Expense) error {
	existing, ok := r.expenses[e.ID]
	if !ok || existing.UserID != e.UserID {
		return apperr.ErrNotFound
	}
	r.expenses[e.ID] = e
	return nil
}

func (r *fakeExpenseRepo) Delete(_ context.Context, userID, id int64) error {
	e, ok := r.expenses[id]
	if !ok || e.UserID != userID {
		return apperr.ErrNotFound
	}
	delete(r.expenses, id)
	return nil
}

func (r *fakeExpenseRepo) Total(_ context.Context, f model.ExpenseFilter) (float64, error) {
	r.lastFilter = f
	var total float64
	for _, e := range r.expenses {
		if e.UserID == f.UserID {
			total += e.Amount
		}
	}
	return total, nil
}

func (r *fakeExpenseRepo) SummaryByCategory(_ context.Context, f model.ExpenseFilter) ([]model.CategoryTotal, error) {
	r.lastFilter = f
	byCategory := map[int64]*model.CategoryTotal{}
	for _, e := range r.expenses {
		if e.UserID != f.UserID {
			continue
		}
		t, ok := byCategory[e.CategoryID]
		if !ok {
			t = &model.CategoryTotal{CategoryID: e.CategoryID, Name: e.CategoryName}
			byCategory[e.CategoryID] = t
		}
		t.Total += e.Amount
		t.Count++
	}

	out := make([]model.CategoryTotal, 0, len(byCategory))
	for _, t := range byCategory {
		out = append(out, *t)
	}
	return out, nil
}
