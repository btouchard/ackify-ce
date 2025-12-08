// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

type userService interface {
	GetUser(r *http.Request) (*models.User, error)
}

type AuthMiddleware struct {
	userService userService
	baseURL     string
}

// SecureHeaders Enforce baseline security headers (CSP, XFO, etc.) to mitigate clickjacking, MIME sniffing, and unsafe embedding by default.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Check if this is an embed route - allow iframe embedding
		isEmbedRoute := strings.HasPrefix(r.URL.Path, "/embed/") || strings.HasPrefix(r.URL.Path, "/embed")

		if isEmbedRoute {
			// Allow embedding from any origin for embed pages
			// Do not set X-Frame-Options to allow iframe embedding
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com; "+
					"script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
					"img-src 'self' data: https://cdn.simpleicons.org; "+
					"connect-src 'self'; "+
					"frame-ancestors *") // Allow embedding from any origin
		} else {
			// Strict headers for non-embed routes
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com; "+
					"script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; "+
					"img-src 'self' data: https://cdn.simpleicons.org; "+
					"connect-src 'self'; "+
					"frame-ancestors 'self'")
		}

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
