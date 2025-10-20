// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"
	"errors"
	"testing"
)

// Mock sender for testing
type mockSender struct {
	sendFunc func(ctx context.Context, msg Message) error
	lastMsg  *Message
}

func (m *mockSender) Send(ctx context.Context, msg Message) error {
	m.lastMsg = &msg
	if m.sendFunc != nil {
		return m.sendFunc(ctx, msg)
	}
	return nil
}

func TestSendEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		template    string
		to          []string
		locale      string
		subject     string
		data        map[string]any
		sendError   error
		expectError bool
	}{
		{
			name:     "Send email successfully",
			template: "test_template",
			to:       []string{"user@example.com"},
			locale:   "en",
			subject:  "Test Subject",
			data: map[string]any{
				"name": "John",
			},
			sendError:   nil,
			expectError: false,
		},
		{
			name:     "Send email with multiple recipients",
			template: "welcome",
			to:       []string{"user1@example.com", "user2@example.com"},
			locale:   "fr",
			subject:  "Bienvenue",
			data: map[string]any{
				"company": "Acme Corp",
			},
			sendError:   nil,
			expectError: false,
		},
		{
			name:        "Send email with error",
			template:    "error_template",
			to:          []string{"user@example.com"},
			locale:      "en",
			subject:     "Error Test",
			data:        nil,
			sendError:   errors.New("SMTP connection failed"),
			expectError: true,
		},
		{
			name:        "Send email with empty data",
			template:    "simple_template",
			to:          []string{"test@example.com"},
			locale:      "en",
			subject:     "Simple Email",
			data:        map[string]any{},
			sendError:   nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mock := &mockSender{
				sendFunc: func(ctx context.Context, msg Message) error {
					return tt.sendError
				},
			}

			err := SendEmail(ctx, mock, tt.template, tt.to, tt.locale, tt.subject, tt.data)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify message was constructed correctly
			if mock.lastMsg == nil {
				t.Fatal("Expected message to be captured")
			}

			if mock.lastMsg.Template != tt.template {
				t.Errorf("Expected template '%s', got '%s'", tt.template, mock.lastMsg.Template)
			}

			if mock.lastMsg.Subject != tt.subject {
				t.Errorf("Expected subject '%s', got '%s'", tt.subject, mock.lastMsg.Subject)
			}

			if mock.lastMsg.Locale != tt.locale {
				t.Errorf("Expected locale '%s', got '%s'", tt.locale, mock.lastMsg.Locale)
			}

			if len(mock.lastMsg.To) != len(tt.to) {
				t.Errorf("Expected %d recipients, got %d", len(tt.to), len(mock.lastMsg.To))
			}
		})
	}
}

func TestSendSignatureReminderEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		to              []string
		locale          string
		docID           string
		docURL          string
		signURL         string
		recipientName   string
		expectedSubject string
		sendError       error
		expectError     bool
	}{
		{
			name:            "Send reminder in English",
			to:              []string{"user@example.com"},
			locale:          "en",
			docID:           "doc123",
			docURL:          "https://example.com/doc.pdf",
			signURL:         "https://example.com/sign?doc=doc123",
			recipientName:   "John Doe",
			expectedSubject: "Reminder: Document reading confirmation required",
			sendError:       nil,
			expectError:     false,
		},
		{
			name:            "Send reminder in French",
			to:              []string{"utilisateur@exemple.fr"},
			locale:          "fr",
			docID:           "doc456",
			docURL:          "https://exemple.fr/document.pdf",
			signURL:         "https://exemple.fr/sign?doc=doc456",
			recipientName:   "Marie Dupont",
			expectedSubject: "Rappel : Confirmation de lecture de document requise",
			sendError:       nil,
			expectError:     false,
		},
		{
			name:            "Send reminder with unknown locale defaults to English",
			to:              []string{"user@example.com"},
			locale:          "es",
			docID:           "doc789",
			docURL:          "https://example.com/doc.pdf",
			signURL:         "https://example.com/sign?doc=doc789",
			recipientName:   "Juan Garcia",
			expectedSubject: "Reminder: Document reading confirmation required",
			sendError:       nil,
			expectError:     false,
		},
		{
			name:            "Send reminder with error",
			to:              []string{"user@example.com"},
			locale:          "en",
			docID:           "doc999",
			docURL:          "https://example.com/doc.pdf",
			signURL:         "https://example.com/sign?doc=doc999",
			recipientName:   "Test User",
			expectedSubject: "Reminder: Document reading confirmation required",
			sendError:       errors.New("email server unavailable"),
			expectError:     true,
		},
		{
			name:            "Send reminder with empty recipient name",
			to:              []string{"user@example.com"},
			locale:          "en",
			docID:           "doc000",
			docURL:          "https://example.com/doc.pdf",
			signURL:         "https://example.com/sign?doc=doc000",
			recipientName:   "",
			expectedSubject: "Reminder: Document reading confirmation required",
			sendError:       nil,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mock := &mockSender{
				sendFunc: func(ctx context.Context, msg Message) error {
					return tt.sendError
				},
			}

			err := SendSignatureReminderEmail(ctx, mock, tt.to, tt.locale, tt.docID, tt.docURL, tt.signURL, tt.recipientName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify message construction
			if mock.lastMsg == nil {
				t.Fatal("Expected message to be captured")
			}

			if mock.lastMsg.Template != "signature_reminder" {
				t.Errorf("Expected template 'signature_reminder', got '%s'", mock.lastMsg.Template)
			}

			if mock.lastMsg.Subject != tt.expectedSubject {
				t.Errorf("Expected subject '%s', got '%s'", tt.expectedSubject, mock.lastMsg.Subject)
			}

			if mock.lastMsg.Locale != tt.locale {
				t.Errorf("Expected locale '%s', got '%s'", tt.locale, mock.lastMsg.Locale)
			}

			// Verify data fields
			if mock.lastMsg.Data == nil {
				t.Fatal("Expected data to be present")
			}

			if docID, ok := mock.lastMsg.Data["DocID"].(string); !ok || docID != tt.docID {
				t.Errorf("Expected DocID '%s', got '%v'", tt.docID, mock.lastMsg.Data["DocID"])
			}

			if docURL, ok := mock.lastMsg.Data["DocURL"].(string); !ok || docURL != tt.docURL {
				t.Errorf("Expected DocURL '%s', got '%v'", tt.docURL, mock.lastMsg.Data["DocURL"])
			}

			if signURL, ok := mock.lastMsg.Data["SignURL"].(string); !ok || signURL != tt.signURL {
				t.Errorf("Expected SignURL '%s', got '%v'", tt.signURL, mock.lastMsg.Data["SignURL"])
			}

			if recipientName, ok := mock.lastMsg.Data["RecipientName"].(string); !ok || recipientName != tt.recipientName {
				t.Errorf("Expected RecipientName '%s', got '%v'", tt.recipientName, mock.lastMsg.Data["RecipientName"])
			}
		})
	}
}
