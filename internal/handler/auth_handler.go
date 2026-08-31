package handler

import (
	"net/http"

	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/middleware"
	"go_expense_service/internal/service"
)

type AuthHandler struct {
	service service.AuthServiceInterface
}

func NewAuthHandler(s service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := decodeAndValidate(r, &req); err != nil {
			helper.WriteError(w, err)
			return
		}

		token, expiresAt, user, err := h.service.Login(r.Context(), req)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapLoginResponse(token, expiresAt, user))
	}
}

func (h *AuthHandler) Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.UserID(r.Context())
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		user, err := h.service.Me(r.Context(), userID)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapUserResponse(user))
	}
}
