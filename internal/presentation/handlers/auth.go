package handlers

import (
	"net/http"
	"net/url"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService authService
	baseURL     string
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService authService, baseURL string) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		baseURL:     baseURL,
	}
}

// HandleLogin handles login requests
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	if next == "" {
		next = h.baseURL + "/"
	}

	authURL := h.authService.GetAuthURL(next)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleLogout handles logout requests
func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.authService.Logout(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleOAuthCallback handles OAuth callback from the configured provider
func (h *AuthHandlers) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, nextURL, err := h.authService.HandleCallback(ctx, code, state)
	if err != nil {
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
