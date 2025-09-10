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

	"ackify/internal/domain/models"
	"ackify/pkg/logger"
)

const sessionName = "ackapp_session"

type OauthService struct {
	oauthConfig   *oauth2.Config
	sessionStore  *sessions.CookieStore
	userInfoURL   string
	allowedDomain string
	secureCookies bool
}

// Config holds OAuth service configuration
type Config struct {
	BaseURL       string
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	UserInfoURL   string
	Scopes        []string
	AllowedDomain string
	CookieSecret  []byte
	SecureCookies bool
}

// NewOAuthService creates a new OAuth service
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
		allowedDomain: config.AllowedDomain,
		secureCookies: config.SecureCookies,
	}
}

func (s *OauthService) GetUser(r *http.Request) (*models.User, error) {
	session, err := s.sessionStore.Get(r, sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	userJSON, ok := session.Values["user"].(string)
	if !ok || userJSON == "" {
		return nil, models.ErrUnauthorized
	}

	var user models.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (s *OauthService) SetUser(w http.ResponseWriter, r *http.Request, user *models.User) error {
	session, _ := s.sessionStore.Get(r, sessionName)

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	session.Values["user"] = string(userJSON)
	session.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteLaxMode,
	}

	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (s *OauthService) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, sessionName)
	session.Options.MaxAge = -1
	_ = session.Save(r, w)
}

func (s *OauthService) GetAuthURL(nextURL string) string {
	state := base64.RawURLEncoding.EncodeToString(securecookie.GenerateRandomKey(20)) +
		":" + base64.RawURLEncoding.EncodeToString([]byte(nextURL))

	return s.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", "select_account"))
}

func (s *OauthService) HandleCallback(ctx context.Context, code, state string) (*models.User, string, error) {
	// Parse state to get next URL
	parts := strings.SplitN(state, ":", 2)
	nextURL := "/"
	if len(parts) == 2 {
		if nb, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
			nextURL = string(nb)
		}
	}

	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, nextURL, fmt.Errorf("oauth exchange failed: %w", err)
	}

	// Get user info
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

	// Check domain restriction
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

// parseUserInfo extracts user information from different OAuth2 providers
func (s *OauthService) parseUserInfo(resp *http.Response) (*models.User, error) {
	var rawUser map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	logger.Logger.Info("Raw OAuth user info received", "data", rawUser)

	user := &models.User{}

	// Extract user ID (sub field or id field depending on provider)
	if sub, ok := rawUser["sub"].(string); ok {
		user.Sub = sub
	} else if id, ok := rawUser["id"]; ok {
		user.Sub = fmt.Sprintf("%v", id) // Convert to string regardless of type
	} else {
		return nil, fmt.Errorf("missing user ID in response")
	}

	// Extract email
	if email, ok := rawUser["email"].(string); ok {
		user.Email = email
	} else {
		// Some providers might have email in a different field or structure
		return nil, fmt.Errorf("missing email in user info response")
	}

	// Extract user name with fallback strategy
	var name string
	if preferredName, ok := rawUser["preferred_username"].(string); ok && preferredName != "" {
		name = preferredName
	} else if firstName, ok := rawUser["given_name"].(string); ok {
		if lastName, ok := rawUser["family_name"].(string); ok {
			name = firstName + " " + lastName
		} else {
			name = firstName
		}
	} else if fullName, ok := rawUser["name"].(string); ok && fullName != "" {
		name = fullName
	} else if cn, ok := rawUser["cn"].(string); ok && cn != "" {
		name = cn
	} else if displayName, ok := rawUser["display_name"].(string); ok && displayName != "" {
		name = displayName
	}

	user.Name = name

	logger.Logger.Info("Extracted user data",
		"sub", user.Sub,
		"email", user.Email,
		"name", user.Name)

	// Validate extracted data
	if !user.IsValid() {
		return nil, fmt.Errorf("invalid user data extracted: sub=%s, email=%s", user.Sub, user.Email)
	}

	return user, nil
}
