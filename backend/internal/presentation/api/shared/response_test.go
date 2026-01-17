// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		data       interface{}
	}{
		{
			name:       "Write simple string data",
			statusCode: http.StatusOK,
			data:       "test data",
		},
		{
			name:       "Write struct data",
			statusCode: http.StatusCreated,
			data: map[string]string{
				"message": "created successfully",
			},
		},
		{
			name:       "Write nil data",
			statusCode: http.StatusOK,
			data:       nil,
		},
		{
			name:       "Write error status",
			statusCode: http.StatusBadRequest,
			data:       map[string]string{"error": "bad request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			WriteJSON(w, tt.statusCode, tt.data)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, w.Code)
			}

			if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var response Response
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Meta should not be present in simple WriteJSON
			if response.Meta != nil {
				t.Error("Expected Meta to be nil")
			}
		})
	}
}

func TestWriteJSONWithMeta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		data       interface{}
		meta       map[string]interface{}
	}{
		{
			name:       "Write with metadata",
			statusCode: http.StatusOK,
			data:       []string{"item1", "item2"},
			meta: map[string]interface{}{
				"count": 2,
				"page":  1,
			},
		},
		{
			name:       "Write with empty meta",
			statusCode: http.StatusOK,
			data:       "test",
			meta:       map[string]interface{}{},
		},
		{
			name:       "Write with nil meta",
			statusCode: http.StatusOK,
			data:       "test",
			meta:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			WriteJSONWithMeta(w, tt.statusCode, tt.data, tt.meta)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, w.Code)
			}

			if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", contentType)
			}

			var response Response
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Check meta is present when provided
			if tt.meta != nil && len(tt.meta) > 0 {
				if response.Meta == nil {
					t.Error("Expected Meta to be present")
				}
			}
		})
	}
}

func TestWritePaginatedJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		data               interface{}
		page               int
		limit              int
		total              int
		expectedTotalPages int
	}{
		{
			name:               "Standard pagination",
			data:               []string{"item1", "item2", "item3"},
			page:               1,
			limit:              10,
			total:              25,
			expectedTotalPages: 3,
		},
		{
			name:               "Exact division",
			data:               []string{"item1"},
			page:               2,
			limit:              5,
			total:              10,
			expectedTotalPages: 2,
		},
		{
			name:               "Zero total",
			data:               []string{},
			page:               1,
			limit:              10,
			total:              0,
			expectedTotalPages: 1, // Minimum 1 page
		},
		{
			name:               "Single item",
			data:               []string{"item1"},
			page:               1,
			limit:              10,
			total:              1,
			expectedTotalPages: 1,
		},
		{
			name:               "Large dataset",
			data:               []string{"item1"},
			page:               5,
			limit:              50,
			total:              500,
			expectedTotalPages: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			WritePaginatedJSON(w, tt.data, tt.page, tt.limit, tt.total)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
			}

			var response Response
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Meta == nil {
				t.Fatal("Expected Meta to be present in paginated response")
			}

			// Check pagination metadata
			if page, ok := response.Meta["page"].(float64); !ok || int(page) != tt.page {
				t.Errorf("Expected page %d, got %v", tt.page, response.Meta["page"])
			}

			if limit, ok := response.Meta["limit"].(float64); !ok || int(limit) != tt.limit {
				t.Errorf("Expected limit %d, got %v", tt.limit, response.Meta["limit"])
			}

			if total, ok := response.Meta["total"].(float64); !ok || int(total) != tt.total {
				t.Errorf("Expected total %d, got %v", tt.total, response.Meta["total"])
			}

			if totalPages, ok := response.Meta["totalPages"].(float64); !ok || int(totalPages) != tt.expectedTotalPages {
				t.Errorf("Expected totalPages %d, got %v", tt.expectedTotalPages, response.Meta["totalPages"])
			}
		})
	}
}
