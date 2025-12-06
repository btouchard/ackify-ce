// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"net/http"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

// HandleVerifyReminderAuthLink handles GET /api/v1/auth/reminder-link/verify
// This endpoint authenticates a user via a reminder auth token and redirects to the document signature page
func (h *Handler) HandleVerifyReminderAuthLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := r.URL.Query().Get("token")

	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	// Extraire IP et User-Agent
	ip := shared.GetClientIP(r)
	userAgent := r.UserAgent()

	// Vérifier le token
	magicToken, err := h.magicLinkService.VerifyReminderAuthToken(ctx, token, ip, userAgent)
	if err != nil {
		logger.Logger.Warn("Reminder auth link verification failed", "error", err, "ip", ip)
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Vérifier si l'utilisateur est déjà authentifié avec le bon compte
	if user, ok := shared.GetUserFromContext(ctx); ok && user.Email == magicToken.Email {
		// Déjà connecté avec le bon compte → redirection directe
		logger.Logger.Info("User already authenticated with correct account",
			"email", magicToken.Email,
			"doc_id", magicToken.DocID)
		http.Redirect(w, r, magicToken.RedirectTo, http.StatusFound)
		return
	}

	// Créer une session utilisateur
	user := &models.User{
		Sub:   magicToken.Email, // Utiliser email comme sub
		Email: magicToken.Email,
		Name:  magicToken.Email, // Par défaut, nom = email
	}

	// Sauvegarder dans la session (réutiliser la logique OAuth existante)
	if err := h.authProvider.SetCurrentUser(w, r, user); err != nil {
		logger.Logger.Error("Failed to create session after reminder auth", "error", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("User authenticated via reminder auth link",
		"email", magicToken.Email,
		"doc_id", magicToken.DocID,
		"redirect_to", magicToken.RedirectTo)

	// Rediriger vers la destination demandée (page de signature)
	http.Redirect(w, r, magicToken.RedirectTo, http.StatusFound)
}
