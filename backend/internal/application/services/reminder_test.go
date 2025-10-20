// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/email"
)

// Mock implementations for testing
type mockExpectedSignerRepository struct {
	listWithStatusFunc func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error)
}

func (m *mockExpectedSignerRepository) ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
	if m.listWithStatusFunc != nil {
		return m.listWithStatusFunc(ctx, docID)
	}
	return nil, nil
}

type mockReminderRepository struct {
	logReminderFunc        func(ctx context.Context, log *models.ReminderLog) error
	getReminderHistoryFunc func(ctx context.Context, docID string) ([]*models.ReminderLog, error)
	getReminderStatsFunc   func(ctx context.Context, docID string) (*models.ReminderStats, error)
}

func (m *mockReminderRepository) LogReminder(ctx context.Context, log *models.ReminderLog) error {
	if m.logReminderFunc != nil {
		return m.logReminderFunc(ctx, log)
	}
	return nil
}

func (m *mockReminderRepository) GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
	if m.getReminderHistoryFunc != nil {
		return m.getReminderHistoryFunc(ctx, docID)
	}
	return nil, nil
}

func (m *mockReminderRepository) GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error) {
	if m.getReminderStatsFunc != nil {
		return m.getReminderStatsFunc(ctx, docID)
	}
	return nil, nil
}

type mockEmailSender struct {
	sendFunc func(ctx context.Context, msg email.Message) error
}

func (m *mockEmailSender) Send(ctx context.Context, msg email.Message) error {
	if m.sendFunc != nil {
		return m.sendFunc(ctx, msg)
	}
	return nil
}

// Test helper function
func TestContainsEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "Email found",
			slice:    []string{"alice@example.com", "bob@example.com", "charlie@example.com"},
			item:     "bob@example.com",
			expected: true,
		},
		{
			name:     "Email not found",
			slice:    []string{"alice@example.com", "bob@example.com"},
			item:     "charlie@example.com",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			item:     "test@example.com",
			expected: false,
		},
		{
			name:     "Case sensitive",
			slice:    []string{"Test@Example.com"},
			item:     "test@example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := containsEmail(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("containsEmail(%v, %q) = %v, want %v", tt.slice, tt.item, result, tt.expected)
			}
		})
	}
}

// Test SendReminders with no pending signers
func TestReminderService_SendReminders_NoPendingSigners(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{
				{ExpectedSigner: models.ExpectedSigner{Email: "signed@example.com"}, HasSigned: true},
			}, nil
		},
	}

	mockReminderRepo := &mockReminderRepository{}
	mockEmailSender := &mockEmailSender{}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", nil, "https://example.com/doc.pdf", "en")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.TotalAttempted != 0 {
		t.Errorf("Expected 0 total attempted, got %d", result.TotalAttempted)
	}

	if result.SuccessfullySent != 0 {
		t.Errorf("Expected 0 successfully sent, got %d", result.SuccessfullySent)
	}
}

// Test SendReminders with successful email send
func TestReminderService_SendReminders_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{
				{ExpectedSigner: models.ExpectedSigner{Email: "pending@example.com", Name: "Pending User"}, HasSigned: false},
			}, nil
		},
	}

	loggedReminder := false
	mockReminderRepo := &mockReminderRepository{
		logReminderFunc: func(ctx context.Context, log *models.ReminderLog) error {
			loggedReminder = true
			if log.Status != "sent" {
				t.Errorf("Expected status 'sent', got '%s'", log.Status)
			}
			return nil
		},
	}

	emailSent := false
	mockEmailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, msg email.Message) error {
			emailSent = true
			if len(msg.To) != 1 || msg.To[0] != "pending@example.com" {
				t.Errorf("Expected email to 'pending@example.com', got %v", msg.To)
			}
			return nil
		},
	}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", nil, "https://example.com/doc.pdf", "en")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.TotalAttempted != 1 {
		t.Errorf("Expected 1 total attempted, got %d", result.TotalAttempted)
	}

	if result.SuccessfullySent != 1 {
		t.Errorf("Expected 1 successfully sent, got %d", result.SuccessfullySent)
	}

	if result.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", result.Failed)
	}

	if !emailSent {
		t.Error("Expected email to be sent")
	}

	if !loggedReminder {
		t.Error("Expected reminder to be logged")
	}
}

// Test SendReminders with email failure
func TestReminderService_SendReminders_EmailFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{
				{ExpectedSigner: models.ExpectedSigner{Email: "pending@example.com", Name: "Pending User"}, HasSigned: false},
			}, nil
		},
	}

	loggedReminder := false
	mockReminderRepo := &mockReminderRepository{
		logReminderFunc: func(ctx context.Context, log *models.ReminderLog) error {
			loggedReminder = true
			if log.Status != "failed" {
				t.Errorf("Expected status 'failed', got '%s'", log.Status)
			}
			if log.ErrorMessage == nil {
				t.Error("Expected error message to be set")
			}
			return nil
		},
	}

	mockEmailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, msg email.Message) error {
			return errors.New("SMTP connection failed")
		},
	}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", nil, "https://example.com/doc.pdf", "en")

	if err != nil {
		t.Fatalf("Expected no error from SendReminders, got: %v", err)
	}

	if result.TotalAttempted != 1 {
		t.Errorf("Expected 1 total attempted, got %d", result.TotalAttempted)
	}

	if result.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", result.Failed)
	}

	if result.SuccessfullySent != 0 {
		t.Errorf("Expected 0 successfully sent, got %d", result.SuccessfullySent)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error message, got %d", len(result.Errors))
	}

	if !loggedReminder {
		t.Error("Expected failed reminder to be logged")
	}
}

// Test SendReminders with specific emails filter
func TestReminderService_SendReminders_SpecificEmails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{
				{ExpectedSigner: models.ExpectedSigner{Email: "pending1@example.com"}, HasSigned: false},
				{ExpectedSigner: models.ExpectedSigner{Email: "pending2@example.com"}, HasSigned: false},
				{ExpectedSigner: models.ExpectedSigner{Email: "pending3@example.com"}, HasSigned: false},
			}, nil
		},
	}

	emailsSent := []string{}
	mockReminderRepo := &mockReminderRepository{
		logReminderFunc: func(ctx context.Context, log *models.ReminderLog) error {
			return nil
		},
	}

	mockEmailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, msg email.Message) error {
			emailsSent = append(emailsSent, msg.To[0])
			return nil
		},
	}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	specificEmails := []string{"pending2@example.com"}
	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", specificEmails, "https://example.com/doc.pdf", "en")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.TotalAttempted != 1 {
		t.Errorf("Expected 1 total attempted, got %d", result.TotalAttempted)
	}

	if len(emailsSent) != 1 || emailsSent[0] != "pending2@example.com" {
		t.Errorf("Expected only 'pending2@example.com' to receive email, got %v", emailsSent)
	}
}

// Test SendReminders with repository error
func TestReminderService_SendReminders_RepositoryError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return nil, errors.New("database connection failed")
		},
	}

	mockReminderRepo := &mockReminderRepository{}
	mockEmailSender := &mockEmailSender{}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", nil, "https://example.com/doc.pdf", "en")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

// Test GetReminderHistory
func TestReminderService_GetReminderHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedLogs := []*models.ReminderLog{
		{
			DocID:          "doc1",
			RecipientEmail: "user@example.com",
			SentAt:         time.Now(),
			SentBy:         "admin@example.com",
			Status:         "sent",
		},
	}

	mockReminderRepo := &mockReminderRepository{
		getReminderHistoryFunc: func(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
			if docID != "doc1" {
				t.Errorf("Expected docID 'doc1', got '%s'", docID)
			}
			return expectedLogs, nil
		},
	}

	service := NewReminderService(&mockExpectedSignerRepository{}, mockReminderRepo, &mockEmailSender{}, "https://example.com")

	logs, err := service.GetReminderHistory(ctx, "doc1")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if logs[0].RecipientEmail != "user@example.com" {
		t.Errorf("Expected recipient 'user@example.com', got '%s'", logs[0].RecipientEmail)
	}
}

// Test GetReminderStats
func TestReminderService_GetReminderStats(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	now := time.Now()
	expectedStats := &models.ReminderStats{
		TotalSent:    5,
		LastSentAt:   &now,
		PendingCount: 2,
	}

	mockReminderRepo := &mockReminderRepository{
		getReminderStatsFunc: func(ctx context.Context, docID string) (*models.ReminderStats, error) {
			if docID != "doc1" {
				t.Errorf("Expected docID 'doc1', got '%s'", docID)
			}
			return expectedStats, nil
		},
	}

	service := NewReminderService(&mockExpectedSignerRepository{}, mockReminderRepo, &mockEmailSender{}, "https://example.com")

	stats, err := service.GetReminderStats(ctx, "doc1")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if stats.TotalSent != 5 {
		t.Errorf("Expected 5 total sent, got %d", stats.TotalSent)
	}

	if stats.PendingCount != 2 {
		t.Errorf("Expected 2 pending, got %d", stats.PendingCount)
	}

	if stats.LastSentAt == nil {
		t.Error("Expected LastSentAt to be set")
	}
}

// Test SendReminders with multiple pending signers
func TestReminderService_SendReminders_MultiplePending(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{
				{ExpectedSigner: models.ExpectedSigner{Email: "pending1@example.com", Name: "User 1"}, HasSigned: false},
				{ExpectedSigner: models.ExpectedSigner{Email: "pending2@example.com", Name: "User 2"}, HasSigned: false},
				{ExpectedSigner: models.ExpectedSigner{Email: "already-signed@example.com", Name: "User 3"}, HasSigned: true},
			}, nil
		},
	}

	emailsSent := 0
	mockReminderRepo := &mockReminderRepository{
		logReminderFunc: func(ctx context.Context, log *models.ReminderLog) error {
			return nil
		},
	}

	mockEmailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, msg email.Message) error {
			emailsSent++
			return nil
		},
	}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", nil, "https://example.com/doc.pdf", "en")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.TotalAttempted != 2 {
		t.Errorf("Expected 2 total attempted, got %d", result.TotalAttempted)
	}

	if result.SuccessfullySent != 2 {
		t.Errorf("Expected 2 successfully sent, got %d", result.SuccessfullySent)
	}

	if emailsSent != 2 {
		t.Errorf("Expected 2 emails sent, got %d", emailsSent)
	}
}

// Test SendReminders with log failure after successful email
func TestReminderService_SendReminders_LogFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockExpectedRepo := &mockExpectedSignerRepository{
		listWithStatusFunc: func(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
			return []*models.ExpectedSignerWithStatus{
				{ExpectedSigner: models.ExpectedSigner{Email: "pending@example.com", Name: "Pending User"}, HasSigned: false},
			}, nil
		},
	}

	mockReminderRepo := &mockReminderRepository{
		logReminderFunc: func(ctx context.Context, log *models.ReminderLog) error {
			return errors.New("database write failed")
		},
	}

	mockEmailSender := &mockEmailSender{
		sendFunc: func(ctx context.Context, msg email.Message) error {
			return nil // Email succeeds
		},
	}

	service := NewReminderService(mockExpectedRepo, mockReminderRepo, mockEmailSender, "https://example.com")

	result, err := service.SendReminders(ctx, "doc1", "admin@example.com", nil, "https://example.com/doc.pdf", "en")

	if err != nil {
		t.Fatalf("Expected no error from SendReminders, got: %v", err)
	}

	// The send should fail because logging failed
	if result.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", result.Failed)
	}

	if result.SuccessfullySent != 0 {
		t.Errorf("Expected 0 successfully sent, got %d", result.SuccessfullySent)
	}
}
