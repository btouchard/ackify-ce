// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestParseCookieSecret(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectError bool
		minLength   int
	}{
		{
			name:        "missing cookie secret generates random",
			envValue:    "",
			expectError: false,
			minLength:   32,
		},
		{
			name:        "valid base64 32-byte secret",
			envValue:    base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expectError: false,
			minLength:   32,
		},
		{
			name:        "valid base64 64-byte secret",
			envValue:    base64.StdEncoding.EncodeToString(make([]byte, 64)),
			expectError: false,
			minLength:   64,
		},
		{
			name:        "raw string secret",
			envValue:    "this-is-a-raw-string-secret-that-should-work",
			expectError: false,
			minLength:   1,
		},
		{
			name:        "short raw string secret",
			envValue:    "short",
			expectError: false,
			minLength:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable
			_ = os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")

			// Set environment variable if specified
			if tt.envValue != "" {
				_ = os.Setenv("ACKIFY_OAUTH_COOKIE_SECRET", tt.envValue)
				defer func() {
					_ = os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")
				}()
			}

			result, err := parseCookieSecret()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) < tt.minLength {
				t.Errorf("Cookie secret length %d is less than minimum %d", len(result), tt.minLength)
			}

			// For empty env value, should generate a random 32-byte secret
			if tt.envValue == "" && len(result) != 32 {
				t.Errorf("Generated secret should be 32 bytes, got %d", len(result))
			}
		})
	}
}

func TestLoad_GoogleProvider(t *testing.T) {
	// Set up environment variables for Google OAuth
	envVars := map[string]string{
		"ACKIFY_BASE_URL":             "https://ackify.example.com",
		"ACKIFY_ORGANISATION":         "Test Organisation",
		"ACKIFY_DB_DSN":               "postgres://user:pass@localhost/test",
		"ACKIFY_OAUTH_CLIENT_ID":      "google-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET":  "google-client-secret",
		"ACKIFY_OAUTH_PROVIDER":       "google",
		"ACKIFY_OAUTH_ALLOWED_DOMAIN": "@example.com",
		"ACKIFY_OAUTH_COOKIE_SECRET":  base64.StdEncoding.EncodeToString(make([]byte, 32)),
		"ACKIFY_LISTEN_ADDR":          ":8080",
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test App config
	if config.App.BaseURL != "https://ackify.example.com" {
		t.Errorf("App.BaseURL = %v, expected https://ackify.example.com", config.App.BaseURL)
	}
	if config.App.Organisation != "Test Organisation" {
		t.Errorf("App.Organisation = %v, expected Test Organisation", config.App.Organisation)
	}
	if !config.App.SecureCookies {
		t.Error("App.SecureCookies should be true for HTTPS base URL")
	}

	// Test Database config
	if config.Database.DSN != "postgres://user:pass@localhost/test" {
		t.Errorf("Database.DSN = %v, expected postgres://user:pass@localhost/test", config.Database.DSN)
	}

	// Test OAuth config for Google
	if config.OAuth.ClientID != "google-client-id" {
		t.Errorf("OAuth.ClientID = %v, expected google-client-id", config.OAuth.ClientID)
	}
	if config.OAuth.AuthURL != "https://accounts.google.com/o/oauth2/auth" {
		t.Errorf("OAuth.AuthURL = %v, expected Google auth URL", config.OAuth.AuthURL)
	}
	if config.OAuth.TokenURL != "https://oauth2.googleapis.com/token" {
		t.Errorf("OAuth.TokenURL = %v, expected Google token URL", config.OAuth.TokenURL)
	}
	if config.OAuth.UserInfoURL != "https://openidconnect.googleapis.com/v1/userinfo" {
		t.Errorf("OAuth.UserInfoURL = %v, expected Google userinfo URL", config.OAuth.UserInfoURL)
	}
	expectedScopes := []string{"openid", "email", "profile"}
	if !equalSlices(config.OAuth.Scopes, expectedScopes) {
		t.Errorf("OAuth.Scopes = %v, expected %v", config.OAuth.Scopes, expectedScopes)
	}
	if config.OAuth.AllowedDomain != "@example.com" {
		t.Errorf("OAuth.AllowedDomain = %v, expected @example.com", config.OAuth.AllowedDomain)
	}
	if len(config.OAuth.CookieSecret) != 32 {
		t.Errorf("OAuth.CookieSecret length = %d, expected 32", len(config.OAuth.CookieSecret))
	}

	// Test Server config
	if config.Server.ListenAddr != ":8080" {
		t.Errorf("Server.ListenAddr = %v, expected :8080", config.Server.ListenAddr)
	}
}

func TestLoad_GitHubProvider(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "http://localhost:8080",
		"ACKIFY_ORGANISATION":        "GitHub Test",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/github",
		"ACKIFY_OAUTH_CLIENT_ID":     "github-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "github-client-secret",
		"ACKIFY_OAUTH_PROVIDER":      "github",
		"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test GitHub-specific OAuth config
	if config.OAuth.AuthURL != "https://github.com/login/oauth/authorize" {
		t.Errorf("OAuth.AuthURL = %v, expected GitHub auth URL", config.OAuth.AuthURL)
	}
	if config.OAuth.TokenURL != "https://github.com/login/oauth/access_token" {
		t.Errorf("OAuth.TokenURL = %v, expected GitHub token URL", config.OAuth.TokenURL)
	}
	if config.OAuth.UserInfoURL != "https://api.github.com/user" {
		t.Errorf("OAuth.UserInfoURL = %v, expected GitHub API user URL", config.OAuth.UserInfoURL)
	}
	expectedScopes := []string{"user:email", "read:user"}
	if !equalSlices(config.OAuth.Scopes, expectedScopes) {
		t.Errorf("OAuth.Scopes = %v, expected %v", config.OAuth.Scopes, expectedScopes)
	}

	// Test that SecureCookies is false for HTTP
	if config.App.SecureCookies {
		t.Error("App.SecureCookies should be false for HTTP base URL")
	}
}

func TestLoad_GitLabProvider(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "https://ackify.gitlab.com",
		"ACKIFY_ORGANISATION":        "GitLab Test",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/gitlab",
		"ACKIFY_OAUTH_CLIENT_ID":     "gitlab-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "gitlab-client-secret",
		"ACKIFY_OAUTH_PROVIDER":      "gitlab",
		"ACKIFY_OAUTH_GITLAB_URL":    "https://gitlab.example.com",
		"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test GitLab-specific OAuth config with custom URL
	if config.OAuth.AuthURL != "https://gitlab.example.com/oauth/authorize" {
		t.Errorf("OAuth.AuthURL = %v, expected custom GitLab auth URL", config.OAuth.AuthURL)
	}
	if config.OAuth.TokenURL != "https://gitlab.example.com/oauth/token" {
		t.Errorf("OAuth.TokenURL = %v, expected custom GitLab token URL", config.OAuth.TokenURL)
	}
	if config.OAuth.UserInfoURL != "https://gitlab.example.com/api/v4/user" {
		t.Errorf("OAuth.UserInfoURL = %v, expected custom GitLab API user URL", config.OAuth.UserInfoURL)
	}
	expectedScopes := []string{"read_user", "profile"}
	if !equalSlices(config.OAuth.Scopes, expectedScopes) {
		t.Errorf("OAuth.Scopes = %v, expected %v", config.OAuth.Scopes, expectedScopes)
	}
}

func TestLoad_GitLabDefaultURL(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "https://ackify.gitlab.com",
		"ACKIFY_ORGANISATION":        "GitLab Test",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/gitlab",
		"ACKIFY_OAUTH_CLIENT_ID":     "gitlab-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "gitlab-client-secret",
		"ACKIFY_OAUTH_PROVIDER":      "gitlab",
		"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	// Ensure OAUTH_GITLAB_URL is not set to test default
	_ = os.Unsetenv("ACKIFY_OAUTH_GITLAB_URL")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test GitLab-specific OAuth config with default URL
	if config.OAuth.AuthURL != "https://gitlab.com/oauth/authorize" {
		t.Errorf("OAuth.AuthURL = %v, expected default GitLab auth URL", config.OAuth.AuthURL)
	}
	if config.OAuth.TokenURL != "https://gitlab.com/oauth/token" {
		t.Errorf("OAuth.TokenURL = %v, expected default GitLab token URL", config.OAuth.TokenURL)
	}
	if config.OAuth.UserInfoURL != "https://gitlab.com/api/v4/user" {
		t.Errorf("OAuth.UserInfoURL = %v, expected default GitLab API user URL", config.OAuth.UserInfoURL)
	}
}

func TestLoad_CustomProvider(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "https://ackify.custom.com",
		"ACKIFY_ORGANISATION":        "Custom Test",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/custom",
		"ACKIFY_OAUTH_CLIENT_ID":     "custom-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "custom-client-secret",
		"ACKIFY_OAUTH_AUTH_URL":      "https://auth.custom.com/oauth/authorize",
		"ACKIFY_OAUTH_TOKEN_URL":     "https://auth.custom.com/oauth/token",
		"ACKIFY_OAUTH_USERINFO_URL":  "https://api.custom.com/user",
		"ACKIFY_OAUTH_SCOPES":        "read,write,admin",
		"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if config.OAuth.AuthURL != "https://auth.custom.com/oauth/authorize" {
		t.Errorf("OAuth.AuthURL = %v, expected custom auth URL", config.OAuth.AuthURL)
	}
	if config.OAuth.TokenURL != "https://auth.custom.com/oauth/token" {
		t.Errorf("OAuth.TokenURL = %v, expected custom token URL", config.OAuth.TokenURL)
	}
	if config.OAuth.UserInfoURL != "https://api.custom.com/user" {
		t.Errorf("OAuth.UserInfoURL = %v, expected custom userinfo URL", config.OAuth.UserInfoURL)
	}
	expectedScopes := []string{"read", "write", "admin"}
	if !equalSlices(config.OAuth.Scopes, expectedScopes) {
		t.Errorf("OAuth.Scopes = %v, expected %v", config.OAuth.Scopes, expectedScopes)
	}
}

func TestLoad_MissingRequiredEnvironmentVariables(t *testing.T) {
	requiredVars := []string{
		"ACKIFY_BASE_URL",
		"ACKIFY_ORGANISATION",
		"ACKIFY_DB_DSN",
	}

	for _, missingVar := range requiredVars {
		t.Run("missing_"+missingVar, func(t *testing.T) {
			envVars := map[string]string{
				"ACKIFY_BASE_URL":            "https://ackify.example.com",
				"ACKIFY_ORGANISATION":        "Test Organisation",
				"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
				"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
				"ACKIFY_OAUTH_PROVIDER":      "google",
				"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
			}

			delete(envVars, missingVar)

			for key, value := range envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range envVars {
					_ = os.Unsetenv(key)
				}
			}()

			_ = os.Unsetenv(missingVar)

			_, err := Load()
			if err == nil {
				t.Errorf("Load() should return error when %s is missing", missingVar)
			}
		})
	}
}

func TestLoad_CustomProviderMissingRequiredVars(t *testing.T) {
	customRequiredVars := []string{
		"ACKIFY_OAUTH_AUTH_URL",
		"ACKIFY_OAUTH_TOKEN_URL",
		"ACKIFY_OAUTH_USERINFO_URL",
	}

	for _, missingVar := range customRequiredVars {
		t.Run("custom_missing_"+missingVar, func(t *testing.T) {
			envVars := map[string]string{
				"ACKIFY_BASE_URL":            "https://ackify.example.com",
				"ACKIFY_ORGANISATION":        "Test Organisation",
				"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
				"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
				"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
				"ACKIFY_OAUTH_AUTH_URL":      "https://auth.custom.com/oauth/authorize",
				"ACKIFY_OAUTH_TOKEN_URL":     "https://auth.custom.com/oauth/token",
				"ACKIFY_OAUTH_USERINFO_URL":  "https://api.custom.com/user",
			}

			delete(envVars, missingVar)

			for key, value := range envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range envVars {
					_ = os.Unsetenv(key)
				}
			}()

			_ = os.Unsetenv(missingVar)

			_, err := Load()
			if err == nil {
				t.Errorf("Load() should return error when %s is missing for custom provider", missingVar)
			}
		})
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "https://ackify.example.com",
		"ACKIFY_ORGANISATION":        "Test Organisation",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
		"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
		"ACKIFY_OAUTH_PROVIDER":      "google",
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	// Ensure optional variables are not set to test defaults
	_ = os.Unsetenv("ACKIFY_OAUTH_ALLOWED_DOMAIN")
	_ = os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")
	_ = os.Unsetenv("ACKIFY_LISTEN_ADDR")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test default values
	if config.OAuth.AllowedDomain != "" {
		t.Errorf("OAuth.AllowedDomain = %v, expected empty string", config.OAuth.AllowedDomain)
	}
	if len(config.OAuth.CookieSecret) != 32 {
		t.Errorf("OAuth.CookieSecret should be generated as 32 bytes, got %d", len(config.OAuth.CookieSecret))
	}
	if config.Server.ListenAddr != ":8080" {
		t.Errorf("Server.ListenAddr = %v, expected :8080", config.Server.ListenAddr)
	}
}

func TestLoad_CustomProviderDefaultScopes(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "https://ackify.custom.com",
		"ACKIFY_ORGANISATION":        "Custom Test",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/custom",
		"ACKIFY_OAUTH_CLIENT_ID":     "custom-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "custom-client-secret",
		"ACKIFY_OAUTH_AUTH_URL":      "https://auth.custom.com/oauth/authorize",
		"ACKIFY_OAUTH_TOKEN_URL":     "https://auth.custom.com/oauth/token",
		"ACKIFY_OAUTH_USERINFO_URL":  "https://api.custom.com/user",
		"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	// Ensure OAUTH_SCOPES is not set to test default
	_ = os.Unsetenv("ACKIFY_OAUTH_SCOPES")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test default scopes for custom provider
	expectedScopes := []string{"openid", "email", "profile"}
	if !equalSlices(config.OAuth.Scopes, expectedScopes) {
		t.Errorf("OAuth.Scopes = %v, expected default %v", config.OAuth.Scopes, expectedScopes)
	}
}

func TestParseCookieSecret_InvalidBase64(t *testing.T) {
	// Test invalid base64 that falls back to raw string
	_ = os.Setenv("ACKIFY_OAUTH_COOKIE_SECRET", "this-is-not-valid-base64!")
	defer func() {
		_ = os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")
	}()

	result, err := parseCookieSecret()
	if err != nil {
		t.Errorf("parseCookieSecret() should not fail for invalid base64: %v", err)
	}

	expected := "this-is-not-valid-base64!"
	if string(result) != expected {
		t.Errorf("parseCookieSecret() = %v, expected %v", string(result), expected)
	}
}

func TestParseCookieSecret_ValidBase64WrongLength(t *testing.T) {
	wrongLength := base64.StdEncoding.EncodeToString(make([]byte, 16)) // 16 bytes instead of 32/64
	_ = os.Setenv("ACKIFY_OAUTH_COOKIE_SECRET", wrongLength)
	defer func() {
		_ = os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")
	}()

	result, err := parseCookieSecret()
	if err != nil {
		t.Errorf("parseCookieSecret() should not fail for wrong length: %v", err)
	}

	if string(result) != wrongLength {
		t.Errorf("parseCookieSecret() should fall back to raw string for wrong length")
	}
}

func TestLoad_ErrorInParseCookieSecret(t *testing.T) {
	envVars := map[string]string{
		"ACKIFY_BASE_URL":            "https://ackify.example.com",
		"ACKIFY_ORGANISATION":        "Test Organisation",
		"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
		"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
		"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
		"ACKIFY_OAUTH_PROVIDER":      "google",
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	_ = os.Setenv("ACKIFY_OAUTH_COOKIE_SECRET", "valid-secret")
	defer func() {
		_ = os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() should not fail: %v", err)
	}

	if config == nil {
		t.Error("Config should not be nil")
	}
}

func TestAppConfig_SecureCookiesLogic(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected bool
	}{
		{
			name:     "HTTPS URL should enable secure cookies",
			baseURL:  "https://ackify.example.com",
			expected: true,
		},
		{
			name:     "HTTP URL should disable secure cookies",
			baseURL:  "http://ackify.example.com",
			expected: false,
		},
		{
			name:     "Mixed case HTTPS should enable secure cookies",
			baseURL:  "HTTPS://ackify.example.com",
			expected: true,
		},
		{
			name:     "Mixed case HTTP should disable secure cookies",
			baseURL:  "HTTP://ackify.example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{
				"ACKIFY_BASE_URL":            tt.baseURL,
				"ACKIFY_ORGANISATION":        "Test Organisation",
				"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
				"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
				"ACKIFY_OAUTH_PROVIDER":      "google",
			}

			for key, value := range envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range envVars {
					_ = os.Unsetenv(key)
				}
			}()

			config, err := Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if config.App.SecureCookies != tt.expected {
				t.Errorf("SecureCookies = %v, expected %v for URL %s",
					config.App.SecureCookies, tt.expected, tt.baseURL)
			}
		})
	}
}

func TestLoad_AdminEmails(t *testing.T) {
	tests := []struct {
		name           string
		adminEmailsEnv string
		expected       []string
	}{
		{
			name:           "single admin email",
			adminEmailsEnv: "admin@example.com",
			expected:       []string{"admin@example.com"},
		},
		{
			name:           "multiple admin emails",
			adminEmailsEnv: "admin@example.com,user@example.com,manager@example.com",
			expected:       []string{"admin@example.com", "user@example.com", "manager@example.com"},
		},
		{
			name:           "admin emails with spaces",
			adminEmailsEnv: "  admin@example.com  ,  user@example.com  ",
			expected:       []string{"admin@example.com", "user@example.com"},
		},
		{
			name:           "admin emails normalized to lowercase",
			adminEmailsEnv: "Admin@Example.COM,User@Test.com",
			expected:       []string{"admin@example.com", "user@test.com"},
		},
		{
			name:           "empty admin emails",
			adminEmailsEnv: "",
			expected:       []string(nil),
		},
		{
			name:           "admin emails with empty values filtered out",
			adminEmailsEnv: "admin@example.com,,user@example.com",
			expected:       []string{"admin@example.com", "user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{
				"ACKIFY_BASE_URL":            "https://ackify.example.com",
				"ACKIFY_ORGANISATION":        "Test Organisation",
				"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
				"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
				"ACKIFY_OAUTH_PROVIDER":      "google",
			}

			if tt.adminEmailsEnv != "" {
				envVars["ACKIFY_ADMIN_EMAILS"] = tt.adminEmailsEnv
			}

			for key, value := range envVars {
				_ = os.Setenv(key, value)
			}
			defer func() {
				for key := range envVars {
					_ = os.Unsetenv(key)
				}
				_ = os.Unsetenv("ACKIFY_ADMIN_EMAILS")
			}()

			config, err := Load()
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if !equalSlices(config.App.AdminEmails, tt.expected) {
				t.Errorf("AdminEmails = %v, expected %v", config.App.AdminEmails, tt.expected)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLoad_MailConfig(t *testing.T) {
	t.Run("mail config with all settings", func(t *testing.T) {
		envVars := map[string]string{
			"ACKIFY_BASE_URL":            "https://ackify.example.com",
			"ACKIFY_ORGANISATION":        "Test Org",
			"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
			"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
			"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
			"ACKIFY_OAUTH_PROVIDER":      "google",
			"ACKIFY_MAIL_HOST":           "smtp.example.com",
			"ACKIFY_MAIL_PORT":           "465",
			"ACKIFY_MAIL_USERNAME":       "noreply@example.com",
			"ACKIFY_MAIL_PASSWORD":       "smtp-password",
			"ACKIFY_MAIL_TLS":            "true",
			"ACKIFY_MAIL_STARTTLS":       "false",
			"ACKIFY_MAIL_TIMEOUT":        "30s",
			"ACKIFY_MAIL_FROM":           "noreply@example.com",
			"ACKIFY_MAIL_FROM_NAME":      "Ackify Notifications",
			"ACKIFY_MAIL_SUBJECT_PREFIX": "[Ackify]",
			"ACKIFY_MAIL_TEMPLATE_DIR":   "/custom/templates/emails",
			"ACKIFY_MAIL_DEFAULT_LOCALE": "fr",
		}

		for key, value := range envVars {
			_ = os.Setenv(key, value)
		}
		defer func() {
			for key := range envVars {
				_ = os.Unsetenv(key)
			}
		}()

		config, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if config.Mail.Host != "smtp.example.com" {
			t.Errorf("Mail.Host = %v, expected smtp.example.com", config.Mail.Host)
		}
		if config.Mail.Port != 465 {
			t.Errorf("Mail.Port = %v, expected 465", config.Mail.Port)
		}
		if config.Mail.Username != "noreply@example.com" {
			t.Errorf("Mail.Username = %v, expected noreply@example.com", config.Mail.Username)
		}
		if config.Mail.Password != "smtp-password" {
			t.Errorf("Mail.Password = %v, expected smtp-password", config.Mail.Password)
		}
		if !config.Mail.TLS {
			t.Error("Mail.TLS should be true")
		}
		if config.Mail.StartTLS {
			t.Error("Mail.StartTLS should be false")
		}
		if config.Mail.Timeout != "30s" {
			t.Errorf("Mail.Timeout = %v, expected 30s", config.Mail.Timeout)
		}
		if config.Mail.From != "noreply@example.com" {
			t.Errorf("Mail.From = %v, expected noreply@example.com", config.Mail.From)
		}
		if config.Mail.FromName != "Ackify Notifications" {
			t.Errorf("Mail.FromName = %v, expected Ackify Notifications", config.Mail.FromName)
		}
		if config.Mail.SubjectPrefix != "[Ackify]" {
			t.Errorf("Mail.SubjectPrefix = %v, expected [Ackify]", config.Mail.SubjectPrefix)
		}
		if config.Mail.TemplateDir != "/custom/templates/emails" {
			t.Errorf("Mail.TemplateDir = %v, expected /custom/templates/emails", config.Mail.TemplateDir)
		}
		if config.Mail.DefaultLocale != "fr" {
			t.Errorf("Mail.DefaultLocale = %v, expected fr", config.Mail.DefaultLocale)
		}
	})

	t.Run("mail config with defaults", func(t *testing.T) {
		envVars := map[string]string{
			"ACKIFY_BASE_URL":            "https://ackify.example.com",
			"ACKIFY_ORGANISATION":        "Test Org",
			"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
			"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
			"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
			"ACKIFY_OAUTH_PROVIDER":      "google",
			"ACKIFY_MAIL_HOST":           "smtp.example.com",
		}

		for key, value := range envVars {
			_ = os.Setenv(key, value)
		}
		defer func() {
			for key := range envVars {
				_ = os.Unsetenv(key)
			}
		}()

		config, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if config.Mail.Port != 587 {
			t.Errorf("Mail.Port = %v, expected default 587", config.Mail.Port)
		}
		if !config.Mail.TLS {
			t.Error("Mail.TLS should default to true")
		}
		if !config.Mail.StartTLS {
			t.Error("Mail.StartTLS should default to true")
		}
		if config.Mail.Timeout != "10s" {
			t.Errorf("Mail.Timeout = %v, expected default 10s", config.Mail.Timeout)
		}
		if config.Mail.FromName != "Test Org" {
			t.Errorf("Mail.FromName = %v, expected organisation name", config.Mail.FromName)
		}
		if config.Mail.TemplateDir != "templates/emails" {
			t.Errorf("Mail.TemplateDir = %v, expected default templates/emails", config.Mail.TemplateDir)
		}
		if config.Mail.DefaultLocale != "en" {
			t.Errorf("Mail.DefaultLocale = %v, expected default en", config.Mail.DefaultLocale)
		}
	})

	t.Run("mail disabled when MAIL_HOST not set", func(t *testing.T) {
		envVars := map[string]string{
			"ACKIFY_BASE_URL":            "https://ackify.example.com",
			"ACKIFY_ORGANISATION":        "Test Org",
			"ACKIFY_DB_DSN":              "postgres://user:pass@localhost/test",
			"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
			"ACKIFY_OAUTH_CLIENT_SECRET": "test-client-secret",
			"ACKIFY_OAUTH_PROVIDER":      "google",
		}

		for key, value := range envVars {
			_ = os.Setenv(key, value)
		}
		defer func() {
			for key := range envVars {
				_ = os.Unsetenv(key)
			}
		}()

		_ = os.Unsetenv("ACKIFY_MAIL_HOST")

		config, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if config.Mail.Host != "" {
			t.Errorf("Mail.Host should be empty when not configured, got %v", config.Mail.Host)
		}
		if config.Mail.Port != 0 {
			t.Errorf("Mail.Port should be 0 when not configured, got %v", config.Mail.Port)
		}
	})
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer",
			key:          "TEST_INT_VAR",
			envValue:     "587",
			defaultValue: 25,
			expected:     587,
		},
		{
			name:         "missing uses default",
			key:          "MISSING_INT_VAR",
			envValue:     "",
			defaultValue: 100,
			expected:     100,
		},
		{
			name:         "invalid integer uses default",
			key:          "INVALID_INT_VAR",
			envValue:     "not-a-number",
			defaultValue: 50,
			expected:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key)
			if tt.envValue != "" {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() {
					_ = os.Unsetenv(tt.key)
				}()
			}

			result := getEnvInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvInt() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true string",
			key:          "TEST_BOOL_VAR",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "1 string",
			key:          "TEST_BOOL_VAR",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false string",
			key:          "TEST_BOOL_VAR",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "missing uses default true",
			key:          "MISSING_BOOL_VAR",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "missing uses default false",
			key:          "MISSING_BOOL_VAR",
			envValue:     "",
			defaultValue: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv(tt.key)
			if tt.envValue != "" {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() {
					_ = os.Unsetenv(tt.key)
				}()
			}

			result := getEnvBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
func TestConfig_AuthValidation(t *testing.T) {
	// Save original env vars
	origOAuthClientID := os.Getenv("ACKIFY_OAUTH_CLIENT_ID")
	origOAuthClientSecret := os.Getenv("ACKIFY_OAUTH_CLIENT_SECRET")
	origMailHost := os.Getenv("ACKIFY_MAIL_HOST")
	origAuthOAuthEnabled := os.Getenv("ACKIFY_AUTH_OAUTH_ENABLED")
	origAuthMagicLinkEnabled := os.Getenv("ACKIFY_AUTH_MAGICLINK_ENABLED")
	origBaseURL := os.Getenv("ACKIFY_BASE_URL")
	origOrg := os.Getenv("ACKIFY_ORGANISATION")
	origDBDSN := os.Getenv("ACKIFY_DB_DSN")
	origCookieSecret := os.Getenv("ACKIFY_OAUTH_COOKIE_SECRET")

	// Cleanup function
	defer func() {
		os.Setenv("ACKIFY_OAUTH_CLIENT_ID", origOAuthClientID)
		os.Setenv("ACKIFY_OAUTH_CLIENT_SECRET", origOAuthClientSecret)
		os.Setenv("ACKIFY_MAIL_HOST", origMailHost)
		os.Setenv("ACKIFY_AUTH_OAUTH_ENABLED", origAuthOAuthEnabled)
		os.Setenv("ACKIFY_AUTH_MAGICLINK_ENABLED", origAuthMagicLinkEnabled)
		os.Setenv("ACKIFY_BASE_URL", origBaseURL)
		os.Setenv("ACKIFY_ORGANISATION", origOrg)
		os.Setenv("ACKIFY_DB_DSN", origDBDSN)
		os.Setenv("ACKIFY_OAUTH_COOKIE_SECRET", origCookieSecret)
	}()

	tests := []struct {
		name          string
		envVars       map[string]string
		expectError   bool
		errorContains string
		checkAuth     func(*testing.T, *Config)
	}{
		{
			name: "OAuth only (auto-detected)",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":            "http://localhost:8080",
				"ACKIFY_ORGANISATION":        "Test Org",
				"ACKIFY_DB_DSN":              "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
				"ACKIFY_OAUTH_CLIENT_ID":     "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET": "test-secret",
				"ACKIFY_OAUTH_PROVIDER":      "google",
			},
			expectError: false,
			checkAuth: func(t *testing.T, cfg *Config) {
				if !cfg.Auth.OAuthEnabled {
					t.Error("OAuth should be enabled")
				}
				if cfg.Auth.MagicLinkEnabled {
					t.Error("MagicLink should be disabled")
				}
			},
		},
		{
			name: "MagicLink only (auto-detected)",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":               "http://localhost:8080",
				"ACKIFY_ORGANISATION":           "Test Org",
				"ACKIFY_DB_DSN":                 "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET":    base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
				"ACKIFY_MAIL_HOST":              "smtp.example.com",
				"ACKIFY_AUTH_MAGICLINK_ENABLED": "true",
			},
			expectError: false,
			checkAuth: func(t *testing.T, cfg *Config) {
				if cfg.Auth.OAuthEnabled {
					t.Error("OAuth should be disabled")
				}
				if !cfg.Auth.MagicLinkEnabled {
					t.Error("MagicLink should be enabled")
				}
				if !cfg.App.SMTPEnabled {
					t.Error("SMTP should be enabled")
				}
			},
		},
		{
			name: "Both OAuth and MagicLink enabled",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":               "http://localhost:8080",
				"ACKIFY_ORGANISATION":           "Test Org",
				"ACKIFY_DB_DSN":                 "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET":    base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
				"ACKIFY_OAUTH_CLIENT_ID":        "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET":    "test-secret",
				"ACKIFY_OAUTH_PROVIDER":         "google",
				"ACKIFY_MAIL_HOST":              "smtp.example.com",
				"ACKIFY_AUTH_MAGICLINK_ENABLED": "true",
			},
			expectError: false,
			checkAuth: func(t *testing.T, cfg *Config) {
				if !cfg.Auth.OAuthEnabled {
					t.Error("OAuth should be enabled")
				}
				if !cfg.Auth.MagicLinkEnabled {
					t.Error("MagicLink should be enabled")
				}
				if !cfg.App.SMTPEnabled {
					t.Error("SMTP should be enabled")
				}
			},
		},
		{
			name: "No authentication method (should fail)",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":            "http://localhost:8080",
				"ACKIFY_ORGANISATION":        "Test Org",
				"ACKIFY_DB_DSN":              "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
			},
			expectError:   true,
			errorContains: "at least one authentication method must be enabled",
		},
		{
			name: "Manual override - OAuth enabled despite missing client ID",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":            "http://localhost:8080",
				"ACKIFY_ORGANISATION":        "Test Org",
				"ACKIFY_DB_DSN":              "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET": base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
				"ACKIFY_AUTH_OAUTH_ENABLED":  "true",
				"ACKIFY_OAUTH_PROVIDER":      "google",
			},
			expectError: false,
			checkAuth: func(t *testing.T, cfg *Config) {
				if !cfg.Auth.OAuthEnabled {
					t.Error("OAuth should be force-enabled via ACKIFY_AUTH_OAUTH_ENABLED")
				}
			},
		},
		{
			name: "Manual override - disable OAuth even with credentials",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":               "http://localhost:8080",
				"ACKIFY_ORGANISATION":           "Test Org",
				"ACKIFY_DB_DSN":                 "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET":    base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
				"ACKIFY_OAUTH_CLIENT_ID":        "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET":    "test-secret",
				"ACKIFY_MAIL_HOST":              "smtp.example.com",
				"ACKIFY_AUTH_OAUTH_ENABLED":     "false",
				"ACKIFY_AUTH_MAGICLINK_ENABLED": "true",
			},
			expectError: false,
			checkAuth: func(t *testing.T, cfg *Config) {
				if cfg.Auth.OAuthEnabled {
					t.Error("OAuth should be disabled via ACKIFY_AUTH_OAUTH_ENABLED=false")
				}
				if !cfg.Auth.MagicLinkEnabled {
					t.Error("MagicLink should still be enabled")
				}
			},
		},
		{
			name: "Manual override - disable MagicLink even with SMTP configured",
			envVars: map[string]string{
				"ACKIFY_BASE_URL":               "http://localhost:8080",
				"ACKIFY_ORGANISATION":           "Test Org",
				"ACKIFY_DB_DSN":                 "postgres://localhost/test",
				"ACKIFY_OAUTH_COOKIE_SECRET":    base64.StdEncoding.EncodeToString([]byte("test-secret-32-bytes-long!!!!!!")),
				"ACKIFY_OAUTH_CLIENT_ID":        "test-client-id",
				"ACKIFY_OAUTH_CLIENT_SECRET":    "test-secret",
				"ACKIFY_OAUTH_PROVIDER":         "google",
				"ACKIFY_MAIL_HOST":              "smtp.example.com",
				"ACKIFY_AUTH_MAGICLINK_ENABLED": "false",
			},
			expectError: false,
			checkAuth: func(t *testing.T, cfg *Config) {
				if !cfg.Auth.OAuthEnabled {
					t.Error("OAuth should still be enabled")
				}
				if cfg.Auth.MagicLinkEnabled {
					t.Error("MagicLink should be disabled via ACKIFY_AUTH_MAGICLINK_ENABLED=false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all auth-related env vars
			os.Unsetenv("ACKIFY_OAUTH_CLIENT_ID")
			os.Unsetenv("ACKIFY_OAUTH_CLIENT_SECRET")
			os.Unsetenv("ACKIFY_OAUTH_PROVIDER")
			os.Unsetenv("ACKIFY_MAIL_HOST")
			os.Unsetenv("ACKIFY_AUTH_OAUTH_ENABLED")
			os.Unsetenv("ACKIFY_AUTH_MAGICLINK_ENABLED")
			os.Unsetenv("ACKIFY_BASE_URL")
			os.Unsetenv("ACKIFY_ORGANISATION")
			os.Unsetenv("ACKIFY_DB_DSN")
			os.Unsetenv("ACKIFY_OAUTH_COOKIE_SECRET")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Try to load config
			cfg, err := Load()

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			// Should not error
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Run auth check if provided
			if tt.checkAuth != nil {
				tt.checkAuth(t, cfg)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
