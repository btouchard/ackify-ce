// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/btouchard/ackify-ce/backend/pkg/config"
	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/email"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/i18n"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/tenant"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/webhook"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/workers"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/handlers"
	"github.com/btouchard/ackify-ce/backend/pkg/crypto"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
	"github.com/btouchard/ackify-ce/backend/pkg/storage"

	sdk "github.com/btouchard/shm/sdk/golang"
)

// Server represents the HTTP server with all its dependencies.
type Server struct {
	httpServer      *http.Server
	db              *sql.DB
	router          *chi.Mux
	emailSender     email.Sender
	emailWorker     *email.Worker
	webhookWorker   *webhook.Worker
	sessionWorker   *auth.SessionWorker
	magicLinkWorker *workers.MagicLinkCleanupWorker
	baseURL         string

	// Capability providers
	authProvider  AuthProvider
	authorizer    Authorizer
	quotaEnforcer QuotaEnforcer
	auditLogger   AuditLogger
}

// ServerBuilder allows dependency injection for extensibility.
// AuthProvider and Authorizer are REQUIRED and must be provided.
// QuotaEnforcer and AuditLogger have sensible defaults for CE.
type ServerBuilder struct {
	cfg      *config.Config
	frontend embed.FS
	version  string

	// Core infrastructure (required)
	db             *sql.DB
	tenantProvider tenant.Provider
	signer         *crypto.Ed25519Signer

	// Capability providers (auth and authorizer are REQUIRED)
	authProvider  AuthProvider
	authorizer    Authorizer
	quotaEnforcer QuotaEnforcer
	auditLogger   AuditLogger

	// Optional infrastructure
	i18nService     *i18n.I18n
	emailSender     email.Sender
	emailRenderer   *email.Renderer
	storageProvider storage.Provider

	// Core services (created internally or injected)
	magicLinkService *services.MagicLinkService
	signatureService *services.SignatureService
	documentService  *services.DocumentService
	adminService     *services.AdminService
	webhookService   *services.WebhookService
	reminderService  *services.ReminderAsyncService
	configService    *services.ConfigService
}

// NewServerBuilder creates a new server builder with the required configuration.
func NewServerBuilder(cfg *config.Config, frontend embed.FS, version string) *ServerBuilder {
	return &ServerBuilder{
		cfg:      cfg,
		frontend: frontend,
		version:  version,
	}
}

// WithDB injects a database connection.
func (b *ServerBuilder) WithDB(db *sql.DB) *ServerBuilder {
	b.db = db
	return b
}

// WithTenantProvider injects a tenant provider.
func (b *ServerBuilder) WithTenantProvider(tp tenant.Provider) *ServerBuilder {
	b.tenantProvider = tp
	return b
}

// WithSigner injects a cryptographic signer.
func (b *ServerBuilder) WithSigner(signer *crypto.Ed25519Signer) *ServerBuilder {
	b.signer = signer
	return b
}

// WithI18nService injects an i18n service.
func (b *ServerBuilder) WithI18nService(i18n *i18n.I18n) *ServerBuilder {
	b.i18nService = i18n
	return b
}

// WithEmailSender injects an email sender.
func (b *ServerBuilder) WithEmailSender(sender email.Sender) *ServerBuilder {
	b.emailSender = sender
	return b
}

// WithEmailRenderer injects an email renderer.
func (b *ServerBuilder) WithEmailRenderer(renderer *email.Renderer) *ServerBuilder {
	b.emailRenderer = renderer
	return b
}

// === Capability Providers ===

// WithAuthProvider injects an authentication provider (REQUIRED).
func (b *ServerBuilder) WithAuthProvider(provider AuthProvider) *ServerBuilder {
	b.authProvider = provider
	return b
}

// WithAuthorizer injects an authorizer (REQUIRED).
func (b *ServerBuilder) WithAuthorizer(authorizer Authorizer) *ServerBuilder {
	b.authorizer = authorizer
	return b
}

// WithQuotaEnforcer injects a quota enforcer (optional, defaults to NoLimit).
func (b *ServerBuilder) WithQuotaEnforcer(enforcer QuotaEnforcer) *ServerBuilder {
	b.quotaEnforcer = enforcer
	return b
}

// WithAuditLogger injects an audit logger (optional, defaults to LogOnly).
func (b *ServerBuilder) WithAuditLogger(logger AuditLogger) *ServerBuilder {
	b.auditLogger = logger
	return b
}

// WithMagicLinkService injects a magic link service.
func (b *ServerBuilder) WithMagicLinkService(service *services.MagicLinkService) *ServerBuilder {
	b.magicLinkService = service
	return b
}

// WithSignatureService injects a signature service.
func (b *ServerBuilder) WithSignatureService(service *services.SignatureService) *ServerBuilder {
	b.signatureService = service
	return b
}

// WithDocumentService injects a document service.
func (b *ServerBuilder) WithDocumentService(service *services.DocumentService) *ServerBuilder {
	b.documentService = service
	return b
}

// WithAdminService injects an admin service.
func (b *ServerBuilder) WithAdminService(service *services.AdminService) *ServerBuilder {
	b.adminService = service
	return b
}

// WithWebhookService injects a webhook service.
func (b *ServerBuilder) WithWebhookService(service *services.WebhookService) *ServerBuilder {
	b.webhookService = service
	return b
}

// WithReminderService injects a reminder service.
func (b *ServerBuilder) WithReminderService(service *services.ReminderAsyncService) *ServerBuilder {
	b.reminderService = service
	return b
}

// WithConfigService injects a configuration service.
func (b *ServerBuilder) WithConfigService(service *services.ConfigService) *ServerBuilder {
	b.configService = service
	return b
}

// Build constructs the server with all dependencies.
func (b *ServerBuilder) Build(ctx context.Context) (*Server, error) {
	if err := b.validateProviders(); err != nil {
		return nil, err
	}

	b.setDefaultProviders()

	if err := b.initializeInfrastructure(); err != nil {
		return nil, err
	}

	repos := b.createRepositories()

	if err := b.initializeTelemetry(ctx); err != nil {
		return nil, err
	}

	whPublisher, whWorker, err := b.initializeWebhookSystem(ctx, repos)
	if err != nil {
		return nil, err
	}

	emailWorker, err := b.initializeEmailWorker(ctx, repos, whPublisher)
	if err != nil {
		return nil, err
	}

	b.initializeCoreServices(repos)
	b.initializeConfigService(ctx, repos)
	magicLinkWorker := b.initializeMagicLinkCleanupWorker(ctx)
	b.initializeReminderService(repos)

	sessionWorker, err := b.initializeSessionWorker(ctx, repos)
	if err != nil {
		return nil, err
	}

	router := b.buildRouter(repos, whPublisher)

	httpServer := &http.Server{
		Addr:    b.cfg.Server.ListenAddr,
		Handler: handlers.RequestLogger(handlers.SecureHeaders(router)),
	}

	return &Server{
		httpServer:      httpServer,
		db:              b.db,
		router:          router,
		emailSender:     b.emailSender,
		emailWorker:     emailWorker,
		webhookWorker:   whWorker,
		sessionWorker:   sessionWorker,
		magicLinkWorker: magicLinkWorker,
		baseURL:         b.cfg.App.BaseURL,
		authProvider:    b.authProvider,
		authorizer:      b.authorizer,
		quotaEnforcer:   b.quotaEnforcer,
		auditLogger:     b.auditLogger,
	}, nil
}

// validateProviders checks that required providers are set.
func (b *ServerBuilder) validateProviders() error {
	if b.authProvider == nil {
		return errors.New("authProvider is required: use WithAuthProvider()")
	}
	if b.authorizer == nil {
		return errors.New("authorizer is required: use WithAuthorizer()")
	}
	return nil
}

// setDefaultProviders sets default implementations for optional providers.
func (b *ServerBuilder) setDefaultProviders() {
	if b.quotaEnforcer == nil {
		b.quotaEnforcer = NewNoLimitQuotaEnforcer()
	}
	if b.auditLogger == nil {
		b.auditLogger = NewLogOnlyAuditLogger()
	}
}

// initializeInfrastructure initializes signer and storage provider.
// i18n and emailSender are expected to be injected from main.go.
func (b *ServerBuilder) initializeInfrastructure() error {
	var err error

	if b.signer == nil {
		b.signer, err = crypto.NewEd25519Signer()
		if err != nil {
			return fmt.Errorf("failed to initialize signer: %w", err)
		}
	}

	if b.storageProvider == nil && b.cfg.Storage.IsEnabled() {
		provider, err := storage.NewProvider(b.cfg.Storage)
		if err != nil {
			return fmt.Errorf("failed to initialize storage provider: %w", err)
		}
		b.storageProvider = provider
		if provider != nil {
			logger.Logger.Info("Storage provider initialized", "type", provider.Type())
		}
	}

	return nil
}

// repositories holds all repository instances.
type repositories struct {
	signature       *database.SignatureRepository
	document        *database.DocumentRepository
	expectedSigner  *database.ExpectedSignerRepository
	reminder        *database.ReminderRepository
	emailQueue      *database.EmailQueueRepository
	webhook         *database.WebhookRepository
	webhookDelivery *database.WebhookDeliveryRepository
	oauthSession    *database.OAuthSessionRepository
	config          *database.ConfigRepository
}

// createRepositories creates all repository instances.
func (b *ServerBuilder) createRepositories() *repositories {
	return &repositories{
		signature:       database.NewSignatureRepository(b.db, b.tenantProvider),
		document:        database.NewDocumentRepository(b.db, b.tenantProvider),
		expectedSigner:  database.NewExpectedSignerRepository(b.db, b.tenantProvider),
		reminder:        database.NewReminderRepository(b.db, b.tenantProvider),
		emailQueue:      database.NewEmailQueueRepository(b.db, b.tenantProvider),
		webhook:         database.NewWebhookRepository(b.db, b.tenantProvider),
		webhookDelivery: database.NewWebhookDeliveryRepository(b.db, b.tenantProvider),
		oauthSession:    database.NewOAuthSessionRepository(b.db, b.tenantProvider),
		config:          database.NewConfigRepository(b.db, b.tenantProvider),
	}
}

func (b *ServerBuilder) initializeTelemetry(ctx context.Context) error {
	telemetry, err := sdk.New(sdk.Config{
		ServerURL:   "https://metrics.kolapsis.com",
		AppName:     "Ackify",
		AppVersion:  b.version,
		Environment: "production",
		Enabled:     b.cfg.Telemetry,
	})
	if err != nil {
		return err
	}
	telemetry.SetProvider(func() map[string]interface{} {
		return map[string]interface{}{
			"documents":     b.documentService.CountDocs(ctx),
			"confirmations": b.signatureService.CountSigns(ctx),
			"webhooks":      b.webhookService.CountWebhooks(ctx),
			"reminds_sent":  b.reminderService.CountSent(ctx),
		}
	})
	go telemetry.Start(context.Background())
	return nil
}

// initializeWebhookSystem initializes webhook publisher and worker.
func (b *ServerBuilder) initializeWebhookSystem(ctx context.Context, repos *repositories) (*services.WebhookPublisher, *webhook.Worker, error) {
	whPublisher := services.NewWebhookPublisher(repos.webhook, repos.webhookDelivery)
	whCfg := webhook.DefaultWorkerConfig()
	whWorker := webhook.NewWorker(repos.webhookDelivery, &http.Client{}, whCfg, ctx, b.db, b.tenantProvider)

	if err := whWorker.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start webhook worker: %w", err)
	}

	return whPublisher, whWorker, nil
}

// initializeEmailWorker initializes email worker for async processing.
// emailRenderer is expected to be injected from main.go via WithEmailRenderer().
func (b *ServerBuilder) initializeEmailWorker(ctx context.Context, repos *repositories, whPublisher *services.WebhookPublisher) (*email.Worker, error) {
	if b.emailSender == nil || b.cfg.Mail.Host == "" || b.emailRenderer == nil {
		return nil, nil
	}

	workerConfig := email.DefaultWorkerConfig()
	emailWorker := email.NewWorker(repos.emailQueue, b.emailSender, b.emailRenderer, workerConfig, ctx, b.db, b.tenantProvider)

	if whPublisher != nil {
		emailWorker.SetPublisher(whPublisher)
	}

	if err := emailWorker.Start(); err != nil {
		return nil, fmt.Errorf("failed to start email worker: %w", err)
	}

	return emailWorker, nil
}

// initializeCoreServices initializes signature, document, admin, and webhook services.
func (b *ServerBuilder) initializeCoreServices(repos *repositories) {
	if b.signatureService == nil {
		b.signatureService = services.NewSignatureService(repos.signature, repos.document, b.signer)
		b.signatureService.SetChecksumConfig(&b.cfg.Checksum)
	}
	if b.documentService == nil {
		b.documentService = services.NewDocumentService(repos.document, repos.expectedSigner, &b.cfg.Checksum)
	}
	if b.adminService == nil {
		b.adminService = services.NewAdminService(repos.document, repos.expectedSigner)
	}
	if b.webhookService == nil {
		b.webhookService = services.NewWebhookService(repos.webhook, repos.webhookDelivery)
	}
}

// initializeConfigService is a no-op when configService is injected from main.go.
// Kept for API compatibility.
func (b *ServerBuilder) initializeConfigService(_ context.Context, _ *repositories) {
	// configService is expected to be injected and initialized from main.go
}

// initializeMagicLinkCleanupWorker starts the cleanup worker for expired magic link tokens.
// magicLinkService is expected to be injected from main.go.
func (b *ServerBuilder) initializeMagicLinkCleanupWorker(ctx context.Context) *workers.MagicLinkCleanupWorker {
	magicLinkWorker := workers.NewMagicLinkCleanupWorker(b.magicLinkService, 1*time.Hour, b.db, b.tenantProvider)
	go magicLinkWorker.Start(ctx)
	return magicLinkWorker
}

// initializeReminderService initializes reminder service.
func (b *ServerBuilder) initializeReminderService(repos *repositories) {
	if b.reminderService == nil {
		b.reminderService = services.NewReminderAsyncService(
			repos.expectedSigner,
			repos.reminder,
			repos.emailQueue,
			b.magicLinkService,
			b.i18nService,
			b.cfg.App.BaseURL,
		)
	}
}

// initializeSessionWorker initializes OAuth session cleanup worker.
func (b *ServerBuilder) initializeSessionWorker(ctx context.Context, repos *repositories) (*auth.SessionWorker, error) {
	if repos.oauthSession == nil {
		return nil, nil
	}

	workerConfig := auth.DefaultSessionWorkerConfig()
	sessionWorker := auth.NewSessionWorker(repos.oauthSession, workerConfig, ctx, b.db, b.tenantProvider)
	if err := sessionWorker.Start(); err != nil {
		return nil, fmt.Errorf("failed to start OAuth session worker: %w", err)
	}

	return sessionWorker, nil
}

// buildRouter creates and configures the main router.
func (b *ServerBuilder) buildRouter(repos *repositories, whPublisher *services.WebhookPublisher) *chi.Mux {
	router := chi.NewRouter()
	router.Use(i18n.Middleware(b.i18nService))
	router.Use(EmbedDocumentMiddleware(b.documentService, whPublisher))

	// Build API router config using unified auth provider
	apiConfig := api.RouterConfig{
		// Database for RLS middleware
		DB:             b.db,
		TenantProvider: b.tenantProvider,

		// Capability providers (AuthProvider handles OIDC + MagicLink dynamically)
		AuthProvider:     b.authProvider,
		Authorizer:       b.authorizer,
		SignatureService: b.signatureService,
		DocumentService:  b.documentService,
		AdminService:     b.adminService,
		ReminderService:  b.reminderService,
		WebhookService:   b.webhookService,
		WebhookPublisher: whPublisher,
		StorageProvider:  b.storageProvider,
		StorageMaxSizeMB: b.cfg.Storage.MaxSizeMB,
		BaseURL:          b.cfg.App.BaseURL,

		// Rate limiting
		AuthRateLimit:     b.cfg.App.AuthRateLimit,
		DocumentRateLimit: b.cfg.App.DocumentRateLimit,
		GeneralRateLimit:  b.cfg.App.GeneralRateLimit,
		ImportMaxSigners:  b.cfg.App.ImportMaxSigners,

		// Config service for dynamic settings
		ConfigService: b.configService,
	}
	apiRouter := api.NewRouter(apiConfig)
	router.Mount("/api/v1", apiRouter)

	router.Get("/oembed", handlers.HandleOEmbed(b.cfg.App.BaseURL))
	router.NotFound(EmbedFolder(b.frontend, "web/dist", b.cfg.App.BaseURL, b.version,
		b.configService, repos.signature))

	return router
}

// === Server Methods ===

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop Magic Link cleanup worker if it exists
	if s.magicLinkWorker != nil {
		s.magicLinkWorker.Stop()
	}

	// Stop OAuth session worker if it exists
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

// GetAuthProvider returns the auth provider.
func (s *Server) GetAuthProvider() AuthProvider {
	return s.authProvider
}

// GetAuthorizer returns the authorizer.
func (s *Server) GetAuthorizer() Authorizer {
	return s.authorizer
}

// GetQuotaEnforcer returns the quota enforcer.
func (s *Server) GetQuotaEnforcer() QuotaEnforcer {
	return s.quotaEnforcer
}

// GetAuditLogger returns the audit logger.
func (s *Server) GetAuditLogger() AuditLogger {
	return s.auditLogger
}

func (s *Server) GetEmailSender() email.Sender {
	return s.emailSender
}

// === Helper Functions ===

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
		"templates",
		"./templates",
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
		"locales",
		"./locales",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "locales"
}
