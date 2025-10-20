// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TESTS - getTemplatesDir
// ============================================================================

func TestGetTemplatesDir_EnvVariable(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Set environment variable
	customPath := "/custom/templates"
	t.Setenv("ACKIFY_TEMPLATES_DIR", customPath)

	result := getTemplatesDir()

	assert.Equal(t, customPath, result, "Should use environment variable")
}

func TestGetTemplatesDir_FallbackToPaths(t *testing.T) {
	// Cannot run in parallel as we need to control environment

	// Create temporary directory
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	err := os.Mkdir(templatesDir, 0755)
	require.NoError(t, err)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Ensure no env variable
	os.Unsetenv("ACKIFY_TEMPLATES_DIR")

	result := getTemplatesDir()

	// Should find the templates directory
	assert.Contains(t, result, "templates")
}

func TestGetTemplatesDir_DefaultFallback(t *testing.T) {
	// Cannot use t.Parallel() when modifying environment

	// Ensure no env variable
	os.Unsetenv("ACKIFY_TEMPLATES_DIR")

	result := getTemplatesDir()

	// Should return default even if path doesn't exist
	assert.NotEmpty(t, result)
	assert.Equal(t, "templates", result, "Should return default path")
}

// ============================================================================
// TESTS - getLocalesDir
// ============================================================================

func TestGetLocalesDir_EnvVariable(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Set environment variable
	customPath := "/custom/locales"
	t.Setenv("ACKIFY_LOCALES_DIR", customPath)

	result := getLocalesDir()

	assert.Equal(t, customPath, result, "Should use environment variable")
}

func TestGetLocalesDir_FallbackToPaths(t *testing.T) {
	// Cannot run in parallel as we need to control environment

	// Create temporary directory
	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "locales")
	err := os.Mkdir(localesDir, 0755)
	require.NoError(t, err)

	// Change to temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Ensure no env variable
	os.Unsetenv("ACKIFY_LOCALES_DIR")

	result := getLocalesDir()

	// Should find the locales directory
	assert.Contains(t, result, "locales")
}

func TestGetLocalesDir_DefaultFallback(t *testing.T) {
	// Cannot use t.Parallel() when modifying environment

	// Ensure no env variable
	os.Unsetenv("ACKIFY_LOCALES_DIR")

	result := getLocalesDir()

	// Should return default even if path doesn't exist
	assert.NotEmpty(t, result)
	assert.Equal(t, "locales", result, "Should return default path")
}

// ============================================================================
// TESTS - Server Accessors
// ============================================================================

func TestServer_Accessors(t *testing.T) {
	t.Parallel()

	// We can't easily create a full server without database,
	// but we can test the accessor methods exist and return correctly

	// This test verifies the Server struct has the expected methods
	// by checking the method signatures at compile time

	// Create a nil server to test method existence
	var s *Server

	// These should compile successfully
	_ = s.GetAddr
	_ = s.Router
	_ = s.GetDB
	_ = s.GetAdminEmails
	_ = s.GetAuthService
	_ = s.GetEmailSender
	_ = s.RegisterRoutes

	// If we got here, all methods exist
	assert.True(t, true, "All accessor methods exist")
}

// ============================================================================
// TESTS - Directory Path Resolution
// ============================================================================

func TestGetTemplatesDir_PathResolution(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv() in subtests

	tests := []struct {
		name        string
		envValue    string
		expectValue string
	}{
		{
			name:        "absolute path",
			envValue:    "/absolute/path/templates",
			expectValue: "/absolute/path/templates",
		},
		{
			name:        "relative path",
			envValue:    "relative/templates",
			expectValue: "relative/templates",
		},
		{
			name:        "empty string falls back",
			envValue:    "",
			expectValue: "templates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv()

			if tt.envValue != "" {
				t.Setenv("ACKIFY_TEMPLATES_DIR", tt.envValue)
			} else {
				os.Unsetenv("ACKIFY_TEMPLATES_DIR")
			}

			result := getTemplatesDir()

			if tt.envValue != "" {
				assert.Equal(t, tt.expectValue, result)
			} else {
				// When empty, should get default
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestGetLocalesDir_PathResolution(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv() in subtests

	tests := []struct {
		name        string
		envValue    string
		expectValue string
	}{
		{
			name:        "absolute path",
			envValue:    "/absolute/path/locales",
			expectValue: "/absolute/path/locales",
		},
		{
			name:        "relative path",
			envValue:    "relative/locales",
			expectValue: "relative/locales",
		},
		{
			name:        "empty string falls back",
			envValue:    "",
			expectValue: "locales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv()

			if tt.envValue != "" {
				t.Setenv("ACKIFY_LOCALES_DIR", tt.envValue)
			} else {
				os.Unsetenv("ACKIFY_LOCALES_DIR")
			}

			result := getLocalesDir()

			if tt.envValue != "" {
				assert.Equal(t, tt.expectValue, result)
			} else {
				// When empty, should get default
				assert.NotEmpty(t, result)
			}
		})
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkGetTemplatesDir(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getTemplatesDir()
	}
}

func BenchmarkGetLocalesDir(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getLocalesDir()
	}
}

func BenchmarkGetTemplatesDir_WithEnv(b *testing.B) {
	os.Setenv("ACKIFY_TEMPLATES_DIR", "/custom/path")
	defer os.Unsetenv("ACKIFY_TEMPLATES_DIR")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getTemplatesDir()
	}
}

func BenchmarkGetLocalesDir_WithEnv(b *testing.B) {
	os.Setenv("ACKIFY_LOCALES_DIR", "/custom/path")
	defer os.Unsetenv("ACKIFY_LOCALES_DIR")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		getLocalesDir()
	}
}
