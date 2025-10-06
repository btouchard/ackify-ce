// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

const sessionName = "ackapp_session"

type OauthService struct {
	oauthConfig   *oauth2.Config
	sessionStore  *sessions.CookieStore
	userInfoURL   string
	logoutURL     string
	allowedDomain string
	secureCookies bool
	baseURL       string
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
}

func NewOAuthService(config Config) *OauthService {
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.BaseURL + "/oauth2/callback",
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	sessionStore := sessions.NewCookieStore(config.CookieSecret)

	return &OauthService{
		oauthConfig:   oauthConfig,
		sessionStore:  sessionStore,
		userInfoURL:   config.UserInfoURL,
		logoutURL:     config.LogoutURL,
		allowedDomain: config.AllowedDomain,
		secureCookies: config.SecureCookies,
		baseURL:       config.BaseURL,
	}
}

func (s *OauthService) GetUser(r *http.Request) (*models.User, error) {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		logger.Logger.Debug("GetUser: failed to get session", "error", err.Error())
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	userJSON, ok := session.Values["user"].(string)
	if !ok || userJSON == "" {
		logger.Logger.Debug("GetUser: no user in session",
			"user_key_exists", ok,
			"user_json_empty", userJSON == "")
		return nil, models.ErrUnauthorized
	}

	var user models.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		logger.Logger.Error("GetUser: failed to unmarshal user", "error", err.Error())
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	logger.Logger.Debug("GetUser: user found", "email", user.Email)
	return &user, nil
}

func (s *OauthService) SetUser(w http.ResponseWriter, r *http.Request, user *models.User) error {
	session, _ := s.sessionStore.Get(r, sessionName)

	userJSON, err := json.Marshal(user)
	if err != nil {
		logger.Logger.Error("SetUser: failed to marshal user", "error", err.Error())
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	logger.Logger.Debug("SetUser: saving user to session",
		"email", user.Email,
		"secure_cookies", s.secureCookies)

	session.Values["user"] = string(userJSON)
	session.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}

	if err := session.Save(r, w); err != nil {
		logger.Logger.Error("SetUser: failed to save session", "error", err.Error())
		return fmt.Errorf("failed to save session: %w", err)
	}

	logger.Logger.Debug("SetUser: session saved successfully")
	return nil
}

func (s *OauthService) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, sessionName)
	session.Options.MaxAge = -1
	_ = session.Save(r, w)
}

// GetLogoutURL returns the SSO logout URL if configured, otherwise returns empty string
func (s *OauthService) GetLogoutURL() string {
	if s.logoutURL == "" {
		return ""
	}

	// For most providers, add post_logout_redirect_uri or continue parameter
	logoutURL := s.logoutURL
	if s.baseURL != "" {
		// Google and OIDC providers use post_logout_redirect_uri
		// GitHub uses a simple redirect
		// GitLab uses a redirect parameter
		logoutURL += "?continue=" + s.baseURL
	}

	return logoutURL
}

func (s *OauthService) GetAuthURL(nextURL string) string {
	state := base64.RawURLEncoding.EncodeToString(securecookie.GenerateRandomKey(20)) +
		":" + base64.RawURLEncoding.EncodeToString([]byte(nextURL))

	return s.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", "select_account"))
}

// CreateAuthURL Persist a CSRF state token server-side to prevent forged OAuth callbacks; encode nextURL to preserve intended redirect.
func (s *OauthService) CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string {
	randPart := securecookie.GenerateRandomKey(20)
	token := base64.RawURLEncoding.EncodeToString(randPart)
	state := token + ":" + base64.RawURLEncoding.EncodeToString([]byte(nextURL))

	logger.Logger.Debug("CreateAuthURL: generating OAuth state",
		"token_length", len(token),
		"next_url", nextURL)

	session, _ := s.sessionStore.Get(r, sessionName)
	session.Values["oauth_state"] = token
	session.Options = &sessions.Options{Path: "/", HttpOnly: true, Secure: s.secureCookies, SameSite: http.SameSiteLaxMode}
	err := session.Save(r, w)
	if err != nil {
		logger.Logger.Error("CreateAuthURL: failed to save session", "error", err.Error())
	}

	// Check if silent login is requested
	promptParam := "select_account"
	if r.URL.Query().Get("silent") == "true" {
		promptParam = "none"
		logger.Logger.Debug("CreateAuthURL: using silent login (prompt=none)")
	}

	authURL := s.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", promptParam))
	logger.Logger.Debug("CreateAuthURL: generated auth URL",
		"prompt", promptParam,
		"url_length", len(authURL))

	return authURL
}

// VerifyState Clear single-use state on success to prevent replay; compare in constant time to avoid timing leaks.
func (s *OauthService) VerifyState(w http.ResponseWriter, r *http.Request, stateToken string) bool {
	session, _ := s.sessionStore.Get(r, sessionName)
	stored, _ := session.Values["oauth_state"].(string)

	logger.Logger.Debug("VerifyState: validating OAuth state",
		"stored_length", len(stored),
		"token_length", len(stateToken),
		"stored_empty", stored == "",
		"token_empty", stateToken == "")

	if stored == "" || stateToken == "" {
		logger.Logger.Warn("VerifyState: empty state tokens")
		return false
	}

	if subtleConstantTimeCompare(stored, stateToken) {
		logger.Logger.Debug("VerifyState: state valid, clearing token")
		delete(session.Values, "oauth_state")
		_ = session.Save(r, w)
		return true
	}

	logger.Logger.Warn("VerifyState: state mismatch")
	return false
}

func subtleConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func (s *OauthService) HandleCallback(ctx context.Context, code, state string) (*models.User, string, error) {
	parts := strings.SplitN(state, ":", 2)
	nextURL := "/"
	if len(parts) == 2 {
		if nb, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
			nextURL = string(nb)
		}
	}

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, nextURL, fmt.Errorf("oauth exchange failed: %w", err)
	}

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get(s.userInfoURL)
	if err != nil || resp.StatusCode != 200 {
		return nil, nextURL, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	user, err := s.parseUserInfo(resp)
	if err != nil {
		return nil, nextURL, fmt.Errorf("failed to parse user info: %w", err)
	}

	if !s.IsAllowedDomain(user.Email) {
		return nil, nextURL, models.ErrDomainNotAllowed
	}

	return user, nextURL, nil
}

func (s *OauthService) IsAllowedDomain(email string) bool {
	if s.allowedDomain == "" {
		return true
	}

	return strings.HasSuffix(
		strings.ToLower(email),
		"@"+strings.ToLower(s.allowedDomain),
	)
}

func (s *OauthService) parseUserInfo(resp *http.Response) (*models.User, error) {
	var rawUser map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Reduce PII in standard logs; log only keys at debug level
	if rawUser != nil {
		keys := make([]string, 0, len(rawUser))
		for k := range rawUser {
			keys = append(keys, k)
		}
		logger.Logger.Debug("OAuth user info received", "keys", keys)
	}

	user := &models.User{}

	if sub, ok := rawUser["sub"].(string); ok {
		user.Sub = sub
	} else if id, ok := rawUser["id"]; ok {
		user.Sub = fmt.Sprintf("%v", id) // Convert to string regardless of type
	} else {
		return nil, fmt.Errorf("missing user ID in response")
	}

	if email, ok := rawUser["email"].(string); ok {
		user.Email = email
	} else {
		return nil, fmt.Errorf("missing email in user info response")
	}

	var name string
	// Priority: full name first, then composite name, then username as fallback
	if fullName, ok := rawUser["name"].(string); ok && fullName != "" {
		name = fullName
	} else if firstName, ok := rawUser["given_name"].(string); ok {
		if lastName, ok := rawUser["family_name"].(string); ok {
			name = firstName + " " + lastName
		} else {
			name = firstName
		}
	} else if cn, ok := rawUser["cn"].(string); ok && cn != "" {
		name = cn
	} else if displayName, ok := rawUser["display_name"].(string); ok && displayName != "" {
		name = displayName
	} else if preferredName, ok := rawUser["preferred_username"].(string); ok && preferredName != "" {
		name = preferredName
	}

	user.Name = name

	logger.Logger.Debug("Extracted OAuth user identifiers",
		"sub", user.Sub,
		"email_present", user.Email != "",
		"name_present", user.Name != "")

	if !user.IsValid() {
		return nil, fmt.Errorf("invalid user data extracted: sub=%s, email=%s", user.Sub, user.Email)
	}

	return user, nil
}
