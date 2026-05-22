CREATE TABLE IF NOT EXISTS payment_orders (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    amount_usdt DECIMAL(12, 4) NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'monthly',
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    merchant_trade_no VARCHAR(64) NOT NULL UNIQUE,
    binance_prepay_id VARCHAR(128) NOT NULL DEFAULT '',
    checkout_url TEXT NOT NULL DEFAULT '',
    pay_method VARCHAR(30) NOT NULL DEFAULT 'binance_pay',
    tx_reference VARCHAR(255) NOT NULL DEFAULT '',
    expires_at TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_merchant_trade_no ON payment_orders(merchant_trade_no);
