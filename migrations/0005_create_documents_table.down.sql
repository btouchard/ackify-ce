-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_documents_updated_at ON documents;
DROP FUNCTION IF EXISTS update_documents_updated_at();

-- Drop table
DROP TABLE IF EXISTS documents;
