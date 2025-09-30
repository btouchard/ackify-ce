// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

type userService interface {
	GetUser(r *http.Request) (*models.User, error)
}

type Middleware struct {
	userService userService
	baseURL     string
	adminEmails []string
	templates   *template.Template
}

func NewAdminMiddleware(userService userService, baseURL string, adminEmails []string, templates *template.Template) *Middleware {
	return &Middleware{
		userService: userService,
		baseURL:     baseURL,
		adminEmails: adminEmails,
		templates:   templates,
	}
}

func (m *Middleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := m.userService.GetUser(r)
		if err != nil {
			nextURL := m.baseURL + r.URL.RequestURI()
			loginURL := "/login?next=" + nextURL
			http.Redirect(w, r, loginURL, http.StatusFound)
			return
		}

		if !m.isAdminUser(user) {
			m.renderForbidden(w, user)
			return
		}

		next(w, r)
	}
}

func (m *Middleware) isAdminUser(user *models.User) bool {
	if len(m.adminEmails) == 0 {
		logger.Logger.Warn("Admin access denied: no admin emails configured")
		return false
	}

	userEmail := strings.ToLower(strings.TrimSpace(user.Email))

	logger.Logger.Debug("Admin access check",
		"user_email", userEmail,
		"configured_admins", m.adminEmails,
		"admin_count", len(m.adminEmails))

	for _, email := range m.adminEmails {
		adminEmail := strings.ToLower(strings.TrimSpace(email))
		if userEmail == adminEmail {
			logger.Logger.Info("Admin access granted", "user_email", userEmail)
			return true
		}
	}

	logger.Logger.Warn("Admin access denied: email not in admin list", "user_email", userEmail)
	return false
}

func (m *Middleware) renderForbidden(w http.ResponseWriter, user *models.User) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		TemplateName string
		User         *models.User
		BaseURL      string
		Year         int
		IsAdmin      bool
		ErrorTitle   string
		ErrorMessage string
		DocID        *string
	}{
		TemplateName: "error",
		User:         user,
		BaseURL:      m.baseURL,
		Year:         time.Now().Year(),
		IsAdmin:      false,
		ErrorTitle:   "Access Denied",
		ErrorMessage: "You do not have permission to access the admin panel.",
		DocID:        nil,
	}

	if err := m.templates.ExecuteTemplate(w, "base", data); err != nil {
		logger.Logger.Error("Failed to render forbidden page", "error", err.Error())
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}

func IsAdminUser(user *models.User, adminEmails []string) bool {
	if user == nil {
		return false
	}

	if len(adminEmails) == 0 {
		return false
	}

	userEmail := strings.ToLower(strings.TrimSpace(user.Email))

	for _, email := range adminEmails {
		adminEmail := strings.ToLower(strings.TrimSpace(email))
		if userEmail == adminEmail {
			return true
		}
	}

	return false
}
