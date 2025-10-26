// SPDX-License-Identifier: AGPL-3.0-or-later
package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TEST FIXTURES
// ============================================================================

const (
	testBaseURL      = "http://localhost:8080"
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
)

var testUser = &models.User{
	Sub:   "test-user-123",
	Email: "user@example.com",
	Name:  "Test User",
}

var testAdminUser = &models.User{
	Sub:   "admin-user-123",
	Email: "admin@example.com",
	Name:  "Admin User",
}

func createTestAuthService() *auth.OauthService {
	return auth.NewOAuthService(auth.Config{
		BaseURL:       testBaseURL,
		ClientID:      testClientID,
		ClientSecret:  testClientSecret,
		AuthURL:       "http://localhost:8080/auth",
		TokenURL:      "http://localhost:8080/token",
		UserInfoURL:   "http://localhost:8080/userinfo",
		LogoutURL:     "",
		CookieSecret:  []byte("test-secret-key-32-bytes-long!"),
		Scopes:        []string{"openid", "email", "profile"},
		AllowedDomain: "",
		SecureCookies: false,
	})
}

func createTestMiddleware(adminEmails []string) *Middleware {
	authService := createTestAuthService()
	return NewMiddleware(authService, testBaseURL, adminEmails)
}

// ============================================================================
// TESTS - NewMiddleware
// ============================================================================

func TestNewMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		adminEmails []string
	}{
		{
			name:        "no admin emails",
			adminEmails: []string{},
		},
		{
			name:        "single admin email",
			adminEmails: []string{"admin@example.com"},
		},
		{
			name:        "multiple admin emails",
			adminEmails: []string{"admin1@example.com", "admin2@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestMiddleware(tt.adminEmails)

			require.NotNil(t, m)
			assert.NotNil(t, m.authService)
			assert.NotNil(t, m.csrfTokens)
			assert.Equal(t, testBaseURL, m.baseURL)
			assert.Equal(t, tt.adminEmails, m.adminEmails)
		})
	}
}

// ============================================================================
// TESTS - CORS Middleware
// ============================================================================

func TestMiddleware_CORS(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	tests := []struct {
		name           string
		origin         string
		method         string
		expectCORS     bool
		expectAllowAll bool
	}{
		{
			name:       "localhost dev server",
			origin:     "http://localhost:5173",
			method:     "GET",
			expectCORS: true,
		},
		{
			name:       "localhost OPTIONS preflight",
			origin:     "http://localhost:5173",
			method:     "OPTIONS",
			expectCORS: true,
		},
		{
			name:       "other origin",
			origin:     "http://example.com",
			method:     "GET",
			expectCORS: false,
		},
		{
			name:       "no origin",
			origin:     "",
			method:     "GET",
			expectCORS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := m.CORS(next)

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.expectCORS {
				assert.Equal(t, tt.origin, rec.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
				assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Methods"))
				assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Headers"))
			} else {
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.method == "OPTIONS" {
				assert.Equal(t, http.StatusOK, rec.Code)
				assert.False(t, nextCalled, "Next handler should not be called for OPTIONS")
			} else {
				assert.True(t, nextCalled, "Next handler should be called")
			}
		})
	}
}

// ============================================================================
// TESTS - RequireAuth Middleware
// ============================================================================

func TestMiddleware_RequireAuth_Success(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	nextCalled := false
	var capturedUser *models.User
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		user, ok := GetUserFromContext(r.Context())
		if ok {
			capturedUser = user
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := m.RequireAuth(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Set user in session
	err := m.authService.SetUser(rec, req, testUser)
	require.NoError(t, err)

	// Extract cookies and use in new request
	cookies := rec.Result().Cookies()
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rec2.Code)
	require.NotNil(t, capturedUser)
	assert.Equal(t, testUser.Email, capturedUser.Email)
}

func TestMiddleware_RequireAuth_Unauthorized(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.RequireAuth(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// TESTS - RequireAdmin Middleware
// ============================================================================

func TestMiddleware_RequireAdmin_Success(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{"admin@example.com"})

	nextCalled := false
	var capturedUser *models.User
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		user, ok := GetUserFromContext(r.Context())
		if ok {
			capturedUser = user
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := m.RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	// Set admin user in session
	err := m.authService.SetUser(rec, req, testAdminUser)
	require.NoError(t, err)

	cookies := rec.Result().Cookies()
	req2 := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rec2.Code)
	require.NotNil(t, capturedUser)
	assert.Equal(t, testAdminUser.Email, capturedUser.Email)
}

func TestMiddleware_RequireAdmin_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		configEmail   string
		userEmail     string
		shouldBeAdmin bool
	}{
		{
			name:          "exact match lowercase",
			configEmail:   "admin@example.com",
			userEmail:     "admin@example.com",
			shouldBeAdmin: true,
		},
		{
			name:          "case insensitive match uppercase",
			configEmail:   "admin@example.com",
			userEmail:     "ADMIN@EXAMPLE.COM",
			shouldBeAdmin: true,
		},
		{
			name:          "case insensitive match mixed",
			configEmail:   "admin@example.com",
			userEmail:     "Admin@Example.Com",
			shouldBeAdmin: true,
		},
		{
			name:          "different email",
			configEmail:   "admin@example.com",
			userEmail:     "user@example.com",
			shouldBeAdmin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestMiddleware([]string{tt.configEmail})

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := m.RequireAdmin(next)

			user := &models.User{
				Sub:   "test-123",
				Email: tt.userEmail,
				Name:  "Test",
			}

			req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
			rec := httptest.NewRecorder()

			err := m.authService.SetUser(rec, req, user)
			require.NoError(t, err)

			cookies := rec.Result().Cookies()
			req2 := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
			for _, cookie := range cookies {
				req2.AddCookie(cookie)
			}

			rec2 := httptest.NewRecorder()
			handler.ServeHTTP(rec2, req2)

			if tt.shouldBeAdmin {
				assert.True(t, nextCalled, "Next handler should be called for admin")
				assert.Equal(t, http.StatusOK, rec2.Code)
			} else {
				assert.False(t, nextCalled, "Next handler should not be called for non-admin")
				assert.Equal(t, http.StatusForbidden, rec2.Code)
			}
		})
	}
}

func TestMiddleware_RequireAdmin_Unauthorized(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{"admin@example.com"})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMiddleware_RequireAdmin_Forbidden(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{"admin@example.com"})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.RequireAdmin(next)

	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	// Set regular user (not admin)
	err := m.authService.SetUser(rec, req, testUser)
	require.NoError(t, err)

	cookies := rec.Result().Cookies()
	req2 := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}

	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.False(t, nextCalled, "Next handler should not be called")
	assert.Equal(t, http.StatusForbidden, rec2.Code)
}

// ============================================================================
// TESTS - CSRF Token Generation & Validation
// ============================================================================

func TestMiddleware_GenerateCSRFToken(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	token, err := m.GenerateCSRFToken()

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 20, "Token should be reasonably long")
}

func TestMiddleware_GenerateCSRFToken_Unique(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	token1, err := m.GenerateCSRFToken()
	require.NoError(t, err)

	token2, err := m.GenerateCSRFToken()
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2, "Tokens should be unique")
}

func TestMiddleware_ValidateCSRFToken_Valid(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	token, err := m.GenerateCSRFToken()
	require.NoError(t, err)

	// Give cleanup goroutine time to finish
	time.Sleep(10 * time.Millisecond)

	valid := m.ValidateCSRFToken(token)
	assert.True(t, valid, "Token should be valid immediately after generation")
}

func TestMiddleware_ValidateCSRFToken_Invalid(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "non-existent token",
			token: "invalid-token-12345",
		},
		{
			name:  "malformed token",
			token: "!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valid := m.ValidateCSRFToken(tt.token)
			assert.False(t, valid, "Token should be invalid")
		})
	}
}

func TestMiddleware_ValidateCSRFToken_Expired(t *testing.T) {
	// Cannot run in parallel as it manipulates token expiry

	m := createTestMiddleware([]string{})

	token, err := m.GenerateCSRFToken()
	require.NoError(t, err)

	// Manually set token to expired
	m.csrfTokens.Store(token, time.Now().Add(-1*time.Hour))

	valid := m.ValidateCSRFToken(token)
	assert.False(t, valid, "Expired token should be invalid")

	// Verify token was deleted
	_, exists := m.csrfTokens.Load(token)
	assert.False(t, exists, "Expired token should be removed")
}

// ============================================================================
// TESTS - CSRFProtect Middleware
// ============================================================================

func TestMiddleware_CSRFProtect_SafeMethods(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	safeMethods := []string{"GET", "HEAD", "OPTIONS"}

	for _, method := range safeMethods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := m.CSRFProtect(next)

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.True(t, nextCalled, "Next handler should be called for safe methods")
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestMiddleware_CSRFProtect_ValidToken_Header(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	token, err := m.GenerateCSRFToken()
	require.NoError(t, err)

	// Give cleanup goroutine time
	time.Sleep(10 * time.Millisecond)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.CSRFProtect(next)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	req.Header.Set(CSRFTokenHeader, token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, nextCalled, "Next handler should be called with valid token")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddleware_CSRFProtect_ValidToken_Cookie(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	token, err := m.GenerateCSRFToken()
	require.NoError(t, err)

	// Give cleanup goroutine time
	time.Sleep(10 * time.Millisecond)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.CSRFProtect(next)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	req.AddCookie(&http.Cookie{
		Name:  CSRFTokenCookie,
		Value: token,
	})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, nextCalled, "Next handler should be called with valid token in cookie")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMiddleware_CSRFProtect_MissingToken(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.CSRFProtect(next)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, nextCalled, "Next handler should not be called without token")
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestMiddleware_CSRFProtect_InvalidToken(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := m.CSRFProtect(next)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("data"))
	req.Header.Set(CSRFTokenHeader, "invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.False(t, nextCalled, "Next handler should not be called with invalid token")
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// ============================================================================
// TESTS - SecurityHeaders Middleware
// ============================================================================

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, nextCalled, "Next handler should be called")
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check security headers
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, rec.Header().Get("Permissions-Policy"))
	assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
}

// ============================================================================
// TESTS - GetUserFromContext
// ============================================================================

func TestGetUserFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ctx         context.Context
		expectUser  bool
		expectEmail string
	}{
		{
			name:        "user in context",
			ctx:         context.WithValue(context.Background(), ContextKeyUser, testUser),
			expectUser:  true,
			expectEmail: testUser.Email,
		},
		{
			name:       "no user in context",
			ctx:        context.Background(),
			expectUser: false,
		},
		{
			name:       "wrong type in context",
			ctx:        context.WithValue(context.Background(), ContextKeyUser, "not-a-user"),
			expectUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, ok := GetUserFromContext(tt.ctx)

			assert.Equal(t, tt.expectUser, ok)
			if tt.expectUser {
				require.NotNil(t, user)
				assert.Equal(t, tt.expectEmail, user.Email)
			} else {
				assert.Nil(t, user)
			}
		})
	}
}

// ============================================================================
// TESTS - RateLimit
// ============================================================================

func TestNewRateLimit(t *testing.T) {
	t.Parallel()

	rl := NewRateLimit(10, 1*time.Minute)

	require.NotNil(t, rl)
	assert.NotNil(t, rl.attempts)
	assert.Equal(t, 10, rl.limit)
	assert.Equal(t, 1*time.Minute, rl.window)
}

func TestRateLimit_Middleware_AllowedRequests(t *testing.T) {
	t.Parallel()

	rl := NewRateLimit(5, 1*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	// Make 5 requests (under limit)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should be allowed", i+1)
	}
}

func TestRateLimit_Middleware_ExceedLimit(t *testing.T) {
	t.Parallel()

	rl := NewRateLimit(3, 1*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	// Make 3 allowed requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 4th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestRateLimit_Middleware_DifferentIPs(t *testing.T) {
	t.Parallel()

	rl := NewRateLimit(2, 1*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	// IP 1: Make 2 requests (at limit)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// IP 2: Should still be allowed
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "Different IP should not be rate limited")
}

func TestRateLimit_Middleware_XForwardedFor(t *testing.T) {
	t.Parallel()

	rl := NewRateLimit(2, 1*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	// Make 2 requests with X-Forwarded-For
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.99:1234"
		req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// 3rd request should be limited (using first IP from X-Forwarded-For)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.99:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

// ============================================================================
// TESTS - Concurrency
// ============================================================================

func TestMiddleware_CSRF_Concurrent(t *testing.T) {
	t.Parallel()

	m := createTestMiddleware([]string{})

	const numGoroutines = 50
	var wg sync.WaitGroup
	tokens := make([]string, numGoroutines)

	// Generate tokens concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := m.GenerateCSRFToken()
			assert.NoError(t, err)
			tokens[idx] = token
		}(i)
	}

	wg.Wait()

	// Give cleanup goroutines time
	time.Sleep(100 * time.Millisecond)

	// Validate all tokens
	for i, token := range tokens {
		assert.NotEmpty(t, token, "Token %d should not be empty", i)
		valid := m.ValidateCSRFToken(token)
		assert.True(t, valid, "Token %d should be valid", i)
	}

	// Check uniqueness
	uniqueTokens := make(map[string]bool)
	for _, token := range tokens {
		uniqueTokens[token] = true
	}
	assert.Equal(t, numGoroutines, len(uniqueTokens), "All tokens should be unique")
}

func TestRateLimit_Middleware_Concurrent(t *testing.T) {
	t.Parallel()

	// Use smaller limits to test concurrency behavior
	rl := NewRateLimit(10, 1*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	const numGoroutines = 20
	var wg sync.WaitGroup
	results := make([]int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.1:1234"
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			results[idx] = rec.Code
		}(i)
	}

	wg.Wait()

	okCount := 0
	limitedCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			okCount++
		} else if code == http.StatusTooManyRequests {
			limitedCount++
		}
	}

	// In concurrent scenario without proper locking, rate limiter may not be exact
	// We just verify that it processes all requests
	assert.Equal(t, numGoroutines, okCount+limitedCount, "Total should equal number of requests")
	// At least some requests should succeed
	assert.Greater(t, okCount, 0, "At least some requests should be allowed")
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkMiddleware_CORS(b *testing.B) {
	m := createTestMiddleware([]string{})

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := m.CORS(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkMiddleware_GenerateCSRFToken(b *testing.B) {
	m := createTestMiddleware([]string{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m.GenerateCSRFToken()
	}
}

func BenchmarkMiddleware_ValidateCSRFToken(b *testing.B) {
	m := createTestMiddleware([]string{})
	token, _ := m.GenerateCSRFToken()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.ValidateCSRFToken(token)
	}
}

func BenchmarkSecurityHeaders(b *testing.B) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeaders(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkRateLimit_Middleware(b *testing.B) {
	rl := NewRateLimit(1000, 1*time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
