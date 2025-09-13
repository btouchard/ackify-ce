package handlers

import (
	"context"
	"net/http"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type authService interface {
	SetUser(w http.ResponseWriter, r *http.Request, user *models.User) error
	Logout(w http.ResponseWriter, r *http.Request)
	GetAuthURL(nextURL string) string
	HandleCallback(ctx context.Context, code, state string) (*models.User, string, error)
}

type userService interface {
	GetUser(r *http.Request) (*models.User, error)
}
