// SPDX-License-Identifier: AGPL-3.0-or-later
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
	"github.com/btouchard/ackify-ce/internal/infrastructure/i18n"
	"github.com/btouchard/ackify-ce/internal/presentation/handlers"
	"github.com/btouchard/ackify-ce/pkg/crypto"
)

type Server struct {
	httpServer  *http.Server
	db          *sql.DB
	router      *chi.Mux
	templates   *template.Template
	baseURL     string
	adminEmails []string
	authService *auth.OauthService
	autoLogin   bool
}

func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	db, tmpl, signer, i18nService, err := initInfrastructure(ctx, cfg)
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
		LogoutURL:     cfg.OAuth.LogoutURL,
		Scopes:        cfg.OAuth.Scopes,
		AllowedDomain: cfg.OAuth.AllowedDomain,
		CookieSecret:  cfg.OAuth.CookieSecret,
		SecureCookies: cfg.App.SecureCookies,
	})

	signatureRepo := database.NewSignatureRepository(db)
	signatureService := services.NewSignatureService(signatureRepo, signer)

	authHandlers := handlers.NewAuthHandlers(authService, cfg.App.BaseURL)
	authMiddleware := handlers.NewAuthMiddleware(authService, cfg.App.BaseURL)
	signatureHandlers := handlers.NewSignatureHandlers(signatureService, authService, tmpl, cfg.App.BaseURL, cfg.App.Organisation, cfg.App.AdminEmails, cfg.OAuth.AutoLogin)
	badgeHandler := handlers.NewBadgeHandler(signatureService)
	oembedHandler := handlers.NewOEmbedHandler(signatureService, tmpl, cfg.App.BaseURL, cfg.App.Organisation)
	healthHandler := handlers.NewHealthHandler()
	langHandlers := handlers.NewLangHandlers(cfg.App.SecureCookies)

	router := setupRouter(authHandlers, authMiddleware, signatureHandlers, badgeHandler, oembedHandler, healthHandler, langHandlers, i18nService, cfg.OAuth.AutoLogin)

	httpServer := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: handlers.RequestLogger(handlers.SecureHeaders(router)),
	}

	return &Server{
		httpServer:  httpServer,
		db:          db,
		router:      router,
		templates:   tmpl,
		baseURL:     cfg.App.BaseURL,
		adminEmails: cfg.App.AdminEmails,
		authService: authService,
		autoLogin:   cfg.OAuth.AutoLogin,
	}, nil
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
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

func (s *Server) GetTemplates() *template.Template {
	return s.templates
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

func initInfrastructure(ctx context.Context, cfg *config.Config) (*sql.DB, *template.Template, *crypto.Ed25519Signer, *i18n.I18n, error) {
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

	localesDir := getLocalesDir()
	i18nService, err := i18n.NewI18n(localesDir)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to initialize i18n: %w", err)
	}

	return db, tmpl, signer, i18nService, nil
}

func setupRouter(
	authHandlers *handlers.AuthHandlers,
	authMiddleware *handlers.AuthMiddleware,
	signatureHandlers *handlers.SignatureHandlers,
	badgeHandler *handlers.BadgeHandler,
	oembedHandler *handlers.OEmbedHandler,
	healthHandler *handlers.HealthHandler,
	langHandlers *handlers.LangHandlers,
	i18nService *i18n.I18n,
	autoLogin bool,
) *chi.Mux {
	router := chi.NewRouter()
	router.Use(i18n.Middleware(i18nService))
	staticDir := getStaticDir()
	fileServer := http.FileServer(http.Dir(staticDir))
	router.Get("/static/*", http.StripPrefix("/static/", fileServer).ServeHTTP)

	router.Get("/", signatureHandlers.HandleIndex)
	router.Get("/login", authHandlers.HandleLogin)
	router.Get("/logout", authHandlers.HandleLogout)
	router.Get("/oauth2/callback", authHandlers.HandleOAuthCallback)

	if autoLogin {
		router.Get("/api/auth/check", authHandlers.HandleAuthCheck)
	}
	router.Get("/status", signatureHandlers.HandleStatusJSON)
	router.Get("/status.png", badgeHandler.HandleStatusPNG)
	router.Get("/oembed", oembedHandler.HandleOEmbed)
	router.Get("/embed", oembedHandler.HandleEmbedView)
	router.Get("/health", healthHandler.HandleHealth)

	router.Get("/lang/{code}", langHandlers.HandleLangSwitch)

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

	additionalTemplates := []string{"index.html.tpl", "sign.html.tpl", "signatures.html.tpl", "embed.html.tpl", "admin_dashboard.html.tpl", "admin_doc_details.html.tpl", "error.html.tpl"}
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

func getStaticDir() string {
	if envDir := os.Getenv("ACKIFY_STATIC_DIR"); envDir != "" {
		return envDir
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		defaultDir := filepath.Join(execDir, "static")
		if _, err := os.Stat(defaultDir); err == nil {
			return defaultDir
		}
	}

	possiblePaths := []string{
		"static",   // When running from project root
		"./static", // Alternative relative path
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "static"
}
