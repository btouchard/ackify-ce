// SPDX-License-Identifier: AGPL-3.0-or-later
package workers

import (
	"context"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// MagicLinkCleanupWorker nettoie périodiquement les tokens expirés
type MagicLinkCleanupWorker struct {
	service  *services.MagicLinkService
	interval time.Duration
	stopChan chan struct{}
}

func NewMagicLinkCleanupWorker(service *services.MagicLinkService, interval time.Duration) *MagicLinkCleanupWorker {
	if interval == 0 {
		interval = 1 * time.Hour // Défaut: toutes les heures
	}

	return &MagicLinkCleanupWorker{
		service:  service,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (w *MagicLinkCleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	logger.Logger.Info("Magic Link cleanup worker started", "interval", w.interval)

	for {
		select {
		case <-ticker.C:
			w.cleanup(ctx)
		case <-w.stopChan:
			logger.Logger.Info("Magic Link cleanup worker stopped")
			return
		case <-ctx.Done():
			logger.Logger.Info("Magic Link cleanup worker context cancelled")
			return
		}
	}
}

func (w *MagicLinkCleanupWorker) Stop() {
	close(w.stopChan)
}

func (w *MagicLinkCleanupWorker) cleanup(ctx context.Context) {
	deleted, err := w.service.CleanupExpiredTokens(ctx)
	if err != nil {
		logger.Logger.Error("Failed to cleanup expired magic link tokens", "error", err)
		return
	}

	if deleted > 0 {
		logger.Logger.Info("Cleaned up expired magic link tokens", "count", deleted)
	}
}
