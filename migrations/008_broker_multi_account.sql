-- Multiple broker accounts per user (one primary for auto-trade)
ALTER TABLE broker_connections ADD COLUMN IF NOT EXISTS is_primary BOOLEAN NOT NULL DEFAULT FALSE;

DROP INDEX IF EXISTS idx_broker_connections_user_active;

CREATE UNIQUE INDEX IF NOT EXISTS idx_broker_connections_user_primary
    ON broker_connections (user_id) WHERE is_primary = TRUE;

CREATE INDEX IF NOT EXISTS idx_broker_connections_user_active
    ON broker_connections (user_id) WHERE is_active = TRUE;
