// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents standardized API error codes
type ErrorCode string

const (
	// Client errors
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeRateLimited  ErrorCode = "RATE_LIMITED"
	ErrCodeCSRFInvalid  ErrorCode = "CSRF_INVALID"

	// Server errors
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func WriteError(w http.ResponseWriter, statusCode int, code ErrorCode, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(response)
}

func WriteValidationError(w http.ResponseWriter, message string, fieldErrors map[string]string) {
	details := make(map[string]interface{})
	if fieldErrors != nil {
		details["fields"] = fieldErrors
	}
	WriteError(w, http.StatusBadRequest, ErrCodeValidation, message, details)
}

func WriteUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	WriteError(w, http.StatusUnauthorized, ErrCodeUnauthorized, message, nil)
}

func WriteForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Access denied"
	}
	WriteError(w, http.StatusForbidden, ErrCodeForbidden, message, nil)
}

func WriteNotFound(w http.ResponseWriter, resource string) {
	message := "Resource not found"
	if resource != "" {
		message = resource + " not found"
	}
	WriteError(w, http.StatusNotFound, ErrCodeNotFound, message, nil)
}

func WriteConflict(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	WriteError(w, http.StatusConflict, ErrCodeConflict, message, nil)
}

func WriteInternalError(w http.ResponseWriter) {
	WriteError(w, http.StatusInternalServerError, ErrCodeInternal, "An internal error occurred", nil)
}
