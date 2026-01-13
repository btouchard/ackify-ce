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

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
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
		"build_date", BuildDate,
		"telemetry", cfg.Telemetry)

	db, err := database.InitDB(ctx, database.Config{DSN: cfg.Database.DSN})
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	tenantProvider, err := tenant.NewSingleTenantProviderWithContext(ctx, db)
	if err != nil {
		log.Fatalf("failed to initialize tenant provider: %v", err)
	}

	// Create repositories needed for auth
	oauthSessionRepo := database.NewOAuthSessionRepository(db, tenantProvider)
	configRepo := database.NewConfigRepository(db, tenantProvider)

	// Create ConfigService (needed for dynamic auth config)
	encryptionKey := cfg.OAuth.CookieSecret
	configService := services.NewConfigService(configRepo, cfg, encryptionKey)

	// Initialize config from DB or ENV
	err = tenant.WithTenantContextFromProvider(ctx, db, tenantProvider, func(txCtx context.Context) error {
		return configService.Initialize(txCtx)
	})
	if err != nil {
		logger.Logger.Warn("Failed to initialize config service, using ENV config", "error", err)
	}

	// Create SessionService (always needed for session management)
	sessionService := auth.NewSessionService(auth.SessionServiceConfig{
		CookieSecret:  cfg.OAuth.CookieSecret,
		SecureCookies: cfg.App.SecureCookies,
		SessionRepo:   oauthSessionRepo,
	})

	// Create DynamicAuthProvider (unified auth for OIDC + MagicLink)
	authProvider := webauth.NewDynamicAuthProvider(webauth.DynamicAuthProviderConfig{
		ConfigProvider: configService,
		SessionService: sessionService,
		BaseURL:        cfg.App.BaseURL,
		// MagicLinkService will be set by ServerBuilder
	})

	// Create authorizer
	authorizer := webauth.NewSimpleAuthorizer(cfg.App.AdminEmails, cfg.App.OnlyAdminCanCreate)

	// === Build Server ===
	server, err := web.NewServerBuilder(cfg, frontend, Version).
		WithDB(db).
		WithTenantProvider(tenantProvider).
		WithAuthProvider(authProvider).
		WithAuthorizer(authorizer).
		WithConfigService(configService).
		Build(ctx)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		log.Printf("Community Edition server starting on %s", server.GetAddr())
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

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
