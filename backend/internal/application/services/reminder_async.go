// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// emailQueueRepository defines minimal interface for email queue operations
type emailQueueRepository interface {
	Enqueue(ctx context.Context, input models.EmailQueueInput) (*models.EmailQueueItem, error)
	GetQueueStats(ctx context.Context) (*models.EmailQueueStats, error)
}

// ReminderAsyncService manages email notifications using asynchronous queue
type ReminderAsyncService struct {
	expectedSignerRepo expectedSignerRepository
	reminderRepo       reminderRepository
	queueRepo          emailQueueRepository
	baseURL            string
	useAsyncQueue      bool // Feature flag to enable/disable async queue
}

// NewReminderAsyncService initializes async reminder service with queue support
func NewReminderAsyncService(
	expectedSignerRepo expectedSignerRepository,
	reminderRepo reminderRepository,
	queueRepo emailQueueRepository,
	baseURL string,
) *ReminderAsyncService {
	return &ReminderAsyncService{
		expectedSignerRepo: expectedSignerRepo,
		reminderRepo:       reminderRepo,
		queueRepo:          queueRepo,
		baseURL:            baseURL,
		useAsyncQueue:      true, // Enable async by default
	}
}

// SendRemindersAsync dispatches email notifications to queue for async processing
func (s *ReminderAsyncService) SendRemindersAsync(
	ctx context.Context,
	docID string,
	sentBy string,
	specificEmails []string,
	docURL string,
	locale string,
) (*models.ReminderSendResult, error) {

	logger.Logger.Info("Starting async reminder queueing process",
		"doc_id", docID,
		"sent_by", sentBy,
		"specific_emails_count", len(specificEmails),
		"locale", locale)

	allSigners, err := s.expectedSignerRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get expected signers for reminders",
			"doc_id", docID,
			"error", err.Error())
		return nil, fmt.Errorf("failed to get expected signers: %w", err)
	}

	logger.Logger.Debug("Retrieved expected signers",
		"doc_id", docID,
		"total_signers", len(allSigners))

	// Filter pending signers
	var pendingSigners []*models.ExpectedSignerWithStatus
	for _, signer := range allSigners {
		if !signer.HasSigned {
			if len(specificEmails) > 0 {
				if containsEmail(specificEmails, signer.Email) {
					pendingSigners = append(pendingSigners, signer)
				}
			} else {
				pendingSigners = append(pendingSigners, signer)
			}
		}
	}

	logger.Logger.Info("Identified pending signers",
		"doc_id", docID,
		"pending_count", len(pendingSigners),
		"total_signers", len(allSigners))

	if len(pendingSigners) == 0 {
		logger.Logger.Info("No pending signers found, no reminders to queue",
			"doc_id", docID)
		return &models.ReminderSendResult{
			TotalAttempted:   0,
			SuccessfullySent: 0,
			Failed:           0,
		}, nil
	}

	result := &models.ReminderSendResult{
		TotalAttempted: len(pendingSigners),
	}

	// Queue emails asynchronously
	for _, signer := range pendingSigners {
		err := s.queueSingleReminder(ctx, docID, signer.Email, signer.Name, sentBy, docURL, locale)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", signer.Email, err))
		} else {
			result.SuccessfullySent++
		}
	}

	logger.Logger.Info("Reminder queueing completed",
		"doc_id", docID,
		"total_attempted", result.TotalAttempted,
		"successfully_queued", result.SuccessfullySent,
		"failed", result.Failed)

	return result, nil
}

// queueSingleReminder queues a reminder for a single signer
func (s *ReminderAsyncService) queueSingleReminder(
	ctx context.Context,
	docID string,
	recipientEmail string,
	recipientName string,
	sentBy string,
	docURL string,
	locale string,
) error {

	logger.Logger.Debug("Queueing reminder for signer",
		"doc_id", docID,
		"recipient_email", recipientEmail,
		"recipient_name", recipientName,
		"sent_by", sentBy)

	signURL := fmt.Sprintf("%s/sign?doc=%s", s.baseURL, docID)

	// Prepare email data (keys must match template variables)
	data := map[string]interface{}{
		"DocID":         docID,
		"DocURL":        docURL,
		"SignURL":       signURL,
		"RecipientName": recipientName,
		"Locale":        locale,
	}

	// Create email queue input
	refType := "signature_reminder"
	input := models.EmailQueueInput{
		ToAddresses:   []string{recipientEmail},
		Subject:       "Reminder: Document signature required",
		Template:      "signature_reminder",
		Locale:        locale,
		Data:          data,
		Priority:      models.EmailPriorityHigh,
		ReferenceType: &refType,
		ReferenceID:   &docID,
		CreatedBy:     &sentBy,
		MaxRetries:    5, // More retries for important reminders
	}

	// Queue the email
	item, err := s.queueRepo.Enqueue(ctx, input)
	if err != nil {
		logger.Logger.Warn("Failed to queue reminder email",
			"doc_id", docID,
			"recipient_email", recipientEmail,
			"error", err.Error())

		// Log the failure
		log := &models.ReminderLog{
			DocID:          docID,
			RecipientEmail: recipientEmail,
			SentAt:         time.Now(),
			SentBy:         sentBy,
			TemplateUsed:   "signature_reminder",
			Status:         "failed",
		}
		errMsg := fmt.Sprintf("Failed to queue: %v", err)
		log.ErrorMessage = &errMsg

		if logErr := s.reminderRepo.LogReminder(ctx, log); logErr != nil {
			logger.Logger.Error("Failed to log reminder queue error",
				"doc_id", docID,
				"recipient_email", recipientEmail,
				"log_error", logErr.Error(),
				"original_error", err.Error())
		}

		return fmt.Errorf("failed to queue email: %w", err)
	}

	logger.Logger.Info("Reminder email queued successfully",
		"doc_id", docID,
		"recipient_email", recipientEmail,
		"queue_id", item.ID)

	// Log successful queueing
	log := &models.ReminderLog{
		DocID:          docID,
		RecipientEmail: recipientEmail,
		SentAt:         time.Now(),
		SentBy:         sentBy,
		TemplateUsed:   "signature_reminder",
		Status:         "queued", // New status for queued emails
	}

	if err := s.reminderRepo.LogReminder(ctx, log); err != nil {
		logger.Logger.Error("Failed to log successful reminder queueing",
			"doc_id", docID,
			"recipient_email", recipientEmail,
			"error", err.Error())
		// Non-critical error, email is already queued
	}

	return nil
}

// GetQueueStats returns current email queue statistics
func (s *ReminderAsyncService) GetQueueStats(ctx context.Context) (*models.EmailQueueStats, error) {
	return s.queueRepo.GetQueueStats(ctx)
}

// GetReminderStats retrieves aggregated reminder metrics for monitoring dashboard
func (s *ReminderAsyncService) GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error) {
	return s.reminderRepo.GetReminderStats(ctx, docID)
}

// GetReminderHistory retrieves complete email send log with success/failure tracking
func (s *ReminderAsyncService) GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
	return s.reminderRepo.GetReminderHistory(ctx, docID)
}

// EnableAsync enables or disables async queue processing
func (s *ReminderAsyncService) EnableAsync(enabled bool) {
	s.useAsyncQueue = enabled
	logger.Logger.Info("Async queue processing toggled", "enabled", enabled)
}

// SendReminders is a compatibility method that calls SendRemindersAsync
// This allows the service to work with existing interfaces expecting SendReminders
func (s *ReminderAsyncService) SendReminders(
	ctx context.Context,
	docID string,
	sentBy string,
	specificEmails []string,
	docURL string,
	locale string,
) (*models.ReminderSendResult, error) {
	return s.SendRemindersAsync(ctx, docID, sentBy, specificEmails, docURL, locale)
}
