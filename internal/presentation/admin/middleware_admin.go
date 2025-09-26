package admin

import (
	"net/http"
	"os"
	"strings"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type userService interface {
	GetUser(r *http.Request) (*models.User, error)
}

// AdminMiddleware provides admin authentication middleware
type AdminMiddleware struct {
	userService userService
	baseURL     string
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(userService userService, baseURL string) *AdminMiddleware {
	return &AdminMiddleware{
		userService: userService,
		baseURL:     baseURL,
	}
}

// RequireAdmin wraps a handler to require admin authentication
func (m *AdminMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := m.userService.GetUser(r)
		if err != nil {
			nextURL := m.baseURL + r.URL.RequestURI()
			loginURL := "/login?next=" + nextURL
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		if !m.isAdminUser(user) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// isAdminUser checks if the user is in the admin emails list
func (m *AdminMiddleware) isAdminUser(user *models.User) bool {
	adminEmails := os.Getenv("ACKIFY_ADMIN_EMAILS")
	if adminEmails == "" {
		return false
	}

	userEmail := strings.ToLower(strings.TrimSpace(user.Email))
	emails := strings.Split(adminEmails, ",")

	for _, email := range emails {
		adminEmail := strings.ToLower(strings.TrimSpace(email))
		if userEmail == adminEmail {
			return true
		}
	}

	return false
}

// IsAdminUser is a public function to check if a user is admin (for templates)
func IsAdminUser(user *models.User) bool {
	if user == nil {
		return false
	}

	adminEmails := os.Getenv("ACKIFY_ADMIN_EMAILS")
	if adminEmails == "" {
		return false
	}

	userEmail := strings.ToLower(strings.TrimSpace(user.Email))
	emails := strings.Split(adminEmails, ",")

	for _, email := range emails {
		adminEmail := strings.ToLower(strings.TrimSpace(email))
		if userEmail == adminEmail {
			return true
		}
	}

	return false
}
