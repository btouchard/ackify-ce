//go:build integration
// +build integration

// SPDX-License-Identifier: AGPL-3.0-or-later
package admin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/admin"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/btouchard/ackify-ce/backend/pkg/config"
	"github.com/go-chi/chi/v5"
)

func setupConfigTestDB(t *testing.T) *database.TestDB {
	testDB := database.SetupTestDB(t)
	return testDB
}

func createTestConfigService(t *testing.T, testDB *database.TestDB) *services.ConfigService {
	configRepo := database.NewConfigRepository(testDB.DB, testDB.TenantProvider)

	envConfig := &config.Config{
		App: config.AppConfig{
			Organisation:       "Test Org",
			OnlyAdminCanCreate: false,
		},
		Auth: config.AuthConfig{
			OAuthEnabled:     true,
			MagicLinkEnabled: false,
		},
		OAuth: config.OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
			Scopes:       []string{"openid", "email", "profile"},
		},
		Mail: config.MailConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "test@example.com",
			Password: "smtp-password",
			TLS:      false,
			StartTLS: true,
			From:     "noreply@example.com",
			FromName: "Test App",
			Timeout:  "10s",
		},
		Storage: config.StorageConfig{
			Type:      "local",
			MaxSizeMB: 50,
			LocalPath: "/data/documents",
		},
	}

	encryptionKey := make([]byte, 32)
	for i := range encryptionKey {
		encryptionKey[i] = byte(i)
	}

	svc := services.NewConfigService(configRepo, envConfig, encryptionKey)

	ctx := context.Background()
	if err := svc.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize config service: %v", err)
	}

	return svc
}

func createTestUser() *models.User {
	return &models.User{
		Sub:     "test-admin-sub",
		Email:   "admin@example.com",
		Name:    "Test Admin",
		IsAdmin: true,
	}
}

func TestSettingsHandler_GetSettings(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response struct {
		Data struct {
			General   models.GeneralConfig   `json:"general"`
			OIDC      admin.OIDCResponse     `json:"oidc"`
			MagicLink models.MagicLinkConfig `json:"magiclink"`
			SMTP      admin.SMTPResponse     `json:"smtp"`
			Storage   admin.StorageResponse  `json:"storage"`
			UpdatedAt string                 `json:"updated_at"`
		} `json:"data"`
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify general config
	if response.Data.General.Organisation != "Test Org" {
		t.Errorf("Expected organisation 'Test Org', got '%s'", response.Data.General.Organisation)
	}

	// Verify OIDC config - secret should be masked
	if response.Data.OIDC.ClientSecret != models.SecretMask {
		t.Errorf("Expected OIDC client_secret to be masked, got '%s'", response.Data.OIDC.ClientSecret)
	}

	// Verify SMTP config - password should be masked
	if response.Data.SMTP.Password != models.SecretMask {
		t.Errorf("Expected SMTP password to be masked, got '%s'", response.Data.SMTP.Password)
	}
}

func TestSettingsHandler_UpdateSection_General(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	// Create request with user context
	body := `{"organisation": "Updated Org", "only_admin_can_create": true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/general", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Add chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("section", "general")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add user context
	user := createTestUser()
	req = req.WithContext(shared.SetUserInContext(req.Context(), user))

	w := httptest.NewRecorder()
	handler.HandleUpdateSection(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the update
	cfg := configService.GetConfig()
	if cfg.General.Organisation != "Updated Org" {
		t.Errorf("Expected organisation 'Updated Org', got '%s'", cfg.General.Organisation)
	}
	if !cfg.General.OnlyAdminCanCreate {
		t.Error("Expected OnlyAdminCanCreate to be true")
	}
}

func TestSettingsHandler_UpdateSection_InvalidSection(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	body := `{"foo": "bar"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/invalid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("section", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := createTestUser()
	req = req.WithContext(shared.SetUserInContext(req.Context(), user))

	w := httptest.NewRecorder()
	handler.HandleUpdateSection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}

func TestSettingsHandler_UpdateSection_ValidationError(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	// Try to disable all auth methods
	body := `{"enabled": false, "provider": ""}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/oidc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("section", "oidc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := createTestUser()
	req = req.WithContext(shared.SetUserInContext(req.Context(), user))

	w := httptest.NewRecorder()
	handler.HandleUpdateSection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsHandler_UpdateSection_NoAuth(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	body := `{"organisation": "Test"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/general", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("section", "general")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// No user context

	w := httptest.NewRecorder()
	handler.HandleUpdateSection(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", w.Code)
	}
}

func TestSettingsHandler_UpdateSection_InvalidJSON(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/general", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("section", "general")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := createTestUser()
	req = req.WithContext(shared.SetUserInContext(req.Context(), user))

	w := httptest.NewRecorder()
	handler.HandleUpdateSection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}

func TestSettingsHandler_ResetFromENV(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	// First modify the config
	ctx := context.Background()
	input := json.RawMessage(`{"organisation": "Modified Org", "only_admin_can_create": true}`)
	_ = configService.UpdateSection(ctx, models.ConfigCategoryGeneral, input, "admin@test.com")

	// Verify modification
	cfg := configService.GetConfig()
	if cfg.General.Organisation != "Modified Org" {
		t.Fatalf("Setup failed: expected 'Modified Org'")
	}

	// Call reset
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/reset", nil)
	user := createTestUser()
	req = req.WithContext(shared.SetUserInContext(req.Context(), user))

	w := httptest.NewRecorder()
	handler.HandleResetFromENV(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify reset to ENV values
	cfg = configService.GetConfig()
	if cfg.General.Organisation != "Test Org" {
		t.Errorf("Expected organisation 'Test Org' after reset, got '%s'", cfg.General.Organisation)
	}
}

func TestSettingsHandler_ResetFromENV_NoAuth(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/reset", nil)
	// No user context

	w := httptest.NewRecorder()
	handler.HandleResetFromENV(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", w.Code)
	}
}

func TestSettingsHandler_TestConnection_InvalidType(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/settings/test/invalid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.HandleTestConnection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}

func TestSettingsHandler_UpdateSection_PreserveMaskedSecrets(t *testing.T) {
	testDB := setupConfigTestDB(t)
	configService := createTestConfigService(t, testDB)
	handler := admin.NewSettingsHandler(configService)

	// Verify initial secret exists
	cfg := configService.GetConfig()
	if cfg.OIDC.ClientSecret != "test-client-secret" {
		t.Fatalf("Initial secret not set")
	}

	// Update with masked secret
	body := `{"enabled": true, "provider": "google", "client_id": "new-id", "client_secret": "********"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings/oidc", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("section", "oidc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	user := createTestUser()
	req = req.WithContext(shared.SetUserInContext(req.Context(), user))

	w := httptest.NewRecorder()
	handler.HandleUpdateSection(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify original secret is preserved
	cfg = configService.GetConfig()
	if cfg.OIDC.ClientSecret != "test-client-secret" {
		t.Errorf("Expected secret to be preserved, got '%s'", cfg.OIDC.ClientSecret)
	}
	if cfg.OIDC.ClientID != "new-id" {
		t.Errorf("Expected client_id to be updated, got '%s'", cfg.OIDC.ClientID)
	}
}
