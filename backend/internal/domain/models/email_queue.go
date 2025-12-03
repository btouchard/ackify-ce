// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EmailQueueStatus represents the status of an email in the queue
type EmailQueueStatus string

const (
	EmailStatusPending    EmailQueueStatus = "pending"
	EmailStatusProcessing EmailQueueStatus = "processing"
	EmailStatusSent       EmailQueueStatus = "sent"
	EmailStatusFailed     EmailQueueStatus = "failed"
	EmailStatusCancelled  EmailQueueStatus = "cancelled"
)

// EmailPriority represents email priority levels
type EmailPriority int

const (
	EmailPriorityNormal EmailPriority = 0
	EmailPriorityHigh   EmailPriority = 10
	EmailPriorityUrgent EmailPriority = 100
)

// EmailQueueItem represents an email in the processing queue
type EmailQueueItem struct {
	ID            int64            `json:"id"`
	TenantID      uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	ToAddresses   []string         `json:"to_addresses"`
	CcAddresses   []string         `json:"cc_addresses,omitempty"`
	BccAddresses  []string         `json:"bcc_addresses,omitempty"`
	Subject       string           `json:"subject"`
	Template      string           `json:"template"`
	Locale        string           `json:"locale"`
	Data          json.RawMessage  `json:"data"`
	Headers       NullRawMessage   `json:"headers,omitempty"`
	Status        EmailQueueStatus `json:"status"`
	Priority      EmailPriority    `json:"priority"`
	RetryCount    int              `json:"retry_count"`
	MaxRetries    int              `json:"max_retries"`
	CreatedAt     time.Time        `json:"created_at"`
	ScheduledFor  time.Time        `json:"scheduled_for"`
	ProcessedAt   *time.Time       `json:"processed_at,omitempty"`
	NextRetryAt   *time.Time       `json:"next_retry_at,omitempty"`
	LastError     *string          `json:"last_error,omitempty"`
	ErrorDetails  NullRawMessage   `json:"error_details,omitempty"`
	ReferenceType *string          `json:"reference_type,omitempty"`
	ReferenceID   *string          `json:"reference_id,omitempty"`
	CreatedBy     *string          `json:"created_by,omitempty"`
}

// EmailQueueInput represents the input for creating a new email queue item
type EmailQueueInput struct {
	ToAddresses   []string               `json:"to_addresses"`
	CcAddresses   []string               `json:"cc_addresses,omitempty"`
	BccAddresses  []string               `json:"bcc_addresses,omitempty"`
	Subject       string                 `json:"subject"`
	Template      string                 `json:"template"`
	Locale        string                 `json:"locale"`
	Data          map[string]interface{} `json:"data"`
	Headers       map[string]string      `json:"headers,omitempty"`
	Priority      EmailPriority          `json:"priority"`
	ScheduledFor  *time.Time             `json:"scheduled_for,omitempty"` // nil = immediate
	ReferenceType *string                `json:"reference_type,omitempty"`
	ReferenceID   *string                `json:"reference_id,omitempty"`
	CreatedBy     *string                `json:"created_by,omitempty"`
	MaxRetries    int                    `json:"max_retries"` // 0 = use default (3)
}

// EmailQueueStats represents aggregated statistics for the email queue
type EmailQueueStats struct {
	TotalPending    int              `json:"total_pending"`
	TotalProcessing int              `json:"total_processing"`
	TotalSent       int              `json:"total_sent"`
	TotalFailed     int              `json:"total_failed"`
	OldestPending   *time.Time       `json:"oldest_pending,omitempty"`
	AverageRetries  float64          `json:"average_retries"`
	ByStatus        map[string]int   `json:"by_status"`
	ByPriority      map[string]int   `json:"by_priority"`
	Last24Hours     EmailPeriodStats `json:"last_24_hours"`
}

// EmailPeriodStats represents email statistics for a time period
type EmailPeriodStats struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
	Queued int `json:"queued"`
}

// JSONB is a helper type for handling JSONB columns
type JSONB map[string]interface{}

// Value implements driver.Valuer
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data = []byte("{}")
	}

	return json.Unmarshal(data, j)
}

// NullRawMessage is a nullable json.RawMessage for database scanning
type NullRawMessage struct {
	RawMessage json.RawMessage
	Valid      bool
}

// Scan implements sql.Scanner
func (n *NullRawMessage) Scan(value interface{}) error {
	if value == nil {
		n.RawMessage = nil
		n.Valid = false
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		n.RawMessage = nil
		n.Valid = false
		return nil
	}

	n.RawMessage = data
	n.Valid = true
	return nil
}

// Value implements driver.Valuer
func (n NullRawMessage) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.RawMessage, nil
}
