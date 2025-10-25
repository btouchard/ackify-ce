// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/handlers"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// Handler handles authentication API requests
type Handler struct {
	authService *auth.OauthService
	middleware  *shared.Middleware
	baseURL     string
}

// NewHandler creates a new auth handler
func NewHandler(authService *auth.OauthService, middleware *shared.Middleware, baseURL string) *Handler {
	return &Handler{
		authService: authService,
		middleware:  middleware,
		baseURL:     baseURL,
	}
}

// HandleGetCSRFToken handles GET /api/v1/csrf
func (h *Handler) HandleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token, err := h.middleware.GenerateCSRFToken()
	if err != nil {
		shared.WriteInternalError(w)
		return
	}

	// Set cookie for the token
	http.SetCookie(w, &http.Cookie{
		Name:     shared.CSRFTokenCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Allow JS to read it
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	shared.WriteJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// HandleStartOAuth handles POST /api/v1/auth/start
func (h *Handler) HandleStartOAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RedirectTo string `json:"redirectTo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body, that's fine, use default redirect
		req.RedirectTo = "/"
	}

	// Default to home if no redirect specified
	if req.RedirectTo == "" {
		req.RedirectTo = "/"
	}

	// Generate OAuth URL and save state in session
	// This is critical - CreateAuthURL saves the state token in session
	// which will be validated when Google redirects to /api/v1/auth/callback
	authURL := h.authService.CreateAuthURL(w, r, req.RedirectTo)

	// Return redirect URL for SPA to handle
	shared.WriteJSON(w, http.StatusOK, map[string]string{
		"redirectUrl": authURL,
	})
}

func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	oauthError := r.URL.Query().Get("error")
	errorDescription := r.URL.Query().Get("error_description")

	logger.Logger.Debug("HandleOAuthCallback: received callback",
		"code_present", code != "",
		"state_present", state != "",
		"error", oauthError,
		"query_params", r.URL.Query().Encode())

	// Gérer les erreurs OAuth (ex: prompt=none sans session active)
	if oauthError != "" {
		logger.Logger.Debug("HandleOAuthCallback: OAuth error received",
			"error", oauthError,
			"description", errorDescription)

		// Si c'est une erreur de silent login (prompt=none), rediriger silencieusement
		if oauthError == "login_required" || oauthError == "interaction_required" || oauthError == "consent_required" {
			// Extraire next_url du state
			parts := strings.SplitN(state, ":", 2)
			nextURL := "/"
			if len(parts) == 2 {
				if nb, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
					nextURL = string(nb)
				}
			}

			logger.Logger.Debug("HandleOAuthCallback: silent login failed, redirecting to original URL",
				"next_url", nextURL)
			http.Redirect(w, r, nextURL, http.StatusFound)
			return
		}

		// Pour d'autres erreurs, afficher un message
		http.Error(w, "OAuth error: "+oauthError, http.StatusBadRequest)
		return
	}

	if code == "" {
		logger.Logger.Warn("HandleOAuthCallback: missing authorization code")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Validate OAuth state for CSRF protection
	parts := strings.SplitN(state, ":", 2)
	token := ""
	if len(parts) > 0 {
		token = parts[0]
	}

	logger.Logger.Debug("HandleOAuthCallback: validating state",
		"token_length", len(token),
		"state_parts", len(parts))

	if token == "" || !h.authService.VerifyState(w, r, token) {
		logger.Logger.Warn("HandleOAuthCallback: invalid OAuth state",
			"token_empty", token == "")
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, nextURL, err := h.authService.HandleCallback(ctx, w, r, code, state)
	if err != nil {
		logger.Logger.Error("OAuth callback failed", "error", err.Error())
		handlers.HandleError(w, err)
		return
	}

	logger.Logger.Debug("HandleOAuthCallback: user authenticated",
		"user_email", user.Email,
		"next_url", nextURL)

	if err := h.authService.SetUser(w, r, user); err != nil {
		logger.Logger.Error("HandleOAuthCallback: failed to set user session", "error", err.Error())
		http.Error(w, "Failed to set user session", http.StatusInternalServerError)
		return
	}

	if nextURL == "" {
		nextURL = "/"
	}

	if parsedURL, err := url.Parse(nextURL); err != nil ||
		(parsedURL.Host != "" && parsedURL.Host != r.Host) {
		logger.Logger.Debug("HandleOAuthCallback: invalid nextURL, using /",
			"original_next", nextURL,
			"parse_error", err != nil)
		nextURL = "/"
	}

	logger.Logger.Debug("HandleOAuthCallback: redirecting user",
		"final_next_url", nextURL)

	http.Redirect(w, r, nextURL, http.StatusFound)
}

// HandleLogout handles GET /api/v1/auth/logout
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session
	h.authService.Logout(w, r)

	// Check if SSO logout is configured
	logoutURL := h.authService.GetLogoutURL()
	if logoutURL != "" {
		returnURL := h.baseURL + "/"
		fullLogoutURL := logoutURL + "?post_logout_redirect_uri=" + url.QueryEscape(returnURL)

		shared.WriteJSON(w, http.StatusOK, map[string]string{
			"message":     "Successfully logged out",
			"redirectUrl": fullLogoutURL,
		})
	} else {
		shared.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "Successfully logged out",
		})
	}
}

// HandleAuthCheck handles GET /api/v1/auth/check
func (h *Handler) HandleAuthCheck(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.GetUser(r)
	if err != nil || user == nil {
		shared.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"user": map[string]interface{}{
			"id":    user.Sub,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}
