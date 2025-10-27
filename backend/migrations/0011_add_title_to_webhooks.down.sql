-- SPDX-License-Identifier: AGPL-3.0-or-later

ALTER TABLE webhooks
    DROP COLUMN IF EXISTS title;

