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

func TestErrorWrapping(t *testing.T) {
	originalErr := ErrSignatureNotFound
	wrappedErr := errors.Join(originalErr, errors.New("additional context"))

    if !errors.Is(wrappedErr, originalErr) {
        t.Error("Original error should be detectable in wrapped error")
    }

    wrappedMsg := wrappedErr.Error()
    if !contains(wrappedMsg, "signature not found") {
        t.Errorf("Wrapped error should contain original message: %v", wrappedMsg)
    }
    if !contains(wrappedMsg, "additional context") {
        t.Errorf("Wrapped error should contain additional context: %v", wrappedMsg)
    }
}

func TestErrorTypes(t *testing.T) {
    errors := []error{
		ErrSignatureNotFound,
		ErrSignatureAlreadyExists,
		ErrInvalidUser,
		ErrInvalidDocument,
		ErrDatabaseConnection,
		ErrUnauthorized,
		ErrDomainNotAllowed,
	}

	for i, err := range errors {
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
    errors := map[string]error{
		"signature not found":       ErrSignatureNotFound,
		"signature already exists":  ErrSignatureAlreadyExists,
		"invalid user":              ErrInvalidUser,
		"invalid document ID":       ErrInvalidDocument,
		"database connection error": ErrDatabaseConnection,
		"unauthorized":              ErrUnauthorized,
		"domain not allowed":        ErrDomainNotAllowed,
	}

	messages := make(map[string]bool)
	for msg, err := range errors {
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

func TestErrorSentinelValues(t *testing.T) {
    if ErrSignatureNotFound != ErrSignatureNotFound {
		t.Error("ErrSignatureNotFound should be a sentinel value")
	}
	if ErrSignatureAlreadyExists != ErrSignatureAlreadyExists {
		t.Error("ErrSignatureAlreadyExists should be a sentinel value")
	}
	if ErrInvalidUser != ErrInvalidUser {
		t.Error("ErrInvalidUser should be a sentinel value")
	}
	if ErrInvalidDocument != ErrInvalidDocument {
		t.Error("ErrInvalidDocument should be a sentinel value")
	}
	if ErrDatabaseConnection != ErrDatabaseConnection {
		t.Error("ErrDatabaseConnection should be a sentinel value")
	}
	if ErrUnauthorized != ErrUnauthorized {
		t.Error("ErrUnauthorized should be a sentinel value")
	}
	if ErrDomainNotAllowed != ErrDomainNotAllowed {
		t.Error("ErrDomainNotAllowed should be a sentinel value")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if s[start+i] != substr[i] {
			if start+1 < len(s) {
				return containsAt(s, substr, start+1)
			}
			return false
		}
	}
	return true
}
