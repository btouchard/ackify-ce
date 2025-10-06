// SPDX-License-Identifier: AGPL-3.0-or-later
package handlers

import (
	"context"
	"net/http"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type authService interface {
	GetUser(r *http.Request) (*models.User, error)
	SetUser(w http.ResponseWriter, r *http.Request, user *models.User) error
	Logout(w http.ResponseWriter, r *http.Request)
	GetLogoutURL() string
	GetAuthURL(nextURL string) string
	CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string
	VerifyState(w http.ResponseWriter, r *http.Request, stateToken string) bool
	HandleCallback(ctx context.Context, code, state string) (*models.User, string, error)
}

type userService interface {
	GetUser(r *http.Request) (*models.User, error)
}
