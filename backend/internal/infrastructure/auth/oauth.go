// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

const sessionName = "ackapp_session"

// SessionRepository defines the interface for OAuth session storage
type SessionRepository interface {
	Create(ctx context.Context, session *models.OAuthSession) error
	GetBySessionID(ctx context.Context, sessionID string) (*models.OAuthSession, error)
	UpdateRefreshToken(ctx context.Context, sessionID string, encryptedToken []byte, expiresAt time.Time) error
	DeleteBySessionID(ctx context.Context, sessionID string) error
	DeleteExpired(ctx context.Context, olderThan time.Duration) (int64, error)
}

// OauthService is a wrapper that composes SessionService and OAuthProvider
// SessionService is ALWAYS present (required for all auth methods)
// OAuthProvider is OPTIONAL (nil if OAuth is disabled)
type OauthService struct {
	SessionService *SessionService // ALWAYS present - manages user sessions
	OAuthProvider  *OAuthProvider  // OPTIONAL - nil if OAuth disabled
}

type Config struct {
	BaseURL       string
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	UserInfoURL   string
	LogoutURL     string
	Scopes        []string
	AllowedDomain string
	CookieSecret  []byte
	SecureCookies bool
	SessionRepo   SessionRepository
}

func NewOAuthService(config Config) *OauthService {
	// Create SessionService (ALWAYS required)
	sessionService := NewSessionService(SessionServiceConfig{
		CookieSecret:  config.CookieSecret,
		SecureCookies: config.SecureCookies,
		SessionRepo:   config.SessionRepo,
	})

	// Create OAuthProvider (only if OAuth is configured)
	// For now, we always create it for backward compatibility
	// Later, this will be conditional based on config flags
	var oauthProvider *OAuthProvider
	if config.ClientID != "" && config.ClientSecret != "" {
		oauthProvider = NewOAuthProvider(OAuthProviderConfig{
			BaseURL:       config.BaseURL,
			ClientID:      config.ClientID,
			ClientSecret:  config.ClientSecret,
			AuthURL:       config.AuthURL,
			TokenURL:      config.TokenURL,
			UserInfoURL:   config.UserInfoURL,
			LogoutURL:     config.LogoutURL,
			Scopes:        config.Scopes,
			AllowedDomain: config.AllowedDomain,
			SessionSvc:    sessionService,
		})
		logger.Logger.Info("OAuth service configured with OAuth provider")
	} else {
		logger.Logger.Info("OAuth service configured WITHOUT OAuth provider (session-only mode)")
	}

	return &OauthService{
		SessionService: sessionService,
		OAuthProvider:  oauthProvider,
	}
}

// Session management methods - delegate to SessionService

func (s *OauthService) GetUser(r *http.Request) (*models.User, error) {
	return s.SessionService.GetUser(r)
}

func (s *OauthService) SetUser(w http.ResponseWriter, r *http.Request, user *models.User) error {
	return s.SessionService.SetUser(w, r, user)
}

func (s *OauthService) Logout(w http.ResponseWriter, r *http.Request) {
	s.SessionService.Logout(w, r)
}

// OAuth methods - delegate to OAuthProvider (nil-safe)

func (s *OauthService) GetLogoutURL() string {
	if s.OAuthProvider == nil {
		return ""
	}
	return s.OAuthProvider.GetLogoutURL()
}

func (s *OauthService) CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string {
	if s.OAuthProvider == nil {
		logger.Logger.Error("CreateAuthURL called but OAuth provider is nil")
		return ""
	}
	return s.OAuthProvider.CreateAuthURL(w, r, nextURL)
}

func (s *OauthService) VerifyState(w http.ResponseWriter, r *http.Request, stateToken string) bool {
	if s.OAuthProvider == nil {
		logger.Logger.Error("VerifyState called but OAuth provider is nil")
		return false
	}
	return s.OAuthProvider.VerifyState(w, r, stateToken)
}

func (s *OauthService) HandleCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, code, state string) (*models.User, string, error) {
	if s.OAuthProvider == nil {
		logger.Logger.Error("HandleCallback called but OAuth provider is nil")
		return nil, "/", models.ErrUnauthorized
	}
	return s.OAuthProvider.HandleCallback(ctx, w, r, code, state)
}

func (s *OauthService) IsAllowedDomain(email string) bool {
	if s.OAuthProvider == nil {
		// If no OAuth provider, allow all domains (used for MagicLink)
		return true
	}
	return s.OAuthProvider.IsAllowedDomain(email)
}
