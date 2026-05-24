-- Subscription tier usage (calendar month UTC)
CREATE TABLE IF NOT EXISTS user_plan_usage (
    user_id BIGINT NOT NULL,
    period_start DATE NOT NULL,
    signals_received INT NOT NULL DEFAULT 0,
    long_trades INT NOT NULL DEFAULT 0,
    short_trades INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, period_start)
);

CREATE INDEX IF NOT EXISTS idx_user_plan_usage_user_id ON user_plan_usage(user_id);
