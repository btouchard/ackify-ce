// SPDX-License-Identifier: AGPL-3.0-or-later
package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/presentation/api/shared"
	"github.com/go-chi/chi/v5"
)

// Minimal interfaces for repositories used by this handler
type webhookRepository interface {
	Create(ctx context.Context, input models.WebhookInput) (*models.Webhook, error)
	Update(ctx context.Context, id int64, input models.WebhookInput) (*models.Webhook, error)
	SetActive(ctx context.Context, id int64, active bool) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*models.Webhook, error)
	List(ctx context.Context, limit, offset int) ([]*models.Webhook, error)
}

type webhookDeliveryRepository interface {
	ListByWebhook(ctx context.Context, webhookID int64, limit, offset int) ([]*models.WebhookDelivery, error)
}

// WebhooksHandler groups operations on webhooks
type WebhooksHandler struct {
	repo       webhookRepository
	deliveries webhookDeliveryRepository
}

func NewWebhooksHandler(repo webhookRepository, deliveries webhookDeliveryRepository) *WebhooksHandler {
	return &WebhooksHandler{repo: repo, deliveries: deliveries}
}

type CreateWebhookRequest struct {
	Title       string            `json:"title"`
	TargetURL   string            `json:"targetUrl"`
	Secret      string            `json:"secret"`
	Active      bool              `json:"active"`
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers,omitempty"`
	Description string            `json:"description,omitempty"`
}

func (h *WebhooksHandler) HandleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", nil)
		return
	}
	if req.Title == "" || req.TargetURL == "" || req.Secret == "" || len(req.Events) == 0 {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "title, targetUrl, secret and events are required", nil)
		return
	}
	user, _ := shared.GetUserFromContext(ctx)
	input := models.WebhookInput{Title: req.Title, TargetURL: req.TargetURL, Secret: req.Secret, Active: req.Active, Events: req.Events, Headers: req.Headers, Description: req.Description}
	if user != nil {
		input.CreatedBy = user.Email
	}
	wh, err := h.repo.Create(ctx, input)
	if err != nil {
		shared.WriteInternalError(w)
		return
	}
	shared.WriteJSON(w, http.StatusCreated, wh)
}

func (h *WebhooksHandler) HandleListWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	limit := 100
	offset := 0
	list, err := h.repo.List(ctx, limit, offset)
	if err != nil {
		shared.WriteInternalError(w)
		return
	}
	meta := map[string]interface{}{"total": len(list), "limit": limit, "offset": offset}
	shared.WriteJSONWithMeta(w, http.StatusOK, list, meta)
}

func (h *WebhooksHandler) HandleGetWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	wh, err := h.repo.GetByID(ctx, id)
	if err != nil {
		shared.WriteError(w, http.StatusNotFound, shared.ErrCodeNotFound, "Webhook not found", nil)
		return
	}
	shared.WriteJSON(w, http.StatusOK, wh)
}

func (h *WebhooksHandler) HandleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "Invalid request body", nil)
		return
	}
	// For updates, allow empty secret to keep current value
	if req.Title == "" || req.TargetURL == "" || len(req.Events) == 0 {
		shared.WriteError(w, http.StatusBadRequest, shared.ErrCodeBadRequest, "title, targetUrl and events are required", nil)
		return
	}
	input := models.WebhookInput{Title: req.Title, TargetURL: req.TargetURL, Secret: req.Secret, Active: req.Active, Events: req.Events, Headers: req.Headers, Description: req.Description}
	wh, err := h.repo.Update(ctx, id, input)
	if err != nil {
		shared.WriteInternalError(w)
		return
	}
	shared.WriteJSON(w, http.StatusOK, wh)
}

func (h *WebhooksHandler) HandleToggleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	enable := chi.URLParam(r, "action") == "enable"
	if err := h.repo.SetActive(ctx, id, enable); err != nil {
		shared.WriteInternalError(w)
		return
	}
	status := "disabled"
	if enable {
		status = "enabled"
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "Webhook " + status})
}

func (h *WebhooksHandler) HandleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		shared.WriteInternalError(w)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"message": "Webhook deleted"})
}

func (h *WebhooksHandler) HandleListDeliveries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	deliveries, err := h.deliveries.ListByWebhook(ctx, id, 100, 0)
	if err != nil {
		shared.WriteInternalError(w)
		return
	}
	shared.WriteJSON(w, http.StatusOK, deliveries)
}
