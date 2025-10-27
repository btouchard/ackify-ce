// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/lib/pq"
)

type WebhookRepository struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(ctx context.Context, input models.WebhookInput) (*models.Webhook, error) {
	headersIn := []byte("{}")
	if input.Headers != nil {
		if data, err := json.Marshal(input.Headers); err == nil {
			headersIn = data
		}
	}

	query := `
        INSERT INTO webhooks (title, target_url, secret, active, events, headers, description, created_by)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        RETURNING id, title, target_url, secret, active, events, headers, description, created_by, created_at, updated_at, last_delivered_at, failure_count
    `
	wh := &models.Webhook{}
	var headersOut models.NullRawMessage
	err := r.db.QueryRowContext(ctx, query,
		input.Title,
		input.TargetURL,
		input.Secret,
		input.Active,
		pq.Array(input.Events),
		headersIn,
		input.Description,
		input.CreatedBy,
	).Scan(
		&wh.ID, &wh.Title, &wh.TargetURL, &wh.Secret, &wh.Active, pq.Array(&wh.Events), &headersOut, &wh.Description, &wh.CreatedBy,
		&wh.CreatedAt, &wh.UpdatedAt, &wh.LastDeliveredAt, &wh.FailureCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	if headersOut.Valid && len(headersOut.RawMessage) > 0 {
		_ = json.Unmarshal(headersOut.RawMessage, &wh.Headers)
	}
	return wh, nil
}

func (r *WebhookRepository) Update(ctx context.Context, id int64, input models.WebhookInput) (*models.Webhook, error) {
	headersJSON := []byte("{}")
	if input.Headers != nil {
		if data, err := json.Marshal(input.Headers); err == nil {
			headersJSON = data
		}
	}

	query := `
        UPDATE webhooks
        SET title=$1, target_url=$2, secret=COALESCE(NULLIF($3,''), secret), active=$4, events=$5, headers=$6, description=$7, updated_at=now()
        WHERE id=$8
        RETURNING id, title, target_url, secret, active, events, headers, description, created_by, created_at, updated_at, last_delivered_at, failure_count
    `
	wh := &models.Webhook{}
	var headersOut models.NullRawMessage
	err := r.db.QueryRowContext(ctx, query,
		input.Title,
		input.TargetURL,
		input.Secret,
		input.Active,
		pq.Array(input.Events),
		headersJSON,
		input.Description,
		id,
	).Scan(
		&wh.ID, &wh.Title, &wh.TargetURL, &wh.Secret, &wh.Active, pq.Array(&wh.Events), &headersOut, &wh.Description, &wh.CreatedBy,
		&wh.CreatedAt, &wh.UpdatedAt, &wh.LastDeliveredAt, &wh.FailureCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	if headersOut.Valid && len(headersOut.RawMessage) > 0 {
		_ = json.Unmarshal(headersOut.RawMessage, &wh.Headers)
	}
	return wh, nil
}

func (r *WebhookRepository) SetActive(ctx context.Context, id int64, active bool) error {
	res, err := r.db.ExecContext(ctx, `UPDATE webhooks SET active=$1, updated_at=now() WHERE id=$2`, active, id)
	if err != nil {
		return fmt.Errorf("failed to set active: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *WebhookRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM webhooks WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}

func (r *WebhookRepository) GetByID(ctx context.Context, id int64) (*models.Webhook, error) {
	query := `
        SELECT id, title, target_url, secret, active, events, headers, description, created_by, created_at, updated_at, last_delivered_at, failure_count
        FROM webhooks
        WHERE id=$1
    `
	wh := &models.Webhook{}
	var events []string
	var headersJSON models.NullRawMessage
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wh.ID, &wh.Title, &wh.TargetURL, &wh.Secret, &wh.Active, pq.Array(&events), &headersJSON, &wh.Description, &wh.CreatedBy,
		&wh.CreatedAt, &wh.UpdatedAt, &wh.LastDeliveredAt, &wh.FailureCount,
	)
	if err != nil {
		return nil, err
	}
	wh.Events = events
	if headersJSON.Valid && len(headersJSON.RawMessage) > 0 {
		_ = json.Unmarshal(headersJSON.RawMessage, &wh.Headers)
	}
	return wh, nil
}

func (r *WebhookRepository) List(ctx context.Context, limit, offset int) ([]*models.Webhook, error) {
	query := `
        SELECT id, title, target_url, secret, active, events, headers, description, created_by, created_at, updated_at, last_delivered_at, failure_count
        FROM webhooks
        ORDER BY id DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var out []*models.Webhook
	for rows.Next() {
		wh := &models.Webhook{}
		var events []string
		var headersJSON models.NullRawMessage
		if err := rows.Scan(
			&wh.ID, &wh.Title, &wh.TargetURL, &wh.Secret, &wh.Active, pq.Array(&events), &headersJSON, &wh.Description, &wh.CreatedBy,
			&wh.CreatedAt, &wh.UpdatedAt, &wh.LastDeliveredAt, &wh.FailureCount,
		); err != nil {
			return nil, err
		}
		wh.Events = events
		if headersJSON.Valid && len(headersJSON.RawMessage) > 0 {
			_ = json.Unmarshal(headersJSON.RawMessage, &wh.Headers)
		}
		out = append(out, wh)
	}
	return out, nil
}

// ListActiveByEvent returns active webhooks subscribed to a given event type
func (r *WebhookRepository) ListActiveByEvent(ctx context.Context, event string) ([]*models.Webhook, error) {
	// Use ANY($1) against events[]
	query := `
        SELECT id, title, target_url, secret, active, events, headers, description, created_by, created_at, updated_at, last_delivered_at, failure_count
        FROM webhooks
        WHERE active = TRUE AND $1 = ANY(events)
    `
	rows, err := r.db.QueryContext(ctx, query, event)
	if err != nil {
		return nil, fmt.Errorf("failed to list active webhooks: %w", err)
	}
	defer rows.Close()

	var res []*models.Webhook
	for rows.Next() {
		wh := &models.Webhook{}
		var events []string
		var headersJSON models.NullRawMessage
		if err := rows.Scan(
			&wh.ID, &wh.Title, &wh.TargetURL, &wh.Secret, &wh.Active, pq.Array(&events), &headersJSON, &wh.Description, &wh.CreatedBy,
			&wh.CreatedAt, &wh.UpdatedAt, &wh.LastDeliveredAt, &wh.FailureCount,
		); err != nil {
			return nil, err
		}
		wh.Events = events
		if headersJSON.Valid && len(headersJSON.RawMessage) > 0 {
			_ = json.Unmarshal(headersJSON.RawMessage, &wh.Headers)
		}
		res = append(res, wh)
	}
	return res, nil
}
