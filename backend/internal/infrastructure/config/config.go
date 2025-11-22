// SPDX-License-Identifier: AGPL-3.0-or-later
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/btouchard/ackify-ce/backend/pkg/logger"
	"github.com/gorilla/securecookie"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Auth     AuthConfig
	OAuth    OAuthConfig
	Server   ServerConfig
	Logger   LoggerConfig
	Mail     MailConfig
	Checksum ChecksumConfig
}

type AuthConfig struct {
	OAuthEnabled     bool
	MagicLinkEnabled bool
}

type AppConfig struct {
	BaseURL            string
	Organisation       string
	SecureCookies      bool
	AdminEmails        []string
	OnlyAdminCanCreate bool
	SMTPEnabled        bool // True if SMTP is configured (for email reminders)
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
	LogoutURL     string
	Scopes        []string
	AllowedDomain string
	CookieSecret  []byte
	AutoLogin     bool
}

type ServerConfig struct {
	ListenAddr string
}

type LoggerConfig struct {
	Level  string
	Format string // "classic" or "json"
}

type MailConfig struct {
	Host               string
	Port               int
	Username           string
	Password           string
	TLS                bool
	StartTLS           bool
	InsecureSkipVerify bool
	Timeout            string
	From               string
	FromName           string
	SubjectPrefix      string
	TemplateDir        string
	DefaultLocale      string
}

type ChecksumConfig struct {
	MaxBytes           int64
	TimeoutMs          int
	MaxRedirects       int
	AllowedContentType []string
	SkipSSRFCheck      bool // For testing only - DO NOT use in production
	InsecureSkipVerify bool // For testing only - DO NOT use in production
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{}

	baseURL := mustGetEnv("ACKIFY_BASE_URL")
	config.App.BaseURL = baseURL
	config.App.Organisation = mustGetEnv("ACKIFY_ORGANISATION")
	config.App.SecureCookies = strings.HasPrefix(strings.ToLower(baseURL), "https://")

	config.Database.DSN = mustGetEnv("ACKIFY_DB_DSN")

	// OAuth configuration - now OPTIONAL
	config.OAuth.ClientID = getEnv("ACKIFY_OAUTH_CLIENT_ID", "")
	config.OAuth.ClientSecret = getEnv("ACKIFY_OAUTH_CLIENT_SECRET", "")
	config.OAuth.AllowedDomain = getEnv("ACKIFY_OAUTH_ALLOWED_DOMAIN", "")
	config.OAuth.AutoLogin = getEnvBool("ACKIFY_OAUTH_AUTO_LOGIN", false)

	// Auto-detect OAuth enabled: true if ClientID and ClientSecret are provided
	oauthConfigured := config.OAuth.ClientID != "" && config.OAuth.ClientSecret != ""

	// Allow manual override via environment variable
	if oauthEnabledStr := getEnv("ACKIFY_AUTH_OAUTH_ENABLED", ""); oauthEnabledStr != "" {
		config.Auth.OAuthEnabled = getEnvBool("ACKIFY_AUTH_OAUTH_ENABLED", false)
	} else {
		config.Auth.OAuthEnabled = oauthConfigured
	}

	// Only configure OAuth URLs if OAuth is enabled
	if config.Auth.OAuthEnabled {
		provider := strings.ToLower(getEnv("ACKIFY_OAUTH_PROVIDER", ""))
		switch provider {
		case "google":
			config.OAuth.AuthURL = "https://accounts.google.com/o/oauth2/auth"
			config.OAuth.TokenURL = "https://oauth2.googleapis.com/token"
			config.OAuth.UserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
			config.OAuth.LogoutURL = "https://accounts.google.com/Logout"
			config.OAuth.Scopes = []string{"openid", "email", "profile"}
		case "github":
			config.OAuth.AuthURL = "https://github.com/login/oauth/authorize"
			config.OAuth.TokenURL = "https://github.com/login/oauth/access_token"
			config.OAuth.UserInfoURL = "https://api.github.com/user"
			config.OAuth.LogoutURL = "https://github.com/logout"
			config.OAuth.Scopes = []string{"user:email", "read:user"}
		case "gitlab":
			gitlabURL := getEnv("ACKIFY_OAUTH_GITLAB_URL", "https://gitlab.com")
			config.OAuth.AuthURL = fmt.Sprintf("%s/oauth/authorize", gitlabURL)
			config.OAuth.TokenURL = fmt.Sprintf("%s/oauth/token", gitlabURL)
			config.OAuth.UserInfoURL = fmt.Sprintf("%s/api/v4/user", gitlabURL)
			config.OAuth.LogoutURL = fmt.Sprintf("%s/users/sign_out", gitlabURL)
			config.OAuth.Scopes = []string{"read_user", "profile"}
		default:
			// Custom OAuth provider - require URLs
			config.OAuth.AuthURL = mustGetEnv("ACKIFY_OAUTH_AUTH_URL")
			config.OAuth.TokenURL = mustGetEnv("ACKIFY_OAUTH_TOKEN_URL")
			config.OAuth.UserInfoURL = mustGetEnv("ACKIFY_OAUTH_USERINFO_URL")
			config.OAuth.LogoutURL = getEnv("ACKIFY_OAUTH_LOGOUT_URL", "")
			scopesStr := getEnv("ACKIFY_OAUTH_SCOPES", "openid,email,profile")
			config.OAuth.Scopes = strings.Split(scopesStr, ",")
		}
	}

	cookieSecret, err := parseCookieSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to parse cookie secret: %w", err)
	}
	config.OAuth.CookieSecret = cookieSecret

	config.Server.ListenAddr = getEnv("ACKIFY_LISTEN_ADDR", ":8080")

	config.Logger.Level = getEnv("ACKIFY_LOG_LEVEL", "info")
	config.Logger.Format = getEnv("ACKIFY_LOG_FORMAT", "classic")

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

	// Parse admin-only document creation flag
	config.App.OnlyAdminCanCreate = getEnvBool("ACKIFY_ONLY_ADMIN_CAN_CREATE", false)

	// Parse mail config (optional, service disabled if MAIL_HOST not set)
	mailHost := getEnv("ACKIFY_MAIL_HOST", "")
	if mailHost != "" {
		config.Mail.Host = mailHost
		config.Mail.Port = getEnvInt("ACKIFY_MAIL_PORT", 587)
		config.Mail.Username = getEnv("ACKIFY_MAIL_USERNAME", "")
		config.Mail.Password = getEnv("ACKIFY_MAIL_PASSWORD", "")
		config.Mail.TLS = getEnvBool("ACKIFY_MAIL_TLS", true)
		config.Mail.StartTLS = getEnvBool("ACKIFY_MAIL_STARTTLS", true)
		config.Mail.InsecureSkipVerify = getEnvBool("ACKIFY_MAIL_INSECURE_SKIP_VERIFY", false)
		config.Mail.Timeout = getEnv("ACKIFY_MAIL_TIMEOUT", "10s")
		config.Mail.From = getEnv("ACKIFY_MAIL_FROM", "")
		config.Mail.FromName = getEnv("ACKIFY_MAIL_FROM_NAME", config.App.Organisation)
		config.Mail.SubjectPrefix = getEnv("ACKIFY_MAIL_SUBJECT_PREFIX", "")
		config.Mail.TemplateDir = getEnv("ACKIFY_MAIL_TEMPLATE_DIR", "templates/emails")
		config.Mail.DefaultLocale = getEnv("ACKIFY_MAIL_DEFAULT_LOCALE", "en")
	}

	// Parse checksum config (automatic checksum computation for remote URLs)
	config.Checksum.MaxBytes = getEnvInt64("ACKIFY_CHECKSUM_MAX_BYTES", 10*1024*1024) // 10 MB default
	config.Checksum.TimeoutMs = getEnvInt("ACKIFY_CHECKSUM_TIMEOUT_MS", 5000)         // 5 seconds default
	config.Checksum.MaxRedirects = getEnvInt("ACKIFY_CHECKSUM_MAX_REDIRECTS", 3)

	// Parse allowed content types
	allowedTypesStr := getEnv("ACKIFY_CHECKSUM_ALLOWED_TYPES", "application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.oasis.opendocument.*")
	if allowedTypesStr != "" {
		types := strings.Split(allowedTypesStr, ",")
		for _, typ := range types {
			trimmed := strings.TrimSpace(typ)
			if trimmed != "" {
				config.Checksum.AllowedContentType = append(config.Checksum.AllowedContentType, trimmed)
			}
		}
	}

	smtpConfigured := mailHost != ""
	config.App.SMTPEnabled = smtpConfigured
	config.Auth.MagicLinkEnabled = getEnvBool("ACKIFY_AUTH_MAGICLINK_ENABLED", false) && smtpConfigured

	// Validation: At least one authentication method must be enabled
	if !config.Auth.OAuthEnabled && !config.Auth.MagicLinkEnabled {
		return nil, fmt.Errorf("at least one authentication method must be enabled: set ACKIFY_OAUTH_CLIENT_ID/CLIENT_SECRET for OAuth or ACKIFY_MAIL_HOST for MagicLink")
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
		logger.Logger.Warn("OAuth cookie secret not set, sessions will reset on restart")
		return secret, nil
	}

	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && (len(decoded) == 32 || len(decoded) == 64) {
		return decoded, nil
	}

	return []byte(raw), nil
}

func getEnvInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
		return result
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true" || value == "1"
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	var result int64
	if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
		return result
	}
	return defaultValue
}
