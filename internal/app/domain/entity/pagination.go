package entity

import "math"

type PageMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NormalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func NewPageMeta(page, limit, total int) PageMeta {
	p, l := NormalizePagination(page, limit)
	if total < 0 {
		total = 0
	}
	totalPages := int(math.Ceil(float64(total) / float64(l)))
	if totalPages < 1 {
		totalPages = 1
	}
	return PageMeta{
		Page:       p,
		Limit:      l,
		Total:      total,
		TotalPages: totalPages,
	}
}
