// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/btouchard/ackify-ce/backend/pkg/models"
	"github.com/stretchr/testify/assert"
)

type fakeAuthService struct {
	shouldFailSetUser  bool
	shouldFailCallback bool
	shouldFailGetUser  bool
	setUserError       error
	getUserError       error
	callbackUser       *models.User
	callbackNextURL    string
	callbackError      error
	authURL            string
	logoutURL          string
	logoutCalled       bool

	verifyStateResult bool
	lastVerifyToken   string
	currentUser       *models.User
}

func newFakeAuthService() *fakeAuthService {
	return &fakeAuthService{
		authURL:           "https://oauth.example.com/auth",
		callbackUser:      &models.User{Sub: "test-user", Email: "test@example.com", Name: "Test User"},
		callbackNextURL:   "/",
		verifyStateResult: true,
	}
}

func (f *fakeAuthService) GetUser(_ *http.Request) (*models.User, error) {
	if f.shouldFailGetUser {
		return nil, f.getUserError
	}
	return f.currentUser, nil
}

func (f *fakeAuthService) SetUser(_ http.ResponseWriter, _ *http.Request, user *models.User) error {
	if f.shouldFailSetUser {
		return f.setUserError
	}
	f.currentUser = user
	return nil
}

func (f *fakeAuthService) Logout(_ http.ResponseWriter, _ *http.Request) {
	f.logoutCalled = true
	f.currentUser = nil
}

func (f *fakeAuthService) GetLogoutURL() string {
	return f.logoutURL
}

func (f *fakeAuthService) GetAuthURL(nextURL string) string {
	return f.authURL + "?next=" + url.QueryEscape(nextURL)
}

func (f *fakeAuthService) CreateAuthURL(_ http.ResponseWriter, _ *http.Request, nextURL string) string {
	return f.GetAuthURL(nextURL)
}

func (f *fakeAuthService) VerifyState(_ http.ResponseWriter, _ *http.Request, token string) bool {
	f.lastVerifyToken = token
	return f.verifyStateResult
}

func (f *fakeAuthService) HandleCallback(_ context.Context, _, _ string) (*models.User, string, error) {
	if f.shouldFailCallback {
		return nil, "", f.callbackError
	}
	return f.callbackUser, f.callbackNextURL, nil
}

type fakeUserService struct {
	user         *models.User
	shouldFail   bool
	getUserError error
}

func newFakeUserService() *fakeUserService {
	return &fakeUserService{
		user: &models.User{Sub: "test-user", Email: "test@example.com", Name: "Test User"},
	}
}

func (f *fakeUserService) GetUser(_ *http.Request) (*models.User, error) {
	if f.shouldFail {
		return nil, f.getUserError
	}
	return f.user, nil
}

func TestHandleOEmbed_Success(t *testing.T) {
	t.Parallel()

	baseURL := "https://example.com"
	handler := HandleOEmbed(baseURL)

	tests := []struct {
		name     string
		docID    string
		referrer string
	}{
		{"simple doc", "doc123", ""},
		{"with referrer", "doc456", "github"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reqURL := baseURL + "/?doc=" + tt.docID
			if tt.referrer != "" {
				reqURL += "&referrer=" + tt.referrer
			}

			req := httptest.NewRequest(http.MethodGet, "/oembed?url="+url.QueryEscape(reqURL), nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			var response OEmbedResponse
			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response.Type != "rich" {
				t.Errorf("Expected type 'rich', got %s", response.Type)
			}
			if response.Version != "1.0" {
				t.Errorf("Expected version '1.0', got %s", response.Version)
			}
			if response.ProviderName != "Ackify" {
				t.Errorf("Expected provider 'Ackify', got %s", response.ProviderName)
			}
			if response.Height != 200 {
				t.Errorf("Expected height 200, got %d", response.Height)
			}
			if !strings.Contains(response.HTML, "iframe") {
				t.Error("Expected HTML to contain iframe")
			}
			if !strings.Contains(response.HTML, tt.docID) {
				t.Errorf("Expected HTML to contain doc ID %s", tt.docID)
			}
		})
	}
}

func TestHandleOEmbed_MissingURLParam(t *testing.T) {
	t.Parallel()

	handler := HandleOEmbed("https://example.com")
	req := httptest.NewRequest(http.MethodGet, "/oembed", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandleOEmbed_InvalidURL(t *testing.T) {
	t.Parallel()

	handler := HandleOEmbed("https://example.com")
	req := httptest.NewRequest(http.MethodGet, "/oembed?url=:::invalid", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandleOEmbed_MissingDocParam(t *testing.T) {
	t.Parallel()

	handler := HandleOEmbed("https://example.com")
	req := httptest.NewRequest(http.MethodGet, "/oembed?url="+url.QueryEscape("https://example.com/"), nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestValidateOEmbedURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		urlStr   string
		baseURL  string
		expected bool
	}{
		{"valid same host", "https://example.com/?doc=123", "https://example.com", true},
		{"valid with port", "https://example.com:443/?doc=123", "https://example.com", true},
		{"different host", "https://other.com/?doc=123", "https://example.com", false},
		{"localhost variations", "http://localhost:8080/?doc=123", "http://127.0.0.1:8080", true},
		{"localhost to 127.0.0.1", "http://127.0.0.1/?doc=123", "http://localhost", true},
		{"invalid URL", ":::invalid", "https://example.com", false},
		{"invalid base URL", "https://example.com/?doc=123", ":::invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateOEmbedURL(tt.urlStr, tt.baseURL)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkHandleOEmbed(b *testing.B) {
	handler := HandleOEmbed("https://example.com")
	reqURL := url.QueryEscape("https://example.com/?doc=test123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/oembed?url="+reqURL, nil)
		rec := httptest.NewRecorder()
		handler(rec, req)
	}
}

func BenchmarkValidateOEmbedURL(b *testing.B) {
	urlStr := "https://example.com/?doc=test123"
	baseURL := "https://example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateOEmbedURL(urlStr, baseURL)
	}
}

// ============================================================================
// TESTS - Middleware: SecureHeaders
// ============================================================================

func TestSecureHeaders_NonEmbedRoute(t *testing.T) {
	t.Parallel()

	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Contains(t, rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'self'")
}

func TestSecureHeaders_EmbedRoute(t *testing.T) {
	t.Parallel()

	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/embed/doc123", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
	assert.Empty(t, rec.Header().Get("X-Frame-Options"), "Embed routes should not have X-Frame-Options")
	assert.Contains(t, rec.Header().Get("Content-Security-Policy"), "frame-ancestors *")
}

func TestSecureHeaders_EmbedRootRoute(t *testing.T) {
	t.Parallel()

	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/embed", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("X-Frame-Options"))
	assert.Contains(t, rec.Header().Get("Content-Security-Policy"), "frame-ancestors *")
}

func TestSecureHeaders_CSPContent(t *testing.T) {
	t.Parallel()

	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "script-src 'self'")
	assert.Contains(t, csp, "style-src 'self'")
	assert.Contains(t, csp, "https://cdn.tailwindcss.com")
	assert.Contains(t, csp, "https://cdn.simpleicons.org")
}

// ============================================================================
// TESTS - Middleware: RequestLogger
// ============================================================================

func TestRequestLogger_Success(t *testing.T) {
	t.Parallel()

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestRequestLogger_WithError(t *testing.T) {
	t.Parallel()

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/fail", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "error", rec.Body.String())
}

func TestRequestLogger_StatusRecorder(t *testing.T) {
	t.Parallel()

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		// Verify the status recorder is working by checking the wrapper
		if sr, ok := w.(*statusRecorder); ok {
			assert.Equal(t, http.StatusCreated, sr.status)
		}
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestRequestLogger_DifferentMethods(t *testing.T) {
	t.Parallel()

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// ============================================================================
// TESTS - HandleError
// ============================================================================

func TestHandleError_Unauthorized(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Unauthorized")
}

func TestHandleError_SignatureNotFound(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrSignatureNotFound)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Signature not found")
}

func TestHandleError_SignatureAlreadyExists(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrSignatureAlreadyExists)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "Signature already exists")
}

func TestHandleError_InvalidUser(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrInvalidUser)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid user")
}

func TestHandleError_InvalidDocument(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrInvalidDocument)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid document ID")
}

func TestHandleError_DomainNotAllowed(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrDomainNotAllowed)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "Domain not allowed")
}

func TestHandleError_DatabaseConnection(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, models.ErrDatabaseConnection)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Database error")
}

func TestHandleError_UnknownError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	HandleError(rec, errors.New("unknown error"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Internal server error")
}

func TestHandleError_WrappedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			"wrapped unauthorized",
			fmt.Errorf("auth failed: %w", models.ErrUnauthorized),
			http.StatusUnauthorized,
			"Unauthorized",
		},
		{
			"wrapped domain error",
			fmt.Errorf("validation failed: %w", models.ErrDomainNotAllowed),
			http.StatusForbidden,
			"Domain not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			HandleError(rec, tt.err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedMsg)
		})
	}
}

// ============================================================================
// TESTS - statusRecorder
// ============================================================================

func TestStatusRecorder_WriteHeader(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	sr.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, sr.status)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestStatusRecorder_DefaultStatus(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	// Don't call WriteHeader, should keep default
	assert.Equal(t, http.StatusOK, sr.status)
}

func TestStatusRecorder_MultipleWriteHeader(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	// First call
	sr.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, sr.status)

	// Second call (should be ignored by http.ResponseWriter)
	sr.WriteHeader(http.StatusInternalServerError)
	// Status recorder updates but ResponseWriter doesn't change
	assert.Equal(t, http.StatusInternalServerError, sr.status)
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkSecureHeaders(b *testing.B) {
	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkRequestLogger(b *testing.B) {
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHandleError(b *testing.B) {
	err := models.ErrUnauthorized

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		HandleError(rec, err)
	}
}
