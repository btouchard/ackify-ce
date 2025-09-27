// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
    "database/sql"
    "html/template"
    "log"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
)

// RegisterAdminRoutes returns a function that registers admin routes
func RegisterAdminRoutes(baseURL string, templates *template.Template, db *sql.DB) func(r *chi.Mux) {
    return func(r *chi.Mux) {

        // Initialize admin repository by reusing the existing DB connection
        adminRepo := database.NewAdminRepositoryFromDB(db)

        // Initialize OAuth service for user authentication
        cfg, err := config.Load()
        if err != nil {
            log.Printf("Failed to load config for admin routes: %v", err)
			return
		}

		authService := auth.NewOAuthService(auth.Config{
			BaseURL:       cfg.App.BaseURL,
			ClientID:      cfg.OAuth.ClientID,
			ClientSecret:  cfg.OAuth.ClientSecret,
			AuthURL:       cfg.OAuth.AuthURL,
			TokenURL:      cfg.OAuth.TokenURL,
			UserInfoURL:   cfg.OAuth.UserInfoURL,
			Scopes:        cfg.OAuth.Scopes,
			AllowedDomain: cfg.OAuth.AllowedDomain,
			CookieSecret:  cfg.OAuth.CookieSecret,
			SecureCookies: cfg.App.SecureCookies,
		})

		// Initialize middleware and handlers
		adminMiddleware := NewAdminMiddleware(authService, baseURL)
		adminHandlers := NewAdminHandlers(adminRepo, authService, templates, baseURL)

		// Register admin routes
		r.Get("/admin", adminMiddleware.RequireAdmin(adminHandlers.HandleDashboard))
		r.Get("/admin/docs/{docID}", adminMiddleware.RequireAdmin(adminHandlers.HandleDocumentDetails))
		r.Get("/admin/api/chain-integrity/{docID}", adminMiddleware.RequireAdmin(adminHandlers.HandleChainIntegrityAPI))
	}
}
