-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Rollback to original exponential backoff without jitter

CREATE OR REPLACE FUNCTION calculate_next_retry_time(retry_count INT)
RETURNS TIMESTAMPTZ AS $$
BEGIN
    -- Exponential backoff: 1min, 2min, 4min, 8min, 16min, 32min...
    RETURN now() + (interval '1 minute' * power(2, retry_count));
END;
$$ LANGUAGE plpgsql;