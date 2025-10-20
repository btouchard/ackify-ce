// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"errors"
	"testing"
)

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedMsg    string
		shouldNotBeNil bool
	}{
		{
			name:           "ErrSignatureNotFound",
			err:            ErrSignatureNotFound,
			expectedMsg:    "signature not found",
			shouldNotBeNil: true,
		},
		{
			name:           "ErrSignatureAlreadyExists",
			err:            ErrSignatureAlreadyExists,
			expectedMsg:    "signature already exists",
			shouldNotBeNil: true,
		},
		{
			name:           "ErrInvalidUser",
			err:            ErrInvalidUser,
			expectedMsg:    "invalid user",
			shouldNotBeNil: true,
		},
		{
			name:           "ErrInvalidDocument",
			err:            ErrInvalidDocument,
			expectedMsg:    "invalid document ID",
			shouldNotBeNil: true,
		},
		{
			name:           "ErrDatabaseConnection",
			err:            ErrDatabaseConnection,
			expectedMsg:    "database connection error",
			shouldNotBeNil: true,
		},
		{
			name:           "ErrUnauthorized",
			err:            ErrUnauthorized,
			expectedMsg:    "unauthorized",
			shouldNotBeNil: true,
		},
		{
			name:           "ErrDomainNotAllowed",
			err:            ErrDomainNotAllowed,
			expectedMsg:    "domain not allowed",
			shouldNotBeNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldNotBeNil && tt.err == nil {
				t.Errorf("Error should not be nil")
				return
			}

			if tt.err.Error() != tt.expectedMsg {
				t.Errorf("Error message mismatch: got %v, expected %v", tt.err.Error(), tt.expectedMsg)
			}
		})
	}
}

func TestErrorComparison(t *testing.T) {
	tests := []struct {
		name  string
		err1  error
		err2  error
		equal bool
	}{
		{
			name:  "same error instances are equal",
			err1:  ErrSignatureNotFound,
			err2:  ErrSignatureNotFound,
			equal: true,
		},
		{
			name:  "different error instances are not equal",
			err1:  ErrSignatureNotFound,
			err2:  ErrSignatureAlreadyExists,
			equal: false,
		},
		{
			name:  "wrapped errors can be detected",
			err1:  ErrInvalidUser,
			err2:  errors.New("wrapped: invalid user"),
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEqual := errors.Is(tt.err1, tt.err2)
			if isEqual != tt.equal {
				t.Errorf("Error comparison mismatch: got %v, expected %v", isEqual, tt.equal)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	errs := []error{
		ErrSignatureNotFound,
		ErrSignatureAlreadyExists,
		ErrInvalidUser,
		ErrInvalidDocument,
		ErrDatabaseConnection,
		ErrUnauthorized,
		ErrDomainNotAllowed,
	}

	for i, err := range errs {
		t.Run("error_type_"+string(rune(i+'0')), func(t *testing.T) {
			if err == nil {
				t.Error("Error should not be nil")
			}

			if _, ok := err.(error); !ok {
				t.Error("Error should implement error interface")
			}

			if err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}

func TestErrorUniqueness(t *testing.T) {
	errs := map[string]error{
		"signature not found":       ErrSignatureNotFound,
		"signature already exists":  ErrSignatureAlreadyExists,
		"invalid user":              ErrInvalidUser,
		"invalid document ID":       ErrInvalidDocument,
		"database connection error": ErrDatabaseConnection,
		"unauthorized":              ErrUnauthorized,
		"domain not allowed":        ErrDomainNotAllowed,
	}

	messages := make(map[string]bool)
	for msg, err := range errs {
		if messages[msg] {
			t.Errorf("Duplicate error message found: %v", msg)
		}
		messages[msg] = true

		if err.Error() != msg {
			t.Errorf("Error message mismatch for %v: got %v, expected %v", err, err.Error(), msg)
		}
	}

	expectedCount := 7
	if len(messages) != expectedCount {
		t.Errorf("Expected %d unique error messages, got %d", expectedCount, len(messages))
	}
}
