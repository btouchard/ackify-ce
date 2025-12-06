// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"net/http"

	"github.com/btouchard/ackify-ce/internal/application/services"
	infraAuth "github.com/btouchard/ackify-ce/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/pkg/web"
)

// MagicLinkProvider adapts the MagicLinkService to the web.MagicLinkAuthProvider interface.
// It requires a SessionService for session management (shared with OAuth if both are enabled).
type MagicLinkProvider struct {
	service        *services.MagicLinkService
	sessionService *infraAuth.SessionService
	enabled        bool
}

// NewMagicLinkProvider creates a new MagicLink provider adapter.
func NewMagicLinkProvider(service *services.MagicLinkService, sessionService *infraAuth.SessionService, enabled bool) *MagicLinkProvider {
	return &MagicLinkProvider{
		service:        service,
		sessionService: sessionService,
		enabled:        enabled,
	}
}

// GetCurrentUser implements web.AuthProvider.
func (p *MagicLinkProvider) GetCurrentUser(r *http.Request) (*web.User, error) {
	if p.sessionService == nil {
		return nil, web.ErrNotAuthenticated
	}

	user, err := p.sessionService.GetUser(r)
	if err != nil {
		return nil, web.ErrNotAuthenticated
	}

	// No conversion needed: models.User = web.User = types.User
	return user, nil
}

// SetCurrentUser implements web.AuthProvider.
func (p *MagicLinkProvider) SetCurrentUser(w http.ResponseWriter, r *http.Request, user *web.User) error {
	if p.sessionService == nil {
		return web.ErrProviderDisabled
	}

	// No conversion needed: models.User = web.User = types.User
	return p.sessionService.SetUser(w, r, user)
}

// Logout implements web.AuthProvider.
func (p *MagicLinkProvider) Logout(w http.ResponseWriter, r *http.Request) {
	if p.sessionService != nil {
		p.sessionService.Logout(w, r)
	}
}

// IsConfigured implements web.AuthProvider.
func (p *MagicLinkProvider) IsConfigured() bool {
	return p.enabled && p.service != nil && p.sessionService != nil
}

// RequestMagicLink implements web.MagicLinkAuthProvider.
func (p *MagicLinkProvider) RequestMagicLink(ctx context.Context, email, redirectTo, ip, userAgent, locale string) error {
	if p.service == nil {
		return web.ErrProviderDisabled
	}
	return p.service.RequestMagicLink(ctx, email, redirectTo, ip, userAgent, locale)
}

// VerifyMagicLink implements web.MagicLinkAuthProvider.
func (p *MagicLinkProvider) VerifyMagicLink(ctx context.Context, token, ip, userAgent string) (*web.MagicLinkResult, error) {
	if p.service == nil {
		return nil, web.ErrProviderDisabled
	}

	magicToken, err := p.service.VerifyMagicLink(ctx, token, ip, userAgent)
	if err != nil {
		return nil, err
	}

	return &web.MagicLinkResult{
		Email:      magicToken.Email,
		RedirectTo: magicToken.RedirectTo,
		DocID:      magicToken.DocID,
	}, nil
}

// VerifyReminderAuthToken implements web.MagicLinkAuthProvider.
func (p *MagicLinkProvider) VerifyReminderAuthToken(ctx context.Context, token, ip, userAgent string) (*web.MagicLinkResult, error) {
	if p.service == nil {
		return nil, web.ErrProviderDisabled
	}

	magicToken, err := p.service.VerifyReminderAuthToken(ctx, token, ip, userAgent)
	if err != nil {
		return nil, err
	}

	return &web.MagicLinkResult{
		Email:      magicToken.Email,
		RedirectTo: magicToken.RedirectTo,
		DocID:      magicToken.DocID,
	}, nil
}

// CreateReminderAuthToken implements web.MagicLinkAuthProvider.
func (p *MagicLinkProvider) CreateReminderAuthToken(ctx context.Context, email, docID string) (string, error) {
	if p.service == nil {
		return "", web.ErrProviderDisabled
	}
	return p.service.CreateReminderAuthToken(ctx, email, docID)
}

// GetService returns the underlying MagicLinkService for backward compatibility.
func (p *MagicLinkProvider) GetService() *services.MagicLinkService {
	return p.service
}

// GetSessionService returns the underlying SessionService for backward compatibility.
func (p *MagicLinkProvider) GetSessionService() *infraAuth.SessionService {
	return p.sessionService
}

// Compile-time interface checks.
var (
	_ web.AuthProvider          = (*MagicLinkProvider)(nil)
	_ web.MagicLinkAuthProvider = (*MagicLinkProvider)(nil)
)
