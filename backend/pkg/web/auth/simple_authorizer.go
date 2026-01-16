// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"strings"

	"github.com/btouchard/ackify-ce/backend/pkg/providers"
)

// SimpleAuthorizer is an authorization implementation based on a list of admin emails.
// This is the default authorizer for Community Edition.
type SimpleAuthorizer struct {
	adminEmails        map[string]bool
	onlyAdminCanCreate bool
}

// NewSimpleAuthorizer creates a new simple authorizer.
func NewSimpleAuthorizer(adminEmails []string, onlyAdminCanCreate bool) *SimpleAuthorizer {
	emailMap := make(map[string]bool, len(adminEmails))
	for _, email := range adminEmails {
		normalized := strings.ToLower(strings.TrimSpace(email))
		if normalized != "" {
			emailMap[normalized] = true
		}
	}
	return &SimpleAuthorizer{
		adminEmails:        emailMap,
		onlyAdminCanCreate: onlyAdminCanCreate,
	}
}

// IsAdmin implements providers.Authorizer.
func (a *SimpleAuthorizer) IsAdmin(_ context.Context, userEmail string) bool {
	normalized := strings.ToLower(strings.TrimSpace(userEmail))
	return a.adminEmails[normalized]
}

// CanCreateDocument implements providers.Authorizer.
func (a *SimpleAuthorizer) CanCreateDocument(ctx context.Context, userEmail string) bool {
	if !a.onlyAdminCanCreate {
		return true
	}
	return a.IsAdmin(ctx, userEmail)
}

// Compile-time interface check.
var _ providers.Authorizer = (*SimpleAuthorizer)(nil)
