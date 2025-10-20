// SPDX-License-Identifier: AGPL-3.0-or-later
package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	apiAdmin "github.com/btouchard/ackify-ce/backend/internal/presentation/api/admin"
	apiAuth "github.com/btouchard/ackify-ce/backend/internal/presentation/api/auth"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/documents"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/health"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/signatures"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/users"
)

// RouterConfig holds configuration for the API router
type RouterConfig struct {
	AuthService              *auth.OauthService
	SignatureService         *services.SignatureService
	DocumentService          *services.DocumentService
	DocumentRepository       *database.DocumentRepository
	ExpectedSignerRepository *database.ExpectedSignerRepository
	ReminderService          *services.ReminderAsyncService // Now using async service
	BaseURL                  string
	AdminEmails              []string
	AutoLogin                bool
}

// NewRouter creates and configures the API v1 router
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Initialize middleware
	apiMiddleware := shared.NewMiddleware(cfg.AuthService, cfg.BaseURL, cfg.AdminEmails)

	// Rate limiters
	authRateLimit := shared.NewRateLimit(5, time.Minute)      // 5 attempts per minute for auth
	documentRateLimit := shared.NewRateLimit(10, time.Minute) // 10 documents per minute
	generalRateLimit := shared.NewRateLimit(100, time.Minute) // 100 requests per minute general

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(shared.AddRequestIDToContext)
	r.Use(middleware.RealIP)
	r.Use(shared.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(shared.SecurityHeaders)
	r.Use(apiMiddleware.CORS)
	r.Use(generalRateLimit.Middleware)

	// Initialize handlers
	healthHandler := health.NewHandler()
	authHandler := apiAuth.NewHandler(cfg.AuthService, apiMiddleware, cfg.BaseURL)
	usersHandler := users.NewHandler(cfg.AdminEmails)
	documentsHandler := documents.NewHandler(cfg.SignatureService, cfg.DocumentService)
	signaturesHandler := signatures.NewHandler(cfg.SignatureService)

	// Public routes
	r.Group(func(r chi.Router) {
		// Health check
		r.Get("/health", healthHandler.HandleHealth)

		// CSRF token
		r.Get("/csrf", authHandler.HandleGetCSRFToken)

		// Auth endpoints
		r.Route("/auth", func(r chi.Router) {
			r.Use(authRateLimit.Middleware)

			r.Post("/start", authHandler.HandleStartOAuth)
			r.Get("/callback", authHandler.HandleOAuthCallback)
			r.Get("/logout", authHandler.HandleLogout)

			if cfg.AutoLogin {
				r.Get("/check", authHandler.HandleAuthCheck)
			}
		})

		// Public document endpoints
		r.Route("/documents", func(r chi.Router) {
			// Document creation (with CSRF and stricter rate limiting)
			r.Group(func(r chi.Router) {
				r.Use(apiMiddleware.CSRFProtect)
				r.Use(documentRateLimit.Middleware)
				r.Post("/", documentsHandler.HandleCreateDocument)
			})

			// Read-only document endpoints
			r.Get("/", documentsHandler.HandleListDocuments)
			r.Get("/{docId}", documentsHandler.HandleGetDocument)
			r.Get("/{docId}/signatures", documentsHandler.HandleGetDocumentSignatures)
			r.Get("/{docId}/expected-signers", documentsHandler.HandleGetExpectedSigners)

			// Find or create document by reference (public for embed support, but with optional auth)
			r.Group(func(r chi.Router) {
				r.Use(apiMiddleware.OptionalAuth)
				r.Get("/find-or-create", documentsHandler.HandleFindOrCreateDocument)
			})
		})
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(apiMiddleware.RequireAuth)
		r.Use(apiMiddleware.CSRFProtect)

		// User endpoints
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", usersHandler.HandleGetCurrentUser)
		})

		// Signature endpoints
		r.Route("/signatures", func(r chi.Router) {
			r.Get("/", signaturesHandler.HandleGetUserSignatures)
			r.Post("/", signaturesHandler.HandleCreateSignature)
		})

		// Document signature status (authenticated)
		r.Get("/documents/{docId}/signatures/status", signaturesHandler.HandleGetSignatureStatus)
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(apiMiddleware.RequireAdmin)
		r.Use(apiMiddleware.CSRFProtect)

		// Initialize admin handler
		adminHandler := apiAdmin.NewHandler(cfg.DocumentRepository, cfg.ExpectedSignerRepository, cfg.ReminderService, cfg.SignatureService, cfg.BaseURL)

		r.Route("/admin", func(r chi.Router) {
			// Document management
			r.Route("/documents", func(r chi.Router) {
				r.Get("/", adminHandler.HandleListDocuments)
				r.Get("/{docId}", adminHandler.HandleGetDocument)
				r.Get("/{docId}/signers", adminHandler.HandleGetDocumentWithSigners)
				r.Get("/{docId}/status", adminHandler.HandleGetDocumentStatus)

				// Document metadata
				r.Put("/{docId}/metadata", adminHandler.HandleUpdateDocumentMetadata)

				// Document deletion
				r.Delete("/{docId}", adminHandler.HandleDeleteDocument)

				// Expected signers management
				r.Post("/{docId}/signers", adminHandler.HandleAddExpectedSigner)
				r.Delete("/{docId}/signers/{email}", adminHandler.HandleRemoveExpectedSigner)

				// Reminder management
				r.Post("/{docId}/reminders", adminHandler.HandleSendReminders)
				r.Get("/{docId}/reminders", adminHandler.HandleGetReminderHistory)
			})
		})
	})

	// Serve OpenAPI spec
	r.Get("/openapi.json", serveOpenAPISpec)

	return r
}

// serveOpenAPISpec serves the OpenAPI specification
func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// TODO: Read and serve the OpenAPI YAML file as JSON
	// For now, return a simple response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"info":{"title":"Ackify API","version":"1.0.0"}}`))
}
