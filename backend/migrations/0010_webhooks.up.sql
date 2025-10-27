-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Webhooks configuration table
CREATE TABLE webhooks (
    id BIGSERIAL PRIMARY KEY,
    target_url TEXT NOT NULL,
    secret TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    events TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    headers JSONB,
    description TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_delivered_at TIMESTAMPTZ,
    failure_count INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE webhooks IS 'Third-party webhook subscriptions';
COMMENT ON COLUMN webhooks.events IS 'Array of event types the webhook subscribes to';

CREATE INDEX idx_webhooks_active ON webhooks(active);
CREATE INDEX idx_webhooks_events_gin ON webhooks USING GIN (events);

-- Webhook deliveries/queue table
CREATE TABLE webhook_deliveries (
    id BIGSERIAL PRIMARY KEY,
    webhook_id BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',

    -- Queue management
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','delivered','failed','cancelled')),
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 6,
    priority INT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,

    -- Request/response metadata (for observability, truncated by code)
    request_headers JSONB,
    response_status INT,
    response_headers JSONB,
    response_body TEXT,
    last_error TEXT
);

CREATE INDEX idx_webhook_deliveries_status_scheduled
    ON webhook_deliveries(status, scheduled_for)
    WHERE status IN ('pending','processing');

CREATE INDEX idx_webhook_deliveries_priority_scheduled
    ON webhook_deliveries(priority DESC, scheduled_for ASC)
    WHERE status = 'pending';

CREATE INDEX idx_webhook_deliveries_retry
    ON webhook_deliveries(next_retry_at)
    WHERE status = 'processing' AND retry_count < max_retries;

CREATE INDEX idx_webhook_deliveries_webhook_id
    ON webhook_deliveries(webhook_id);

CREATE INDEX idx_webhook_deliveries_event_type
    ON webhook_deliveries(event_type);

-- Trigger to auto-update next_retry_at on status change to processing (reuse existing function)
CREATE OR REPLACE FUNCTION update_webhook_retry_time()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'processing' AND OLD.status != 'processing' THEN
        NEW.next_retry_at = calculate_next_retry_time(NEW.retry_count);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_webhook_retry
    BEFORE UPDATE ON webhook_deliveries
    FOR EACH ROW
    EXECUTE FUNCTION update_webhook_retry_time();

