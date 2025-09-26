package web

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/btouchard/ackify-ce/internal/application/services"
	"github.com/btouchard/ackify-ce/internal/infrastructure/auth"
	"github.com/btouchard/ackify-ce/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/internal/presentation/handlers"
	"github.com/btouchard/ackify-ce/pkg/crypto"
)

// Server represents the Ackify web server
type Server struct {
	httpServer *http.Server
	db         *sql.DB
	router     *chi.Mux
	templates  *template.Template
	baseURL    string
}

// NewServer creates a new Ackify server instance
func NewServer(ctx context.Context) (*Server, error) {
	cfg, db, tmpl, signer, err := initInfrastructure(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

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

	signatureRepo := database.NewSignatureRepository(db)
	signatureService := services.NewSignatureService(signatureRepo, signer)

	authHandlers := handlers.NewAuthHandlers(authService, cfg.App.BaseURL)
	authMiddleware := handlers.NewAuthMiddleware(authService, cfg.App.BaseURL)
	signatureHandlers := handlers.NewSignatureHandlers(signatureService, authService, tmpl, cfg.App.BaseURL, cfg.App.Organisation)
	badgeHandler := handlers.NewBadgeHandler(signatureService)
	oembedHandler := handlers.NewOEmbedHandler(signatureService, tmpl, cfg.App.BaseURL, cfg.App.Organisation)
	healthHandler := handlers.NewHealthHandler()

	router := setupRouter(authHandlers, authMiddleware, signatureHandlers, badgeHandler, oembedHandler, healthHandler)

	httpServer := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: handlers.SecureHeaders(router),
	}

	return &Server{
		httpServer: httpServer,
		db:         db,
		router:     router,
		templates:  tmpl,
		baseURL:    cfg.App.BaseURL,
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

// GetTemplates returns the server templates
func (s *Server) GetTemplates() *template.Template {
	return s.templates
}

// GetBaseURL returns the server base URL
func (s *Server) GetBaseURL() string {
	return s.baseURL
}

func initInfrastructure(ctx context.Context) (*config.Config, *sql.DB, *template.Template, *crypto.Ed25519Signer, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := database.InitDB(ctx, database.Config{
		DSN: cfg.Database.DSN,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	tmpl, err := initTemplates()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize templates: %w", err)
	}

	signer, err := crypto.NewEd25519Signer()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize signer: %w", err)
	}

	return cfg, db, tmpl, signer, nil
}

func setupRouter(
	authHandlers *handlers.AuthHandlers,
	authMiddleware *handlers.AuthMiddleware,
	signatureHandlers *handlers.SignatureHandlers,
	badgeHandler *handlers.BadgeHandler,
	oembedHandler *handlers.OEmbedHandler,
	healthHandler *handlers.HealthHandler,
) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/", signatureHandlers.HandleIndex)
	router.Get("/login", authHandlers.HandleLogin)
	router.Get("/logout", authHandlers.HandleLogout)
	router.Get("/oauth2/callback", authHandlers.HandleOAuthCallback)
	router.Get("/status", signatureHandlers.HandleStatusJSON)
	router.Get("/status.png", badgeHandler.HandleStatusPNG)
	router.Get("/oembed", oembedHandler.HandleOEmbed)
	router.Get("/embed", oembedHandler.HandleEmbedView)
	router.Get("/health", healthHandler.HandleHealth)
	// Alias to match documentation and install script
	router.Get("/healthz", healthHandler.HandleHealth)

	router.Get("/sign", authMiddleware.RequireAuth(signatureHandlers.HandleSignGET))
	router.Post("/sign", authMiddleware.RequireAuth(signatureHandlers.HandleSignPOST))
	router.Get("/signatures", authMiddleware.RequireAuth(signatureHandlers.HandleUserSignatures))

	return router
}

func initTemplates() (*template.Template, error) {
	templatesDir := getTemplatesDir()

	baseTemplatePath := filepath.Join(templatesDir, "base.html.tpl")
	tmpl, err := template.New("base").ParseFiles(baseTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %w", err)
	}

	additionalTemplates := []string{"index.html.tpl", "sign.html.tpl", "signatures.html.tpl", "embed.html.tpl", "admin_dashboard.html.tpl", "admin_doc_details.html.tpl"}
	for _, templateFile := range additionalTemplates {
		templatePath := filepath.Join(templatesDir, templateFile)
		_, err = tmpl.ParseFiles(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
		}
	}

	return tmpl, nil
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
