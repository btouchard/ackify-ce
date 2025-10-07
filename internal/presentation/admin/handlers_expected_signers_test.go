// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"testing"
)

func TestParseContactsFromText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ParsedContact
	}{
		{
			name:  "newline separated plain emails",
			input: "user1@example.com\nuser2@example.com\nuser3@example.com",
			expected: []ParsedContact{
				{Email: "user1@example.com", Name: ""},
				{Email: "user2@example.com", Name: ""},
				{Email: "user3@example.com", Name: ""},
			},
		},
		{
			name:  "comma separated plain emails",
			input: "user1@example.com,user2@example.com,user3@example.com",
			expected: []ParsedContact{
				{Email: "user1@example.com", Name: ""},
				{Email: "user2@example.com", Name: ""},
				{Email: "user3@example.com", Name: ""},
			},
		},
		{
			name:  "with names format",
			input: "Benjamin Touchard <benjamin@example.com>\nMarie Dupont <marie@example.com>",
			expected: []ParsedContact{
				{Email: "benjamin@example.com", Name: "Benjamin Touchard"},
				{Email: "marie@example.com", Name: "Marie Dupont"},
			},
		},
		{
			name:  "mixed formats",
			input: "Benjamin Touchard <benjamin@example.com>\njohn@doe.fr\nMarie Dupont <marie@example.com>",
			expected: []ParsedContact{
				{Email: "benjamin@example.com", Name: "Benjamin Touchard"},
				{Email: "john@doe.fr", Name: ""},
				{Email: "marie@example.com", Name: "Marie Dupont"},
			},
		},
		{
			name:  "with extra whitespace in names",
			input: "  Benjamin Touchard  <  benjamin@example.com  >  ",
			expected: []ParsedContact{
				{Email: "benjamin@example.com", Name: "Benjamin Touchard"},
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []ParsedContact{},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			expected: []ParsedContact{},
		},
		{
			name:  "single email",
			input: "user@example.com",
			expected: []ParsedContact{
				{Email: "user@example.com", Name: ""},
			},
		},
		{
			name:  "single email with name",
			input: "John Doe <john@example.com>",
			expected: []ParsedContact{
				{Email: "john@example.com", Name: "John Doe"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContactsFromText(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d contacts, got %d", len(tt.expected), len(result))
				return
			}

			for i, contact := range result {
				if contact.Email != tt.expected[i].Email {
					t.Errorf("at index %d: expected email %s, got %s", i, tt.expected[i].Email, contact.Email)
				}
				if contact.Name != tt.expected[i].Name {
					t.Errorf("at index %d: expected name %s, got %s", i, tt.expected[i].Name, contact.Name)
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
