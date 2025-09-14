package web

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/application/services"
	"github.com/btouchard/ackify-ce/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/internal/presentation/handlers"
	"github.com/btouchard/ackify-ce/internal/presentation/templates"
	"github.com/btouchard/ackify-ce/pkg/crypto"
)

// Server represents the Ackify CE web server
type Server struct {
	httpServer *http.Server
	db         *sql.DB
	router     *chi.Mux
}

// NewServer creates a new Ackify CE server instance
func NewServer(ctx context.Context) (*Server, error) {
	// Initialize infrastructure
	cfg, db, tmpl, signer, err := initInfrastructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize services
	authService := auth.NewOAuthService(auth.Config{
		BaseURL:       cfg.App.BaseURL,
		ClientID:      cfg.OAuth.ClientID,
		ClientSecret:  cfg.OAuth.ClientSecret,
		AuthURL:       cfg.OAuth.AuthURL,
		TokenURL:      cfg.OAuth.TokenURL,
		UserInfoURL:   cfg.OAuth.UserInfoURL,
		Scopes:        cfg.OAuth.Scopes,
		AllowedDomain: cfg.OAuth.AllowedDomain,
		CookieSecret:  cfg.OAuth.CookieSecret,
		SecureCookies: cfg.App.SecureCookies,
	})

	// Initialize signatures
	signatureRepo := database.NewSignatureRepository(db)
	signatureService := services.NewSignatureService(signatureRepo, signer)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, cfg.App.BaseURL)
	authMiddleware := handlers.NewAuthMiddleware(authService, cfg.App.BaseURL)
	signatureHandlers := handlers.NewSignatureHandlers(signatureService, authService, tmpl, cfg.App.BaseURL)
	badgeHandler := handlers.NewBadgeHandler(signatureService)
	oembedHandler := handlers.NewOEmbedHandler(signatureService, tmpl, cfg.App.BaseURL, cfg.App.Organisation)
	healthHandler := handlers.NewHealthHandler()

	// Setup HTTP router
	router := setupRouter(authHandlers, authMiddleware, signatureHandlers, badgeHandler, oembedHandler, healthHandler)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: handlers.SecureHeaders(router),
	}

	return &Server{
		httpServer: httpServer,
		db:         db,
		router:     router,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}

// Router returns the underlying Chi router for composition
func (s *Server) Router() *chi.Mux {
	return s.router
}

// RegisterRoutes allows external packages to register additional routes
func (s *Server) RegisterRoutes(fn func(r *chi.Mux)) {
	fn(s.router)
}

// initInfrastructure initializes the basic infrastructure components
func initInfrastructure(ctx context.Context) (*config.Config, *sql.DB, *template.Template, *crypto.Ed25519Signer, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	db, err := database.InitDB(ctx, database.Config{
		DSN: cfg.Database.DSN,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize templates
	tmpl, err := templates.InitTemplates()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize templates: %w", err)
	}

	// Initialize cryptographic signer
	signer, err := crypto.NewEd25519Signer()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize signer: %w", err)
	}

	return cfg, db, tmpl, signer, nil
}

// setupRouter configures all HTTP routes
func setupRouter(
	authHandlers *handlers.AuthHandlers,
	authMiddleware *handlers.AuthMiddleware,
	signatureHandlers *handlers.SignatureHandlers,
	badgeHandler *handlers.BadgeHandler,
	oembedHandler *handlers.OEmbedHandler,
	healthHandler *handlers.HealthHandler,
) *chi.Mux {
	router := chi.NewRouter()

	// Public routes
	router.Get("/", signatureHandlers.HandleIndex)
	router.Get("/login", authHandlers.HandleLogin)
	router.Get("/logout", authHandlers.HandleLogout)
	router.Get("/oauth2/callback", authHandlers.HandleOAuthCallback)
	router.Get("/status", signatureHandlers.HandleStatusJSON)
	router.Get("/status.png", badgeHandler.HandleStatusPNG)
	router.Get("/oembed", oembedHandler.HandleOEmbed)
	router.Get("/embed", oembedHandler.HandleEmbedView)
	router.Get("/health", healthHandler.HandleHealth)

	// Protected routes (require authentication)
	router.Get("/sign", authMiddleware.RequireAuth(signatureHandlers.HandleSignGET))
	router.Post("/sign", authMiddleware.RequireAuth(signatureHandlers.HandleSignPOST))
	router.Get("/signatures", authMiddleware.RequireAuth(signatureHandlers.HandleUserSignatures))

	// Note: Enterprise routes can be added via RegisterRoutes method

	return router
}
