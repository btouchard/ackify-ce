// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/btouchard/ackify-ce/backend/pkg/providers"
)

// Common errors for capability providers.
var (
	ErrNotAuthenticated = errors.New("user not authenticated")
	ErrNotAuthorized    = errors.New("user not authorized")
	ErrQuotaExceeded    = errors.New("quota exceeded")
	ErrProviderDisabled = errors.New("provider is disabled")
)

// Re-export interfaces from pkg/providers for backward compatibility.
// This allows pkg/web users to continue using web.AuthProvider, etc.
type AuthProvider = providers.AuthProvider
type OAuthAuthProvider = providers.OAuthAuthProvider
type Authorizer = providers.Authorizer

// MagicLinkAuthProvider extends AuthProvider with magic link-specific methods.
// Used when magic link authentication is enabled.
type MagicLinkAuthProvider interface {
	providers.AuthProvider

	// RequestMagicLink sends a magic link to the specified email.
	RequestMagicLink(ctx context.Context, email, redirectTo, ip, userAgent, locale string) error

	// VerifyMagicLink verifies a magic link token and returns the associated user info.
	VerifyMagicLink(ctx context.Context, token, ip, userAgent string) (*MagicLinkResult, error)

	// VerifyReminderAuthToken verifies a reminder auth token.
	VerifyReminderAuthToken(ctx context.Context, token, ip, userAgent string) (*MagicLinkResult, error)

	// CreateReminderAuthToken creates an auth token for reminder emails.
	CreateReminderAuthToken(ctx context.Context, email, docID string) (string, error)
}

// MagicLinkResult represents the result of verifying a magic link.
type MagicLinkResult struct {
	Email      string
	RedirectTo string
	DocID      *string // Non-nil for reminder auth tokens
}

// QuotaEnforcer defines the interface for quota management.
// CE: NoLimitQuotaEnforcer (no limits).
// SaaS: PlanBasedQuotaEnforcer (limits based on subscription plan).
type QuotaEnforcer interface {
	// Check verifies if the action is allowed under current quotas.
	// Returns ErrQuotaExceeded if the quota would be exceeded.
	Check(ctx context.Context, tenantID string, action QuotaAction) error

	// Record records that an action was performed (for tracking usage).
	// Should be called after the action succeeds.
	Record(ctx context.Context, tenantID string, action QuotaAction) error

	// GetUsage returns the current usage metrics for a tenant.
	GetUsage(ctx context.Context, tenantID string) (*QuotaUsage, error)
}

// AuditLogger defines the interface for audit logging.
// CE: LogOnlyAuditLogger (logs to standard logger).
// SaaS: DatabaseAuditLogger (stores in database with search/export).
type AuditLogger interface {
	// Log records an audit event.
	Log(ctx context.Context, event AuditEvent) error
}

// CompositeAuthProvider combines multiple auth providers (OAuth + MagicLink).
// This is the typical setup for CE where both methods may be enabled.
type CompositeAuthProvider struct {
	OAuth     OAuthAuthProvider
	MagicLink MagicLinkAuthProvider
	// Primary provider for GetCurrentUser (uses session which is shared)
	sessionProvider AuthProvider
}

// NewCompositeAuthProvider creates a new composite auth provider.
func NewCompositeAuthProvider(oauth OAuthAuthProvider, magicLink MagicLinkAuthProvider, sessionProvider AuthProvider) *CompositeAuthProvider {
	return &CompositeAuthProvider{
		OAuth:           oauth,
		MagicLink:       magicLink,
		sessionProvider: sessionProvider,
	}
}

// GetCurrentUser implements AuthProvider.
func (c *CompositeAuthProvider) GetCurrentUser(r *http.Request) (*User, error) {
	if c.sessionProvider != nil {
		return c.sessionProvider.GetCurrentUser(r)
	}
	// Fallback to OAuth if available
	if c.OAuth != nil && c.OAuth.IsConfigured() {
		return c.OAuth.GetCurrentUser(r)
	}
	// Fallback to MagicLink if available
	if c.MagicLink != nil && c.MagicLink.IsConfigured() {
		return c.MagicLink.GetCurrentUser(r)
	}
	return nil, ErrNotAuthenticated
}

// SetCurrentUser implements AuthProvider.
func (c *CompositeAuthProvider) SetCurrentUser(w http.ResponseWriter, r *http.Request, user *User) error {
	if c.sessionProvider != nil {
		return c.sessionProvider.SetCurrentUser(w, r, user)
	}
	if c.OAuth != nil && c.OAuth.IsConfigured() {
		return c.OAuth.SetCurrentUser(w, r, user)
	}
	if c.MagicLink != nil && c.MagicLink.IsConfigured() {
		return c.MagicLink.SetCurrentUser(w, r, user)
	}
	return ErrProviderDisabled
}

// Logout implements AuthProvider.
func (c *CompositeAuthProvider) Logout(w http.ResponseWriter, r *http.Request) {
	if c.sessionProvider != nil {
		c.sessionProvider.Logout(w, r)
		return
	}
	if c.OAuth != nil && c.OAuth.IsConfigured() {
		c.OAuth.Logout(w, r)
	}
}

// IsConfigured implements AuthProvider.
func (c *CompositeAuthProvider) IsConfigured() bool {
	return (c.OAuth != nil && c.OAuth.IsConfigured()) ||
		(c.MagicLink != nil && c.MagicLink.IsConfigured())
}

// OAuthEnabled returns true if OAuth is configured.
func (c *CompositeAuthProvider) OAuthEnabled() bool {
	return c.OAuth != nil && c.OAuth.IsConfigured()
}

// MagicLinkEnabled returns true if MagicLink is configured.
func (c *CompositeAuthProvider) MagicLinkEnabled() bool {
	return c.MagicLink != nil && c.MagicLink.IsConfigured()
}
