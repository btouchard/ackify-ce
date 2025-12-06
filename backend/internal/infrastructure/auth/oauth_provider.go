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
	"golang.org/x/oauth2"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/crypto"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

// OAuthProvider handles OAuth2 authentication flow
// This component is optional and can be nil if OAuth is disabled
type OAuthProvider struct {
	oauthConfig   *oauth2.Config
	userInfoURL   string
	logoutURL     string
	allowedDomain string
	baseURL       string
	sessionSvc    *SessionService // Reference to session service for state management
}

// OAuthProviderConfig holds configuration for the OAuth provider
type OAuthProviderConfig struct {
	BaseURL       string
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	UserInfoURL   string
	LogoutURL     string
	Scopes        []string
	AllowedDomain string
	SessionSvc    *SessionService
}

// NewOAuthProvider creates a new OAuth provider
func NewOAuthProvider(config OAuthProviderConfig) *OAuthProvider {
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

	logger.Logger.Info("OAuth provider configured successfully")

	return &OAuthProvider{
		oauthConfig:   oauthConfig,
		userInfoURL:   config.UserInfoURL,
		logoutURL:     config.LogoutURL,
		allowedDomain: config.AllowedDomain,
		baseURL:       config.BaseURL,
		sessionSvc:    config.SessionSvc,
	}
}

// GetLogoutURL returns the SSO logout URL if configured
func (p *OAuthProvider) GetLogoutURL() string {
	if p.logoutURL == "" {
		return ""
	}

	logoutURL := p.logoutURL
	if p.baseURL != "" {
		logoutURL += "?continue=" + p.baseURL
	}

	return logoutURL
}

// CreateAuthURL creates an OAuth authorization URL with PKCE
func (p *OAuthProvider) CreateAuthURL(w http.ResponseWriter, r *http.Request, nextURL string) string {
	// Generate PKCE code verifier and challenge
	codeVerifier, err := crypto.GenerateCodeVerifier()
	if err != nil {
		logger.Logger.Error("Failed to generate PKCE code verifier", "error", err.Error())
		// Fallback to OAuth flow without PKCE for backward compatibility
		return p.createAuthURLWithoutPKCE(w, r, nextURL)
	}

	codeChallenge := crypto.GenerateCodeChallenge(codeVerifier)
	logger.Logger.Debug("Generated PKCE parameters for OAuth flow")

	// Generate state token
	randPart := securecookie.GenerateRandomKey(20)
	token := base64.RawURLEncoding.EncodeToString(randPart)
	state := token + ":" + base64.RawURLEncoding.EncodeToString([]byte(nextURL))

	promptParam := "select_account"
	isSilent := r.URL.Query().Get("silent") == "true"
	if isSilent {
		promptParam = "none"
	}

	logger.Logger.Info("Starting OAuth flow with PKCE",
		"next_url", nextURL,
		"silent", isSilent,
		"state_token_length", len(token))

	session, err := p.sessionSvc.GetSession(r)
	if err != nil {
		logger.Logger.Error("CreateAuthURL: failed to get session from store", "error", err.Error())
		// Create a new empty session if Get fails
		session, _ = p.sessionSvc.GetNewSession(r)
	}

	// Store state and code_verifier in session
	session.Values["oauth_state"] = token
	session.Values["code_verifier"] = codeVerifier

	err = session.Save(r, w)
	if err != nil {
		logger.Logger.Error("CreateAuthURL: failed to save session", "error", err.Error())
	}

	// Generate OAuth URL with PKCE parameters
	authURL := p.oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("prompt", promptParam),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))

	logger.Logger.Debug("CreateAuthURL: generated auth URL with PKCE",
		"prompt", promptParam,
		"url_length", len(authURL))

	return authURL
}

// createAuthURLWithoutPKCE is a fallback method for OAuth without PKCE
// Used for backward compatibility if PKCE generation fails
func (p *OAuthProvider) createAuthURLWithoutPKCE(w http.ResponseWriter, r *http.Request, nextURL string) string {
	randPart := securecookie.GenerateRandomKey(20)
	token := base64.RawURLEncoding.EncodeToString(randPart)
	state := token + ":" + base64.RawURLEncoding.EncodeToString([]byte(nextURL))

	promptParam := "select_account"
	isSilent := r.URL.Query().Get("silent") == "true"
	if isSilent {
		promptParam = "none"
	}

	logger.Logger.Warn("Starting OAuth flow WITHOUT PKCE (fallback mode)",
		"next_url", nextURL,
		"silent", isSilent)

	session, err := p.sessionSvc.GetSession(r)
	if err != nil {
		session, _ = p.sessionSvc.GetNewSession(r)
	}

	session.Values["oauth_state"] = token

	err = session.Save(r, w)
	if err != nil {
		logger.Logger.Error("CreateAuthURL: failed to save session", "error", err.Error())
	}

	authURL := p.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("prompt", promptParam))

	return authURL
}

// VerifyState validates the OAuth state token for CSRF protection
func (p *OAuthProvider) VerifyState(w http.ResponseWriter, r *http.Request, stateToken string) bool {
	session, _ := p.sessionSvc.GetSession(r)
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

// subtleConstantTimeCompare performs a timing-safe string comparison
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

// HandleCallback processes the OAuth callback and returns the authenticated user
func (p *OAuthProvider) HandleCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, code, state string) (*models.User, string, error) {
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

	// Retrieve code_verifier from session for PKCE
	session, _ := p.sessionSvc.GetSession(r)
	codeVerifier, hasPKCE := session.Values["code_verifier"].(string)

	// Clean up code_verifier immediately after retrieval
	if hasPKCE {
		delete(session.Values, "code_verifier")
		_ = session.Save(r, w)
	}

	// Exchange authorization code for token (with or without PKCE)
	var token *oauth2.Token
	var err error

	if hasPKCE && codeVerifier != "" {
		logger.Logger.Info("OAuth token exchange with PKCE")
		token, err = p.oauthConfig.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	} else {
		logger.Logger.Warn("OAuth token exchange without PKCE (legacy session or fallback)")
		token, err = p.oauthConfig.Exchange(ctx, code)
	}

	if err != nil {
		logger.Logger.Error("OAuth token exchange failed",
			"error", err.Error(),
			"with_pkce", hasPKCE)
		return nil, nextURL, fmt.Errorf("oauth exchange failed: %w", err)
	}

	logger.Logger.Info("OAuth token exchange successful", "with_pkce", hasPKCE)

	client := p.oauthConfig.Client(ctx, token)
	resp, err := client.Get(p.userInfoURL)
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

	user, err := p.parseUserInfo(resp)
	if err != nil {
		logger.Logger.Error("Failed to parse user info",
			"error", err.Error())
		return nil, nextURL, fmt.Errorf("failed to parse user info: %w", err)
	}

	if !p.IsAllowedDomain(user.Email) {
		logger.Logger.Warn("User domain not allowed")
		return nil, nextURL, models.ErrDomainNotAllowed
	}

	logger.Logger.Info("OAuth callback successful")

	// Store refresh token if available
	if token.RefreshToken != "" && p.sessionSvc.sessionRepo != nil {
		if err := p.sessionSvc.StoreRefreshToken(ctx, w, r, token, user); err != nil {
			// Log error but don't fail the authentication
			logger.Logger.Error("Failed to store refresh token (non-fatal)", "error", err.Error())
		}
	}

	return user, nextURL, nil
}

// IsAllowedDomain checks if the user's email domain is allowed
func (p *OAuthProvider) IsAllowedDomain(email string) bool {
	if p.allowedDomain == "" {
		return true
	}

	domain := strings.ToLower(p.allowedDomain)
	// If domain already has @ prefix, don't add another one
	if !strings.HasPrefix(domain, "@") {
		domain = "@" + domain
	}

	return strings.HasSuffix(strings.ToLower(email), domain)
}

// parseUserInfo extracts user information from the OAuth provider response
func (p *OAuthProvider) parseUserInfo(resp *http.Response) (*models.User, error) {
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

	// Check for email in various provider-specific fields
	// - "email": Standard OIDC claim (Google, GitHub, GitLab, etc.)
	// - "mail": Microsoft Graph API
	// - "userPrincipalName": Microsoft fallback (UPN format)
	if email, ok := rawUser["email"].(string); ok && email != "" {
		user.Email = email
	} else if mail, ok := rawUser["mail"].(string); ok && mail != "" {
		user.Email = mail
	} else if upn, ok := rawUser["userPrincipalName"].(string); ok && upn != "" {
		user.Email = upn
	} else {
		return nil, fmt.Errorf("missing email in user info response (checked: email, mail, userPrincipalName)")
	}

	// Extract display name from various provider-specific fields
	var name string
	if fullName, ok := rawUser["name"].(string); ok && fullName != "" {
		name = fullName
	} else if firstName, ok := rawUser["given_name"].(string); ok {
		if lastName, ok := rawUser["family_name"].(string); ok {
			name = firstName + " " + lastName
		} else {
			name = firstName
		}
	} else if displayName, ok := rawUser["displayName"].(string); ok && displayName != "" {
		name = displayName
	} else if cn, ok := rawUser["cn"].(string); ok && cn != "" {
		name = cn
	} else if displayNameSnake, ok := rawUser["display_name"].(string); ok && displayNameSnake != "" {
		name = displayNameSnake
	} else if preferredName, ok := rawUser["preferred_username"].(string); ok && preferredName != "" {
		name = preferredName
	}

	user.Name = name

	logger.Logger.Debug("Extracted OAuth user identifiers",
		"has_sub", user.Sub != "",
		"has_email", user.Email != "",
		"has_name", user.Name != "")

	if !user.IsValid() {
		return nil, fmt.Errorf("invalid user data extracted")
	}

	return user, nil
}
