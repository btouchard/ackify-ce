// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/gorilla/securecookie"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	OAuth    OAuthConfig
	Server   ServerConfig
	Logger   LoggerConfig
}

type AppConfig struct {
	BaseURL       string
	Organisation  string
	SecureCookies bool
	AdminEmails   []string
}

type DatabaseConfig struct {
	DSN string
}

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

type ServerConfig struct {
	ListenAddr string
}

type LoggerConfig struct {
	Level string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{}

	baseURL := mustGetEnv("ACKIFY_BASE_URL")
	config.App.BaseURL = baseURL
	config.App.Organisation = mustGetEnv("ACKIFY_ORGANISATION")
	config.App.SecureCookies = strings.HasPrefix(strings.ToLower(baseURL), "https://")

	config.Database.DSN = mustGetEnv("ACKIFY_DB_DSN")

	config.OAuth.ClientID = mustGetEnv("ACKIFY_OAUTH_CLIENT_ID")
	config.OAuth.ClientSecret = mustGetEnv("ACKIFY_OAUTH_CLIENT_SECRET")
	config.OAuth.AllowedDomain = os.Getenv("ACKIFY_OAUTH_ALLOWED_DOMAIN")

	provider := strings.ToLower(getEnv("ACKIFY_OAUTH_PROVIDER", ""))
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
		gitlabURL := getEnv("ACKIFY_OAUTH_GITLAB_URL", "https://gitlab.com")
		config.OAuth.AuthURL = fmt.Sprintf("%s/oauth/authorize", gitlabURL)
		config.OAuth.TokenURL = fmt.Sprintf("%s/oauth/token", gitlabURL)
		config.OAuth.UserInfoURL = fmt.Sprintf("%s/api/v4/user", gitlabURL)
		config.OAuth.Scopes = []string{"read_user", "profile"}
	default:
		config.OAuth.AuthURL = mustGetEnv("ACKIFY_OAUTH_AUTH_URL")
		config.OAuth.TokenURL = mustGetEnv("ACKIFY_OAUTH_TOKEN_URL")
		config.OAuth.UserInfoURL = mustGetEnv("ACKIFY_OAUTH_USERINFO_URL")
		scopesStr := getEnv("ACKIFY_OAUTH_SCOPES", "openid,email,profile")
		config.OAuth.Scopes = strings.Split(scopesStr, ",")
	}

	cookieSecret, err := parseCookieSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to parse cookie secret: %w", err)
	}
	config.OAuth.CookieSecret = cookieSecret

	config.Server.ListenAddr = getEnv("ACKIFY_LISTEN_ADDR", ":8080")

	config.Logger.Level = getEnv("ACKIFY_LOG_LEVEL", "info")

	// Parse admin emails
	adminEmailsStr := getEnv("ACKIFY_ADMIN_EMAILS", "")
	if adminEmailsStr != "" {
		emails := strings.Split(strings.ToLower(adminEmailsStr), ",")
		for _, email := range emails {
			trimmed := strings.TrimSpace(email)
			if trimmed != "" {
				config.App.AdminEmails = append(config.App.AdminEmails, trimmed)
			}
		}
	}

	return config, nil
}

func mustGetEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		panic(fmt.Sprintf("missing required environment variable: %s", key))
	}
	return value
}

func getEnv(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func parseCookieSecret() ([]byte, error) {
	raw := os.Getenv("ACKIFY_OAUTH_COOKIE_SECRET")
	if raw == "" {
		secret := securecookie.GenerateRandomKey(32)
		fmt.Println("[WARN] ACKIFY_OAUTH_COOKIE_SECRET not set, generated volatile secret (sessions reset on restart)")
		return secret, nil
	}

	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && (len(decoded) == 32 || len(decoded) == 64) {
		return decoded, nil
	}

	return []byte(raw), nil
}
