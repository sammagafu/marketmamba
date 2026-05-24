-- Per-user signal asset classes: forex, indexes (synthetics/indices), crypto

CREATE TABLE IF NOT EXISTS user_signal_preferences (
    user_id BIGINT PRIMARY KEY,
    forex BOOLEAN NOT NULL DEFAULT TRUE,
    indexes BOOLEAN NOT NULL DEFAULT TRUE,
    crypto BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
