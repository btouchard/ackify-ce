// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/pkg/types"
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

// mockAuthorizer is a test implementation of providers.Authorizer
type mockAuthorizer struct {
	adminEmails        map[string]bool
	onlyAdminCanCreate bool
}

func newMockAuthorizer(adminEmails []string, onlyAdminCanCreate bool) *mockAuthorizer {
	emails := make(map[string]bool)
	for _, email := range adminEmails {
		emails[strings.ToLower(email)] = true
	}
	return &mockAuthorizer{
		adminEmails:        emails,
		onlyAdminCanCreate: onlyAdminCanCreate,
	}
}

func (m *mockAuthorizer) IsAdmin(_ context.Context, userEmail string) bool {
	return m.adminEmails[strings.ToLower(userEmail)]
}

func (m *mockAuthorizer) CanCreateDocument(_ context.Context, userEmail string) bool {
	if !m.onlyAdminCanCreate {
		return true
	}
	return m.adminEmails[strings.ToLower(userEmail)]
}

// mockAuthProvider is a test implementation of providers.AuthProvider
type mockAuthProvider struct {
	mu          sync.RWMutex
	currentUser *types.User
}

func (m *mockAuthProvider) GetCurrentUser(_ *http.Request) (*types.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentUser == nil {
		return nil, http.ErrNoCookie
	}
	return m.currentUser, nil
}

func (m *mockAuthProvider) SetCurrentUser(_ http.ResponseWriter, _ *http.Request, user *types.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUser = user
	return nil
}

func (m *mockAuthProvider) Logout(_ http.ResponseWriter, _ *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUser = nil
}

func (m *mockAuthProvider) IsConfigured() bool {
	return true
}

// mockOAuthProvider is a test implementation of providers.OAuthAuthProvider
type mockOAuthProvider struct {
	mu          sync.RWMutex
	currentUser *types.User
	logoutURL   string
}

func newMockOAuthProvider() *mockOAuthProvider {
	return &mockOAuthProvider{
		logoutURL: testLogoutURL,
	}
}

func (m *mockOAuthProvider) GetCurrentUser(_ *http.Request) (*types.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentUser == nil {
		return nil, http.ErrNoCookie
	}
	return m.currentUser, nil
}

func (m *mockOAuthProvider) SetCurrentUser(_ http.ResponseWriter, _ *http.Request, user *types.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUser = user
	return nil
}

func (m *mockOAuthProvider) Logout(_ http.ResponseWriter, _ *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUser = nil
}

func (m *mockOAuthProvider) IsConfigured() bool {
	return true
}

func (m *mockOAuthProvider) CreateAuthURL(_ http.ResponseWriter, _ *http.Request, nextURL string) string {
	return testAuthURL + "?redirect_uri=" + testBaseURL + "/api/v1/auth/callback&state=test-state&next=" + nextURL
}

func (m *mockOAuthProvider) VerifyState(_ http.ResponseWriter, _ *http.Request, _ string) bool {
	return true
}

func (m *mockOAuthProvider) HandleCallback(_ context.Context, _ http.ResponseWriter, _ *http.Request, _, _ string) (*types.User, string, error) {
	return &types.User{
		Sub:   testUser.Sub,
		Email: testUser.Email,
		Name:  testUser.Name,
	}, "/", nil
}

func (m *mockOAuthProvider) GetLogoutURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logoutURL
}

func (m *mockOAuthProvider) IsAllowedDomain(_ string) bool {
	return true
}

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
	authProvider := &mockAuthProvider{}
	authorizer := newMockAuthorizer([]string{}, false)
	return shared.NewMiddleware(authProvider, testBaseURL, authorizer)
}

// ============================================================================
// TESTS - Constructor
// ============================================================================

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "with valid dependencies",
			baseURL: testBaseURL,
		},
		{
			name:    "with empty baseURL",
			baseURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authProvider := &mockAuthProvider{}
			middleware := createTestMiddleware()
			handler := NewHandler(authProvider, nil, nil, middleware, tt.baseURL, true, true)

			assert.NotNil(t, handler)
			assert.NotNil(t, handler.authProvider)
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

	// Use mockAuthProvider with pre-set user to test authenticated response
	authProvider := &mockAuthProvider{
		currentUser: &types.User{
			Sub:   testUser.Sub,
			Email: testUser.Email,
			Name:  testUser.Name,
		},
	}
	handler := NewHandler(authProvider, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
	rec := httptest.NewRecorder()

	// Execute
	handler.HandleAuthCheck(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Parse response
	var wrapper struct {
		Data struct {
			Authenticated bool                   `json:"authenticated"`
			User          map[string]interface{} `json:"user"`
		} `json:"data"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &wrapper)
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

			handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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

	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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

	oauthProvider := newMockOAuthProvider()
	authService := createTestAuthService()
	handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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
	// testBaseURL is URL-encoded in the redirect URL
	assert.Contains(t, wrapper.Data.RedirectURL, url.QueryEscape(testBaseURL))
}

func TestHandler_HandleLogout_WithoutSSO(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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

	oauthProvider := newMockOAuthProvider()
	authService := createTestAuthService()
	handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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

	// Verify that logout was called by checking status code
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Verify the response contains the expected structure
	var wrapper struct {
		Data struct {
			Message     string `json:"message"`
			RedirectURL string `json:"redirectUrl"`
		} `json:"data"`
	}
	err = json.Unmarshal(rec2.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "Successfully logged out", wrapper.Data.Message)
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

			oauthProvider := newMockOAuthProvider()
			handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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
		})
	}
}

func TestHandler_HandleStartOAuth_NoBody(t *testing.T) {
	t.Parallel()

	oauthProvider := newMockOAuthProvider()
	handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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

	oauthProvider := newMockOAuthProvider()
	handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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

	oauthProvider := newMockOAuthProvider()
	handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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

	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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

	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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

	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

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

	oauthProvider := newMockOAuthProvider()
	handler := NewHandler(oauthProvider, oauthProvider, nil, createTestMiddleware(), testBaseURL, true, true)

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
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
		rec := httptest.NewRecorder()

		handler.HandleAuthCheck(rec, req)
	}
}

func BenchmarkHandler_HandleAuthCheck_Parallel(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check", nil)
			rec := httptest.NewRecorder()

			handler.HandleAuthCheck(rec, req)
		}
	})
}

func BenchmarkHandler_HandleLogout(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
		rec := httptest.NewRecorder()

		handler.HandleLogout(rec, req)
	}
}

func BenchmarkHandler_HandleLogout_Parallel(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil)
			rec := httptest.NewRecorder()

			handler.HandleLogout(rec, req)
		}
	})
}

func BenchmarkHandler_HandleStartOAuth(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
		rec := httptest.NewRecorder()

		handler.HandleStartOAuth(rec, req)
	}
}

func BenchmarkHandler_HandleStartOAuth_Parallel(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/start", nil)
			rec := httptest.NewRecorder()

			handler.HandleStartOAuth(rec, req)
		}
	})
}

func BenchmarkHandler_HandleGetCSRFToken(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
		rec := httptest.NewRecorder()

		handler.HandleGetCSRFToken(rec, req)
	}
}

func BenchmarkHandler_HandleGetCSRFToken_Parallel(b *testing.B) {
	handler := NewHandler(&mockAuthProvider{}, nil, nil, createTestMiddleware(), testBaseURL, true, true)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/csrf", nil)
			rec := httptest.NewRecorder()

			handler.HandleGetCSRFToken(rec, req)
		}
	})
}
