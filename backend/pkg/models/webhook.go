// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID              int64             `json:"id"`
	TenantID        uuid.UUID         `json:"tenant_id" db:"tenant_id"`
	Title           string            `json:"title"`
	TargetURL       string            `json:"targetUrl"`
	Secret          string            `json:"-"`
	Active          bool              `json:"active"`
	Events          []string          `json:"events"`
	Headers         map[string]string `json:"headers,omitempty"`
	Description     string            `json:"description,omitempty"`
	CreatedBy       string            `json:"createdBy,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	LastDeliveredAt *time.Time        `json:"lastDeliveredAt,omitempty"`
	FailureCount    int               `json:"failureCount"`
}

type WebhookInput struct {
	Title       string            `json:"title"`
	TargetURL   string            `json:"targetUrl"`
	Secret      string            `json:"secret"`
	Active      bool              `json:"active"`
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers,omitempty"`
	Description string            `json:"description,omitempty"`
	CreatedBy   string            `json:"createdBy,omitempty"`
}

// NullRawMessage mirrors Null handling used elsewhere for JSONB columns
// Note: uses existing NullRawMessage from models/email_queue.go

type WebhookDeliveryStatus string

const (
	WebhookStatusPending    WebhookDeliveryStatus = "pending"
	WebhookStatusProcessing WebhookDeliveryStatus = "processing"
	WebhookStatusDelivered  WebhookDeliveryStatus = "delivered"
	WebhookStatusFailed     WebhookDeliveryStatus = "failed"
	WebhookStatusCancelled  WebhookDeliveryStatus = "cancelled"
)

type WebhookDelivery struct {
	ID              int64                 `json:"id"`
	TenantID        uuid.UUID             `json:"tenant_id" db:"tenant_id"`
	WebhookID       int64                 `json:"webhookId"`
	EventType       string                `json:"eventType"`
	EventID         string                `json:"eventId"`
	Payload         json.RawMessage       `json:"payload"`
	Status          WebhookDeliveryStatus `json:"status"`
	RetryCount      int                   `json:"retryCount"`
	MaxRetries      int                   `json:"maxRetries"`
	Priority        int                   `json:"priority"`
	CreatedAt       time.Time             `json:"createdAt"`
	ScheduledFor    time.Time             `json:"scheduledFor"`
	ProcessedAt     *time.Time            `json:"processedAt,omitempty"`
	NextRetryAt     *time.Time            `json:"nextRetryAt,omitempty"`
	RequestHeaders  NullRawMessage        `json:"requestHeaders,omitempty"`
	ResponseStatus  *int                  `json:"responseStatus,omitempty"`
	ResponseHeaders NullRawMessage        `json:"responseHeaders,omitempty"`
	ResponseBody    *string               `json:"responseBody,omitempty"`
	LastError       *string               `json:"lastError,omitempty"`
}

type WebhookDeliveryInput struct {
	WebhookID    int64
	EventType    string
	EventID      string
	Payload      map[string]interface{}
	Priority     int
	MaxRetries   int
	ScheduledFor *time.Time
}
