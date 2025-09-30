// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"database/sql"
	"html/template"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
)

// RegisterAdminRoutes returns a function that registers admin routes
func RegisterAdminRoutes(cfg *config.Config, templates *template.Template, db *sql.DB, authService userService) func(r *chi.Mux) {
	return func(r *chi.Mux) {
		// Initialize admin repository by reusing the existing DB connection
		adminRepo := database.NewAdminRepository(db)

		// Initialize middleware and handlers
		adminMiddleware := NewAdminMiddleware(authService, cfg.App.BaseURL, cfg.App.AdminEmails, templates)
		adminHandlers := NewAdminHandlers(adminRepo, authService, templates, cfg.App.BaseURL)

		// Register admin routes
		r.Get("/admin", adminMiddleware.RequireAdmin(adminHandlers.HandleDashboard))
		r.Get("/admin/docs/{docID}", adminMiddleware.RequireAdmin(adminHandlers.HandleDocumentDetails))
		r.Get("/admin/api/chain-integrity/{docID}", adminMiddleware.RequireAdmin(adminHandlers.HandleChainIntegrityAPI))
	}
}
