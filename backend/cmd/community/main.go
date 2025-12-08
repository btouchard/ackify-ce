package main

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/tenant"
	"github.com/btouchard/ackify-ce/backend/pkg/config"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
	"github.com/btouchard/ackify-ce/backend/pkg/web"
	webauth "github.com/btouchard/ackify-ce/backend/pkg/web/auth"
)

// Build-time variables set via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

//go:embed all:web/dist
var frontend embed.FS

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.SetLevelAndFormat(logger.ParseLevel(cfg.Logger.Level), cfg.Logger.Format)
	logger.Logger.Info("Starting Ackify Community Edition",
		"version", Version,
		"commit", Commit,
		"build_date", BuildDate)

	// Initialize DB
	db, err := database.InitDB(ctx, database.Config{DSN: cfg.Database.DSN})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Initialize tenant provider
	tenantProvider, err := tenant.NewSingleTenantProviderWithContext(ctx, db)
	if err != nil {
		log.Fatalf("failed to initialize tenant provider: %v", err)
	}

	// Create OAuth session repository
	oauthSessionRepo := database.NewOAuthSessionRepository(db, tenantProvider)

	// Create OAuth service (internal infrastructure)
	var oauthService *auth.OauthService
	if cfg.Auth.OAuthEnabled {
		oauthService = auth.NewOAuthService(auth.Config{
			BaseURL:       cfg.App.BaseURL,
			ClientID:      cfg.OAuth.ClientID,
			ClientSecret:  cfg.OAuth.ClientSecret,
			AuthURL:       cfg.OAuth.AuthURL,
			TokenURL:      cfg.OAuth.TokenURL,
			UserInfoURL:   cfg.OAuth.UserInfoURL,
			LogoutURL:     cfg.OAuth.LogoutURL,
			Scopes:        cfg.OAuth.Scopes,
			AllowedDomain: cfg.OAuth.AllowedDomain,
			CookieSecret:  cfg.OAuth.CookieSecret,
			SecureCookies: cfg.App.SecureCookies,
			SessionRepo:   oauthSessionRepo,
		})
	}

	// Create OAuth provider adapter
	oauthProvider := webauth.NewOAuthProvider(oauthService, cfg.Auth.OAuthEnabled)

	// Create Authorizer
	authorizer := webauth.NewSimpleAuthorizer(cfg.App.AdminEmails, cfg.App.OnlyAdminCanCreate)

	// === Build Server ===
	server, err := web.NewServerBuilder(cfg, frontend, Version).
		WithDB(db).
		WithTenantProvider(tenantProvider).
		WithOAuthProvider(oauthProvider). // OAuth provider for OAuth-specific operations
		WithAuthorizer(authorizer).       // Authorization decisions
		// QuotaEnforcer and AuditLogger use defaults (NoLimit, LogOnly)
		Build(ctx)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	go func() {
		log.Printf("Community Edition server starting on %s", server.GetAddr())
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Community Edition server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Community Edition server exited")
}
