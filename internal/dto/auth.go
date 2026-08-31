package dto

import (
	"time"

	"go_expense_service/internal/model"
)

type LoginRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=4"`
}

type UserResponse struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresAt string        `json:"expiresAt"`
	User      *UserResponse `json:"user"`
}

func MapUserResponse(u *model.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:    u.ID,
		Login: u.Login,
		Name:  u.Name,
	}
}

func MapLoginResponse(token string, expiresAt time.Time, u *model.User) *LoginResponse {
	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
		User:      MapUserResponse(u),
	}
}
