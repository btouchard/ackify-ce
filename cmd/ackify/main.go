package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"

	"ackify/internal/application/services"
	"ackify/internal/infrastructure/auth"
	"ackify/internal/infrastructure/config"
	"ackify/internal/infrastructure/database"
	"ackify/internal/presentation/handlers"
	"ackify/internal/presentation/templates"
	"ackify/pkg/crypto"
)

func main() {
	ctx := context.Background()

	// Initialize dependencies
	cfg, db, tmpl, signer, err := initInfrastructure(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize infrastructure: %v", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

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
	server := &http.Server{
		Addr:    cfg.Server.ListenAddr,
		Handler: handlers.SecureHeaders(router),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", cfg.Server.ListenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
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
) *httprouter.Router {
	router := httprouter.New()

	// Public routes
	router.GET("/", signatureHandlers.HandleIndex)
	router.GET("/login", authHandlers.HandleLogin)
	router.GET("/logout", authHandlers.HandleLogout)
	router.GET("/oauth2/callback", authHandlers.HandleOAuthCallback)
	router.GET("/status", signatureHandlers.HandleStatusJSON)
	router.GET("/status.png", badgeHandler.HandleStatusPNG)
	router.GET("/oembed", oembedHandler.HandleOEmbed)
	router.GET("/embed", oembedHandler.HandleEmbedView)
	router.GET("/health", healthHandler.HandleHealth)

	// Protected routes (require authentication)
	router.GET("/sign", authMiddleware.RequireAuth(signatureHandlers.HandleSignGET))
	router.POST("/sign", authMiddleware.RequireAuth(signatureHandlers.HandleSignPOST))
	router.GET("/signatures", authMiddleware.RequireAuth(signatureHandlers.HandleUserSignatures))

	return router
}
