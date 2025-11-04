// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
)

// ============================================================================
// TEST FIXTURES
// ============================================================================

const (
	testBaseURL      = "https://example.com"
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
	testAuthURL      = "https://oauth.example.com/authorize"
	testTokenURL     = "https://oauth.example.com/token"
	testUserInfoURL  = "https://oauth.example.com/userinfo"
	testLogoutURL    = "https://oauth.example.com/logout"
)

var (
	testCookieSecret = securecookie.GenerateRandomKey(32)

	testUser = &models.User{
		Sub:   "oauth2|123456789",
		Email: "user@example.com",
		Name:  "Test User",
	}
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func createTestAuthService() *auth.OauthService {
	return auth.NewOAuthService(auth.Config{
		BaseURL:       testBaseURL,
		ClientID:      testClientID,
		ClientSecret:  testClientSecret,
		AuthURL:       testAuthURL,
		TokenURL:      testTokenURL,
		UserInfoURL:   testUserInfoURL,
		LogoutURL:     testLogoutURL,
		Scopes:        []string{"openid", "email", "profile"},
		AllowedDomain: "",
		CookieSecret:  testCookieSecret,
		SecureCookies: false, // false for testing (no HTTPS)
	})
}

func createTestMiddleware() *shared.Middleware {
	authService := createTestAuthService()
	return shared.NewMiddleware(authService, testBaseURL, []string{})
}

// ============================================================================
// TESTS - Constructor
// ============================================================================

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		authService *auth.OauthService
		middleware  *shared.Middleware
		baseURL     string
	}{
		{
			name:        "with valid dependencies",
			authService: createTestAuthService(),
			middleware:  createTestMiddleware(),
			baseURL:     testBaseURL,
		},
		{
			name:        "with empty baseURL",
			authService: createTestAuthService(),
			middleware:  createTestMiddleware(),
			baseURL:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(tt.authService, nil, tt.middleware, tt.baseURL, true, true)

			assert.NotNil(t, handler)
			assert.NotNil(t, handler.authService)
			assert.NotNil(t, handler.middleware)
			assert.Equal(t, tt.baseURL, handler.baseURL)
		})
	}
}

// ============================================================================
// TESTS - HandleAuthCheck
// ============================================================================

func TestHandler_HandleAuthCheck_Authenticated(t *testing.T) {
	t.Parallel()

	authService := createTestAuthService()
	handler := NewHandler(authService, nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
	rec := httptest.NewRecorder()

	// Set user in session
	err := authService.SetUser(rec, req, testUser)
	require.NoError(t, err)

	// Get the session cookie from the recorder
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies, "Session cookie should be set")

	// Create a new request with the session cookie
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	rec2 := httptest.NewRecorder()

	// Execute
	handler.HandleAuthCheck(rec2, req2)

	// Assert
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "application/json", rec2.Header().Get("Content-Type"))

	// Parse response
	var wrapper struct {
		Data struct {
			Authenticated bool                   `json:"authenticated"`
			User          map[string]interface{} `json:"user"`
		} `json:"data"`
	}
	err = json.Unmarshal(rec2.Body.Bytes(), &wrapper)
	require.NoError(t, err, "Response should be valid JSON")

	// Validate fields
	assert.True(t, wrapper.Data.Authenticated)
	assert.NotNil(t, wrapper.Data.User)
	assert.Equal(t, testUser.Sub, wrapper.Data.User["id"])
	assert.Equal(t, testUser.Email, wrapper.Data.User["email"])
	assert.Equal(t, testUser.Name, wrapper.Data.User["name"])
}

func TestHandler_HandleAuthCheck_NotAuthenticated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(*http.Request) *http.Request
	}{
		{
			name: "no session cookie",
			setupFunc: func(req *http.Request) *http.Request {
				return req // No modifications
			},
		},
		{
			name: "invalid session cookie",
			setupFunc: func(req *http.Request) *http.Request {
				req.AddCookie(&http.Cookie{
					Name:  "ackapp_session",
					Value: "invalid-cookie-value",
				})
				return req
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
			req = tt.setupFunc(req)
			rec := httptest.NewRecorder()

			// Execute
			handler.HandleAuthCheck(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			// Parse response
			var wrapper struct {
				Data struct {
					Authenticated bool `json:"authenticated"`
				} `json:"data"`
			}
			err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
			require.NoError(t, err)

			assert.False(t, wrapper.Data.Authenticated)
		})
	}
}

func TestHandler_HandleAuthCheck_ResponseFormat(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
	rec := httptest.NewRecorder()

	handler.HandleAuthCheck(rec, req)

	// Check Content-Type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)

	// Validate JSON structure
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check wrapper structure
	assert.Contains(t, response, "data")

	// Get data object
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")

	// Check required field
	assert.Contains(t, data, "authenticated")

	// Validate field type
	_, ok = data["authenticated"].(bool)
	assert.True(t, ok, "authenticated should be a boolean")
}

// ============================================================================
// TESTS - HandleLogout
// ============================================================================

func TestHandler_HandleLogout_WithSSO(t *testing.T) {
	t.Parallel()

	authService := createTestAuthService()
	handler := NewHandler(authService, nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	// Set user in session first
	err := authService.SetUser(rec, req, testUser)
	require.NoError(t, err)

	// Get the session cookie
	cookies := rec.Result().Cookies()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	rec2 := httptest.NewRecorder()

	// Execute logout
	handler.HandleLogout(rec2, req2)

	// Assert
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "application/json", rec2.Header().Get("Content-Type"))

	// Parse response
	var wrapper struct {
		Data struct {
			Message     string `json:"message"`
			RedirectURL string `json:"redirectUrl"`
		} `json:"data"`
	}
	err = json.Unmarshal(rec2.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	assert.Equal(t, "Successfully logged out", wrapper.Data.Message)
	assert.Contains(t, wrapper.Data.RedirectURL, testLogoutURL)
	assert.Contains(t, wrapper.Data.RedirectURL, "post_logout_redirect_uri")
	assert.Contains(t, wrapper.Data.RedirectURL, testBaseURL)
}

func TestHandler_HandleLogout_WithoutSSO(t *testing.T) {
	t.Parallel()

	// Create auth service without logout URL
	authService := auth.NewOAuthService(auth.Config{
		BaseURL:       testBaseURL,
		ClientID:      testClientID,
		ClientSecret:  testClientSecret,
		AuthURL:       testAuthURL,
		TokenURL:      testTokenURL,
		UserInfoURL:   testUserInfoURL,
		LogoutURL:     "", // No SSO logout
		Scopes:        []string{"openid", "email", "profile"},
		CookieSecret:  testCookieSecret,
		SecureCookies: false,
	})

	handler := NewHandler(authService, nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	// Execute
	handler.HandleLogout(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var wrapper struct {
		Data struct {
			Message     string `json:"message"`
			RedirectURL string `json:"redirectUrl,omitempty"`
		} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	assert.Equal(t, "Successfully logged out", wrapper.Data.Message)
	assert.Empty(t, wrapper.Data.RedirectURL)
}

func TestHandler_HandleLogout_ClearsSession(t *testing.T) {
	t.Parallel()

	authService := createTestAuthService()
	handler := NewHandler(authService, nil, createTestMiddleware(), testBaseURL, true, true)

	// Set user in session
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	err := authService.SetUser(rec, req, testUser)
	require.NoError(t, err)

	// Get the session cookie
	cookies := rec.Result().Cookies()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	rec2 := httptest.NewRecorder()

	// Execute logout
	handler.HandleLogout(rec2, req2)

	// Verify session is cleared by checking the Set-Cookie header
	setCookieHeaders := rec2.Header().Values("Set-Cookie")
	assert.NotEmpty(t, setCookieHeaders, "Should set cookie to clear session")

	// Check that MaxAge is negative (cookie deletion)
	foundMaxAge := false
	for _, setCookie := range setCookieHeaders {
		if strings.Contains(setCookie, "Max-Age") && strings.Contains(setCookie, "ackapp_session") {
			foundMaxAge = true
			// Should contain negative Max-Age or Max-Age=0
			assert.True(t, strings.Contains(setCookie, "Max-Age=-1") || strings.Contains(setCookie, "Max-Age=0"),
				"Cookie should be deleted with negative Max-Age")
		}
	}
	assert.True(t, foundMaxAge, "Should set Max-Age for session cookie")
}

// ============================================================================
// TESTS - HandleStartOAuth
// ============================================================================

func TestHandler_HandleStartOAuth_WithRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		requestBody map[string]string
		expectedURL string
	}{
		{
			name:        "with custom redirect path",
			requestBody: map[string]string{"redirectTo": "/dashboard"},
			expectedURL: "/dashboard",
		},
		{
			name:        "with root redirect",
			requestBody: map[string]string{"redirectTo": "/"},
			expectedURL: "/",
		},
		{
			name:        "with empty redirect",
			requestBody: map[string]string{"redirectTo": ""},
			expectedURL: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			handler.HandleStartOAuth(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			// Parse response
			var wrapper struct {
				Data struct {
					RedirectURL string `json:"redirectUrl"`
				} `json:"data"`
			}
			err = json.Unmarshal(rec.Body.Bytes(), &wrapper)
			require.NoError(t, err)

			// Validate redirect URL contains OAuth provider URL
			assert.NotEmpty(t, wrapper.Data.RedirectURL)
			assert.Contains(t, wrapper.Data.RedirectURL, testAuthURL)
			assert.Contains(t, wrapper.Data.RedirectURL, "client_id="+testClientID)
			assert.Contains(t, wrapper.Data.RedirectURL, "redirect_uri=")
			assert.Contains(t, wrapper.Data.RedirectURL, "state=")

			// Check that session cookie was set (for state verification)
			cookies := rec.Result().Cookies()
			assert.NotEmpty(t, cookies, "Session cookie should be set for OAuth state")
		})
	}
}

func TestHandler_HandleStartOAuth_NoBody(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
	rec := httptest.NewRecorder()

	// Execute
	handler.HandleStartOAuth(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var wrapper struct {
		Data struct {
			RedirectURL string `json:"redirectUrl"`
		} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	// Should default to root redirect
	assert.NotEmpty(t, wrapper.Data.RedirectURL)
	assert.Contains(t, wrapper.Data.RedirectURL, testAuthURL)
}

func TestHandler_HandleStartOAuth_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Execute
	handler.HandleStartOAuth(rec, req)

	// Assert - should still succeed and default to "/"
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var wrapper struct {
		Data struct {
			RedirectURL string `json:"redirectUrl"`
		} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	assert.NotEmpty(t, wrapper.Data.RedirectURL)
}

func TestHandler_HandleStartOAuth_ResponseFormat(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
	rec := httptest.NewRecorder()

	handler.HandleStartOAuth(rec, req)

	// Check Content-Type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)

	// Validate JSON structure
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check wrapper structure
	assert.Contains(t, response, "data")

	// Get data object
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")

	// Check required field
	assert.Contains(t, data, "redirectUrl")

	// Validate field type
	redirectURL, ok := data["redirectUrl"].(string)
	assert.True(t, ok, "redirectUrl should be a string")
	assert.NotEmpty(t, redirectURL)
}

// ============================================================================
// TESTS - HandleGetCSRFToken
// ============================================================================

func TestHandler_HandleGetCSRFToken_Success(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
	rec := httptest.NewRecorder()

	// Execute
	handler.HandleGetCSRFToken(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Parse response
	var wrapper struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
	require.NoError(t, err)

	// Validate token
	assert.NotEmpty(t, wrapper.Data.Token)
	assert.Greater(t, len(wrapper.Data.Token), 20, "CSRF token should be sufficiently long")

	// Check cookie was set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == shared.CSRFTokenCookie {
			csrfCookie = cookie
			break
		}
	}

	require.NotNil(t, csrfCookie, "CSRF cookie should be set")
	assert.Equal(t, wrapper.Data.Token, csrfCookie.Value)
	assert.Equal(t, "/", csrfCookie.Path)
	assert.False(t, csrfCookie.HttpOnly, "CSRF cookie should be readable by JS")
	assert.Equal(t, http.SameSiteLaxMode, csrfCookie.SameSite)
	assert.Equal(t, 86400, csrfCookie.MaxAge, "CSRF token should have 24h lifetime")
}

func TestHandler_HandleGetCSRFToken_ResponseFormat(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
	rec := httptest.NewRecorder()

	handler.HandleGetCSRFToken(rec, req)

	// Check Content-Type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)

	// Validate JSON structure
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check wrapper structure
	assert.Contains(t, response, "data")

	// Get data object
	data, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "data should be an object")

	// Check required field
	assert.Contains(t, data, "token")

	// Validate field type
	token, ok := data["token"].(string)
	assert.True(t, ok, "token should be a string")
	assert.NotEmpty(t, token)
}

// ============================================================================
// TESTS - Concurrency
// ============================================================================

func TestHandler_HandleAuthCheck_Concurrent(t *testing.T) {
	t.Parallel()

	authService := createTestAuthService()
	handler := NewHandler(authService, nil, createTestMiddleware(), testBaseURL, true, true)

	const numRequests = 100
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	// Spawn concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			defer func() { done <- true }()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
			rec := httptest.NewRecorder()

			// Half with session, half without
			if id%2 == 0 {
				err := authService.SetUser(rec, req, testUser)
				if err != nil {
					errors <- err
					return
				}
				cookies := rec.Result().Cookies()
				req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
				for _, cookie := range cookies {
					req2.AddCookie(cookie)
				}
				rec2 := httptest.NewRecorder()
				handler.HandleAuthCheck(rec2, req2)

				if rec2.Code != http.StatusOK {
					errors <- assert.AnError
				}
			} else {
				handler.HandleAuthCheck(rec, req)
				if rec.Code != http.StatusOK {
					errors <- assert.AnError
				}
			}
		}(i)
	}

	// Wait for all requests
	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errCount++
	}

	assert.Equal(t, 0, errCount, "All concurrent requests should succeed")
}

func TestHandler_HandleLogout_Concurrent(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	const numRequests = 100
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	// Spawn concurrent logout requests
	for i := 0; i < numRequests; i++ {
		go func() {
			defer func() { done <- true }()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
			rec := httptest.NewRecorder()

			handler.HandleLogout(rec, req)

			if rec.Code != http.StatusOK {
				errors <- assert.AnError
			}

			var wrapper struct {
				Data struct {
					Message string `json:"message"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &wrapper); err != nil {
				errors <- err
			}
		}()
	}

	// Wait for all requests
	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errCount++
	}

	assert.Equal(t, 0, errCount, "All concurrent requests should succeed")
}

func TestHandler_HandleStartOAuth_Concurrent(t *testing.T) {
	t.Parallel()

	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	const numRequests = 100
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	// Spawn concurrent OAuth start requests
	for i := 0; i < numRequests; i++ {
		go func() {
			defer func() { done <- true }()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
			rec := httptest.NewRecorder()

			handler.HandleStartOAuth(rec, req)

			if rec.Code != http.StatusOK {
				errors <- assert.AnError
			}

			var wrapper struct {
				Data struct {
					RedirectURL string `json:"redirectUrl"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &wrapper); err != nil {
				errors <- err
				return
			}

			if wrapper.Data.RedirectURL == "" {
				errors <- assert.AnError
			}
		}()
	}

	// Wait for all requests
	for i := 0; i < numRequests; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	var errCount int
	for err := range errors {
		t.Logf("Concurrent request error: %v", err)
		errCount++
	}

	assert.Equal(t, 0, errCount, "All concurrent requests should succeed")
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkHandler_HandleAuthCheck(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
		rec := httptest.NewRecorder()

		handler.HandleAuthCheck(rec, req)
	}
}

func BenchmarkHandler_HandleAuthCheck_Parallel(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
			rec := httptest.NewRecorder()

			handler.HandleAuthCheck(rec, req)
		}
	})
}

func BenchmarkHandler_HandleLogout(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
		rec := httptest.NewRecorder()

		handler.HandleLogout(rec, req)
	}
}

func BenchmarkHandler_HandleLogout_Parallel(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
			rec := httptest.NewRecorder()

			handler.HandleLogout(rec, req)
		}
	})
}

func BenchmarkHandler_HandleStartOAuth(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
		rec := httptest.NewRecorder()

		handler.HandleStartOAuth(rec, req)
	}
}

func BenchmarkHandler_HandleStartOAuth_Parallel(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
			rec := httptest.NewRecorder()

			handler.HandleStartOAuth(rec, req)
		}
	})
}

func BenchmarkHandler_HandleGetCSRFToken(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
		rec := httptest.NewRecorder()

		handler.HandleGetCSRFToken(rec, req)
	}
}

func BenchmarkHandler_HandleGetCSRFToken_Parallel(b *testing.B) {
	handler := NewHandler(createTestAuthService(), nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
			rec := httptest.NewRecorder()

			handler.HandleGetCSRFToken(rec, req)
		}
	})
}
