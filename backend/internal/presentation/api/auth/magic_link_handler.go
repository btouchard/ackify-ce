// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"encoding/json"
	"net/http"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// HandleRequestMagicLink handles POST /api/v1/auth/magic-link/request
func (h *Handler) HandleRequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email      string `json:"email"`
		RedirectTo string `json:"redirectTo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}

	if req.Email == "" {
		shared.WriteError(w, http.StatusBadRequest, "missing_email", "Email is required", nil)
		return
	}

	if req.RedirectTo == "" {
		req.RedirectTo = "/"
	}

	// Extraire IP et User-Agent
	ip := shared.GetClientIP(r)
	userAgent := r.UserAgent()

	// Demander le Magic Link
	err := h.magicLinkService.RequestMagicLink(r.Context(), req.Email, req.RedirectTo, ip, userAgent)

	// IMPORTANT: Ne jamais révéler si l'email existe ou non (protection contre énumération)
	// Toujours retourner succès, même en cas d'erreur de rate limiting
	if err != nil {
		logger.Logger.Error("Magic Link request failed", "email", req.Email, "error", err)
		// On log l'erreur mais on retourne succès au client
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "If the email is valid, a magic link has been sent",
	})
}

// HandleVerifyMagicLink handles GET /api/v1/auth/magic-link/verify
func (h *Handler) HandleVerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	// Extraire IP et User-Agent
	ip := shared.GetClientIP(r)
	userAgent := r.UserAgent()

	// Vérifier le token
	magicToken, err := h.magicLinkService.VerifyMagicLink(r.Context(), token, ip, userAgent)
	if err != nil {
		logger.Logger.Warn("Magic Link verification failed", "error", err, "ip", ip)
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Créer une session utilisateur
	user := &models.User{
		Sub:   magicToken.Email, // Utiliser email comme sub
		Email: magicToken.Email,
		Name:  magicToken.Email, // Par défaut, nom = email
	}

	// Sauvegarder dans la session (réutiliser la logique OAuth existante)
	if err := h.authService.SetUser(w, r, user); err != nil {
		logger.Logger.Error("Failed to create session after Magic Link", "error", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	logger.Logger.Info("User authenticated via Magic Link",
		"email", magicToken.Email,
		"redirect_to", magicToken.RedirectTo)

	// Rediriger vers la destination demandée
	http.Redirect(w, r, magicToken.RedirectTo, http.StatusFound)
}
