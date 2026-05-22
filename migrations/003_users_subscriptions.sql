CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(50) PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255) NOT NULL DEFAULT '',
    first_name VARCHAR(255) NOT NULL DEFAULT '',
    last_name VARCHAR(255) NOT NULL DEFAULT '',
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    last_seen_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'trial',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP,
    notes TEXT NOT NULL DEFAULT '',
    activated_by VARCHAR(50) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

CREATE TABLE IF NOT EXISTS payment_records (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    amount DECIMAL(15, 2),
    currency VARCHAR(10) NOT NULL DEFAULT 'TZS',
    method VARCHAR(50) NOT NULL DEFAULT 'manual',
    reference VARCHAR(255) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_by_admin BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
