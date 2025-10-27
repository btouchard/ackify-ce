-- SPDX-License-Identifier: AGPL-3.0-or-later

ALTER TABLE webhooks
    ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';

-- Backfill could be added here if needed (e.g., copy description into title when empty)

