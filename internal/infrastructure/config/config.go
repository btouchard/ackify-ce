package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/securecookie"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	OAuth    OAuthConfig
	Server   ServerConfig
}

// AppConfig holds general application settings
type AppConfig struct {
	BaseURL       string
	Organisation  string
	SecureCookies bool
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	DSN string
}

// OAuthConfig holds OAuth authentication settings
type OAuthConfig struct {
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	UserInfoURL   string
	Scopes        []string
	AllowedDomain string
	CookieSecret  []byte
}

// ServerConfig holds server settings
type ServerConfig struct {
	ListenAddr string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{}

	// App config
	baseURL := mustGetEnv("APP_BASE_URL")
	config.App.BaseURL = baseURL
	config.App.Organisation = mustGetEnv("APP_ORGANISATION")
	config.App.SecureCookies = strings.HasPrefix(strings.ToLower(baseURL), "https://")

	// Database config
	config.Database.DSN = mustGetEnv("DB_DSN")

	// OAuth config
	config.OAuth.ClientID = mustGetEnv("OAUTH_CLIENT_ID")
	config.OAuth.ClientSecret = mustGetEnv("OAUTH_CLIENT_SECRET")
	config.OAuth.AllowedDomain = os.Getenv("OAUTH_ALLOWED_DOMAIN")

	// Configure OAuth endpoints based on provider or use custom URLs
	provider := strings.ToLower(getEnv("OAUTH_PROVIDER", ""))
	switch provider {
	case "google":
		config.OAuth.AuthURL = "https://accounts.google.com/o/oauth2/auth"
		config.OAuth.TokenURL = "https://oauth2.googleapis.com/token"
		config.OAuth.UserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
		config.OAuth.Scopes = []string{"openid", "email", "profile"}
	case "github":
		config.OAuth.AuthURL = "https://github.com/login/oauth/authorize"
		config.OAuth.TokenURL = "https://github.com/login/oauth/access_token"
		config.OAuth.UserInfoURL = "https://api.github.com/user"
		config.OAuth.Scopes = []string{"user:email", "read:user"}
	case "gitlab":
		gitlabURL := getEnv("OAUTH_GITLAB_URL", "https://gitlab.com")
		config.OAuth.AuthURL = fmt.Sprintf("%s/oauth/authorize", gitlabURL)
		config.OAuth.TokenURL = fmt.Sprintf("%s/oauth/token", gitlabURL)
		config.OAuth.UserInfoURL = fmt.Sprintf("%s/api/v4/user", gitlabURL)
		config.OAuth.Scopes = []string{"read_user", "profile"}
	default:
		// Custom OAuth provider - all URLs must be explicitly set
		config.OAuth.AuthURL = mustGetEnv("OAUTH_AUTH_URL")
		config.OAuth.TokenURL = mustGetEnv("OAUTH_TOKEN_URL")
		config.OAuth.UserInfoURL = mustGetEnv("OAUTH_USERINFO_URL")
		scopesStr := getEnv("OAUTH_SCOPES", "openid,email,profile")
		config.OAuth.Scopes = strings.Split(scopesStr, ",")
	}

	cookieSecret, err := parseCookieSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to parse cookie secret: %w", err)
	}
	config.OAuth.CookieSecret = cookieSecret

	// Server config
	config.Server.ListenAddr = getEnv("LISTEN_ADDR", ":8080")

	return config, nil
}

// mustGetEnv gets an environment variable or panics if not found
func mustGetEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		panic(fmt.Sprintf("missing required environment variable: %s", key))
	}
	return value
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

// parseCookieSecret parses the cookie secret from environment
func parseCookieSecret() ([]byte, error) {
	raw := os.Getenv("OAUTH_COOKIE_SECRET")
	if raw == "" {
		// Generate random 32 bytes for development
		secret := securecookie.GenerateRandomKey(32)
		fmt.Println("[WARN] OAUTH_COOKIE_SECRET not set, generated volatile secret (sessions reset on restart)")
		return secret, nil
	}

	// Try base64 decoding first
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && (len(decoded) == 32 || len(decoded) == 64) {
		return decoded, nil
	}

	// Fallback to raw bytes
	return []byte(raw), nil
}
