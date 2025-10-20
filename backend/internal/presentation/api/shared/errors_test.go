// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		message     string
		fieldErrors map[string]string
	}{
		{
			name:    "Validation error with field errors",
			message: "Invalid input",
			fieldErrors: map[string]string{
				"email": "Invalid email format",
				"age":   "Must be positive",
			},
		},
		{
			name:        "Validation error without field errors",
			message:     "Invalid request",
			fieldErrors: nil,
		},
		{
			name:        "Validation error with empty field errors",
			message:     "Validation failed",
			fieldErrors: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			WriteValidationError(w, tt.message, tt.fieldErrors)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
			}

			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Error.Message != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, response.Error.Message)
			}

			if response.Error.Code != ErrCodeValidation {
				t.Errorf("Expected code '%s', got '%s'", ErrCodeValidation, response.Error.Code)
			}
		})
	}
}

func TestWriteNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		resource        string
		expectedMessage string
	}{
		{
			name:            "Not found with resource name",
			resource:        "User",
			expectedMessage: "User not found",
		},
		{
			name:            "Not found without resource name",
			resource:        "",
			expectedMessage: "Resource not found",
		},
		{
			name:            "Not found with document resource",
			resource:        "Document",
			expectedMessage: "Document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			WriteNotFound(w, tt.resource)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
			}

			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Error.Message != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, response.Error.Message)
			}

			if response.Error.Code != ErrCodeNotFound {
				t.Errorf("Expected code '%s', got '%s'", ErrCodeNotFound, response.Error.Code)
			}
		})
	}
}

func TestWriteConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "Conflict with custom message",
			message:         "Email already exists",
			expectedMessage: "Email already exists",
		},
		{
			name:            "Conflict with empty message",
			message:         "",
			expectedMessage: "Resource conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			WriteConflict(w, tt.message)

			if w.Code != http.StatusConflict {
				t.Errorf("Expected status code %d, got %d", http.StatusConflict, w.Code)
			}

			var response ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Error.Message != tt.expectedMessage {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, response.Error.Message)
			}

			if response.Error.Code != ErrCodeConflict {
				t.Errorf("Expected code '%s', got '%s'", ErrCodeConflict, response.Error.Code)
			}
		})
	}
}

func TestWriteInternalError(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	WriteInternalError(w)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error.Message != "An internal error occurred" {
		t.Errorf("Expected message 'An internal error occurred', got '%s'", response.Error.Message)
	}

	if response.Error.Code != ErrCodeInternal {
		t.Errorf("Expected code '%s', got '%s'", ErrCodeInternal, response.Error.Code)
	}
}
