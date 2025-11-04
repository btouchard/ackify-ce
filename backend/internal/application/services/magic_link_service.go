// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/email"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// MagicLinkRepository définit les opérations sur les tokens Magic Link
type MagicLinkRepository interface {
	CreateToken(ctx context.Context, token *models.MagicLinkToken) error
	GetByToken(ctx context.Context, token string) (*models.MagicLinkToken, error)
	MarkAsUsed(ctx context.Context, token string, ip string, userAgent string) error
	DeleteExpired(ctx context.Context) (int64, error)

	LogAttempt(ctx context.Context, attempt *models.MagicLinkAuthAttempt) error
	CountRecentAttempts(ctx context.Context, email string, since time.Time) (int, error)
	CountRecentAttemptsByIP(ctx context.Context, ip string, since time.Time) (int, error)
}

// MagicLinkService gère l'authentification par Magic Link
type MagicLinkService struct {
	repo           MagicLinkRepository
	emailSender    email.Sender
	baseURL        string
	appName        string
	allowedDomains []string // Domaines email autorisés (vide = tous)
	tokenValidity  time.Duration
}

// MagicLinkServiceConfig pour le service Magic Link
type MagicLinkServiceConfig struct {
	Repository     MagicLinkRepository
	EmailSender    email.Sender
	BaseURL        string
	AppName        string
	AllowedDomains []string
	TokenValidity  time.Duration // Défaut: 15 minutes
}

func NewMagicLinkService(cfg MagicLinkServiceConfig) *MagicLinkService {
	if cfg.TokenValidity == 0 {
		cfg.TokenValidity = 15 * time.Minute
	}

	if cfg.AppName == "" {
		cfg.AppName = "Ackify"
	}

	return &MagicLinkService{
		repo:           cfg.Repository,
		emailSender:    cfg.EmailSender,
		baseURL:        cfg.BaseURL,
		appName:        cfg.AppName,
		allowedDomains: cfg.AllowedDomains,
		tokenValidity:  cfg.TokenValidity,
	}
}

// RequestMagicLink génère et envoie un Magic Link par email
func (s *MagicLinkService) RequestMagicLink(
	ctx context.Context,
	emailAddr string,
	redirectTo string,
	ip string,
	userAgent string,
) error {
	// Normaliser l'email
	emailAddr = strings.ToLower(strings.TrimSpace(emailAddr))

	// Valider le format email
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		s.logAttempt(ctx, emailAddr, false, "invalid_email_format", ip, userAgent)
		return fmt.Errorf("invalid email format")
	}

	// Vérifier le domaine autorisé si configuré
	if len(s.allowedDomains) > 0 {
		allowed := false
		for _, domain := range s.allowedDomains {
			if strings.HasSuffix(emailAddr, "@"+domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			s.logAttempt(ctx, emailAddr, false, "domain_not_allowed", ip, userAgent)
			return fmt.Errorf("email domain not allowed")
		}
	}

	// Rate limiting par email (max 3/heure)
	since := time.Now().Add(-1 * time.Hour)
	count, err := s.repo.CountRecentAttempts(ctx, emailAddr, since)
	if err != nil {
		logger.Logger.Error("Failed to check rate limit for email", "email", emailAddr, "error", err)
		return fmt.Errorf("rate limit check failed")
	}
	if count >= 3 {
		s.logAttempt(ctx, emailAddr, false, "rate_limit_exceeded_email", ip, userAgent)
		// Ne pas révéler le rate limiting pour éviter l'énumération
		logger.Logger.Warn("Magic Link rate limit exceeded", "email", emailAddr, "count", count)
		// On retourne success pour ne pas révéler qu'on a bloqué
		return nil
	}

	// Rate limiting par IP (max 10/heure)
	countIP, err := s.repo.CountRecentAttemptsByIP(ctx, ip, since)
	if err != nil {
		logger.Logger.Error("Failed to check rate limit for IP", "ip", ip, "error", err)
		return fmt.Errorf("rate limit check failed")
	}
	if countIP >= 10 {
		s.logAttempt(ctx, emailAddr, false, "rate_limit_exceeded_ip", ip, userAgent)
		logger.Logger.Warn("Magic Link IP rate limit exceeded", "ip", ip, "count", countIP)
		return nil
	}

	// Générer un token cryptographiquement sécurisé
	token, err := s.generateSecureToken()
	if err != nil {
		s.logAttempt(ctx, emailAddr, false, "token_generation_failed", ip, userAgent)
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Créer le token en DB
	magicToken := &models.MagicLinkToken{
		Token:              token,
		Email:              emailAddr,
		ExpiresAt:          time.Now().Add(s.tokenValidity),
		RedirectTo:         redirectTo,
		CreatedByIP:        ip,
		CreatedByUserAgent: userAgent,
	}

	if err := s.repo.CreateToken(ctx, magicToken); err != nil {
		s.logAttempt(ctx, emailAddr, false, "database_error", ip, userAgent)
		return fmt.Errorf("failed to create token: %w", err)
	}

	// Construire le lien magique avec URL encoding du redirect
	redirectEncoded := url.QueryEscape(redirectTo)
	magicLink := fmt.Sprintf("%s/api/v1/auth/magic-link/verify?token=%s&redirect=%s", s.baseURL, token, redirectEncoded)

	// Déterminer la locale (TODO: implémenter détection de langue préférée)
	locale := "en" // Défaut

	// Envoyer l'email
	msg := email.Message{
		To:       []string{emailAddr},
		Subject:  "Your login link",
		Template: "magic_link",
		Locale:   locale,
		Data: map[string]interface{}{
			"AppName":   s.appName,
			"Email":     emailAddr,
			"MagicLink": magicLink,
			"ExpiresIn": int(s.tokenValidity.Minutes()),
			"BaseURL":   s.baseURL,
		},
	}

	if err := s.emailSender.Send(ctx, msg); err != nil {
		s.logAttempt(ctx, emailAddr, false, "email_send_failed", ip, userAgent)
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Log succès
	s.logAttempt(ctx, emailAddr, true, "", ip, userAgent)

	logger.Logger.Info("Magic Link sent successfully",
		"email", emailAddr,
		"expires_in", s.tokenValidity,
		"ip", ip)

	return nil
}

// VerifyMagicLink vérifie et consomme un token Magic Link
func (s *MagicLinkService) VerifyMagicLink(
	ctx context.Context,
	token string,
	ip string,
	userAgent string,
) (*models.MagicLinkToken, error) {
	// Récupérer le token
	magicToken, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		logger.Logger.Warn("Magic Link token not found", "token_prefix", token[:min(8, len(token))])
		return nil, fmt.Errorf("invalid token")
	}

	// Vérifier la validité
	if !magicToken.IsValid() {
		if magicToken.UsedAt != nil {
			logger.Logger.Warn("Magic Link token already used",
				"email", magicToken.Email,
				"used_at", magicToken.UsedAt)
			return nil, fmt.Errorf("token already used")
		}
		logger.Logger.Warn("Magic Link token expired",
			"email", magicToken.Email,
			"expires_at", magicToken.ExpiresAt)
		return nil, fmt.Errorf("token expired")
	}

	// Marquer comme utilisé
	if err := s.repo.MarkAsUsed(ctx, token, ip, userAgent); err != nil {
		logger.Logger.Error("Failed to mark token as used", "error", err)
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	logger.Logger.Info("Magic Link verified successfully",
		"email", magicToken.Email,
		"ip", ip)

	return magicToken, nil
}

// generateSecureToken génère un token cryptographiquement sécurisé
func (s *MagicLinkService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Base64 URL-safe encoding (sans padding)
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// logAttempt enregistre une tentative d'authentification
func (s *MagicLinkService) logAttempt(
	ctx context.Context,
	email string,
	success bool,
	failureReason string,
	ip string,
	userAgent string,
) {
	attempt := &models.MagicLinkAuthAttempt{
		Email:         email,
		Success:       success,
		FailureReason: failureReason,
		IPAddress:     ip,
		UserAgent:     userAgent,
	}

	if err := s.repo.LogAttempt(ctx, attempt); err != nil {
		logger.Logger.Error("Failed to log Magic Link attempt", "error", err)
	}
}

// CleanupExpiredTokens supprime les tokens expirés (à appeler périodiquement)
func (s *MagicLinkService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpired(ctx)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
