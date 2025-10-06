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
		// Initialize repositories by reusing the existing DB connection
		adminRepo := database.NewAdminRepository(db)
		expectedSignerRepo := database.NewExpectedSignerRepository(db)

		// Initialize middleware and handlers
		adminMiddleware := NewAdminMiddleware(authService, cfg.App.BaseURL, cfg.App.AdminEmails, templates)
		adminHandlers := NewAdminHandlers(adminRepo, authService, templates, cfg.App.BaseURL)
		expectedHandlers := NewExpectedSignersHandlers(expectedSignerRepo, adminRepo, authService, templates, cfg.App.BaseURL)

		// Register admin routes
		r.Get("/admin", adminMiddleware.RequireAdmin(adminHandlers.HandleDashboard))
		r.Get("/admin/docs/{docID}", adminMiddleware.RequireAdmin(expectedHandlers.HandleDocumentDetailsWithExpected))
		r.Post("/admin/docs/{docID}/expected", adminMiddleware.RequireAdmin(expectedHandlers.HandleAddExpectedSigners))
		r.Post("/admin/docs/{docID}/expected/remove", adminMiddleware.RequireAdmin(expectedHandlers.HandleRemoveExpectedSigner))
		r.Get("/admin/docs/{docID}/status.json", adminMiddleware.RequireAdmin(expectedHandlers.HandleGetDocumentStatusJSON))
		r.Get("/admin/api/chain-integrity/{docID}", adminMiddleware.RequireAdmin(adminHandlers.HandleChainIntegrityAPI))
	}
}
