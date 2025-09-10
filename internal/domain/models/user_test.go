package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUser_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected bool
	}{
		{
			name: "valid user with all fields",
			user: User{
				Sub:   "google-oauth2|123456789",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expected: true,
		},
		{
			name: "valid user without name",
			user: User{
				Sub:   "github|987654321",
				Email: "user@github.com",
				Name:  "",
			},
			expected: true,
		},
		{
			name: "invalid user - missing sub",
			user: User{
				Sub:   "",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expected: false,
		},
		{
			name: "invalid user - missing email",
			user: User{
				Sub:   "google-oauth2|123456789",
				Email: "",
				Name:  "Test User",
			},
			expected: false,
		},
		{
			name: "invalid user - missing both sub and email",
			user: User{
				Sub:   "",
				Email: "",
				Name:  "Test User",
			},
			expected: false,
		},
		{
			name: "invalid user - whitespace only sub",
			user: User{
				Sub:   "   ",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expected: false,
		},
		{
			name: "invalid user - whitespace only email",
			user: User{
				Sub:   "google-oauth2|123456789",
				Email: "   ",
				Name:  "Test User",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsValid()
			if result != tt.expected {
				t.Errorf("User.IsValid() = %v, expected %v for user %+v", result, tt.expected, tt.user)
			}
		})
	}
}

func TestUser_NormalizedEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "lowercase email",
			email:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "uppercase email",
			email:    "TEST@EXAMPLE.COM",
			expected: "test@example.com",
		},
		{
			name:     "mixed case email",
			email:    "TeSt@ExAmPlE.CoM",
			expected: "test@example.com",
		},
		{
			name:     "email with mixed domain",
			email:    "user@GitHub.COM",
			expected: "user@github.com",
		},
		{
			name:     "empty email",
			email:    "",
			expected: "",
		},
		{
			name:     "email with special characters",
			email:    "User+Tag@DOMAIN.ORG",
			expected: "user+tag@domain.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Email: tt.email}
			result := user.NormalizedEmail()
			if result != tt.expected {
				t.Errorf("User.NormalizedEmail() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestUser_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		user     User
		expected string
	}{
		{
			name: "complete user",
			user: User{
				Sub:   "google-oauth2|123456789",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expected: `{"sub":"google-oauth2|123456789","email":"test@example.com","name":"Test User"}`,
		},
		{
			name: "user without name",
			user: User{
				Sub:   "github|987654321",
				Email: "user@github.com",
				Name:  "",
			},
			expected: `{"sub":"github|987654321","email":"user@github.com","name":""}`,
		},
		{
			name: "user with special characters",
			user: User{
				Sub:   "gitlab|special-chars-123",
				Email: "user+tag@domain.org",
				Name:  "Nom avec accénts",
			},
			expected: `{"sub":"gitlab|special-chars-123","email":"user+tag@domain.org","name":"Nom avec accénts"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test serialization
			data, err := json.Marshal(tt.user)
			if err != nil {
				t.Fatalf("Failed to marshal user: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("JSON serialization mismatch:\ngot:      %s\nexpected: %s", string(data), tt.expected)
			}

			// Test deserialization
			var user User
			err = json.Unmarshal(data, &user)
			if err != nil {
				t.Fatalf("Failed to unmarshal user: %v", err)
			}

			if user.Sub != tt.user.Sub || user.Email != tt.user.Email || user.Name != tt.user.Name {
				t.Errorf("Deserialized user mismatch:\ngot:      %+v\nexpected: %+v", user, tt.user)
			}
		})
	}
}

func TestUser_JSONDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected User
		wantErr  bool
	}{
		{
			name:     "valid JSON",
			jsonData: `{"sub":"google-oauth2|123456789","email":"test@example.com","name":"Test User"}`,
			expected: User{
				Sub:   "google-oauth2|123456789",
				Email: "test@example.com",
				Name:  "Test User",
			},
			wantErr: false,
		},
		{
			name:     "JSON with missing name",
			jsonData: `{"sub":"github|987654321","email":"user@github.com"}`,
			expected: User{
				Sub:   "github|987654321",
				Email: "user@github.com",
				Name:  "",
			},
			wantErr: false,
		},
		{
			name:     "JSON with null values",
			jsonData: `{"sub":"gitlab|123","email":"test@example.com","name":null}`,
			expected: User{
				Sub:   "gitlab|123",
				Email: "test@example.com",
				Name:  "",
			},
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{"sub":"invalid"email":"missing-comma"}`,
			expected: User{},
			wantErr:  true,
		},
		{
			name:     "empty JSON object",
			jsonData: `{}`,
			expected: User{
				Sub:   "",
				Email: "",
				Name:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user User
			err := json.Unmarshal([]byte(tt.jsonData), &user)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if user.Sub != tt.expected.Sub || user.Email != tt.expected.Email || user.Name != tt.expected.Name {
				t.Errorf("Deserialized user mismatch:\ngot:      %+v\nexpected: %+v", user, tt.expected)
			}
		})
	}
}

func TestUser_EmailValidationRules(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectValid bool
	}{
		{
			name:        "valid standard email",
			email:       "test@example.com",
			expectValid: true,
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			expectValid: true,
		},
		{
			name:        "valid email with plus sign",
			email:       "user+tag@example.com",
			expectValid: true,
		},
		{
			name:        "valid email with dots",
			email:       "first.last@example.com",
			expectValid: true,
		},
		{
			name:        "empty email is invalid",
			email:       "",
			expectValid: false,
		},
		{
			name:        "whitespace email is invalid",
			email:       "   ",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Sub:   "test-sub",
				Email: tt.email,
			}

			isValid := user.IsValid()
			if isValid != tt.expectValid {
				t.Errorf("User with email '%s' validation = %v, expected %v", tt.email, isValid, tt.expectValid)
			}

			// Test normalized email
			normalized := user.NormalizedEmail()
			if tt.email != "" {
				expectedNormalized := strings.ToLower(tt.email)
				if normalized != expectedNormalized {
					t.Errorf("NormalizedEmail() = %v, expected %v", normalized, expectedNormalized)
				}
			}
		})
	}
}

func TestUser_SubValidationRules(t *testing.T) {
	tests := []struct {
		name        string
		sub         string
		expectValid bool
	}{
		{
			name:        "valid Google OAuth2 sub",
			sub:         "google-oauth2|123456789012345678901",
			expectValid: true,
		},
		{
			name:        "valid GitHub sub",
			sub:         "github|12345678",
			expectValid: true,
		},
		{
			name:        "valid GitLab sub",
			sub:         "gitlab|987654321",
			expectValid: true,
		},
		{
			name:        "valid custom provider sub",
			sub:         "custom-provider|user-123",
			expectValid: true,
		},
		{
			name:        "empty sub is invalid",
			sub:         "",
			expectValid: false,
		},
		{
			name:        "whitespace sub is invalid",
			sub:         "   ",
			expectValid: false,
		},
		{
			name:        "numeric sub is valid",
			sub:         "123456789",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Sub:   tt.sub,
				Email: "test@example.com",
			}

			isValid := user.IsValid()
			if isValid != tt.expectValid {
				t.Errorf("User with sub '%s' validation = %v, expected %v", tt.sub, isValid, tt.expectValid)
			}
		})
	}
}
