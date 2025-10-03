// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
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

	// Persist CSRF state in session when generating auth URL
	authURL := h.authService.CreateAuthURL(w, r, next)
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

func (h *AuthHandlers) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Validate OAuth state for CSRF protection
	parts := strings.SplitN(state, ":", 2)
	token := ""
	if len(parts) > 0 {
		token = parts[0]
	}
	if token == "" || !h.authService.VerifyState(w, r, token) {
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

	if err := h.authService.SetUser(w, r, user); err != nil {
		http.Error(w, "Failed to set user session", http.StatusInternalServerError)
		return
	}

	if nextURL == "" {
		nextURL = "/"
	}

	if parsedURL, err := url.Parse(nextURL); err != nil ||
		(parsedURL.Host != "" && parsedURL.Host != r.Host) {
		nextURL = "/"
	}

	http.Redirect(w, r, nextURL, http.StatusFound)
}
