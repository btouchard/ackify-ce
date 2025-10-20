// SPDX-License-Identifier: AGPL-3.0-or-later
package users

import (
	"net/http"
	"strings"

	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
)

// Handler handles user API requests
type Handler struct {
	adminEmails []string
}

// NewHandler creates a new users handler
func NewHandler(adminEmails []string) *Handler {
	return &Handler{
		adminEmails: adminEmails,
	}
}

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
}

// HandleGetCurrentUser handles GET /api/v1/users/me
func (h *Handler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := shared.GetUserFromContext(r.Context())
	if !ok {
		shared.WriteUnauthorized(w, "")
		return
	}

	// Check if user is admin
	isAdmin := false
	for _, adminEmail := range h.adminEmails {
		if strings.EqualFold(user.Email, adminEmail) {
			isAdmin = true
			break
		}
	}

	userDTO := UserDTO{
		ID:      user.Sub,
		Email:   user.Email,
		Name:    user.Name,
		IsAdmin: isAdmin,
	}

	shared.WriteJSON(w, http.StatusOK, userDTO)
}
