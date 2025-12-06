// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

func TestOAuthProvider_IsAllowedDomain(t *testing.T) {
	sessionSvc := NewSessionService(SessionServiceConfig{
		CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
		SecureCookies: false,
	})

	tests := []struct {
		name          string
		allowedDomain string
		email         string
		expected      bool
	}{
		{
			name:          "allowed domain match",
			allowedDomain: "@example.com",
			email:         "user@example.com",
			expected:      true,
		},
		{
			name:          "allowed domain mismatch",
			allowedDomain: "@example.com",
			email:         "user@other.com",
			expected:      false,
		},
		{
			name:          "no restriction",
			allowedDomain: "",
			email:         "user@any.com",
			expected:      true,
		},
		{
			name:          "case insensitive match",
			allowedDomain: "@Example.Com",
			email:         "user@EXAMPLE.com",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &OAuthProvider{
				allowedDomain: tt.allowedDomain,
				sessionSvc:    sessionSvc,
			}

			result := provider.IsAllowedDomain(tt.email)
			if result != tt.expected {
				t.Errorf("IsAllowedDomain(%v) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestOAuthProvider_CreateAuthURL(t *testing.T) {
	sessionSvc := NewSessionService(SessionServiceConfig{
		CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
		SecureCookies: false,
	})

	config := OAuthProviderConfig{
		BaseURL:      "https://ackify.example.com",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		AuthURL:      "https://provider.com/auth",
		TokenURL:     "https://provider.com/token",
		UserInfoURL:  "https://provider.com/userinfo",
		Scopes:       []string{"openid", "email"},
		SessionSvc:   sessionSvc,
	}

	provider := NewOAuthProvider(config)

	t.Run("creates auth URL with PKCE", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		authURL := provider.CreateAuthURL(rec, req, "/dashboard")

		if authURL == "" {
			t.Fatal("CreateAuthURL() returned empty string")
		}

		// Parse the URL and check parameters
		parsedURL, err := url.Parse(authURL)
		if err != nil {
			t.Fatalf("Failed to parse auth URL: %v", err)
		}

		query := parsedURL.Query()

		// Check for required OAuth parameters
		if query.Get("client_id") == "" {
			t.Error("client_id parameter missing")
		}
		if query.Get("redirect_uri") == "" {
			t.Error("redirect_uri parameter missing")
		}
		if query.Get("response_type") != "code" {
			t.Error("response_type should be 'code'")
		}
		if query.Get("state") == "" {
			t.Error("state parameter missing")
		}

		// Check for PKCE parameters
		if query.Get("code_challenge") == "" {
			t.Error("code_challenge parameter missing (PKCE should be enabled)")
		}
		if query.Get("code_challenge_method") != "S256" {
			t.Error("code_challenge_method should be 'S256'")
		}
	})

	t.Run("creates auth URL with silent flag", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/?silent=true", nil)

		authURL := provider.CreateAuthURL(rec, req, "/dashboard")

		parsedURL, _ := url.Parse(authURL)
		query := parsedURL.Query()

		if query.Get("prompt") != "none" {
			t.Errorf("prompt parameter = %v, expected 'none' for silent auth", query.Get("prompt"))
		}
	})
}

func TestOAuthProvider_VerifyState(t *testing.T) {
	sessionSvc := NewSessionService(SessionServiceConfig{
		CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
		SecureCookies: false,
	})

	config := OAuthProviderConfig{
		BaseURL:      "https://ackify.example.com",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		AuthURL:      "https://provider.com/auth",
		TokenURL:     "https://provider.com/token",
		UserInfoURL:  "https://provider.com/userinfo",
		Scopes:       []string{"openid"},
		SessionSvc:   sessionSvc,
	}

	provider := NewOAuthProvider(config)

	t.Run("valid state", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		// Create auth URL to generate state
		_ = provider.CreateAuthURL(rec, req, "/")

		// Get session with state
		req2 := httptest.NewRequest("GET", "/", nil)
		for _, cookie := range rec.Result().Cookies() {
			req2.AddCookie(cookie)
		}

		session, _ := sessionSvc.GetSession(req2)
		storedState, ok := session.Values["oauth_state"].(string)
		if !ok {
			t.Fatal("State not stored in session")
		}

		// Verify state
		rec2 := httptest.NewRecorder()
		valid := provider.VerifyState(rec2, req2, storedState)
		if !valid {
			t.Error("VerifyState() should return true for valid state")
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		valid := provider.VerifyState(rec, req, "invalid-state-token")
		if valid {
			t.Error("VerifyState() should return false for invalid state")
		}
	})

	t.Run("empty state", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		valid := provider.VerifyState(rec, req, "")
		if valid {
			t.Error("VerifyState() should return false for empty state")
		}
	})
}

func TestOAuthProvider_parseUserInfo(t *testing.T) {
	sessionSvc := NewSessionService(SessionServiceConfig{
		CookieSecret:  []byte("32-byte-secret-for-secure-cookies"),
		SecureCookies: false,
	})

	provider := &OAuthProvider{
		sessionSvc: sessionSvc,
	}

	tests := []struct {
		name        string
		responseObj map[string]interface{}
		wantErr     bool
		checkUser   func(*testing.T, *models.User)
	}{
		{
			name: "complete user info with sub",
			responseObj: map[string]interface{}{
				"sub":   "12345",
				"email": "user@example.com",
				"name":  "Test User",
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				if user.Sub != "12345" {
					t.Errorf("Sub = %v, expected 12345", user.Sub)
				}
				if user.Email != "user@example.com" {
					t.Errorf("Email = %v, expected user@example.com", user.Email)
				}
				if user.Name != "Test User" {
					t.Errorf("Name = %v, expected Test User", user.Name)
				}
			},
		},
		{
			name: "user info with id instead of sub",
			responseObj: map[string]interface{}{
				"id":    67890,
				"email": "user@example.com",
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				if user.Sub != "67890" {
					t.Errorf("Sub = %v, expected 67890", user.Sub)
				}
			},
		},
		{
			name: "user info with given_name and family_name",
			responseObj: map[string]interface{}{
				"sub":         "12345",
				"email":       "user@example.com",
				"given_name":  "John",
				"family_name": "Doe",
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				if user.Name != "John Doe" {
					t.Errorf("Name = %v, expected 'John Doe'", user.Name)
				}
			},
		},
		{
			name: "Microsoft Graph API - mail field",
			responseObj: map[string]interface{}{
				"id":                "microsoft-id-12345",
				"mail":              "user@company.com",
				"displayName":       "Microsoft User",
				"userPrincipalName": "user@company.onmicrosoft.com",
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				if user.Sub != "microsoft-id-12345" {
					t.Errorf("Sub = %v, expected microsoft-id-12345", user.Sub)
				}
				if user.Email != "user@company.com" {
					t.Errorf("Email = %v, expected user@company.com (from mail field)", user.Email)
				}
				if user.Name != "Microsoft User" {
					t.Errorf("Name = %v, expected Microsoft User (from displayName)", user.Name)
				}
			},
		},
		{
			name: "Microsoft Graph API - userPrincipalName fallback",
			responseObj: map[string]interface{}{
				"id":                "microsoft-id-67890",
				"displayName":       "UPN User",
				"userPrincipalName": "user@company.onmicrosoft.com",
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				if user.Email != "user@company.onmicrosoft.com" {
					t.Errorf("Email = %v, expected user@company.onmicrosoft.com (from userPrincipalName)", user.Email)
				}
			},
		},
		{
			name: "email field takes priority over mail",
			responseObj: map[string]interface{}{
				"sub":   "12345",
				"email": "primary@example.com",
				"mail":  "secondary@example.com",
			},
			wantErr: false,
			checkUser: func(t *testing.T, user *models.User) {
				if user.Email != "primary@example.com" {
					t.Errorf("Email = %v, expected primary@example.com (email should take priority)", user.Email)
				}
			},
		},
		{
			name: "missing email",
			responseObj: map[string]interface{}{
				"sub":  "12345",
				"name": "Test User",
			},
			wantErr: true,
		},
		{
			name: "missing sub and id",
			responseObj: map[string]interface{}{
				"email": "user@example.com",
				"name":  "Test User",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP response
			jsonData, _ := json.Marshal(tt.responseObj)
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(jsonData)),
			}

			user, err := provider.parseUserInfo(resp)

			if tt.wantErr {
				if err == nil {
					t.Error("parseUserInfo() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseUserInfo() unexpected error: %v", err)
			}

			if user == nil {
				t.Fatal("parseUserInfo() returned nil user")
			}

			if tt.checkUser != nil {
				tt.checkUser(t, user)
			}
		})
	}
}

func TestSubtleConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "secret123",
			b:        "secret123",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "secret123",
			b:        "secret456",
			expected: false,
		},
		{
			name:     "different lengths",
			a:        "short",
			b:        "longer string",
			expected: false,
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "",
			expected: true,
		},
		{
			name:     "one empty",
			a:        "string",
			b:        "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subtleConstantTimeCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("subtleConstantTimeCompare(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
