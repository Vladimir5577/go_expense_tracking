package handler

import (
	"net/http"

	"go_expense_service/internal/dto"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/middleware"
	"go_expense_service/internal/service"
)

type ReportHandler struct {
	service service.ReportServiceInterface
}

func NewReportHandler(s service.ReportServiceInterface) *ReportHandler {
	return &ReportHandler{service: s}
}

func (h *ReportHandler) Summary() http.HandlerFunc {
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

		summary, err := h.service.Summary(r.Context(), userID, query)
		if err != nil {
			helper.WriteError(w, err)
			return
		}

		helper.WriteJSON(w, http.StatusOK, dto.MapSummaryResponse(summary))
	}
}
