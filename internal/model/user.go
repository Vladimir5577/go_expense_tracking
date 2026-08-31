package model

import "time"

// User — пользователь сервиса. Заводится только консольной утилитой useradm,
// регистрации через API нет.
type User struct {
	ID             int64
	Login          string
	PasswordHash   string
	Name           string
	FailedAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsLocked сообщает, заблокирован ли вход на момент now.
func (u *User) IsLocked(now time.Time) bool {
	return u.LockedUntil != nil && u.LockedUntil.After(now)
}
