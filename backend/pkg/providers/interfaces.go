// SPDX-License-Identifier: AGPL-3.0-or-later
// Package providers defines capability interfaces for dependency injection.
// These interfaces are in a separate package to avoid import cycles.
package providers

import (
	"context"
	"net/http"

	"github.com/btouchard/ackify-ce/backend/pkg/types"
)

// Common errors for capability providers.
// Defined as strings to avoid import cycles - implementations can wrap these.
const (
	ErrNotAuthenticatedMsg = "user not authenticated"
	ErrNotAuthorizedMsg    = "user not authorized"
	ErrQuotaExceededMsg    = "quota exceeded"
	ErrProviderDisabledMsg = "provider is disabled"
)

// AuthProvider defines the interface for authentication providers.
// Implementations: OAuth2Provider, MagicLinkProvider, CompositeAuthProvider (CE),
// Auth0Provider, KeycloakProvider (SaaS), etc.
type AuthProvider interface {
	// GetCurrentUser returns the authenticated user from the request context/session.
	// Returns error if no user is authenticated.
	GetCurrentUser(r *http.Request) (*types.User, error)

	// SetCurrentUser stores the authenticated user in the session.
	SetCurrentUser(w http.ResponseWriter, r *http.Request, user *types.User) error

	// Logout clears the user session.
	Logout(w http.ResponseWriter, r *http.Request)

	// IsConfigured returns true if this provider is properly configured and enabled.
	IsConfigured() bool
}

// OAuthAuthProvider extends AuthProvider with OAuth2-specific methods.
// Used when OAuth2 authentication is enabled.
type OAuthAuthProvider interface {
	AuthProvider

	// CreateAuthURL generates the OAuth2 authorization URL.
	// The nextURL parameter specifies where to redirect after successful auth.
	CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string

	// VerifyState verifies the OAuth2 state token to prevent CSRF.
	VerifyState(w http.ResponseWriter, r *http.Request, stateToken string) bool

	// HandleCallback processes the OAuth2 callback.
	// Returns the authenticated user and the redirect URL.
	HandleCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, code, state string) (*types.User, string, error)

	// GetLogoutURL returns the OAuth2 provider's logout URL if available.
	GetLogoutURL() string

	// IsAllowedDomain checks if the email domain is allowed.
	IsAllowedDomain(email string) bool
}

// Authorizer defines the interface for authorization decisions.
// CE: SimpleAuthorizer based on admin email list.
// SaaS: RBACAuthorizer with roles and permissions.
type Authorizer interface {
	// IsAdmin returns true if the user is an administrator.
	IsAdmin(ctx context.Context, userEmail string) bool

	// CanCreateDocument returns true if the user can create documents.
	CanCreateDocument(ctx context.Context, userEmail string) bool
}
