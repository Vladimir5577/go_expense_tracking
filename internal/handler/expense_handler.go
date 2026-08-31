package handler

import (
	"net/http"

	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/middleware"
	"go_expense_service/internal/service"
)

type ExpenseHandler struct {
	service service.ExpenseServiceInterface
}

func NewExpenseHandler(s service.ExpenseServiceInterface) *ExpenseHandler {
	return &ExpenseHandler{service: s}
}

func (h *ExpenseHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.UserID(r.Context())
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		query, err := parseExpenseQuery(r)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		list, err := h.service.List(r.Context(), userID, query)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, &dto.ExpenseListResponse{
			Items:      dto.MapExpensesResponse(list.Items),
			Page:       list.Page,
			Limit:      list.Limit,
			TotalItems: list.TotalItems,
		})
	}
}

func (h *ExpenseHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, id, err := userAndID(r)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		expense, err := h.service.Get(r.Context(), userID, id)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapExpenseResponse(expense))
	}
}

func (h *ExpenseHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.UserID(r.Context())
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		var req dto.CreateExpenseRequest
		if err := decodeAndValidate(r, &req); err != nil {
			helper.WriteError(w, err)
			return
		}

		expense, err := h.service.Create(r.Context(), userID, req)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusCreated, dto.MapExpenseResponse(expense))
	}
}

func (h *ExpenseHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, id, err := userAndID(r)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		var req dto.UpdateExpenseRequest
		if err := decodeAndValidate(r, &req); err != nil {
			helper.WriteError(w, err)
			return
		}

		expense, err := h.service.Update(r.Context(), userID, id, req)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapExpenseResponse(expense))
	}
}

func (h *ExpenseHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, id, err := userAndID(r)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		if err := h.service.Delete(r.Context(), userID, id); err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusNoContent, nil)
	}
}
