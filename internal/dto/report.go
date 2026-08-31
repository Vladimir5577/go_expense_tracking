package dto

import "go_expense_service/internal/model"

type CategorySummaryResponse struct {
	CategoryID int64   `json:"categoryId"`
	Name       string  `json:"name"`
	Total      float64 `json:"total"`
	Count      int64   `json:"count"`
	// Share — доля категории в общей сумме, в процентах.
	Share float64 `json:"share"`
}

type SummaryResponse struct {
	From       string                     `json:"from"`
	To         string                     `json:"to"`
	Total      float64                    `json:"total"`
	ByCategory []*CategorySummaryResponse `json:"byCategory"`
}

func MapSummaryResponse(s *model.Summary) *SummaryResponse {
	if s == nil {
		return nil
	}

	byCategory := make([]*CategorySummaryResponse, 0, len(s.ByCategory))
	for i := range s.ByCategory {
		t := &s.ByCategory[i]
		byCategory = append(byCategory, &CategorySummaryResponse{
			CategoryID: t.CategoryID,
			Name:       t.Name,
			Total:      t.Total,
			Count:      t.Count,
			Share:      t.Share,
		})
	}

	return &SummaryResponse{
		From:       s.From,
		To:         s.To,
		Total:      s.Total,
		ByCategory: byCategory,
	}
}
