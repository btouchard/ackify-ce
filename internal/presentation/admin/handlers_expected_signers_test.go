// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"testing"
)

func TestParseEmailsFromText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "newline separated",
			input:    "user1@example.com\nuser2@example.com\nuser3@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:     "comma separated",
			input:    "user1@example.com,user2@example.com,user3@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:     "semicolon separated",
			input:    "user1@example.com;user2@example.com;user3@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com"},
		},
		{
			name:     "mixed separators",
			input:    "user1@example.com\nuser2@example.com,user3@example.com;user4@example.com",
			expected: []string{"user1@example.com", "user2@example.com", "user3@example.com", "user4@example.com"},
		},
		{
			name:     "with extra whitespace",
			input:    "  user1@example.com  \n  user2@example.com  ",
			expected: []string{"user1@example.com", "user2@example.com"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []string{},
		},
		{
			name:     "single email",
			input:    "user@example.com",
			expected: []string{"user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEmailsFromText(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d emails, got %d", len(tt.expected), len(result))
				return
			}

			for i, email := range result {
				if email != tt.expected[i] {
					t.Errorf("at index %d: expected %s, got %s", i, tt.expected[i], email)
				}
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "valid email",
			email: "user@example.com",
			valid: true,
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
			valid: true,
		},
		{
			name:  "valid email with plus",
			email: "user+tag@example.com",
			valid: true,
		},
		{
			name:  "valid email with dots",
			email: "first.last@example.com",
			valid: true,
		},
		{
			name:  "missing @",
			email: "userexample.com",
			valid: false,
		},
		{
			name:  "missing domain",
			email: "user@",
			valid: false,
		},
		{
			name:  "missing username",
			email: "@example.com",
			valid: false,
		},
		{
			name:  "no TLD",
			email: "user@example",
			valid: false,
		},
		{
			name:  "empty string",
			email: "",
			valid: false,
		},
		{
			name:  "whitespace",
			email: "  ",
			valid: false,
		},
		{
			name:  "multiple @",
			email: "user@@example.com",
			valid: false,
		},
		{
			name:  "spaces in email",
			email: "user name@example.com",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.valid {
				t.Errorf("expected %v, got %v for email: %s", tt.valid, result, tt.email)
			}
		})
	}
}
