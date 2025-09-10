package models

import "strings"

// User represents an authenticated user
type User struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// IsValid checks if the user has valid credentials
func (u *User) IsValid() bool {
	return strings.TrimSpace(u.Sub) != "" && strings.TrimSpace(u.Email) != ""
}

// NormalizedEmail returns the email in lowercase
func (u *User) NormalizedEmail() string {
	return strings.ToLower(u.Email)
}
