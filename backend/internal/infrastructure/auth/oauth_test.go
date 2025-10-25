// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
)

func TestNewOAuthService(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "complete config",
			config: Config{
				BaseURL:       "https://ackify.example.com",
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				AuthURL:       "https://provider.com/auth",
				TokenURL:      "https://provider.com/token",
				UserInfoURL:   "https://provider.com/userinfo",
				Scopes:        []string{"openid", "email", "profile"},
				AllowedDomain: "@example.com",
				CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
				SecureCookies: true,
			},
		},
		{
			name: "minimal config",
			config: Config{
				BaseURL:      "http://localhost:8080",
				ClientID:     "minimal-client",
				ClientSecret: "minimal-secret",
				AuthURL:      "https://auth.com/oauth",
				TokenURL:     "https://auth.com/token",
				UserInfoURL:  "https://api.com/user",
				Scopes:       []string{"user"},
				CookieSecret: []byte("test-secret"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOAuthService(tt.config)

			if service == nil {
				t.Fatal("NewOAuthService() returned nil")
			}

			// Test that OAuth config is properly initialized
			if service.oauthConfig == nil {
				t.Error("OAuth config should not be nil")
			}
			if service.oauthConfig.ClientID != tt.config.ClientID {
				t.Errorf("ClientID = %v, expected %v", service.oauthConfig.ClientID, tt.config.ClientID)
			}
			if service.oauthConfig.ClientSecret != tt.config.ClientSecret {
				t.Errorf("ClientSecret = %v, expected %v", service.oauthConfig.ClientSecret, tt.config.ClientSecret)
			}

			expectedRedirectURL := tt.config.BaseURL + "/api/v1/auth/callback"
			if service.oauthConfig.RedirectURL != expectedRedirectURL {
				t.Errorf("RedirectURL = %v, expected %v", service.oauthConfig.RedirectURL, expectedRedirectURL)
			}

			if len(service.oauthConfig.Scopes) != len(tt.config.Scopes) {
				t.Errorf("Scopes length = %v, expected %v", len(service.oauthConfig.Scopes), len(tt.config.Scopes))
			}

			if service.oauthConfig.Endpoint.AuthURL != tt.config.AuthURL {
				t.Errorf("AuthURL = %v, expected %v", service.oauthConfig.Endpoint.AuthURL, tt.config.AuthURL)
			}
			if service.oauthConfig.Endpoint.TokenURL != tt.config.TokenURL {
				t.Errorf("TokenURL = %v, expected %v", service.oauthConfig.Endpoint.TokenURL, tt.config.TokenURL)
			}

			// Test service fields
			if service.userInfoURL != tt.config.UserInfoURL {
				t.Errorf("userInfoURL = %v, expected %v", service.userInfoURL, tt.config.UserInfoURL)
			}
			if service.allowedDomain != tt.config.AllowedDomain {
				t.Errorf("allowedDomain = %v, expected %v", service.allowedDomain, tt.config.AllowedDomain)
			}
			if service.secureCookies != tt.config.SecureCookies {
				t.Errorf("secureCookies = %v, expected %v", service.secureCookies, tt.config.SecureCookies)
			}

			// Test session store
			if service.sessionStore == nil {
				t.Error("Session store should not be nil")
			}
		})
	}
}

func TestOauthService_GetUser(t *testing.T) {
	service := createTestService()

	tests := []struct {
		name          string
		setupSession  func(*httptest.ResponseRecorder, *http.Request)
		expectError   bool
		expectedError error
		expectedUser  *models.User
	}{
		{
			name: "valid user session",
			setupSession: func(w *httptest.ResponseRecorder, r *http.Request) {
				user := &models.User{
					Sub:   "test-sub",
					Email: "test@example.com",
					Name:  "Test User",
				}
				err := service.SetUser(w, r, user)
				if err != nil {
					t.Fatalf("Failed to set user: %v", err)
				}
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "test-sub",
				Email: "test@example.com",
				Name:  "Test User",
			},
		},
		{
			name:          "no session",
			setupSession:  func(w *httptest.ResponseRecorder, r *http.Request) {},
			expectError:   true,
			expectedError: models.ErrUnauthorized,
		},
		{
			name: "invalid JSON in session",
			setupSession: func(w *httptest.ResponseRecorder, r *http.Request) {
				session, _ := service.sessionStore.Get(r, sessionName)
				session.Values["user"] = "invalid-json"
				session.Save(r, w)
			},
			expectError: true,
		},
		{
			name: "empty user value in session",
			setupSession: func(w *httptest.ResponseRecorder, r *http.Request) {
				session, _ := service.sessionStore.Get(r, sessionName)
				session.Values["user"] = ""
				session.Save(r, w)
			},
			expectError:   true,
			expectedError: models.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			// Setup session if needed
			tt.setupSession(w, r)

			// Add cookies from the setup response to the request
			if len(w.Result().Cookies()) > 0 {
				for _, cookie := range w.Result().Cookies() {
					r.AddCookie(cookie)
				}
			}

			user, err := service.GetUser(r)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.expectedError != nil && err != tt.expectedError {
					t.Errorf("Error = %v, expected %v", err, tt.expectedError)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("User should not be nil")
				return
			}

			if user.Sub != tt.expectedUser.Sub {
				t.Errorf("User.Sub = %v, expected %v", user.Sub, tt.expectedUser.Sub)
			}
			if user.Email != tt.expectedUser.Email {
				t.Errorf("User.Email = %v, expected %v", user.Email, tt.expectedUser.Email)
			}
			if user.Name != tt.expectedUser.Name {
				t.Errorf("User.Name = %v, expected %v", user.Name, tt.expectedUser.Name)
			}
		})
	}
}

func TestOauthService_SetUser(t *testing.T) {
	tests := []struct {
		name        string
		service     *OauthService
		user        *models.User
		expectError bool
	}{
		{
			name:    "valid user with secure cookies",
			service: createTestServiceWithSecure(true),
			user: &models.User{
				Sub:   "test-sub",
				Email: "test@example.com",
				Name:  "Test User",
			},
			expectError: false,
		},
		{
			name:    "valid user without secure cookies",
			service: createTestServiceWithSecure(false),
			user: &models.User{
				Sub:   "github|123",
				Email: "user@github.com",
				Name:  "GitHub User",
			},
			expectError: false,
		},
		{
			name:    "user with special characters",
			service: createTestService(),
			user: &models.User{
				Sub:   "google-oauth2|123456789",
				Email: "user+test@example.com",
				Name:  "Üser Námé",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			err := tt.service.SetUser(w, r, tt.user)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify that user can be retrieved
			for _, cookie := range w.Result().Cookies() {
				r.AddCookie(cookie)
			}

			retrievedUser, err := tt.service.GetUser(r)
			if err != nil {
				t.Errorf("Failed to retrieve user: %v", err)
				return
			}

			if retrievedUser.Sub != tt.user.Sub {
				t.Errorf("Retrieved user Sub = %v, expected %v", retrievedUser.Sub, tt.user.Sub)
			}
			if retrievedUser.Email != tt.user.Email {
				t.Errorf("Retrieved user Email = %v, expected %v", retrievedUser.Email, tt.user.Email)
			}
			if retrievedUser.Name != tt.user.Name {
				t.Errorf("Retrieved user Name = %v, expected %v", retrievedUser.Name, tt.user.Name)
			}

			// Verify cookie properties
			cookies := w.Result().Cookies()
			if len(cookies) == 0 {
				t.Error("No cookies set")
				return
			}

			sessionCookie := cookies[0]
			if sessionCookie.HttpOnly != true {
				t.Error("Cookie should be HttpOnly")
			}
			if sessionCookie.Secure != tt.service.secureCookies {
				t.Errorf("Cookie Secure = %v, expected %v", sessionCookie.Secure, tt.service.secureCookies)
			}
			if sessionCookie.SameSite != http.SameSiteLaxMode {
				t.Errorf("Cookie SameSite = %v, expected %v", sessionCookie.SameSite, http.SameSiteLaxMode)
			}
		})
	}
}

func TestOauthService_Logout(t *testing.T) {
	service := createTestService()

	// First, set a user
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	user := &models.User{
		Sub:   "test-sub",
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := service.SetUser(w, r, user)
	if err != nil {
		t.Fatalf("Failed to set user: %v", err)
	}

	// Add cookies to request
	for _, cookie := range w.Result().Cookies() {
		r.AddCookie(cookie)
	}

	// Verify user exists
	retrievedUser, err := service.GetUser(r)
	if err != nil {
		t.Fatalf("Failed to get user before logout: %v", err)
	}
	if retrievedUser == nil {
		t.Fatal("User should exist before logout")
	}

	// Logout
	w2 := httptest.NewRecorder()
	service.Logout(w2, r)

	// Verify logout cookie has MaxAge = -1
	cookies := w2.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("No logout cookies set")
		return
	}

	logoutCookie := cookies[0]
	if logoutCookie.MaxAge != -1 {
		t.Errorf("Logout cookie MaxAge = %v, expected -1", logoutCookie.MaxAge)
	}

	// Test that logout doesn't fail even with no session
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/", nil)
	service.Logout(w3, r3) // Should not panic
}

func TestOauthService_GetAuthURL(t *testing.T) {
	service := createTestService()

	tests := []struct {
		name    string
		nextURL string
	}{
		{
			name:    "root next URL",
			nextURL: "/",
		},
		{
			name:    "specific page next URL",
			nextURL: "/sign?doc=test-doc",
		},
		{
			name:    "empty next URL",
			nextURL: "",
		},
		{
			name:    "complex next URL with parameters",
			nextURL: "/sign?doc=test-doc&referrer=github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authURL := service.GetAuthURL(tt.nextURL)

			if authURL == "" {
				t.Error("Auth URL should not be empty")
				return
			}

			// Parse the URL to verify it's valid
			parsedURL, err := url.Parse(authURL)
			if err != nil {
				t.Errorf("Invalid auth URL: %v", err)
				return
			}

			// Verify it contains the expected OAuth parameters
			query := parsedURL.Query()
			if query.Get("client_id") != "test-client-id" {
				t.Errorf("client_id = %v, expected test-client-id", query.Get("client_id"))
			}
			if query.Get("response_type") != "code" {
				t.Errorf("response_type = %v, expected code", query.Get("response_type"))
			}
			if query.Get("redirect_uri") == "" {
				t.Error("redirect_uri should not be empty")
			}
			if query.Get("scope") == "" {
				t.Error("scope should not be empty")
			}
			if query.Get("state") == "" {
				t.Error("state should not be empty")
			}
			if query.Get("prompt") != "select_account" {
				t.Errorf("prompt = %v, expected select_account", query.Get("prompt"))
			}

			// Verify state contains the next URL (basic check)
			state := query.Get("state")
			if !strings.Contains(state, ":") {
				t.Error("State should contain ':' separator")
			}
		})
	}
}

func TestOauthService_IsAllowedDomain(t *testing.T) {
	tests := []struct {
		name          string
		allowedDomain string
		email         string
		expected      bool
	}{
		{
			name:          "no domain restriction",
			allowedDomain: "",
			email:         "user@anywhere.com",
			expected:      true,
		},
		{
			name:          "matching domain",
			allowedDomain: "example.com",
			email:         "user@example.com",
			expected:      true,
		},
		{
			name:          "non-matching domain",
			allowedDomain: "example.com",
			email:         "user@other.com",
			expected:      false,
		},
		{
			name:          "case insensitive matching",
			allowedDomain: "EXAMPLE.COM",
			email:         "user@example.com",
			expected:      true,
		},
		{
			name:          "case insensitive email",
			allowedDomain: "example.com",
			email:         "USER@EXAMPLE.COM",
			expected:      true,
		},
		{
			name:          "subdomain not allowed",
			allowedDomain: "example.com",
			email:         "user@sub.example.com",
			expected:      false,
		},
		{
			name:          "partial domain match not allowed",
			allowedDomain: "example.com",
			email:         "user@notexample.com",
			expected:      false,
		},
		{
			name:          "domain without @ prefix",
			allowedDomain: "example.com",
			email:         "user@example.com",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &OauthService{
				allowedDomain: tt.allowedDomain,
			}

			result := service.IsAllowedDomain(tt.email)
			if result != tt.expected {
				t.Errorf("IsAllowedDomain() = %v, expected %v for email %s with domain %s",
					result, tt.expected, tt.email, tt.allowedDomain)
			}
		})
	}
}

func TestOauthService_parseUserInfo(t *testing.T) {
	service := createTestService()

	tests := []struct {
		name         string
		responseBody map[string]interface{}
		expectError  bool
		expectedUser *models.User
	}{
		{
			name: "Google OAuth response",
			responseBody: map[string]interface{}{
				"sub":   "google-oauth2|123456789",
				"email": "test@example.com",
				"name":  "Test User",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "google-oauth2|123456789",
				Email: "test@example.com",
				Name:  "Test User",
			},
		},
		{
			name: "GitHub OAuth response",
			responseBody: map[string]interface{}{
				"id":    float64(12345), // JSON numbers become float64
				"email": "user@github.com",
				"name":  "GitHub User",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "12345",
				Email: "user@github.com",
				Name:  "GitHub User",
			},
		},
		{
			name: "GitLab OAuth response",
			responseBody: map[string]interface{}{
				"id":                 float64(987),
				"email":              "user@gitlab.com",
				"preferred_username": "gitlabuser",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "987",
				Email: "user@gitlab.com",
				Name:  "gitlabuser",
			},
		},
		{
			name: "OAuth with first/last names",
			responseBody: map[string]interface{}{
				"sub":         "oauth2|12345",
				"email":       "user@example.com",
				"given_name":  "John",
				"family_name": "Doe",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "oauth2|12345",
				Email: "user@example.com",
				Name:  "John Doe",
			},
		},
		{
			name: "OAuth with only first name",
			responseBody: map[string]interface{}{
				"sub":        "oauth2|12345",
				"email":      "user@example.com",
				"given_name": "John",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "oauth2|12345",
				Email: "user@example.com",
				Name:  "John",
			},
		},
		{
			name: "OAuth with CN field",
			responseBody: map[string]interface{}{
				"sub":   "ldap|12345",
				"email": "user@company.com",
				"cn":    "Common Name",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "ldap|12345",
				Email: "user@company.com",
				Name:  "Common Name",
			},
		},
		{
			name: "OAuth with display_name",
			responseBody: map[string]interface{}{
				"sub":          "custom|12345",
				"email":        "user@custom.com",
				"display_name": "Display Name",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "custom|12345",
				Email: "user@custom.com",
				Name:  "Display Name",
			},
		},
		{
			name: "OAuth without name fields",
			responseBody: map[string]interface{}{
				"sub":   "minimal|12345",
				"email": "user@minimal.com",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "minimal|12345",
				Email: "user@minimal.com",
				Name:  "",
			},
		},
		{
			name: "missing sub and id",
			responseBody: map[string]interface{}{
				"email": "user@example.com",
				"name":  "Test User",
			},
			expectError: true,
		},
		{
			name: "missing email",
			responseBody: map[string]interface{}{
				"sub":  "test|12345",
				"name": "Test User",
			},
			expectError: true,
		},
		{
			name: "string ID",
			responseBody: map[string]interface{}{
				"id":    "string-id-123",
				"email": "user@example.com",
				"name":  "String ID User",
			},
			expectError: false,
			expectedUser: &models.User{
				Sub:   "string-id-123",
				Email: "user@example.com",
				Name:  "String ID User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP response
			jsonBody, _ := json.Marshal(tt.responseBody)
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(jsonBody)),
			}

			user, err := service.parseUserInfo(resp)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("User should not be nil")
				return
			}

			if user.Sub != tt.expectedUser.Sub {
				t.Errorf("User.Sub = %v, expected %v", user.Sub, tt.expectedUser.Sub)
			}
			if user.Email != tt.expectedUser.Email {
				t.Errorf("User.Email = %v, expected %v", user.Email, tt.expectedUser.Email)
			}
			if user.Name != tt.expectedUser.Name {
				t.Errorf("User.Name = %v, expected %v", user.Name, tt.expectedUser.Name)
			}
		})
	}
}

func TestOauthService_parseUserInfo_InvalidJSON(t *testing.T) {
	service := createTestService()

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}

	_, err := service.parseUserInfo(resp)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to decode user info") {
		t.Errorf("Error should mention decoding failure: %v", err)
	}
}

func TestOauthService_HandleCallback_StateDecoding(t *testing.T) {
	service := createTestService()

	tests := []struct {
		name        string
		state       string
		expectedURL string
	}{
		{
			name:        "valid state with next URL",
			state:       "randomstate:L3NpZ24_ZG9jPXRlc3Q", // base64 for "/sign?doc=test"
			expectedURL: "/sign?doc=test",
		},
		{
			name:        "state without separator",
			state:       "invalidstate",
			expectedURL: "/",
		},
		{
			name:        "state with invalid base64",
			state:       "randomstate:invalid-base64!",
			expectedURL: "/",
		},
		{
			name:        "empty state",
			state:       "",
			expectedURL: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the full HandleCallback without mocking OAuth2 exchange
			// So we test the state parsing logic by calling with invalid code
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			_, nextURL, _ := service.HandleCallback(context.Background(), w, r, "invalid-code", tt.state)

			if nextURL != tt.expectedURL {
				t.Errorf("NextURL = %v, expected %v", nextURL, tt.expectedURL)
			}
		})
	}
}

// Helper functions
func createTestService() *OauthService {
	return createTestServiceWithSecure(false)
}

func TestOauthService_HandleCallback_DomainRestriction(t *testing.T) {
	// Create service with domain restriction
	config := Config{
		BaseURL:       "https://test.example.com",
		ClientID:      "test-client-id",
		ClientSecret:  "test-client-secret",
		AuthURL:       "https://provider.com/auth",
		TokenURL:      "https://provider.com/token",
		UserInfoURL:   "https://provider.com/userinfo",
		Scopes:        []string{"openid", "email", "profile"},
		AllowedDomain: "example.com",
		CookieSecret:  []byte("test-secret-32-bytes-long-key!"),
		SecureCookies: false,
	}
	service := NewOAuthService(config)

	// Test with disallowed domain - this will fail during OAuth exchange
	// but we can test the domain check logic by calling IsAllowedDomain directly
	if service.IsAllowedDomain("user@other.com") {
		t.Error("Domain restriction should reject other.com emails")
	}
	if !service.IsAllowedDomain("user@example.com") {
		t.Error("Domain restriction should allow example.com emails")
	}
}

func TestOauthService_GetUser_SessionError(t *testing.T) {
	// Test with invalid cookie secret to trigger session errors
	config := Config{
		BaseURL:      "https://test.example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		AuthURL:      "https://provider.com/auth",
		TokenURL:     "https://provider.com/token",
		UserInfoURL:  "https://provider.com/userinfo",
		Scopes:       []string{"openid", "email"},
		CookieSecret: []byte("short"), // Too short, might cause issues
	}
	service := NewOAuthService(config)

	r := httptest.NewRequest("GET", "/", nil)
	// Add a malformed cookie to trigger session error
	r.AddCookie(&http.Cookie{
		Name:  sessionName,
		Value: "malformed-session-data",
	})

	_, err := service.GetUser(r)
	if err == nil {
		t.Error("Expected error with malformed session")
	}
}

func TestConfig_Structure(t *testing.T) {
	config := Config{
		BaseURL:       "https://ackify.example.com",
		ClientID:      "test-client-id",
		ClientSecret:  "test-client-secret",
		AuthURL:       "https://auth.provider.com/oauth/authorize",
		TokenURL:      "https://auth.provider.com/oauth/token",
		UserInfoURL:   "https://api.provider.com/user",
		Scopes:        []string{"openid", "email", "profile"},
		AllowedDomain: "example.com",
		CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
		SecureCookies: true,
	}

	// Test that config fields are accessible and correct
	if config.BaseURL != "https://ackify.example.com" {
		t.Errorf("BaseURL = %v, expected https://ackify.example.com", config.BaseURL)
	}
	if config.ClientID != "test-client-id" {
		t.Errorf("ClientID = %v, expected test-client-id", config.ClientID)
	}
	if len(config.Scopes) != 3 {
		t.Errorf("Scopes length = %v, expected 3", len(config.Scopes))
	}
	if !config.SecureCookies {
		t.Error("SecureCookies should be true")
	}
	if len(config.CookieSecret) == 0 {
		t.Error("CookieSecret should not be empty")
	}
}

func TestOauthService_SetUser_MarshalError(t *testing.T) {
	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Test with a user that would cause JSON marshal issues
	// In Go, it's hard to make json.Marshal fail with basic types
	// but we can test with a valid user to ensure the path works
	user := &models.User{
		Sub:   "test-sub",
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := service.SetUser(w, r, user)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify session was set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("No cookies set")
	}
}

func createTestServiceWithSecure(secure bool) *OauthService {
	config := Config{
		BaseURL:       "https://test.example.com",
		ClientID:      "test-client-id",
		ClientSecret:  "test-client-secret",
		AuthURL:       "https://provider.com/auth",
		TokenURL:      "https://provider.com/token",
		UserInfoURL:   "https://provider.com/userinfo",
		Scopes:        []string{"openid", "email", "profile"},
		AllowedDomain: "example.com",
		CookieSecret:  []byte("test-secret-32-bytes-long-key!"),
		SecureCookies: secure,
	}
	return NewOAuthService(config)
}

// ============================================================================
// TESTS - VerifyState
// ============================================================================

func TestOauthService_VerifyState_Success(t *testing.T) {
	t.Parallel()

	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// First, create a session with an oauth_state
	session, _ := service.sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = "test-state-token-123"
	_ = session.Save(r, w)

	// Get cookies from response
	cookies := w.Result().Cookies()
	r2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		r2.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	result := service.VerifyState(w2, r2, "test-state-token-123")

	if !result {
		t.Error("VerifyState should return true for matching state")
	}
}

func TestOauthService_VerifyState_Mismatch(t *testing.T) {
	t.Parallel()

	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set state in session
	session, _ := service.sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = "correct-state"
	_ = session.Save(r, w)

	cookies := w.Result().Cookies()
	r2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		r2.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	result := service.VerifyState(w2, r2, "wrong-state")

	if result {
		t.Error("VerifyState should return false for mismatched state")
	}
}

func TestOauthService_VerifyState_EmptyStored(t *testing.T) {
	t.Parallel()

	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Don't set any state in session (empty)
	result := service.VerifyState(w, r, "some-token")

	if result {
		t.Error("VerifyState should return false when stored state is empty")
	}
}

func TestOauthService_VerifyState_EmptyToken(t *testing.T) {
	t.Parallel()

	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set state in session
	session, _ := service.sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = "some-state"
	_ = session.Save(r, w)

	cookies := w.Result().Cookies()
	r2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		r2.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	result := service.VerifyState(w2, r2, "")

	if result {
		t.Error("VerifyState should return false when token is empty")
	}
}

func TestOauthService_VerifyState_BothEmpty(t *testing.T) {
	t.Parallel()

	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	result := service.VerifyState(w, r, "")

	if result {
		t.Error("VerifyState should return false when both are empty")
	}
}

// ============================================================================
// TESTS - subtleConstantTimeCompare
// ============================================================================

func TestSubtleConstantTimeCompare_Equal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
	}{
		{"identical strings", "hello", "hello"},
		{"identical long strings", "this-is-a-very-long-state-token-12345", "this-is-a-very-long-state-token-12345"},
		{"empty strings", "", ""},
		{"special characters", "abc!@#$%^&*()", "abc!@#$%^&*()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !subtleConstantTimeCompare(tt.a, tt.b) {
				t.Errorf("subtleConstantTimeCompare(%q, %q) should return true", tt.a, tt.b)
			}
		})
	}
}

func TestSubtleConstantTimeCompare_NotEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		b    string
	}{
		{"different strings", "hello", "world"},
		{"different lengths", "short", "longer-string"},
		{"one empty", "hello", ""},
		{"other empty", "", "world"},
		{"similar but different", "state123", "state124"},
		{"case sensitive", "Hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if subtleConstantTimeCompare(tt.a, tt.b) {
				t.Errorf("subtleConstantTimeCompare(%q, %q) should return false", tt.a, tt.b)
			}
		})
	}
}

func TestSubtleConstantTimeCompare_TimingSafety(t *testing.T) {
	t.Parallel()

	// Test that comparison takes similar time regardless of where difference occurs
	// This is a basic test - true timing attack resistance requires more sophisticated testing
	a := "this-is-a-long-state-token-with-many-characters"
	b1 := "Xhis-is-a-long-state-token-with-many-characters" // Differs at start
	b2 := "this-is-a-long-state-token-with-many-characterX" // Differs at end

	// Both should return false
	if subtleConstantTimeCompare(a, b1) {
		t.Error("Should return false for b1")
	}
	if subtleConstantTimeCompare(a, b2) {
		t.Error("Should return false for b2")
	}

	// The function should have similar behavior regardless of where the difference is
	// (This is ensured by the XOR loop that always iterates through entire string)
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkVerifyState(b *testing.B) {
	service := createTestService()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Setup session
	session, _ := service.sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = "test-state-token"
	_ = session.Save(r, w)

	cookies := w.Result().Cookies()
	r2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range cookies {
		r2.AddCookie(cookie)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		_ = service.VerifyState(w, r2, "test-state-token")
	}
}

func BenchmarkSubtleConstantTimeCompare(b *testing.B) {
	a := "this-is-a-state-token-123456789"
	b1 := "this-is-a-state-token-123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = subtleConstantTimeCompare(a, b1)
	}
}

func BenchmarkSubtleConstantTimeCompare_Different(b *testing.B) {
	a := "this-is-a-state-token-123456789"
	b1 := "this-is-a-state-token-987654321"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = subtleConstantTimeCompare(a, b1)
	}
}
