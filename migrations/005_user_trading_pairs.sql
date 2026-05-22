-- Per-user pair preferences (user_id = telegram_id)

CREATE TABLE IF NOT EXISTS user_trading_pairs (
    user_id BIGINT NOT NULL,
    symbol VARCHAR(16) NOT NULL,
    receive_signals BOOLEAN NOT NULL DEFAULT TRUE,
    auto_trade BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, symbol)
);

CREATE INDEX IF NOT EXISTS idx_user_trading_pairs_user ON user_trading_pairs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_trading_pairs_symbol_signals
    ON user_trading_pairs(symbol) WHERE receive_signals = TRUE;
