// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/btouchard/ackify-ce/pkg/logger"
)

type AuthHandlers struct {
	authService authService
	baseURL     string
}

func NewAuthHandlers(authService authService, baseURL string) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		baseURL:     baseURL,
	}
}

func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = h.baseURL + "/"
	}

	logger.Logger.Debug("HandleLogin: starting OAuth flow",
		"next_url", next,
		"query_params", r.URL.Query().Encode())

	// Persist CSRF state in session when generating auth URL
	authURL := h.authService.CreateAuthURL(w, r, next)

	logger.Logger.Debug("HandleLogin: redirecting to OAuth provider",
		"auth_url", authURL)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.authService.Logout(w, r)

	// Redirect to SSO logout if configured, otherwise redirect to home
	ssoLogoutURL := h.authService.GetLogoutURL()
	if ssoLogoutURL != "" {
		http.Redirect(w, r, ssoLogoutURL, http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandlers) HandleAuthCheck(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.GetUser(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"authenticated":false}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"authenticated": true,
		"user": map[string]string{
			"email": user.Email,
			"name":  user.Name,
		},
	}

	if jsonBytes, err := json.Marshal(response); err == nil {
		w.Write(jsonBytes)
	} else {
		w.Write([]byte(`{"authenticated":false}`))
	}
}

func (h *AuthHandlers) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
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
	user, nextURL, err := h.authService.HandleCallback(ctx, code, state)
	if err != nil {
		logger.Logger.Error("OAuth callback failed", "error", err.Error())
		HandleError(w, err)
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
