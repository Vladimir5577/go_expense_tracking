package handler

import (
	"net/http"

	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/middleware"
	"go_expense_service/internal/service"
)

type CategoryHandler struct {
	service service.CategoryServiceInterface
}

func NewCategoryHandler(s service.CategoryServiceInterface) *CategoryHandler {
	return &CategoryHandler{service: s}
}

func (h *CategoryHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.UserID(r.Context())
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		categories, err := h.service.List(r.Context(), userID)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapCategoriesResponse(categories))
	}
}

func (h *CategoryHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, id, err := userAndID(r)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		category, err := h.service.Get(r.Context(), userID, id)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapCategoryResponse(category))
	}
}

func (h *CategoryHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := middleware.UserID(r.Context())
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		var req dto.CreateCategoryRequest
		if err := decodeAndValidate(r, &req); err != nil {
			helper.WriteError(w, err)
			return
		}

		category, err := h.service.Create(r.Context(), userID, req)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusCreated, dto.MapCategoryResponse(category))
	}
}

func (h *CategoryHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, id, err := userAndID(r)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		var req dto.UpdateCategoryRequest
		if err := decodeAndValidate(r, &req); err != nil {
			helper.WriteError(w, err)
			return
		}

		category, err := h.service.Update(r.Context(), userID, id, req)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapCategoryResponse(category))
	}
}

func (h *CategoryHandler) Delete() http.HandlerFunc {
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

// userAndID достаёт id текущего пользователя и {id} из пути — эта пара нужна
// почти каждому хендлеру.
func userAndID(r *http.Request) (int64, int64, error) {
	userID, err := middleware.UserID(r.Context())
	if err != nil {
		return 0, 0, err
	}
	id, err := helper.IDParam(r, "id")
	if err != nil {
		return 0, 0, err
	}
	return userID, id, nil
}
