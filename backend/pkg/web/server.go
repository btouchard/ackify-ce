// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/email"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/i18n"
	whworker "github.com/btouchard/ackify-ce/backend/internal/infrastructure/webhook"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/workers"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/handlers"
	"github.com/btouchard/ackify-ce/backend/pkg/crypto"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

type Server struct {
	httpServer      *http.Server
	db              *sql.DB
	router          *chi.Mux
	emailSender     email.Sender
	emailWorker     *email.Worker
	webhookWorker   *whworker.Worker
	sessionWorker   *auth.SessionWorker
	magicLinkWorker *workers.MagicLinkCleanupWorker
	baseURL         string
	adminEmails     []string
	authService     *auth.OauthService
	autoLogin       bool
}

func NewServer(ctx context.Context, cfg *config.Config, frontend embed.FS, version string) (*Server, error) {
	db, signer, i18nService, emailSender, err := initInfrastructure(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize repositories
	signatureRepo := database.NewSignatureRepository(db)
	documentRepo := database.NewDocumentRepository(db)
	expectedSignerRepo := database.NewExpectedSignerRepository(db)
	reminderRepo := database.NewReminderRepository(db)
	emailQueueRepo := database.NewEmailQueueRepository(db)
	webhookRepo := database.NewWebhookRepository(db)
	webhookDeliveryRepo := database.NewWebhookDeliveryRepository(db)
	magicLinkRepo := database.NewMagicLinkRepository(db)

	// Initialize webhook publisher and worker
	webhookPublisher := services.NewWebhookPublisher(webhookRepo, webhookDeliveryRepo)
	whCfg := whworker.DefaultWorkerConfig()
	webhookWorker := whworker.NewWorker(webhookDeliveryRepo, &http.Client{}, whCfg)
	oauthSessionRepo := database.NewOAuthSessionRepository(db)

	// Initialize OAuth auth service with session repository
	// Note: SessionService is ALWAYS created, OAuthProvider is OPTIONAL (based on credentials)
	authService := auth.NewOAuthService(auth.Config{
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

	// Log authentication method status
	if cfg.Auth.OAuthEnabled {
		logger.Logger.Info("OAuth authentication enabled")
	} else {
		logger.Logger.Info("OAuth authentication disabled")
	}

	// Initialize services
	signatureService := services.NewSignatureService(signatureRepo, documentRepo, signer)
	signatureService.SetChecksumConfig(&cfg.Checksum)
	documentService := services.NewDocumentService(documentRepo, &cfg.Checksum)

	// Initialize email worker for async processing
	var emailWorker *email.Worker
	if emailSender != nil && cfg.Mail.Host != "" {
		renderer := email.NewRenderer(getTemplatesDir(), cfg.App.BaseURL, cfg.App.Organisation, cfg.Mail.FromName, cfg.Mail.From, "fr", i18nService)
		workerConfig := email.DefaultWorkerConfig()
		emailWorker = email.NewWorker(emailQueueRepo, emailSender, renderer, workerConfig)
		// Attach webhook event publisher so reminder events can be emitted
		if webhookPublisher != nil {
			emailWorker.SetPublisher(webhookPublisher)
		}
		// Start the worker
		if err := emailWorker.Start(); err != nil {
			return nil, fmt.Errorf("failed to start email worker: %w", err)
		}
	}

	// Start webhook worker
	if err := webhookWorker.Start(); err != nil {
		return nil, fmt.Errorf("failed to start webhook worker: %w", err)
	}

	magicLinkService := services.NewMagicLinkService(services.MagicLinkServiceConfig{
		Repository:        magicLinkRepo,
		EmailSender:       emailSender,
		BaseURL:           cfg.App.BaseURL,
		AppName:           cfg.App.Organisation,
		RateLimitPerEmail: cfg.Auth.MagicLinkRateLimitEmail,
		RateLimitPerIP:    cfg.Auth.MagicLinkRateLimitIP,
	})

	// Initialize Magic Link cleanup worker
	var magicLinkWorker *workers.MagicLinkCleanupWorker
	if cfg.Auth.MagicLinkEnabled {
		logger.Logger.Info("Magic Link authentication enabled")
		magicLinkWorker = workers.NewMagicLinkCleanupWorker(magicLinkService, 1*time.Hour)
		go magicLinkWorker.Start(ctx)
	} else {
		logger.Logger.Info("Magic Link authentication disabled")
	}

	// Initialize reminder service with async support (needs magicLinkService)
	reminderService := services.NewReminderAsyncService(
		expectedSignerRepo,
		reminderRepo,
		emailQueueRepo,
		magicLinkService,
		cfg.App.BaseURL,
	)

	// Initialize OAuth session cleanup worker
	var sessionWorker *auth.SessionWorker
	if oauthSessionRepo != nil {
		workerConfig := auth.DefaultSessionWorkerConfig()
		sessionWorker = auth.NewSessionWorker(oauthSessionRepo, workerConfig)
		if err := sessionWorker.Start(); err != nil {
			return nil, fmt.Errorf("failed to start OAuth session worker: %w", err)
		}
	}

	router := chi.NewRouter()

	router.Use(i18n.Middleware(i18nService))

	// Embed middleware: intercepts /embed to ensure document exists (with strict rate limit)
	// This runs BEFORE the SPA is served, allowing Notion/Outline embeds to work
	router.Use(EmbedDocumentMiddleware(
		documentService,
		webhookPublisher,
	))

	apiConfig := api.RouterConfig{
		AuthService:               authService,
		MagicLinkService:          magicLinkService,
		SignatureService:          signatureService,
		DocumentService:           documentService,
		DocumentRepository:        documentRepo,
		ExpectedSignerRepository:  expectedSignerRepo,
		ReminderService:           reminderService,
		WebhookRepository:         webhookRepo,
		WebhookDeliveryRepository: webhookDeliveryRepo,
		WebhookPublisher:          webhookPublisher,
		BaseURL:                   cfg.App.BaseURL,
		AdminEmails:               cfg.App.AdminEmails,
		AutoLogin:                 cfg.OAuth.AutoLogin,
		OAuthEnabled:              cfg.Auth.OAuthEnabled,
		MagicLinkEnabled:          cfg.Auth.MagicLinkEnabled,
		OnlyAdminCanCreate:        cfg.App.OnlyAdminCanCreate,
		AuthRateLimit:             cfg.App.AuthRateLimit,
		DocumentRateLimit:         cfg.App.DocumentRateLimit,
		GeneralRateLimit:          cfg.App.GeneralRateLimit,
		ImportMaxSigners:          cfg.App.ImportMaxSigners,
	}
	apiRouter := api.NewRouter(apiConfig)
	router.Mount("/api/v1", apiRouter)

	router.Get("/oembed", handlers.HandleOEmbed(cfg.App.BaseURL))

	router.NotFound(EmbedFolder(frontend, "web/dist", cfg.App.BaseURL, version, cfg.Auth.OAuthEnabled, cfg.Auth.MagicLinkEnabled, cfg.App.SMTPEnabled, cfg.App.OnlyAdminCanCreate, signatureRepo))

	httpServer := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: handlers.RequestLogger(handlers.SecureHeaders(router)),
	}

	return &Server{
		httpServer:      httpServer,
		db:              db,
		router:          router,
		emailSender:     emailSender,
		emailWorker:     emailWorker,
		webhookWorker:   webhookWorker,
		sessionWorker:   sessionWorker,
		magicLinkWorker: magicLinkWorker,
		baseURL:         cfg.App.BaseURL,
		adminEmails:     cfg.App.AdminEmails,
		authService:     authService,
		autoLogin:       cfg.OAuth.AutoLogin,
	}, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop Magic Link cleanup worker if it exists
	if s.magicLinkWorker != nil {
		s.magicLinkWorker.Stop()
	}

	// Stop OAuth session worker first if it exists
	if s.sessionWorker != nil {
		if err := s.sessionWorker.Stop(); err != nil {
			logger.Logger.Warn("Failed to stop OAuth session worker", "error", err)
		}
	}

	// Stop email worker if it exists
	if s.emailWorker != nil {
		if err := s.emailWorker.Stop(); err != nil {
			logger.Logger.Warn("Failed to stop email worker", "error", err)
		}
	}

	// Stop webhook worker
	if s.webhookWorker != nil {
		if err := s.webhookWorker.Stop(); err != nil {
			logger.Logger.Warn("Failed to stop webhook worker", "error", err)
		}
	}

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	// Close database connection
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) RegisterRoutes(fn func(r *chi.Mux)) {
	fn(s.router)
}

func (s *Server) GetDB() *sql.DB {
	return s.db
}

func (s *Server) GetAdminEmails() []string {
	return s.adminEmails
}

func (s *Server) GetAuthService() *auth.OauthService {
	return s.authService
}

func (s *Server) GetEmailSender() email.Sender {
	return s.emailSender
}

func initInfrastructure(ctx context.Context, cfg *config.Config) (*sql.DB, *crypto.Ed25519Signer, *i18n.I18n, email.Sender, error) {
	db, err := database.InitDB(ctx, database.Config{
		DSN: cfg.Database.DSN,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	signer, err := crypto.NewEd25519Signer()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize signer: %w", err)
	}

	localesDir := getLocalesDir()
	i18nService, err := i18n.NewI18n(localesDir)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize i18n: %w", err)
	}

	emailTemplatesDir := getTemplatesDir()
	renderer := email.NewRenderer(emailTemplatesDir, cfg.App.BaseURL, cfg.App.Organisation, cfg.Mail.FromName, cfg.Mail.From, "fr", i18nService)
	emailSender := email.NewSMTPSender(cfg.Mail, renderer)

	return db, signer, i18nService, emailSender, nil
}

func getTemplatesDir() string {
	if envDir := os.Getenv("ACKIFY_TEMPLATES_DIR"); envDir != "" {
		return envDir
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		defaultDir := filepath.Join(execDir, "templates")
		if _, err := os.Stat(defaultDir); err == nil {
			return defaultDir
		}
	}

	possiblePaths := []string{
		"templates",   // When running from project root
		"./templates", // Alternative relative path
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "templates"
}

func getLocalesDir() string {
	if envDir := os.Getenv("ACKIFY_LOCALES_DIR"); envDir != "" {
		return envDir
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		defaultDir := filepath.Join(execDir, "locales")
		if _, err := os.Stat(defaultDir); err == nil {
			return defaultDir
		}
	}

	possiblePaths := []string{
		"locales",   // When running from project root
		"./locales", // Alternative relative path
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "locales"
}
