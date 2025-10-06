// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
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

type fakeSignatureService struct {
	shouldFailCreate       bool
	shouldFailGetStatus    bool
	shouldFailGetByDocUser bool
	shouldFailGetDoc       bool
	shouldFailGetUser      bool
	shouldFailCheck        bool
	createError            error
	statusResult           *models.SignatureStatus
	getStatusError         error
	signature              *models.Signature
	getSignatureError      error
	docSignatures          []*models.Signature
	getDocError            error
	userSignatures         []*models.Signature
	getUserError           error
	checkResult            bool
	checkError             error
}

func newFakeSignatureService() *fakeSignatureService {
	return &fakeSignatureService{
		statusResult: &models.SignatureStatus{
			DocID:     "test-doc",
			UserEmail: "test@example.com",
			IsSigned:  false,
			SignedAt:  nil,
		},
		signature: &models.Signature{
			ID:          1,
			DocID:       "test-doc",
			UserSub:     "test-user",
			UserEmail:   "test@example.com",
			SignedAtUTC: time.Now().UTC(),
		},
		docSignatures: []*models.Signature{
			{
				ID:          1,
				DocID:       "test-doc",
				UserSub:     "test-user",
				UserEmail:   "test@example.com",
				SignedAtUTC: time.Now().UTC(),
			},
		},
		userSignatures: []*models.Signature{
			{
				ID:          1,
				DocID:       "test-doc",
				UserSub:     "test-user",
				UserEmail:   "test@example.com",
				SignedAtUTC: time.Now().UTC(),
			},
		},
		checkResult: true,
	}
}

func (f *fakeSignatureService) CreateSignature(_ context.Context, _ *models.SignatureRequest) error {
	if f.shouldFailCreate {
		return f.createError
	}
	return nil
}

func (f *fakeSignatureService) GetSignatureStatus(_ context.Context, _ string, _ *models.User) (*models.SignatureStatus, error) {
	if f.shouldFailGetStatus {
		return nil, f.getStatusError
	}
	return f.statusResult, nil
}

func (f *fakeSignatureService) GetSignatureByDocAndUser(_ context.Context, _ string, _ *models.User) (*models.Signature, error) {
	if f.shouldFailGetByDocUser {
		return nil, f.getSignatureError
	}
	return f.signature, nil
}

func (f *fakeSignatureService) GetDocumentSignatures(_ context.Context, _ string) ([]*models.Signature, error) {
	if f.shouldFailGetDoc {
		return nil, f.getDocError
	}
	return f.docSignatures, nil
}

func (f *fakeSignatureService) GetUserSignatures(_ context.Context, _ *models.User) ([]*models.Signature, error) {
	if f.shouldFailGetUser {
		return nil, f.getUserError
	}
	return f.userSignatures, nil
}

func (f *fakeSignatureService) CheckUserSignature(_ context.Context, _, _ string) (bool, error) {
	if f.shouldFailCheck {
		return false, f.checkError
	}
	return f.checkResult, nil
}

func createTestTemplate() *template.Template {
	tmpl := template.New("test")
	template.Must(tmpl.New("base").Parse(`<html><body>{{.TemplateName}}</body></html>`))
	return tmpl
}

func TestAuthHandlers_NewAuthHandlers(t *testing.T) {
	authService := newFakeAuthService()
	baseURL := "https://example.com"

	handlers := NewAuthHandlers(authService, baseURL)

	if handlers == nil {
		t.Error("NewAuthHandlers should not return nil")
	} else if handlers.authService != authService {
		t.Error("AuthService not set correctly")
	} else if handlers.baseURL != baseURL {
		t.Error("BaseURL not set correctly")
	}
}

func TestAuthHandlers_HandleLogin(t *testing.T) {
	tests := []struct {
		name        string
		nextParam   string
		expectedURL string
	}{
		{
			name:        "login with next parameter",
			nextParam:   "/sign?doc=test",
			expectedURL: "https://oauth.example.com/auth?next=" + url.QueryEscape("/sign?doc=test"),
		},
		{
			name:        "login without next parameter",
			nextParam:   "",
			expectedURL: "https://oauth.example.com/auth?next=" + url.QueryEscape("https://example.com/"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := newFakeAuthService()
			handlers := NewAuthHandlers(authService, "https://example.com")

			req := httptest.NewRequest("GET", "/login", nil)
			if tt.nextParam != "" {
				q := req.URL.Query()
				q.Set("next", tt.nextParam)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			handlers.HandleLogin(w, req)

			if w.Code != http.StatusFound {
				t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
			}

			location := w.Header().Get("Location")
			if location != tt.expectedURL {
				t.Errorf("Expected redirect to %s, got %s", tt.expectedURL, location)
			}
		})
	}
}

func TestAuthHandlers_HandleLogout(t *testing.T) {
	t.Run("logout without SSO logout URL redirects to home", func(t *testing.T) {
		authService := newFakeAuthService()
		handlers := NewAuthHandlers(authService, "https://example.com")

		req := httptest.NewRequest("GET", "/logout", nil)
		w := httptest.NewRecorder()

		handlers.HandleLogout(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/" {
			t.Errorf("Expected redirect to /, got %s", location)
		}

		if !authService.logoutCalled {
			t.Error("Logout should have been called on auth service")
		}
	})

	t.Run("logout with SSO logout URL redirects to SSO", func(t *testing.T) {
		authService := newFakeAuthService()
		authService.logoutURL = "https://accounts.google.com/Logout?continue=https://example.com"
		handlers := NewAuthHandlers(authService, "https://example.com")

		req := httptest.NewRequest("GET", "/logout", nil)
		w := httptest.NewRecorder()

		handlers.HandleLogout(w, req)

		if w.Code != http.StatusFound {
			t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
		}

		location := w.Header().Get("Location")
		expectedLocation := "https://accounts.google.com/Logout?continue=https://example.com"
		if location != expectedLocation {
			t.Errorf("Expected redirect to %s, got %s", expectedLocation, location)
		}

		if !authService.logoutCalled {
			t.Error("Logout should have been called on auth service")
		}
	})
}

func TestAuthHandlers_HandleOAuthCallback(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		state            string
		setupAuth        func(*fakeAuthService)
		expectedStatus   int
		expectedRedirect string
	}{
		{
			name:             "successful callback",
			code:             "test-code",
			state:            "test-state",
			setupAuth:        func(a *fakeAuthService) {},
			expectedStatus:   http.StatusFound,
			expectedRedirect: "/",
		},
		{
			name:           "missing state",
			code:           "test-code",
			state:          "",
			setupAuth:      func(a *fakeAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "invalid state - verify fails",
			code:  "test-code",
			state: "abc123:Lw",
			setupAuth: func(a *fakeAuthService) {
				a.verifyStateResult = false
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing code",
			code:           "",
			state:          "test-state",
			setupAuth:      func(a *fakeAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "callback fails",
			code:  "test-code",
			state: "test-state",
			setupAuth: func(a *fakeAuthService) {
				a.shouldFailCallback = true
				a.callbackError = models.ErrDomainNotAllowed
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:  "set user fails",
			code:  "test-code",
			state: "test-state",
			setupAuth: func(a *fakeAuthService) {
				a.shouldFailSetUser = true
				a.setUserError = errors.New("session error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := newFakeAuthService()
			tt.setupAuth(authService)
			handlers := NewAuthHandlers(authService, "https://example.com")

			req := httptest.NewRequest("GET", "/oauth2/callback", nil)
			q := req.URL.Query()
			if tt.code != "" {
				q.Set("code", tt.code)
			}
			if tt.state != "" {
				q.Set("state", tt.state)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			handlers.HandleOAuthCallback(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedRedirect != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedRedirect {
					t.Errorf("Expected redirect to %s, got %s", tt.expectedRedirect, location)
				}
			}
		})
	}
}

func TestAuthHandlers_HandleOAuthCallback_VerifyStateToken(t *testing.T) {
	authService := newFakeAuthService()
	handlers := NewAuthHandlers(authService, "https://example.com")

	token := "abc123"
	state := token + ":" + base64.RawURLEncoding.EncodeToString([]byte("/"))

	req := httptest.NewRequest("GET", "/oauth2/callback?code=ok&state="+url.QueryEscape(state), nil)
	w := httptest.NewRecorder()
	handlers.HandleOAuthCallback(w, req)

	if authService.lastVerifyToken != token {
		t.Errorf("expected VerifyState to receive token %q, got %q", token, authService.lastVerifyToken)
	}
}

func TestSignatureHandlers_NewSignatureHandlers(t *testing.T) {
	signatureService := newFakeSignatureService()
	userService := newFakeUserService()
	tmpl := createTestTemplate()
	baseURL := "https://example.com"
	organisation := "Organisation"
	adminEmails := []string{"admin@example.com"}

	autoLogin := false
	handlers := NewSignatureHandlers(signatureService, userService, tmpl, baseURL, organisation, adminEmails, autoLogin)

	if handlers == nil {
		t.Error("NewSignatureHandlers should not return nil")
	} else if handlers.signatureService != signatureService {
		t.Error("SignatureService not set correctly")
	} else if handlers.userService != userService {
		t.Error("UserService not set correctly")
	} else if handlers.template != tmpl {
		t.Error("Template not set correctly")
	} else if handlers.baseURL != baseURL {
		t.Error("BaseURL not set correctly")
	} else if handlers.organisation != organisation {
		t.Error("Organisation not set correctly")
	} else if handlers.autoLogin != autoLogin {
		t.Error("AutoLogin not set correctly")
	}
}

func TestSignatureHandlers_HandleIndex(t *testing.T) {
	signatureService := newFakeSignatureService()
	userService := newFakeUserService()
	tmpl := createTestTemplate()
	handlers := NewSignatureHandlers(signatureService, userService, tmpl, "https://example.com", "Organisation", []string{}, false)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handlers.HandleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "index") {
		t.Error("Response should contain template name 'index'")
	}
}

func TestSignatureHandlers_HandleSignGET(t *testing.T) {
	tests := []struct {
		name           string
		docParam       string
		setupUser      func(*fakeUserService)
		setupSig       func(*fakeSignatureService)
		expectedStatus int
		shouldRedirect bool
	}{
		{
			name:      "successful sign page load - not signed",
			docParam:  "test-doc",
			setupUser: func(u *fakeUserService) {},
			setupSig: func(s *fakeSignatureService) {
				s.statusResult.IsSigned = false
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "successful sign page load - already signed",
			docParam:  "test-doc",
			setupUser: func(u *fakeUserService) {},
			setupSig: func(s *fakeSignatureService) {
				s.statusResult.IsSigned = true
				signedAt := time.Now().UTC()
				s.statusResult.SignedAt = &signedAt
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "user not authenticated",
			docParam: "test-doc",
			setupUser: func(u *fakeUserService) {
				u.shouldFail = true
				u.getUserError = models.ErrUnauthorized
			},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing doc parameter",
			docParam:       "",
			setupUser:      func(u *fakeUserService) {},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusFound,
			shouldRedirect: true,
		},
		{
			name:      "signature service fails",
			docParam:  "test-doc",
			setupUser: func(u *fakeUserService) {},
			setupSig: func(s *fakeSignatureService) {
				s.shouldFailGetStatus = true
				s.getStatusError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureService := newFakeSignatureService()
			userService := newFakeUserService()
			tt.setupUser(userService)
			tt.setupSig(signatureService)

			tmpl := createTestTemplate()
			handlers := NewSignatureHandlers(signatureService, userService, tmpl, "https://example.com", "Organisation", []string{}, false)

			req := httptest.NewRequest("GET", "/sign", nil)
			if tt.docParam != "" {
				q := req.URL.Query()
				q.Set("doc", tt.docParam)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			handlers.HandleSignGET(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.shouldRedirect {
				location := w.Header().Get("Location")
				if location == "" {
					t.Error("Expected redirect but no Location header found")
				}
			}
		})
	}
}

func TestSignatureHandlers_HandleSignPOST(t *testing.T) {
	tests := []struct {
		name           string
		formData       map[string]string
		setupUser      func(*fakeUserService)
		setupSig       func(*fakeSignatureService)
		expectedStatus int
		shouldRedirect bool
	}{
		{
			name: "successful signature creation",
			formData: map[string]string{
				"doc": "test-doc",
			},
			setupUser:      func(u *fakeUserService) {},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusFound,
			shouldRedirect: true,
		},
		{
			name: "signature already exists",
			formData: map[string]string{
				"doc": "test-doc",
			},
			setupUser: func(u *fakeUserService) {},
			setupSig: func(s *fakeSignatureService) {
				s.shouldFailCreate = true
				s.createError = models.ErrSignatureAlreadyExists
			},
			expectedStatus: http.StatusFound,
			shouldRedirect: true,
		},
		{
			name: "user not authenticated",
			formData: map[string]string{
				"doc": "test-doc",
			},
			setupUser: func(u *fakeUserService) {
				u.shouldFail = true
				u.getUserError = models.ErrUnauthorized
			},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusFound,
			shouldRedirect: true,
		},
		{
			name:           "missing doc parameter",
			formData:       map[string]string{},
			setupUser:      func(u *fakeUserService) {},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "signature service fails",
			formData: map[string]string{
				"doc": "test-doc",
			},
			setupUser: func(u *fakeUserService) {},
			setupSig: func(s *fakeSignatureService) {
				s.shouldFailCreate = true
				s.createError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureService := newFakeSignatureService()
			userService := newFakeUserService()
			tt.setupUser(userService)
			tt.setupSig(signatureService)

			tmpl := createTestTemplate()
			handlers := NewSignatureHandlers(signatureService, userService, tmpl, "https://example.com", "Organisation", []string{}, false)

			form := url.Values{}
			for key, value := range tt.formData {
				form.Set(key, value)
			}

			req := httptest.NewRequest("POST", "/sign", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			handlers.HandleSignPOST(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.shouldRedirect {
				location := w.Header().Get("Location")
				if location == "" {
					t.Error("Expected redirect but no Location header found")
				}
			}
		})
	}
}

func TestSignatureHandlers_HandleStatusJSON(t *testing.T) {
	tests := []struct {
		name           string
		docParam       string
		setupSig       func(*fakeSignatureService)
		expectedStatus int
	}{
		{
			name:           "successful status JSON",
			docParam:       "test-doc",
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing doc parameter",
			docParam:       "",
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "service fails",
			docParam: "test-doc",
			setupSig: func(s *fakeSignatureService) {
				s.shouldFailGetDoc = true
				s.getDocError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureService := newFakeSignatureService()
			userService := newFakeUserService()
			tt.setupSig(signatureService)

			tmpl := createTestTemplate()
			handlers := NewSignatureHandlers(signatureService, userService, tmpl, "https://example.com", "Organisation", []string{}, false)

			req := httptest.NewRequest("GET", "/status", nil)
			if tt.docParam != "" {
				q := req.URL.Query()
				q.Set("doc", tt.docParam)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			handlers.HandleStatusJSON(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Errorf("Expected JSON content type, got %s", contentType)
				}
			}
		})
	}
}

func TestSignatureHandlers_HandleUserSignatures(t *testing.T) {
	tests := []struct {
		name           string
		setupUser      func(*fakeUserService)
		setupSig       func(*fakeSignatureService)
		expectedStatus int
	}{
		{
			name:           "successful user signatures",
			setupUser:      func(u *fakeUserService) {},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "user not authenticated",
			setupUser: func(u *fakeUserService) {
				u.shouldFail = true
				u.getUserError = models.ErrUnauthorized
			},
			setupSig:       func(s *fakeSignatureService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "service fails",
			setupUser: func(u *fakeUserService) {},
			setupSig: func(s *fakeSignatureService) {
				s.shouldFailGetUser = true
				s.getUserError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signatureService := newFakeSignatureService()
			userService := newFakeUserService()
			tt.setupUser(userService)
			tt.setupSig(signatureService)

			tmpl := createTestTemplate()
			handlers := NewSignatureHandlers(signatureService, userService, tmpl, "https://example.com", "Organisation", []string{}, false)

			req := httptest.NewRequest("GET", "/signatures", nil)
			w := httptest.NewRecorder()

			handlers.HandleUserSignatures(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, "text/html") {
					t.Errorf("Expected HTML content type, got %s", contentType)
				}
			}
		})
	}
}
