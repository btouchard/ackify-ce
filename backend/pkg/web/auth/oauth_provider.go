// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"net/http"

	infraAuth "github.com/btouchard/ackify-ce/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/pkg/web"
)

// OAuthProvider adapts the internal OauthService to the web.OAuthAuthProvider interface.
// This allows the OAuth authentication to be used as a pluggable capability.
type OAuthProvider struct {
	service *infraAuth.OauthService
	enabled bool
}

// NewOAuthProvider creates a new OAuth provider adapter.
func NewOAuthProvider(service *infraAuth.OauthService, enabled bool) *OAuthProvider {
	return &OAuthProvider{
		service: service,
		enabled: enabled,
	}
}

// GetCurrentUser implements web.AuthProvider.
func (p *OAuthProvider) GetCurrentUser(r *http.Request) (*web.User, error) {
	if p.service == nil {
		return nil, web.ErrNotAuthenticated
	}

	user, err := p.service.GetUser(r)
	if err != nil {
		return nil, web.ErrNotAuthenticated
	}

	// No conversion needed: models.User = web.User = types.User
	return user, nil
}

// SetCurrentUser implements web.AuthProvider.
func (p *OAuthProvider) SetCurrentUser(w http.ResponseWriter, r *http.Request, user *web.User) error {
	if p.service == nil {
		return web.ErrProviderDisabled
	}

	// No conversion needed: models.User = web.User = types.User
	return p.service.SetUser(w, r, user)
}

// Logout implements web.AuthProvider.
func (p *OAuthProvider) Logout(w http.ResponseWriter, r *http.Request) {
	if p.service != nil {
		p.service.Logout(w, r)
	}
}

// IsConfigured implements web.AuthProvider.
func (p *OAuthProvider) IsConfigured() bool {
	return p.enabled && p.service != nil && p.service.OAuthProvider != nil
}

// CreateAuthURL implements web.OAuthAuthProvider.
func (p *OAuthProvider) CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string {
	if p.service == nil {
		return ""
	}
	return p.service.CreateAuthURL(w, r, nextURL)
}

// VerifyState implements web.OAuthAuthProvider.
func (p *OAuthProvider) VerifyState(w http.ResponseWriter, r *http.Request, stateToken string) bool {
	if p.service == nil {
		return false
	}
	return p.service.VerifyState(w, r, stateToken)
}

// HandleCallback implements web.OAuthAuthProvider.
func (p *OAuthProvider) HandleCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, code, state string) (*web.User, string, error) {
	if p.service == nil {
		return nil, "/", web.ErrProviderDisabled
	}

	// No conversion needed: models.User = web.User = types.User
	return p.service.HandleCallback(ctx, w, r, code, state)
}

// GetLogoutURL implements web.OAuthAuthProvider.
func (p *OAuthProvider) GetLogoutURL() string {
	if p.service == nil {
		return ""
	}
	return p.service.GetLogoutURL()
}

// IsAllowedDomain implements web.OAuthAuthProvider.
func (p *OAuthProvider) IsAllowedDomain(email string) bool {
	if p.service == nil {
		return true
	}
	return p.service.IsAllowedDomain(email)
}

// GetService returns the underlying OauthService for backward compatibility.
// This should only be used during migration; prefer using the interface methods.
func (p *OAuthProvider) GetService() *infraAuth.OauthService {
	return p.service
}

// Compile-time interface checks.
var (
	_ web.AuthProvider      = (*OAuthProvider)(nil)
	_ web.OAuthAuthProvider = (*OAuthProvider)(nil)
)
