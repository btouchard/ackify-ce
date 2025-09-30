// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestConfig_Structures(t *testing.T) {
	t.Run("Config structure", func(t *testing.T) {
		config := &Config{
			App: AppConfig{
				BaseURL:       "https://example.com",
				Organisation:  "Test Org",
				SecureCookies: true,
			},
			Database: DatabaseConfig{
				DSN: "postgres://user:pass@localhost/db",
			},
			OAuth: OAuthConfig{
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				AuthURL:       "https://provider.com/auth",
				TokenURL:      "https://provider.com/token",
				UserInfoURL:   "https://provider.com/userinfo",
				Scopes:        []string{"openid", "email"},
				AllowedDomain: "@example.com",
				CookieSecret:  []byte("test-secret"),
			},
			Server: ServerConfig{
				ListenAddr: ":8080",
			},
		}

		// Test that all fields are accessible
		if config.App.BaseURL != "https://example.com" {
			t.Errorf("App.BaseURL mismatch")
		}
		if config.Database.DSN != "postgres://user:pass@localhost/db" {
			t.Errorf("Database.DSN mismatch")
		}
		if config.OAuth.ClientID != "test-client-id" {
			t.Errorf("OAuth.ClientID mismatch")
		}
		if config.Server.ListenAddr != ":8080" {
			t.Errorf("Server.ListenAddr mismatch")
		}
	})

	t.Run("AppConfig structure", func(t *testing.T) {
		app := AppConfig{
			BaseURL:       "https://ackify.example.com",
			Organisation:  "My Company",
			SecureCookies: true,
		}

		if app.BaseURL == "" {
			t.Error("BaseURL should not be empty")
		}
		if app.Organisation == "" {
			t.Error("Organisation should not be empty")
		}
		if !app.SecureCookies {
			t.Error("SecureCookies should be true for HTTPS")
		}
	})

	t.Run("OAuthConfig structure", func(t *testing.T) {
		oauth := OAuthConfig{
			ClientID:      "oauth-client-123",
			ClientSecret:  "oauth-secret-456",
			AuthURL:       "https://oauth.provider.com/auth",
			TokenURL:      "https://oauth.provider.com/token",
			UserInfoURL:   "https://oauth.provider.com/userinfo",
			Scopes:        []string{"openid", "email", "profile"},
			AllowedDomain: "@company.com",
			CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
		}

		// Test required fields
		if oauth.ClientID == "" {
			t.Error("ClientID should not be empty")
		}
		if oauth.ClientSecret == "" {
			t.Error("ClientSecret should not be empty")
		}
		if len(oauth.Scopes) == 0 {
			t.Error("Scopes should not be empty")
		}
		if len(oauth.CookieSecret) == 0 {
			t.Error("CookieSecret should not be empty")
		}
	})
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "existing environment variable",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "custom_value",
			expected:     "custom_value",
		},
		{
			name:         "missing environment variable uses default",
			key:          "MISSING_ENV_VAR",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "empty environment variable uses default",
			key:          "EMPTY_ENV_VAR",
			defaultValue: "default_value",
			envValue:     "",
			expected:     "default_value",
		},
		{
			name:         "whitespace-only environment variable uses default",
			key:          "WHITESPACE_ENV_VAR",
			defaultValue: "default_value",
			envValue:     "   ",
			expected:     "default_value",
		},
		{
			name:         "environment variable with leading/trailing spaces",
			key:          "SPACES_ENV_VAR",
			defaultValue: "default",
			envValue:     "  value_with_spaces  ",
			expected:     "value_with_spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before test
			_ = os.Unsetenv(tt.key)

			// Set environment variable if specified
			if tt.envValue != "" {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func(key string) {
					_ = os.Unsetenv(key)
				}(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMustGetEnv(t *testing.T) {
	t.Run("existing environment variable", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR"
		expected := "test_value"
		_ = os.Setenv(key, expected)
		defer func(key string) {
			_ = os.Unsetenv(key)
		}(key)

		result := mustGetEnv(key)
		if result != expected {
			t.Errorf("mustGetEnv() = %v, expected %v", result, expected)
		}
	})

	t.Run("environment variable with spaces is trimmed", func(t *testing.T) {
		key := "TEST_MUST_ENV_VAR_SPACES"
		_ = os.Setenv(key, "  trimmed_value  ")
		defer func(key string) {
			_ = os.Unsetenv(key)
		}(key)

		result := mustGetEnv(key)
		if result != "trimmed_value" {
			t.Errorf("mustGetEnv() = %v, expected 'trimmed_value'", result)
		}
	})

	t.Run("missing environment variable panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("mustGetEnv() should panic for missing environment variable")
			}
		}()

		mustGetEnv("DEFINITELY_MISSING_ENV_VAR")
	})

	t.Run("empty environment variable panics", func(t *testing.T) {
		key := "TEST_EMPTY_ENV_VAR"
		_ = os.Setenv(key, "")
		defer func(key string) {
			_ = os.Unsetenv(key)
		}(key)

		defer func() {
			if r := recover(); r == nil {
				t.Error("mustGetEnv() should panic for empty environment variable")
			}
		}()

		mustGetEnv(key)
	})

	t.Run("whitespace-only environment variable panics", func(t *testing.T) {
		key := "TEST_WHITESPACE_ENV_VAR"
		_ = os.Setenv(key, "   ")
		defer func(key string) {
			_ = os.Unsetenv(key)
		}(key)

		defer func() {
			if r := recover(); r == nil {
				t.Error("mustGetEnv() should panic for whitespace-only environment variable")
			}
		}()

		mustGetEnv(key)
	})
}

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
		"ACKIFY_OAUTH_CLIENT_ID",
		"ACKIFY_OAUTH_CLIENT_SECRET",
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

			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Load() should panic when %s is missing", missingVar)
				}
			}()

			_, _ = Load()
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

			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Load() should panic when %s is missing for custom provider", missingVar)
				}
			}()

			_, _ = Load()
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
