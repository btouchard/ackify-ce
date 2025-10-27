-- SPDX-License-Identifier: AGPL-3.0-or-later

DROP TRIGGER IF EXISTS trigger_update_webhook_retry ON webhook_deliveries;
DROP FUNCTION IF EXISTS update_webhook_retry_time();

DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;

