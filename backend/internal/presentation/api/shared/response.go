// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Response represents a standardized API response
type Response struct {
	Data interface{}            `json:"data,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page     int `json:"page" schema:"page"`
	PageSize int `json:"page_size" schema:"page_size"`
	Offset   int `json:"-"`
}

func NewPaginationParams(defaultPage, defaultPageSize, maxPageSize int) *PaginationParams {
	if defaultPage < 1 {
		defaultPage = 1
	}
	if defaultPageSize < 1 {
		defaultPageSize = 20
	}
	if maxPageSize < 1 {
		maxPageSize = 100
	}

	return &PaginationParams{
		Page:     defaultPage,
		PageSize: defaultPageSize,
	}
}

func ParsePaginationParams(r *http.Request, defaultPageSize, maxPageSize int) *PaginationParams {
	params := NewPaginationParams(1, defaultPageSize, maxPageSize)

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	pageSizeStr := r.URL.Query().Get("limit")
	if pageSizeStr == "" {
		pageSizeStr = r.URL.Query().Get("page_size")
	}
	if pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}

	params.Validate(maxPageSize)
	return params
}

// Validate validates pagination parameters and calculates offset
func (p *PaginationParams) Validate(maxPageSize int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if maxPageSize > 0 && p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
	p.Offset = (p.Page - 1) * p.PageSize
}

func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
	}

	json.NewEncoder(w).Encode(response)
}

func WriteJSONWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
		Meta: meta,
	}

	json.NewEncoder(w).Encode(response)
}

func WritePaginatedJSON(w http.ResponseWriter, data interface{}, page, limit, total int) {
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	meta := map[string]interface{}{
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": totalPages,
	}

	WriteJSONWithMeta(w, http.StatusOK, data, meta)
}
