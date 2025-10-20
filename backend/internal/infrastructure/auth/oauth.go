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

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
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
		RedirectURL:  config.BaseURL + "/api/v1/auth/callback",
		Scopes:       config.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
	}

	sessionStore := sessions.NewCookieStore(config.CookieSecret)

	// Configure session options globally on the store
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   config.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	}

	logger.Logger.Info("OAuth session store configured",
		"secure_cookies", config.SecureCookies,
		"max_age_days", 7)

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
	// Always create a fresh new session to ensure session ID is generated
	// This fixes an issue where reusing an existing invalid session results in empty session.ID
	session, err := s.sessionStore.New(r, sessionName)
	if err != nil {
		logger.Logger.Error("SetUser: failed to create new session", "error", err.Error())
		return fmt.Errorf("failed to create new session: %w", err)
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		logger.Logger.Error("SetUser: failed to marshal user", "error", err.Error())
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	logger.Logger.Debug("SetUser: saving user to new session",
		"email", user.Email,
		"secure_cookies", s.secureCookies,
		"session_is_new", session.IsNew)

	session.Values["user"] = string(userJSON)

	// Session options are already configured globally on the store
	// No need to set them again here

	if err := session.Save(r, w); err != nil {
		logger.Logger.Error("SetUser: failed to save session",
			"error", err.Error(),
			"session_is_new", session.IsNew,
			"session_id_length", len(session.ID))
		return fmt.Errorf("failed to save session: %w", err)
	}

	logger.Logger.Info("SetUser: session saved successfully",
		"email", user.Email,
		"session_id_length", len(session.ID))
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

func (s *OauthService) CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string {
	randPart := securecookie.GenerateRandomKey(20)
	token := base64.RawURLEncoding.EncodeToString(randPart)
	state := token + ":" + base64.RawURLEncoding.EncodeToString([]byte(nextURL))

	promptParam := "select_account"
	isSilent := r.URL.Query().Get("silent") == "true"
	if isSilent {
		promptParam = "none"
	}

	logger.Logger.Info("Starting OAuth flow",
		"next_url", nextURL,
		"silent", isSilent,
		"state_token_length", len(token))

	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		logger.Logger.Error("CreateAuthURL: failed to get session from store", "error", err.Error())
		// Create a new empty session if Get fails
		session, _ = s.sessionStore.New(r, sessionName)
	}

	session.Values["oauth_state"] = token

	// Session options are already configured globally on the store
	// No need to set them again here

	err = session.Save(r, w)
	if err != nil {
		logger.Logger.Error("CreateAuthURL: failed to save session", "error", err.Error())
	}

	authURL := s.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", promptParam))
	logger.Logger.Debug("CreateAuthURL: generated auth URL",
		"prompt", promptParam,
		"url_length", len(authURL))

	return authURL
}

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

	logger.Logger.Debug("Processing OAuth callback",
		"has_code", code != "",
		"next_url", nextURL)

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Logger.Error("OAuth token exchange failed",
			"error", err.Error())
		return nil, nextURL, fmt.Errorf("oauth exchange failed: %w", err)
	}

	logger.Logger.Debug("OAuth token exchange successful")

	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get(s.userInfoURL)
	if err != nil || resp.StatusCode != 200 {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		logger.Logger.Error("User info request failed",
			"error", err,
			"status_code", statusCode)
		return nil, nextURL, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	logger.Logger.Debug("User info retrieved successfully",
		"status_code", resp.StatusCode)

	user, err := s.parseUserInfo(resp)
	if err != nil {
		logger.Logger.Error("Failed to parse user info",
			"error", err.Error())
		return nil, nextURL, fmt.Errorf("failed to parse user info: %w", err)
	}

	if !s.IsAllowedDomain(user.Email) {
		logger.Logger.Warn("User domain not allowed",
			"user_email", user.Email,
			"allowed_domain", s.allowedDomain)
		return nil, nextURL, models.ErrDomainNotAllowed
	}

	logger.Logger.Info("OAuth callback successful",
		"user_email", user.Email,
		"user_name", user.Name)

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
		user.Sub = fmt.Sprintf("%v", id)
	} else {
		return nil, fmt.Errorf("missing user ID in response")
	}

	if email, ok := rawUser["email"].(string); ok {
		user.Email = email
	} else {
		return nil, fmt.Errorf("missing email in user info response")
	}

	var name string
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
