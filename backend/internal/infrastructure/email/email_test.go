// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/config"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TEST FIXTURES
// ============================================================================

const (
	testBaseURL      = "https://example.com"
	testOrganisation = "Test Org"
	testFromName     = "Test Sender"
	testFromEmail    = "noreply@example.com"
)

func createTestI18n(t *testing.T, tmpDir string) *i18n.I18n {
	t.Helper()

	// Create simple test translations for all supported languages
	translations := map[string]map[string]string{
		"en": {
			"test.title":   "Test Template",
			"test.message": "Message: {{.message}}",
		},
		"fr": {
			"test.title":   "Modèle de Test",
			"test.message": "Message: {{.message}}",
		},
		"de": {
			"test.title":   "Test Vorlage",
			"test.message": "Nachricht: {{.message}}",
		},
		"es": {
			"test.title":   "Plantilla de Prueba",
			"test.message": "Mensaje: {{.message}}",
		},
		"it": {
			"test.title":   "Modello di Test",
			"test.message": "Messaggio: {{.message}}",
		},
	}

	for lang, trans := range translations {
		// Write locale files
		content := "{"
		first := true
		for key, value := range trans {
			if !first {
				content += ","
			}
			content += `"` + key + `":"` + value + `"`
			first = false
		}
		content += "}"

		err := os.WriteFile(filepath.Join(tmpDir, lang+".json"), []byte(content), 0644)
		require.NoError(t, err)
	}

	i18nService, err := i18n.NewI18n(tmpDir)
	require.NoError(t, err)

	return i18nService
}

func createTestRenderer(t *testing.T) (*Renderer, string) {
	t.Helper()

	// Create temporary template directory
	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "locales")
	err := os.MkdirAll(localesDir, 0755)
	require.NoError(t, err)

	// Create i18n service
	i18nService := createTestI18n(t, localesDir)

	// Create base templates
	baseHTML := `{{define "base"}}<!DOCTYPE html>
<html>
<head><title>{{.Organisation}}</title></head>
<body>
{{template "content" .}}
<p>From: {{.FromName}} ({{.FromMail}})</p>
<p>Base URL: {{.BaseURL}}</p>
</body>
</html>{{end}}`

	baseTxt := `{{define "base"}}{{template "content" .}}

From: {{.FromName}} ({{.FromMail}})
Base URL: {{.BaseURL}}
Organisation: {{.Organisation}}{{end}}`

	err = os.WriteFile(filepath.Join(tmpDir, "base.html.tmpl"), []byte(baseHTML), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "base.txt.tmpl"), []byte(baseTxt), 0644)
	require.NoError(t, err)

	// Create unified test templates using i18n
	testHTML := `{{define "content"}}<h1>{{T "test.title"}}</h1><p>{{T "test.message" (dict "message" .Data.message)}}</p>{{end}}`
	testTxt := `{{define "content"}}{{T "test.title"}}
{{T "test.message" (dict "message" .Data.message)}}{{end}}`

	err = os.WriteFile(filepath.Join(tmpDir, "test.html.tmpl"), []byte(testHTML), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "test.txt.tmpl"), []byte(testTxt), 0644)
	require.NoError(t, err)

	renderer := NewRenderer(tmpDir, testBaseURL, testOrganisation, testFromName, testFromEmail, "en", i18nService)

	return renderer, tmpDir
}

// ============================================================================
// TESTS - NewRenderer
// ============================================================================

func TestNewRenderer(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "locales")
	os.MkdirAll(localesDir, 0755)
	i18nService := createTestI18n(t, localesDir)

	renderer := NewRenderer("/tmp/templates", testBaseURL, testOrganisation, testFromName, testFromEmail, "en", i18nService)

	require.NotNil(t, renderer)
	assert.Equal(t, "/tmp/templates", renderer.templateDir)
	assert.Equal(t, testBaseURL, renderer.baseURL)
	assert.Equal(t, testOrganisation, renderer.organisation)
	assert.Equal(t, testFromName, renderer.fromName)
	assert.Equal(t, testFromEmail, renderer.fromMail)
	assert.Equal(t, "en", renderer.defaultLocale)
	assert.NotNil(t, renderer.i18n)
}

// ============================================================================
// TESTS - Renderer.Render
// ============================================================================

func TestRenderer_Render_Success(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	data := map[string]any{
		"message": "Hello World",
	}

	htmlBody, textBody, err := renderer.Render("test", "en", data)

	require.NoError(t, err)
	assert.Contains(t, htmlBody, "Test Template")
	assert.Contains(t, htmlBody, "Hello World")
	assert.Contains(t, htmlBody, testOrganisation)
	assert.Contains(t, htmlBody, testBaseURL)
	assert.Contains(t, htmlBody, testFromName)

	assert.Contains(t, textBody, "Test Template")
	assert.Contains(t, textBody, "Hello World")
	assert.Contains(t, textBody, testOrganisation)
}

func TestRenderer_Render_FrenchLocale(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	data := map[string]any{
		"message": "Bonjour le monde",
	}

	htmlBody, textBody, err := renderer.Render("test", "fr", data)

	require.NoError(t, err)
	assert.Contains(t, htmlBody, "Modèle de Test")
	assert.Contains(t, htmlBody, "Bonjour le monde")

	assert.Contains(t, textBody, "Modèle de Test")
	assert.Contains(t, textBody, "Bonjour le monde")
}

func TestRenderer_Render_DefaultLocale(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	data := map[string]any{
		"message": "Default locale test",
	}

	// Empty locale should use default (en)
	htmlBody, textBody, err := renderer.Render("test", "", data)

	require.NoError(t, err)
	assert.Contains(t, htmlBody, "Test Template")
	assert.Contains(t, textBody, "Default locale test")
}

func TestRenderer_Render_TemplateNotFound(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	data := map[string]any{
		"message": "test",
	}

	_, _, err := renderer.Render("nonexistent", "en", data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestRenderer_Render_InvalidTemplateDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	localesDir := filepath.Join(tmpDir, "locales")
	os.MkdirAll(localesDir, 0755)
	i18nService := createTestI18n(t, localesDir)

	renderer := NewRenderer("/nonexistent/dir", testBaseURL, testOrganisation, testFromName, testFromEmail, "en", i18nService)

	data := map[string]any{
		"message": "test",
	}

	_, _, err := renderer.Render("test", "en", data)

	require.Error(t, err)
}

// ============================================================================
// TESTS - NewSMTPSender
// ============================================================================

func TestNewSMTPSender(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	cfg := config.MailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     testFromEmail,
		FromName: testFromName,
	}

	sender := NewSMTPSender(cfg, renderer)

	require.NotNil(t, sender)
	assert.NotNil(t, sender.config)
	assert.NotNil(t, sender.renderer)
	assert.Equal(t, "smtp.example.com", sender.config.Host)
	assert.Equal(t, 587, sender.config.Port)
}

// ============================================================================
// TESTS - SMTPSender.Send
// ============================================================================

func TestSMTPSender_Send_SMTPNotConfigured(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	// Empty host = SMTP not configured
	cfg := config.MailConfig{
		Host: "",
	}

	sender := NewSMTPSender(cfg, renderer)

	msg := Message{
		To:       []string{"recipient@example.com"},
		Subject:  "Test",
		Template: "test",
		Locale:   "en",
		Data: map[string]any{
			"message": "test",
		},
	}

	// Should not return error when SMTP not configured
	err := sender.Send(context.Background(), msg)
	assert.NoError(t, err)
}

func TestSMTPSender_Send_NoFrom(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	cfg := config.MailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "", // No from address
		FromName: testFromName,
	}

	sender := NewSMTPSender(cfg, renderer)

	msg := Message{
		To:       []string{"recipient@example.com"},
		Subject:  "Test",
		Template: "test",
		Locale:   "en",
		Data: map[string]any{
			"message": "test",
		},
	}

	err := sender.Send(context.Background(), msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ACKIFY_MAIL_FROM not set")
}

func TestSMTPSender_Send_NoRecipients(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	cfg := config.MailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     testFromEmail,
		FromName: testFromName,
	}

	sender := NewSMTPSender(cfg, renderer)

	msg := Message{
		To:       []string{}, // No recipients
		Subject:  "Test",
		Template: "test",
		Locale:   "en",
		Data: map[string]any{
			"message": "test",
		},
	}

	err := sender.Send(context.Background(), msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no recipients specified")
}

func TestSMTPSender_Send_InvalidTemplate(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	cfg := config.MailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     testFromEmail,
		FromName: testFromName,
	}

	sender := NewSMTPSender(cfg, renderer)

	msg := Message{
		To:       []string{"recipient@example.com"},
		Subject:  "Test",
		Template: "nonexistent",
		Locale:   "en",
		Data:     map[string]any{},
	}

	err := sender.Send(context.Background(), msg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render email template")
}

func TestSMTPSender_Send_SubjectPrefix(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	cfg := config.MailConfig{
		Host:          "smtp.example.com",
		Port:          587,
		Username:      "user",
		Password:      "pass",
		From:          testFromEmail,
		FromName:      testFromName,
		SubjectPrefix: "[TEST] ",
	}

	sender := NewSMTPSender(cfg, renderer)

	// We can't actually send email in tests, but we can verify the config is used
	assert.Equal(t, "[TEST] ", sender.config.SubjectPrefix)
}

func TestMessage_Structure(t *testing.T) {
	t.Parallel()

	msg := Message{
		To:       []string{"to@example.com"},
		Cc:       []string{"cc@example.com"},
		Bcc:      []string{"bcc@example.com"},
		Subject:  "Test Subject",
		Template: "test",
		Locale:   "en",
		Data: map[string]any{
			"key": "value",
		},
		Headers: map[string]string{
			"X-Custom": "value",
		},
	}

	assert.Equal(t, []string{"to@example.com"}, msg.To)
	assert.Equal(t, []string{"cc@example.com"}, msg.Cc)
	assert.Equal(t, []string{"bcc@example.com"}, msg.Bcc)
	assert.Equal(t, "Test Subject", msg.Subject)
	assert.Equal(t, "test", msg.Template)
	assert.Equal(t, "en", msg.Locale)
	assert.Equal(t, "value", msg.Data["key"])
	assert.Equal(t, "value", msg.Headers["X-Custom"])
}

// ============================================================================
// TESTS - TemplateData Structure
// ============================================================================

func TestTemplateData_Structure(t *testing.T) {
	t.Parallel()

	data := TemplateData{
		Organisation: "Test Org",
		BaseURL:      "https://example.com",
		FromName:     "Test Sender",
		FromMail:     "test@example.com",
		Data: map[string]any{
			"key1": "value1",
			"key2": 123,
		},
		T: func(key string, args ...map[string]any) string {
			return key
		},
	}

	assert.Equal(t, "Test Org", data.Organisation)
	assert.Equal(t, "https://example.com", data.BaseURL)
	assert.Equal(t, "Test Sender", data.FromName)
	assert.Equal(t, "test@example.com", data.FromMail)
	assert.Equal(t, "value1", data.Data["key1"])
	assert.Equal(t, 123, data.Data["key2"])
	assert.NotNil(t, data.T)
}

// ============================================================================
// TESTS - Concurrency
// ============================================================================

func TestRenderer_Render_Concurrent(t *testing.T) {
	t.Parallel()

	renderer, _ := createTestRenderer(t)

	const numGoroutines = 50

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			data := map[string]any{
				"message": "Concurrent test",
			}

			locale := "en"
			if id%2 == 0 {
				locale = "fr"
			}

			htmlBody, textBody, err := renderer.Render("test", locale, data)

			assert.NoError(t, err)
			assert.NotEmpty(t, htmlBody)
			assert.NotEmpty(t, textBody)
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkRenderer_Render(b *testing.B) {
	renderer, _ := createTestRenderer(&testing.T{})

	data := map[string]any{
		"message": "Benchmark test",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = renderer.Render("test", "en", data)
	}
}

func BenchmarkRenderer_Render_Parallel(b *testing.B) {
	renderer, _ := createTestRenderer(&testing.T{})

	data := map[string]any{
		"message": "Benchmark test",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, _ = renderer.Render("test", "en", data)
		}
	})
}
