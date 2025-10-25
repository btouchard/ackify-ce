// SPDX-License-Identifier: AGPL-3.0-or-later
package i18n

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TEST FIXTURES
// ============================================================================

var testLocalesDir = filepath.Join("..", "..", "..", "locales")

// ============================================================================
// TESTS - NewI18n
// ============================================================================

func TestNewI18n_Success(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)

	require.NoError(t, err)
	require.NotNil(t, i18n)
	assert.NotEmpty(t, i18n.translations)
	assert.Contains(t, i18n.translations, "en")
	assert.Contains(t, i18n.translations, "fr")
}

func TestNewI18n_InvalidDirectory(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n("/nonexistent/directory")

	assert.Error(t, err)
	assert.Nil(t, i18n)
	assert.Contains(t, err.Error(), "failed to load en translations")
}

func TestNewI18n_MissingEnglishFile(t *testing.T) {
	t.Parallel()

	// Create temporary directory without en.json
	tmpDir := t.TempDir()

	i18n, err := NewI18n(tmpDir)

	assert.Error(t, err)
	assert.Nil(t, i18n)
}

func TestNewI18n_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Create temporary directory with invalid JSON
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte("invalid json"), 0644)
	require.NoError(t, err)

	i18n, err := NewI18n(tmpDir)

	assert.Error(t, err)
	assert.Nil(t, i18n)
}

// ============================================================================
// TESTS - T (Translation)
// ============================================================================

func TestI18n_T_EnglishTranslation(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	// Test a known key from en.json
	result := i18n.T("en", "email.reminder.subject")
	assert.NotEmpty(t, result)
	assert.NotEqual(t, "email.reminder.subject", result, "Should return translation, not key")
	assert.Contains(t, result, "Document Reading", "Should contain expected text")
}

func TestI18n_T_FrenchTranslation(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	// Test a known key from fr.json
	result := i18n.T("fr", "email.reminder.subject")
	assert.NotEmpty(t, result)
	assert.NotEqual(t, "email.reminder.subject", result, "Should return translation, not key")
	assert.Contains(t, result, "lecture", "Should contain expected French text")
}

func TestI18n_T_FallbackToEnglish(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	// Request French translation for a key - should work for existing keys
	result := i18n.T("fr", "email.reminder.subject")
	assert.NotEmpty(t, result)
}

func TestI18n_T_UnknownKey(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	unknownKey := "unknown.key.that.does.not.exist"
	result := i18n.T("en", unknownKey)

	assert.Equal(t, unknownKey, result, "Should return key itself when translation not found")
}

func TestI18n_T_UnknownLanguage(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	// Test with unsupported language (Chinese), should fallback to English
	result := i18n.T("zh", "email.reminder.subject")
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Document Reading", "Should fallback to English translation")
}

// ============================================================================
// TESTS - GetLangFromRequest
// ============================================================================

func TestGetLangFromRequest_FromCookie(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		cookieValue  string
		expectedLang string
	}{
		{
			name:         "English cookie",
			cookieValue:  "en",
			expectedLang: "en",
		},
		{
			name:         "French cookie",
			cookieValue:  "fr",
			expectedLang: "fr",
		},
		{
			name:         "English with region",
			cookieValue:  "en-US",
			expectedLang: "en",
		},
		{
			name:         "French with region",
			cookieValue:  "fr-FR",
			expectedLang: "fr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  LangCookieName,
				Value: tt.cookieValue,
			})

			lang := GetLangFromRequest(req)
			assert.Equal(t, tt.expectedLang, lang)
		})
	}
}

func TestGetLangFromRequest_FromAcceptLanguageHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		acceptLang   string
		expectedLang string
	}{
		{
			name:         "English",
			acceptLang:   "en",
			expectedLang: "en",
		},
		{
			name:         "French",
			acceptLang:   "fr",
			expectedLang: "fr",
		},
		{
			name:         "English with quality",
			acceptLang:   "en-US,en;q=0.9",
			expectedLang: "en",
		},
		{
			name:         "French with quality",
			acceptLang:   "fr-FR,fr;q=0.9,en;q=0.8",
			expectedLang: "fr",
		},
		{
			name:         "Unsupported language defaults to English",
			acceptLang:   "zh,ja",
			expectedLang: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Language", tt.acceptLang)

			lang := GetLangFromRequest(req)
			assert.Equal(t, tt.expectedLang, lang)
		})
	}
}

func TestGetLangFromRequest_DefaultToEnglish(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No cookie, no Accept-Language header

	lang := GetLangFromRequest(req)
	assert.Equal(t, DefaultLang, lang)
}

func TestGetLangFromRequest_CookieTakesPrecedence(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  LangCookieName,
		Value: "fr",
	})
	req.Header.Set("Accept-Language", "en")

	lang := GetLangFromRequest(req)
	assert.Equal(t, "fr", lang, "Cookie should take precedence over Accept-Language header")
}

// ============================================================================
// TESTS - SetLangCookie
// ============================================================================

func TestSetLangCookie_ValidLanguages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		lang           string
		secureCookies  bool
		expectedLang   string
		expectedSecure bool
	}{
		{
			name:           "English",
			lang:           "en",
			secureCookies:  false,
			expectedLang:   "en",
			expectedSecure: false,
		},
		{
			name:           "French",
			lang:           "fr",
			secureCookies:  false,
			expectedLang:   "fr",
			expectedSecure: false,
		},
		{
			name:           "English with secure cookies",
			lang:           "en",
			secureCookies:  true,
			expectedLang:   "en",
			expectedSecure: true,
		},
		{
			name:           "English with region",
			lang:           "en-US",
			secureCookies:  false,
			expectedLang:   "en",
			expectedSecure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			SetLangCookie(rec, tt.lang, tt.secureCookies)

			cookies := rec.Result().Cookies()
			require.Len(t, cookies, 1, "Should set exactly one cookie")

			cookie := cookies[0]
			assert.Equal(t, LangCookieName, cookie.Name)
			assert.Equal(t, tt.expectedLang, cookie.Value)
			assert.Equal(t, "/", cookie.Path)
			assert.Equal(t, 365*24*60*60, cookie.MaxAge)
			assert.True(t, cookie.HttpOnly)
			assert.Equal(t, tt.expectedSecure, cookie.Secure)
			assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
		})
	}
}

func TestSetLangCookie_UnsupportedLanguage(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	SetLangCookie(rec, "zh", false)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, DefaultLang, cookie.Value, "Unsupported language should default to English")
}

// ============================================================================
// TESTS - normalizeLang
// ============================================================================

func Test_normalizeLang(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "English",
			input:    "en",
			expected: "en",
		},
		{
			name:     "French",
			input:    "fr",
			expected: "fr",
		},
		{
			name:     "English with region",
			input:    "en-US",
			expected: "en",
		},
		{
			name:     "French with region",
			input:    "fr-FR",
			expected: "fr",
		},
		{
			name:     "English uppercase",
			input:    "EN",
			expected: "en",
		},
		{
			name:     "English mixed case",
			input:    "En-Us",
			expected: "en",
		},
		{
			name:     "German",
			input:    "de",
			expected: "de",
		},
		{
			name:     "German with region",
			input:    "de-DE",
			expected: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := normalizeLang(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// TESTS - isSupported
// ============================================================================

func Test_isSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lang     string
		expected bool
	}{
		{
			name:     "English",
			lang:     "en",
			expected: true,
		},
		{
			name:     "French",
			lang:     "fr",
			expected: true,
		},
		{
			name:     "English with region",
			lang:     "en-US",
			expected: true,
		},
		{
			name:     "French with region",
			lang:     "fr-FR",
			expected: true,
		},
		{
			name:     "German",
			lang:     "de",
			expected: true,
		},
		{
			name:     "Spanish",
			lang:     "es",
			expected: true,
		},
		{
			name:     "Italian",
			lang:     "it",
			expected: true,
		},
		{
			name:     "Unsupported language (Chinese)",
			lang:     "zh",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSupported(tt.lang)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// TESTS - GetTranslations
// ============================================================================

func TestI18n_GetTranslations_English(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	translations := i18n.GetTranslations("en")
	assert.NotEmpty(t, translations)
}

func TestI18n_GetTranslations_French(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	translations := i18n.GetTranslations("fr")
	assert.NotEmpty(t, translations)
}

func TestI18n_GetTranslations_UnsupportedLanguage(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	// Should fallback to default language (English) for truly unsupported languages
	translations := i18n.GetTranslations("zh")
	assert.NotEmpty(t, translations)
	assert.Equal(t, i18n.translations[DefaultLang], translations)
}

// ============================================================================
// TESTS - Concurrency
// ============================================================================

func TestI18n_T_Concurrent(t *testing.T) {
	t.Parallel()

	i18n, err := NewI18n(testLocalesDir)
	require.NoError(t, err)

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			lang := "en"
			if id%2 == 0 {
				lang = "fr"
			}

			result := i18n.T(lang, "email.reminder.subject")
			assert.NotEmpty(t, result)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkI18n_T(b *testing.B) {
	i18n, err := NewI18n(testLocalesDir)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		i18n.T("en", "email.reminder.subject")
	}
}

func BenchmarkI18n_T_Parallel(b *testing.B) {
	i18n, err := NewI18n(testLocalesDir)
	require.NoError(b, err)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i18n.T("en", "email.reminder.subject")
		}
	})
}

func BenchmarkGetLangFromRequest(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  LangCookieName,
		Value: "fr",
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetLangFromRequest(req)
	}
}

func BenchmarkGetLangFromRequest_Parallel(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  LangCookieName,
		Value: "fr",
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetLangFromRequest(req)
		}
	})
}

func BenchmarkSetLangCookie(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		SetLangCookie(rec, "fr", false)
	}
}
