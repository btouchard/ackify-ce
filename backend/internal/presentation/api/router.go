// SPDX-License-Identifier: AGPL-3.0-or-later
package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gopkg.in/yaml.v3"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	apiAdmin "github.com/btouchard/ackify-ce/backend/internal/presentation/api/admin"
	apiAuth "github.com/btouchard/ackify-ce/backend/internal/presentation/api/auth"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/health"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/users"
	"github.com/btouchard/ackify-ce/backend/pkg/coreapp"
)

// RouterConfig holds configuration for the API router
type RouterConfig struct {
	AuthService               *auth.OauthService
	MagicLinkService          *services.MagicLinkService
	WebhookRepository         *database.WebhookRepository
	WebhookDeliveryRepository *database.WebhookDeliveryRepository
	CoreDeps                  coreapp.CoreDeps
	BaseURL                   string
	AdminEmails               []string
	AutoLogin                 bool
	OAuthEnabled              bool
	MagicLinkEnabled          bool
	AuthRateLimit             int
	DocumentRateLimit         int
	GeneralRateLimit          int
}

// NewRouter creates and configures the API v1 router
func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	// Initialize middleware
	apiMiddleware := shared.NewMiddleware(cfg.AuthService, cfg.BaseURL, cfg.AdminEmails)

	// Rate limiters with configurable limits
	authLimit := cfg.AuthRateLimit
	if authLimit == 0 {
		authLimit = 5 // Default: 5 attempts per minute for auth
	}
	documentLimit := cfg.DocumentRateLimit
	if documentLimit == 0 {
		documentLimit = 10 // Default: 10 documents per minute
	}
	generalLimit := cfg.GeneralRateLimit
	if generalLimit == 0 {
		generalLimit = 100 // Default: 100 requests per minute general
	}

	authRateLimit := shared.NewRateLimit(authLimit, time.Minute)
	documentRateLimit := shared.NewRateLimit(documentLimit, time.Minute)
	generalRateLimit := shared.NewRateLimit(generalLimit, time.Minute)

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(shared.AddRequestIDToContext)
	r.Use(middleware.RealIP)
	r.Use(shared.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(shared.SecurityHeaders)
	r.Use(apiMiddleware.CORS)
	r.Use(generalRateLimit.Middleware)

	// Initialize coreapp handler groups (documents/signatures)
	coreGroups := coreapp.NewHandlerGroups(cfg.CoreDeps)

	// Initialize handlers for non-coreapp routes
	healthHandler := health.NewHandler()
	authHandler := apiAuth.NewHandler(cfg.AuthService, cfg.MagicLinkService, apiMiddleware, cfg.BaseURL, cfg.OAuthEnabled, cfg.MagicLinkEnabled)
	usersHandler := users.NewHandler(cfg.AdminEmails)
	webhooksHandler := apiAdmin.NewWebhooksHandler(cfg.WebhookRepository, cfg.WebhookDeliveryRepository)

	// Public routes
	r.Group(func(r chi.Router) {
		// Health check
		r.Get("/health", healthHandler.HandleHealth)

		// CSRF token
		r.Get("/csrf", authHandler.HandleGetCSRFToken)

		// Auth endpoints
		r.Route("/auth", func(r chi.Router) {
			// Public endpoint to expose available authentication methods
			r.Get("/config", authHandler.HandleGetAuthConfig)

			// Apply rate limiting to auth endpoints (except /config which should be fast)
			r.Group(func(r chi.Router) {
				r.Use(authRateLimit.Middleware)

				// OAuth endpoints (conditional)
				if cfg.OAuthEnabled {
					r.Post("/start", authHandler.HandleStartOAuth)
					r.Get("/callback", authHandler.HandleOAuthCallback)

					if cfg.AutoLogin {
						r.Get("/check", authHandler.HandleAuthCheck)
					}
				}

				// Magic Link endpoints (conditional)
				if cfg.MagicLinkEnabled {
					r.Post("/magic-link/request", authHandler.HandleRequestMagicLink)
					r.Get("/magic-link/verify", authHandler.HandleVerifyMagicLink)
					// Reminder auth link (authentification via email de reminder)
					r.Get("/reminder-link/verify", authHandler.HandleVerifyReminderAuthLink)
				}

				// Logout endpoint (available for both OAuth and MagicLink)
				if cfg.OAuthEnabled || cfg.MagicLinkEnabled {
					r.Get("/logout", authHandler.HandleLogout)
				}
			})
		})

		// Public document endpoints (from coreapp)
		coreGroups.RegisterPublic(r)
	})

	// User document routes with special handling
	// - Document creation requires CSRF + rate limiting
	// - find-or-create requires optional auth
	r.Group(func(r chi.Router) {
		r.Use(apiMiddleware.CSRFProtect)
		r.Use(documentRateLimit.Middleware)
		r.Use(apiMiddleware.OptionalAuth)
		coreGroups.RegisterUser(r)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(apiMiddleware.RequireAuth)
		r.Use(apiMiddleware.CSRFProtect)

		// User endpoints (non-coreapp)
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", usersHandler.HandleGetCurrentUser)
		})
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(apiMiddleware.RequireAdmin)
		r.Use(apiMiddleware.CSRFProtect)

		// Admin document routes (from coreapp)
		coreGroups.RegisterAdmin(r)

		// Webhooks management (non-coreapp)
		r.Route("/admin/webhooks", func(r chi.Router) {
			r.Get("/", webhooksHandler.HandleListWebhooks)
			r.Post("/", webhooksHandler.HandleCreateWebhook)
			r.Get("/{id}", webhooksHandler.HandleGetWebhook)
			r.Put("/{id}", webhooksHandler.HandleUpdateWebhook)
			r.Patch("/{id}/{action}", webhooksHandler.HandleToggleWebhook) // action: enable|disable
			r.Delete("/{id}", webhooksHandler.HandleDeleteWebhook)
			r.Get("/{id}/deliveries", webhooksHandler.HandleListDeliveries)
		})
	})

	// Serve OpenAPI spec
	r.Get("/openapi.json", serveOpenAPISpec)

	return r
}

// serveOpenAPISpec serves the OpenAPI specification
func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	// Read the OpenAPI YAML file and convert to JSON
	yamlData, err := os.ReadFile("openapi.yaml")
	if err != nil {
		// Fallback to basic response if file not found
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"info":{"title":"Ackify API","version":"1.0.0"},"message":"OpenAPI spec file not found - see /backend/openapi.yaml"}`))
		return
	}

	// Parse YAML and convert to JSON
	var spec map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &spec); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to parse OpenAPI spec"}`))
		return
	}

	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to convert OpenAPI spec to JSON"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
