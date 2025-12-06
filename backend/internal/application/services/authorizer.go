// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"strings"
)

// AuthorizerService provides authorization decisions for the organization.
// In Community Edition, it uses environment variables for configuration.
type AuthorizerService struct {
	adminEmails        []string
	onlyAdminCanCreate bool
}

// NewAuthorizerService creates a new AuthorizerService with the given configuration.
func NewAuthorizerService(adminEmails []string, onlyAdminCanCreate bool) *AuthorizerService {
	// Normalize admin emails to lowercase for case-insensitive comparison
	normalized := make([]string, len(adminEmails))
	for i, email := range adminEmails {
		normalized[i] = strings.ToLower(strings.TrimSpace(email))
	}
	return &AuthorizerService{
		adminEmails:        normalized,
		onlyAdminCanCreate: onlyAdminCanCreate,
	}
}

// IsAdmin checks if the given email belongs to an administrator.
func (s *AuthorizerService) IsAdmin(email string) bool {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	for _, adminEmail := range s.adminEmails {
		if normalizedEmail == adminEmail {
			return true
		}
	}
	return false
}

// OnlyAdminCanCreate returns whether only administrators can create documents.
func (s *AuthorizerService) OnlyAdminCanCreate() bool {
	return s.onlyAdminCanCreate
}
