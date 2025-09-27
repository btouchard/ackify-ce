// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)


func TestBadgeHandler_NewBadgeHandler(t *testing.T) {
	checkService := newFakeSignatureService()
	handler := NewBadgeHandler(checkService)

	if handler == nil {
		t.Error("NewBadgeHandler should not return nil")
	} else if handler.checkService != checkService {
		t.Error("CheckService not set correctly")
	}
}

func TestBadgeHandler_HandleStatusPNG(t *testing.T) {
	tests := []struct {
		name           string
		docParam       string
		userParam      string
		setupService   func(*fakeSignatureService)
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "successful badge - signed",
			docParam:       "test-doc",
			userParam:      "test@example.com",
			setupService:   func(s *fakeSignatureService) { s.checkResult = true },
			expectedStatus: http.StatusOK,
			expectedType:   "image/png",
		},
		{
			name:           "successful badge - not signed",
			docParam:       "test-doc",
			userParam:      "test@example.com",
			setupService:   func(s *fakeSignatureService) { s.checkResult = false },
			expectedStatus: http.StatusOK,
			expectedType:   "image/png",
		},
		{
			name:           "missing doc parameter",
			docParam:       "",
			userParam:      "test@example.com",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing user parameter",
			docParam:       "test-doc",
			userParam:      "",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "service fails",
			docParam:  "test-doc",
			userParam: "test@example.com",
			setupService: func(s *fakeSignatureService) {
				s.shouldFailCheck = true
				s.checkError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newFakeSignatureService()
			tt.setupService(service)
			handler := NewBadgeHandler(service)

			req := httptest.NewRequest("GET", "/status.png", nil)
			q := req.URL.Query()
			if tt.docParam != "" {
				q.Set("doc", tt.docParam)
			}
			if tt.userParam != "" {
				q.Set("user", tt.userParam)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			handler.HandleStatusPNG(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedType != "" {
				contentType := w.Header().Get("Content-Type")
				if contentType != tt.expectedType {
					t.Errorf("Expected content type %s, got %s", tt.expectedType, contentType)
				}

				cacheControl := w.Header().Get("Cache-Control")
				if cacheControl != "no-store" {
					t.Errorf("Expected Cache-Control: no-store, got %s", cacheControl)
				}
			}
		})
	}
}


func TestHealthHandler_NewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()
	if handler == nil {
		t.Error("NewHealthHandler should not return nil")
	}
}

func TestHealthHandler_HandleHealth(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"ok":true`) {
		t.Error("Response should contain ok:true")
	}
	if !strings.Contains(body, `"time"`) {
		t.Error("Response should contain time field")
	}
}


func TestOEmbedHandler_NewOEmbedHandler(t *testing.T) {
	service := newFakeSignatureService()
	tmpl := createTestTemplate()
	baseURL := "https://example.com"
	org := "Test Org"

	handler := NewOEmbedHandler(service, tmpl, baseURL, org)

	if handler == nil {
		t.Error("NewOEmbedHandler should not return nil")
	} else if handler.signatureService != service {
		t.Error("SignatureService not set correctly")
	} else if handler.template != tmpl {
		t.Error("Template not set correctly")
	} else if handler.baseURL != baseURL {
		t.Error("BaseURL not set correctly")
	} else if handler.organisation != org {
		t.Error("Organisation not set correctly")
	}
}

func TestOEmbedHandler_HandleOEmbed(t *testing.T) {
	tests := []struct {
		name           string
		urlParam       string
		formatParam    string
		setupService   func(*fakeSignatureService)
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "successful oembed",
			urlParam:       "https://example.com/embed?doc=test-doc",
			formatParam:    "json",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "default format (json)",
			urlParam:       "https://example.com/embed?doc=test-doc",
			formatParam:    "",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
		},
		{
			name:           "unsupported format",
			urlParam:       "https://example.com/embed?doc=test-doc",
			formatParam:    "xml",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusNotImplemented,
		},
		{
			name:           "missing url parameter",
			urlParam:       "",
			formatParam:    "json",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid url format",
			urlParam:       "https://example.com/embed",
			formatParam:    "json",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service fails",
			urlParam:    "https://example.com/embed?doc=test-doc",
			formatParam: "json",
			setupService: func(s *fakeSignatureService) {
				s.shouldFailGetDoc = true
				s.getDocError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newFakeSignatureService()
			tt.setupService(service)
			tmpl := createTestTemplate()
			template.Must(tmpl.New("embed").Parse(`<div>{{.DocID}} - {{.Count}} signatures</div>`))
			handler := NewOEmbedHandler(service, tmpl, "https://example.com", "Test Org")

			req := httptest.NewRequest("GET", "/oembed", nil)
			q := req.URL.Query()
			if tt.urlParam != "" {
				q.Set("url", tt.urlParam)
			}
			if tt.formatParam != "" {
				q.Set("format", tt.formatParam)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			handler.HandleOEmbed(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedType != "" {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, tt.expectedType) {
					t.Errorf("Expected content type %s, got %s", tt.expectedType, contentType)
				}
			}
		})
	}
}

func TestOEmbedHandler_HandleEmbedView(t *testing.T) {
	tests := []struct {
		name           string
		docParam       string
		setupService   func(*fakeSignatureService)
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "successful embed view",
			docParam:       "test-doc",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusOK,
			expectedType:   "text/html",
		},
		{
			name:           "missing doc parameter",
			docParam:       "",
			setupService:   func(s *fakeSignatureService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "service fails",
			docParam: "test-doc",
			setupService: func(s *fakeSignatureService) {
				s.shouldFailGetDoc = true
				s.getDocError = errors.New("service error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newFakeSignatureService()
			tt.setupService(service)
			tmpl := createTestTemplate()
			template.Must(tmpl.New("embed").Parse(`<div>{{.DocID}} - {{.Count}} signatures</div>`))

			handler := NewOEmbedHandler(service, tmpl, "https://example.com", "Test Org")

			req := httptest.NewRequest("GET", "/embed", nil)
			if tt.docParam != "" {
				q := req.URL.Query()
				q.Set("doc", tt.docParam)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			handler.HandleEmbedView(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedType != "" {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, tt.expectedType) {
					t.Errorf("Expected content type %s, got %s", tt.expectedType, contentType)
				}

				frameOptions := w.Header().Get("X-Frame-Options")
				if frameOptions != "ALLOWALL" {
					t.Errorf("Expected X-Frame-Options: ALLOWALL, got %s", frameOptions)
				}
			}
		})
	}
}

func TestOEmbedHandler_extractDocIDFromURL(t *testing.T) {
	handler := &OEmbedHandler{}

	tests := []struct {
		name      string
		url       string
		expected  string
		shouldErr bool
	}{
		{
			name:     "extract from query parameter",
			url:      "https://example.com/embed?doc=test-doc",
			expected: "test-doc",
		},
		{
			name:     "extract from embed path",
			url:      "https://example.com/embed/test-doc",
			expected: "test-doc",
		},
		{
			name:     "extract from status path",
			url:      "https://example.com/status/test-doc",
			expected: "test-doc",
		},
		{
			name:     "extract from sign path",
			url:      "https://example.com/sign/test-doc",
			expected: "test-doc",
		},
		{
			name:      "invalid url",
			url:       "not-a-url",
			shouldErr: true,
		},
		{
			name:      "no doc id found",
			url:       "https://example.com/other",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.extractDocIDFromURL(tt.url)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}


func TestAuthMiddleware_NewAuthMiddleware(t *testing.T) {
	userService := newFakeUserService()
	baseURL := "https://example.com"

	middleware := NewAuthMiddleware(userService, baseURL)

	if middleware == nil {
		t.Error("NewAuthMiddleware should not return nil")
	} else if middleware.userService != userService {
		t.Error("UserService not set correctly")
	} else if middleware.baseURL != baseURL {
		t.Error("BaseURL not set correctly")
	}
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		setupUser      func(*fakeUserService)
		expectedStatus int
		shouldRedirect bool
	}{
		{
			name:           "authenticated user",
			setupUser:      func(u *fakeUserService) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthenticated user",
			setupUser: func(u *fakeUserService) {
				u.shouldFail = true
				u.getUserError = models.ErrUnauthorized
			},
			expectedStatus: http.StatusFound,
			shouldRedirect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userService := newFakeUserService()
			tt.setupUser(userService)
			middleware := NewAuthMiddleware(userService, "https://example.com")

			testHandler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}

			wrappedHandler := middleware.RequireAuth(testHandler)

			req := httptest.NewRequest("GET", "/protected", nil)
			w := httptest.NewRecorder()

			wrappedHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.shouldRedirect {
				location := w.Header().Get("Location")
				if location == "" {
					t.Error("Expected redirect but no Location header found")
				}
				if !strings.Contains(location, "/login") {
					t.Error("Expected redirect to login page")
				}
			}
		})
	}
}

func TestSecureHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrapped := SecureHeaders(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	headers := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "no-referrer",
		"Content-Security-Policy": "default-src 'self'",
	}

	for header, expectedValue := range headers {
		actualValue := w.Header().Get(header)
		if !strings.Contains(actualValue, expectedValue) {
			t.Errorf("Expected header %s to contain %s, got %s", header, expectedValue, actualValue)
		}
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedText   string
	}{
		{
			name:           "unauthorized error",
			err:            models.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedText:   "Unauthorized",
		},
		{
			name:           "signature not found error",
			err:            models.ErrSignatureNotFound,
			expectedStatus: http.StatusNotFound,
			expectedText:   "Signature not found",
		},
		{
			name:           "signature already exists error",
			err:            models.ErrSignatureAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedText:   "Signature already exists",
		},
		{
			name:           "invalid user error",
			err:            models.ErrInvalidUser,
			expectedStatus: http.StatusBadRequest,
			expectedText:   "Invalid user",
		},
		{
			name:           "invalid document error",
			err:            models.ErrInvalidDocument,
			expectedStatus: http.StatusBadRequest,
			expectedText:   "Invalid document ID",
		},
		{
			name:           "domain not allowed error",
			err:            models.ErrDomainNotAllowed,
			expectedStatus: http.StatusForbidden,
			expectedText:   "Domain not allowed",
		},
		{
			name:           "database connection error",
			err:            models.ErrDatabaseConnection,
			expectedStatus: http.StatusInternalServerError,
			expectedText:   "Database error",
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedText:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleError(w, tt.err)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if !strings.Contains(body, tt.expectedText) {
				t.Errorf("Expected body to contain %s, got %s", tt.expectedText, body)
			}
		})
	}
}


func TestValidateDocID(t *testing.T) {
	tests := []struct {
		name      string
		setupReq  func() *http.Request
		expected  string
		shouldErr bool
	}{
		{
			name: "from query parameter",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?doc=test-doc", nil)
				return req
			},
			expected: "test-doc",
		},
		{
			name: "from form value",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader("doc=test-doc"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expected: "test-doc",
		},
		{
			name: "trimmed whitespace",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?doc=%20test-doc%20", nil)
				return req
			},
			expected: "test-doc",
		},
		{
			name: "missing doc parameter",
			setupReq: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				return req
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			result, err := validateDocID(req)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestBuildSignURL(t *testing.T) {
	result := buildSignURL("https://example.com", "test doc")
	expected := "https://example.com/sign?doc=test+doc"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestBuildLoginURL(t *testing.T) {
	result := buildLoginURL("https://example.com/sign?doc=test")
	expected := "/login?next=" + url.QueryEscape("https://example.com/sign?doc=test")

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestValidateUserIdentifier(t *testing.T) {
	tests := []struct {
		name      string
		userParam string
		expected  string
		shouldErr bool
	}{
		{
			name:      "valid user identifier",
			userParam: "test@example.com",
			expected:  "test@example.com",
		},
		{
			name:      "trimmed whitespace",
			userParam: " test@example.com ",
			expected:  "test@example.com",
		},
		{
			name:      "missing user parameter",
			userParam: "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.userParam != "" {
				q := req.URL.Query()
				q.Set("user", tt.userParam)
				req.URL.RawQuery = q.Encode()
			}

			result, err := validateUserIdentifier(req)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}
