// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"strings"

	"github.com/btouchard/ackify-ce/backend/pkg/web"
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

// IsAdmin implements web.Authorizer.
func (a *SimpleAuthorizer) IsAdmin(_ context.Context, userEmail string) bool {
	normalized := strings.ToLower(strings.TrimSpace(userEmail))
	return a.adminEmails[normalized]
}

// CanCreateDocument implements web.Authorizer.
func (a *SimpleAuthorizer) CanCreateDocument(ctx context.Context, userEmail string) bool {
	if !a.onlyAdminCanCreate {
		return true
	}
	return a.IsAdmin(ctx, userEmail)
}

// OnlyAdminCanCreate returns whether only administrators can create documents.
// This is provided for backward compatibility with existing code.
func (a *SimpleAuthorizer) OnlyAdminCanCreate() bool {
	return a.onlyAdminCanCreate
}

// Compile-time interface check.
var _ web.Authorizer = (*SimpleAuthorizer)(nil)
