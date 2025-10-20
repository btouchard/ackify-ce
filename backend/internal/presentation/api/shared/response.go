// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"encoding/json"
	"net/http"
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

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteJSONWithMeta writes a JSON response with metadata
func WriteJSONWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
		Meta: meta,
	}

	json.NewEncoder(w).Encode(response)
}

// WritePaginatedJSON writes a paginated JSON response
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

// WriteNoContent writes a 204 No Content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
