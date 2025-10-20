// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"
	"fmt"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
)

// QueueSender implements the Sender interface by queuing emails instead of sending them directly
type QueueSender struct {
	queueRepo QueueRepository
	baseURL   string
}

// NewQueueSender creates a new queue-based email sender
func NewQueueSender(queueRepo QueueRepository, baseURL string) *QueueSender {
	return &QueueSender{
		queueRepo: queueRepo,
		baseURL:   baseURL,
	}
}

// Send queues an email for asynchronous processing
func (q *QueueSender) Send(ctx context.Context, msg Message) error {
	// Convert message data to proper format
	data := msg.Data
	if data == nil {
		data = make(map[string]interface{})
	}

	input := models.EmailQueueInput{
		ToAddresses:  msg.To,
		CcAddresses:  msg.Cc,
		BccAddresses: msg.Bcc,
		Subject:      msg.Subject,
		Template:     msg.Template,
		Locale:       msg.Locale,
		Data:         data,
		Headers:      msg.Headers,
		Priority:     models.EmailPriorityNormal,
	}

	// Set priority based on template type
	switch msg.Template {
	case "signature_reminder":
		input.Priority = models.EmailPriorityHigh
	case "welcome", "notification":
		input.Priority = models.EmailPriorityNormal
	default:
		input.Priority = models.EmailPriorityNormal
	}

	// Queue the email
	_, err := q.queueRepo.Enqueue(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}

	return nil
}

// QueueSignatureReminderEmail queues a signature reminder email
func QueueSignatureReminderEmail(
	ctx context.Context,
	queueRepo QueueRepository,
	recipients []string,
	locale string,
	docID string,
	docURL string,
	signURL string,
	recipientName string,
	sentBy string,
) error {
	data := map[string]interface{}{
		"doc_id":         docID,
		"doc_url":        docURL,
		"sign_url":       signURL,
		"recipient_name": recipientName,
		"locale":         locale,
	}

	// Create a reference for tracking
	refType := "signature_reminder"

	input := models.EmailQueueInput{
		ToAddresses:   recipients,
		Subject:       "Reminder: Document signature required",
		Template:      "signature_reminder",
		Locale:        locale,
		Data:          data,
		Priority:      models.EmailPriorityHigh,
		ReferenceType: &refType,
		ReferenceID:   &docID,
		CreatedBy:     &sentBy,
	}

	_, err := queueRepo.Enqueue(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to queue signature reminder: %w", err)
	}

	return nil
}
