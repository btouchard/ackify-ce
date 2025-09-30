// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

type AuthMiddleware struct {
	userService userService
	baseURL     string
}

func NewAuthMiddleware(userService userService, baseURL string) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
		baseURL:     baseURL,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := m.userService.GetUser(r)
		if err != nil {
			nextURL := m.baseURL + r.URL.RequestURI()
			loginURL := buildLoginURL(nextURL)
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}
		next(w, r)
	}
}

// SecureHeaders Enforce baseline security headers (CSP, XFO, etc.) to mitigate clickjacking, MIME sniffing, and unsafe embedding by default.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
				"script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
				"img-src 'self' data: https://cdn.simpleicons.org; connect-src 'self'; "+
				"frame-ancestors 'self'")
		next.ServeHTTP(w, r)
	})
}

// RequestLogger Minimal structured logging without PII; record latency and status for ops visibility.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(sr, r)
		duration := time.Since(start)
		// Minimal structured log to avoid PII
		logger.Logger.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sr.status,
			"duration_ms", duration.Milliseconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HandleError handles different types of errors and returns appropriate HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrUnauthorized):
		logger.Logger.Warn("Unauthorized access attempt", "error", err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	case errors.Is(err, models.ErrSignatureNotFound):
		logger.Logger.Debug("Signature not found", "error", err.Error())
		http.Error(w, "Signature not found", http.StatusNotFound)
	case errors.Is(err, models.ErrSignatureAlreadyExists):
		logger.Logger.Debug("Duplicate signature attempt", "error", err.Error())
		http.Error(w, "Signature already exists", http.StatusConflict)
	case errors.Is(err, models.ErrInvalidUser):
		logger.Logger.Warn("Invalid user data", "error", err.Error())
		http.Error(w, "Invalid user", http.StatusBadRequest)
	case errors.Is(err, models.ErrInvalidDocument):
		logger.Logger.Warn("Invalid document ID", "error", err.Error())
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
	case errors.Is(err, models.ErrDomainNotAllowed):
		logger.Logger.Warn("Domain not allowed", "error", err.Error())
		http.Error(w, "Domain not allowed", http.StatusForbidden)
	case errors.Is(err, models.ErrDatabaseConnection):
		logger.Logger.Error("Database connection error", "error", err.Error())
		http.Error(w, "Database error", http.StatusInternalServerError)
	default:
		logger.Logger.Error("Unhandled error", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
