// SPDX-License-Identifier: AGPL-3.0-or-later
package types

import "strings"

// User represents an authenticated user across all layers of the application.
// This is the canonical user representation used by auth providers, domain models,
// and API handlers.
type User struct {
	Sub   string `json:"sub"`   // Unique identifier (OAuth sub claim or email for MagicLink)
	Email string `json:"email"` // User's email address
	Name  string `json:"name"`  // Display name (optional)
}

// IsValid returns true if the user has required fields populated.
func (u *User) IsValid() bool {
	return strings.TrimSpace(u.Sub) != "" && strings.TrimSpace(u.Email) != ""
}

// NormalizedEmail returns the email in lowercase for comparison.
func (u *User) NormalizedEmail() string {
	return strings.ToLower(u.Email)
}
