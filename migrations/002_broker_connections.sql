CREATE TABLE IF NOT EXISTS broker_connections (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR(50) NOT NULL,
    label VARCHAR(100) NOT NULL DEFAULT '',
    credentials_enc TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_broker_connections_user_active
    ON broker_connections (user_id) WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_broker_connections_user_id ON broker_connections(user_id);
