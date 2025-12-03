// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"time"

	"github.com/google/uuid"
)

// ReminderLog represents a log entry for an email reminder sent to a signer
type ReminderLog struct {
	ID             int64     `json:"id" db:"id"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	DocID          string    `json:"doc_id" db:"doc_id"`
	RecipientEmail string    `json:"recipient_email" db:"recipient_email"`
	SentAt         time.Time `json:"sent_at" db:"sent_at"`
	SentBy         string    `json:"sent_by" db:"sent_by"`
	TemplateUsed   string    `json:"template_used" db:"template_used"`
	Status         string    `json:"status" db:"status"`
	ErrorMessage   *string   `json:"error_message,omitempty" db:"error_message"`
}

// ReminderStats provides statistics about reminders for a document
type ReminderStats struct {
	TotalSent    int        `json:"total_sent"`
	LastSentAt   *time.Time `json:"last_sent_at,omitempty"`
	PendingCount int        `json:"pending_count"`
}

// ReminderSendResult represents the result of a bulk reminder send operation
type ReminderSendResult struct {
	TotalAttempted   int      `json:"total_attempted"`
	SuccessfullySent int      `json:"successfully_sent"`
	Failed           int      `json:"failed"`
	Errors           []string `json:"errors,omitempty"`
}
