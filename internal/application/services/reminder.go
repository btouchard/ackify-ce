// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/infrastructure/database"
	"github.com/btouchard/ackify-ce/internal/infrastructure/email"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

type ReminderService struct {
	expectedSignerRepo *database.ExpectedSignerRepository
	reminderRepo       *database.ReminderRepository
	emailSender        email.Sender
	baseURL            string
}

func NewReminderService(
	expectedSignerRepo *database.ExpectedSignerRepository,
	reminderRepo *database.ReminderRepository,
	emailSender email.Sender,
	baseURL string,
) *ReminderService {
	return &ReminderService{
		expectedSignerRepo: expectedSignerRepo,
		reminderRepo:       reminderRepo,
		emailSender:        emailSender,
		baseURL:            baseURL,
	}
}

// SendReminders sends reminder emails to pending signers
func (s *ReminderService) SendReminders(
	ctx context.Context,
	docID string,
	sentBy string,
	specificEmails []string,
	docURL string,
) (*models.ReminderSendResult, error) {

	allSigners, err := s.expectedSignerRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expected signers: %w", err)
	}

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

	if len(pendingSigners) == 0 {
		return &models.ReminderSendResult{
			TotalAttempted:   0,
			SuccessfullySent: 0,
			Failed:           0,
		}, nil
	}

	result := &models.ReminderSendResult{
		TotalAttempted: len(pendingSigners),
	}

	for _, signer := range pendingSigners {
		err := s.sendSingleReminder(ctx, docID, signer.Email, signer.Name, sentBy, docURL)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", signer.Email, err))
		} else {
			result.SuccessfullySent++
		}
	}

	return result, nil
}

// sendSingleReminder sends a reminder to a single signer
func (s *ReminderService) sendSingleReminder(
	ctx context.Context,
	docID string,
	recipientEmail string,
	recipientName string,
	sentBy string,
	docURL string,
) error {

	signURL := fmt.Sprintf("%s/sign?doc=%s", s.baseURL, docID)

	log := &models.ReminderLog{
		DocID:          docID,
		RecipientEmail: recipientEmail,
		SentAt:         time.Now(),
		SentBy:         sentBy,
		TemplateUsed:   "signature_reminder",
		Status:         "sent",
	}

	err := email.SendSignatureReminderEmail(ctx, s.emailSender, []string{recipientEmail}, "fr", docID, docURL, signURL, recipientName)
	if err != nil {
		log.Status = "failed"
		errMsg := err.Error()
		log.ErrorMessage = &errMsg

		if logErr := s.reminderRepo.LogReminder(ctx, log); logErr != nil {
			logger.Logger.Error("failed to log reminder error", "error", logErr, "original_error", err)
		}

		return fmt.Errorf("failed to send email: %w", err)
	}

	if err := s.reminderRepo.LogReminder(ctx, log); err != nil {
		logger.Logger.Error("failed to log successful reminder", "error", err)
		return fmt.Errorf("email sent but failed to log: %w", err)
	}

	return nil
}

// GetReminderStats returns reminder statistics for a document
func (s *ReminderService) GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error) {
	return s.reminderRepo.GetReminderStats(ctx, docID)
}

// GetReminderHistory returns reminder history for a document
func (s *ReminderService) GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
	return s.reminderRepo.GetReminderHistory(ctx, docID)
}

func containsEmail(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
