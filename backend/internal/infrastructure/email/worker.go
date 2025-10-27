// SPDX-License-Identifier: AGPL-3.0-or-later
package email

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// QueueRepository defines the interface for email queue operations
type QueueRepository interface {
	Enqueue(ctx context.Context, input models.EmailQueueInput) (*models.EmailQueueItem, error)
	GetNextToProcess(ctx context.Context, limit int) ([]*models.EmailQueueItem, error)
	MarkAsSent(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, err error, shouldRetry bool) error
	GetRetryableEmails(ctx context.Context, limit int) ([]*models.EmailQueueItem, error)
	CleanupOldEmails(ctx context.Context, olderThan time.Duration) (int64, error)
}

// Worker processes emails from the queue asynchronously
type Worker struct {
	queueRepo QueueRepository
	sender    Sender
	renderer  *Renderer
	publisher EventPublisher

	// Worker configuration
	batchSize       int
	pollInterval    time.Duration
	cleanupInterval time.Duration
	cleanupAge      time.Duration
	maxConcurrent   int

	// Control
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	stopChan chan struct{}
	started  bool
	mu       sync.Mutex
}

// WorkerConfig contains configuration for the email worker
type WorkerConfig struct {
	BatchSize       int           // Number of emails to process in each batch (default: 10)
	PollInterval    time.Duration // How often to check for new emails (default: 5s)
	CleanupInterval time.Duration // How often to cleanup old emails (default: 1 hour)
	CleanupAge      time.Duration // Age of emails to cleanup (default: 7 days)
	MaxConcurrent   int           // Maximum concurrent email sends (default: 5)
}

// DefaultWorkerConfig returns default worker configuration
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		BatchSize:       10,
		PollInterval:    5 * time.Second,
		CleanupInterval: 1 * time.Hour,
		CleanupAge:      7 * 24 * time.Hour, // 7 days
		MaxConcurrent:   5,
	}
}

// NewWorker creates a new email worker
func NewWorker(queueRepo QueueRepository, sender Sender, renderer *Renderer, config WorkerConfig) *Worker {
	// Apply defaults
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 1 * time.Hour
	}
	if config.CleanupAge <= 0 {
		config.CleanupAge = 7 * 24 * time.Hour
	}
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 5
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		queueRepo:       queueRepo,
		sender:          sender,
		renderer:        renderer,
		batchSize:       config.BatchSize,
		pollInterval:    config.PollInterval,
		cleanupInterval: config.CleanupInterval,
		cleanupAge:      config.CleanupAge,
		maxConcurrent:   config.MaxConcurrent,
		ctx:             ctx,
		cancel:          cancel,
		stopChan:        make(chan struct{}),
	}
}

// EventPublisher publishes webhook-like events ( decoupled interface )
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]interface{}) error
}

// SetPublisher injects an optional event publisher (e.g., webhooks)
func (w *Worker) SetPublisher(p EventPublisher) { w.publisher = p }

// Start begins processing emails from the queue
func (w *Worker) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return fmt.Errorf("worker already started")
	}

	logger.Logger.Info("Starting email worker",
		"batch_size", w.batchSize,
		"poll_interval", w.pollInterval,
		"max_concurrent", w.maxConcurrent)

	w.started = true

	// Start the main processing loop
	w.wg.Add(1)
	go w.processLoop()

	// Start the cleanup loop
	w.wg.Add(1)
	go w.cleanupLoop()

	return nil
}

// Stop gracefully stops the worker
func (w *Worker) Stop() error {
	w.mu.Lock()
	if !w.started {
		w.mu.Unlock()
		return fmt.Errorf("worker not started")
	}
	w.mu.Unlock()

	logger.Logger.Info("Stopping email worker...")

	// Signal shutdown
	w.cancel()
	close(w.stopChan)

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Logger.Info("Email worker stopped gracefully")
	case <-time.After(30 * time.Second):
		logger.Logger.Warn("Email worker stop timeout, some operations may not have completed")
	}

	w.mu.Lock()
	w.started = false
	w.mu.Unlock()

	return nil
}

// processLoop is the main processing loop
func (w *Worker) processLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Immediate first check
	w.processBatch()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.processBatch()
		}
	}
}

// processBatch processes a batch of emails
func (w *Worker) processBatch() {
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	// Get next batch of emails
	emails, err := w.queueRepo.GetNextToProcess(ctx, w.batchSize)
	if err != nil {
		logger.Logger.Error("Failed to get emails to process", "error", err.Error())
		return
	}

	if len(emails) == 0 {
		// Also check for retryable emails
		emails, err = w.queueRepo.GetRetryableEmails(ctx, w.batchSize)
		if err != nil {
			logger.Logger.Error("Failed to get retryable emails", "error", err.Error())
			return
		}
		if len(emails) == 0 {
			return // Nothing to process
		}
	}

	logger.Logger.Debug("Processing email batch", "count", len(emails))

	// Process emails concurrently with limited concurrency
	sem := make(chan struct{}, w.maxConcurrent)
	var wg sync.WaitGroup

	for _, email := range emails {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(item *models.EmailQueueItem) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			w.processEmail(ctx, item)
		}(email)
	}

	wg.Wait()
}

// processEmail processes a single email
func (w *Worker) processEmail(ctx context.Context, item *models.EmailQueueItem) {
	logger.Logger.Debug("Processing email",
		"id", item.ID,
		"template", item.Template,
		"retry_count", item.RetryCount)

	// Convert data from JSON to map
	var data map[string]interface{}
	if len(item.Data) > 0 {
		if err := json.Unmarshal(item.Data, &data); err != nil {
			logger.Logger.Error("Failed to unmarshal email data",
				"id", item.ID,
				"error", err.Error())
			// Mark as failed without retry (data corruption)
			w.queueRepo.MarkAsFailed(ctx, item.ID, err, false)
			return
		}
	}

	// Convert headers from JSON to map
	var headers map[string]string
	if item.Headers.Valid && len(item.Headers.RawMessage) > 0 {
		if err := json.Unmarshal(item.Headers.RawMessage, &headers); err != nil {
			logger.Logger.Error("Failed to unmarshal email headers",
				"id", item.ID,
				"error", err.Error())
			// Continue without headers
			headers = nil
		}
	}

	// Create message
	msg := Message{
		To:       item.ToAddresses,
		Cc:       item.CcAddresses,
		Bcc:      item.BccAddresses,
		Subject:  item.Subject,
		Template: item.Template,
		Locale:   item.Locale,
		Data:     data,
		Headers:  headers,
	}

	// Send email
	err := w.sender.Send(ctx, msg)
	if err != nil {
		logger.Logger.Warn("Failed to send email",
			"id", item.ID,
			"template", item.Template,
			"error", err.Error(),
			"retry_count", item.RetryCount)

		// Determine if we should retry
		shouldRetry := item.RetryCount < item.MaxRetries && isRetryableError(err)

		// Mark as failed (with or without retry)
		if markErr := w.queueRepo.MarkAsFailed(ctx, item.ID, err, shouldRetry); markErr != nil {
			logger.Logger.Error("Failed to mark email as failed",
				"id", item.ID,
				"error", markErr.Error())
		}

		// Publish reminder.failed event
		if w.publisher != nil {
			payload := map[string]interface{}{
				"template": item.Template,
				"to":       item.ToAddresses,
			}
			if item.ReferenceType != nil && item.ReferenceID != nil && *item.ReferenceType == "signature_reminder" {
				payload["doc_id"] = *item.ReferenceID
			}
			_ = w.publisher.Publish(ctx, "reminder.failed", payload)
		}
		return
	}

	// Mark as sent
	if err := w.queueRepo.MarkAsSent(ctx, item.ID); err != nil {
		logger.Logger.Error("Failed to mark email as sent",
			"id", item.ID,
			"error", err.Error())
		// Email was sent but we failed to update the database
		// This is not critical, the email won't be resent
	}

	logger.Logger.Info("Email sent successfully",
		"id", item.ID,
		"template", item.Template,
		"to", item.ToAddresses)

	// Publish reminder.sent event
	if w.publisher != nil {
		payload := map[string]interface{}{
			"template": item.Template,
			"to":       item.ToAddresses,
		}
		if item.ReferenceType != nil && item.ReferenceID != nil && *item.ReferenceType == "signature_reminder" {
			payload["doc_id"] = *item.ReferenceID
		}
		_ = w.publisher.Publish(ctx, "reminder.sent", payload)
	}
}

// cleanupLoop periodically cleans up old emails
func (w *Worker) cleanupLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.performCleanup()
		}
	}
}

// performCleanup removes old processed emails
func (w *Worker) performCleanup() {
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()

	deleted, err := w.queueRepo.CleanupOldEmails(ctx, w.cleanupAge)
	if err != nil {
		logger.Logger.Error("Failed to cleanup old emails", "error", err.Error())
		return
	}

	if deleted > 0 {
		logger.Logger.Info("Cleaned up old emails", "count", deleted)
	}
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	// TODO: Implement more sophisticated error detection
	// For now, retry all errors except explicit data/template errors
	errStr := err.Error()

	// Don't retry template or data errors
	if contains(errStr, "template") || contains(errStr, "unmarshal") || contains(errStr, "invalid") {
		return false
	}

	// Retry network and timeout errors
	if contains(errStr, "timeout") || contains(errStr, "connection") || contains(errStr, "refused") {
		return true
	}

	// Default to retry
	return true
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
