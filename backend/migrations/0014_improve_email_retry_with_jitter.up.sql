-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Improve email retry logic with jitter to prevent thundering herd
-- Replace existing calculate_next_retry_time function with improved version

CREATE OR REPLACE FUNCTION calculate_next_retry_time(retry_count INT)
RETURNS TIMESTAMPTZ AS $$
BEGIN
    -- Exponential backoff with jitter (0-30% random variation)
    -- This prevents multiple failed emails from retrying at exactly the same time
    -- Base delay: 1min, 2min, 4min, 8min, 16min, 32min...
    -- With jitter: adds 0-30% random variation to spread out retry attempts
    RETURN now() + (
        interval '1 minute' * power(2, retry_count) *
        (1.0 + random() * 0.3)
    );
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION calculate_next_retry_time(INT) IS 'Calculates next retry time with exponential backoff and 0-30% jitter to prevent thundering herd';