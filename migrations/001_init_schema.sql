-- Create trades table
CREATE TABLE IF NOT EXISTS trades (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    type VARCHAR(10) NOT NULL,
    entry_price DECIMAL(15, 5) NOT NULL,
    quantity DECIMAL(15, 2) NOT NULL,
    stop_loss DECIMAL(15, 5) NOT NULL,
    take_profit DECIMAL(15, 5) NOT NULL,
    risk_amount DECIMAL(15, 2),
    reward_amount DECIMAL(15, 2),
    risk_reward_ratio DECIMAL(10, 2),
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    exit_price DECIMAL(15, 5),
    exit_time TIMESTAMP,
    profit DECIMAL(15, 2),
    closure_reason VARCHAR(50),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create positions table
CREATE TABLE IF NOT EXISTS positions (
    id VARCHAR(50) PRIMARY KEY,
    trade_id VARCHAR(50),
    broker_id VARCHAR(100),
    user_id BIGINT NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    type VARCHAR(10) NOT NULL,
    quantity DECIMAL(15, 2) NOT NULL,
    entry_price DECIMAL(15, 5) NOT NULL,
    current_price DECIMAL(15, 5),
    stop_loss DECIMAL(15, 5) NOT NULL,
    take_profit DECIMAL(15, 5) NOT NULL,
    profit DECIMAL(15, 2),
    profit_pct DECIMAL(10, 2),
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (trade_id) REFERENCES trades(id)
);

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    broker_provider VARCHAR(50) NOT NULL,
    balance DECIMAL(15, 2) NOT NULL,
    equity DECIMAL(15, 2) NOT NULL,
    used_margin DECIMAL(15, 2),
    free_margin DECIMAL(15, 2),
    leverage INT DEFAULT 1,
    last_synced_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL
);

-- Create risk_settings table
CREATE TABLE IF NOT EXISTS risk_settings (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    max_risk_per_trade DECIMAL(5, 4) NOT NULL,
    max_daily_loss DECIMAL(5, 4) NOT NULL,
    max_open_trades INT NOT NULL,
    max_trades_per_day INT NOT NULL,
    risk_reward_ratio DECIMAL(10, 2) DEFAULT 1.0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create daily_stats table
CREATE TABLE IF NOT EXISTS daily_stats (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    trading_date DATE NOT NULL,
    trade_count INT DEFAULT 0,
    win_count INT DEFAULT 0,
    loss_count INT DEFAULT 0,
    total_profit DECIMAL(15, 2) DEFAULT 0,
    total_loss DECIMAL(15, 2) DEFAULT 0,
    net_profit DECIMAL(15, 2) DEFAULT 0,
    win_rate DECIMAL(5, 2) DEFAULT 0,
    max_drawdown DECIMAL(5, 2) DEFAULT 0,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(user_id, trading_date)
);

-- Create bot_states table
CREATE TABLE IF NOT EXISTS bot_states (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    is_paused BOOLEAN DEFAULT FALSE,
    auto_trading_active BOOLEAN DEFAULT FALSE,
    daily_loss_hit BOOLEAN DEFAULT FALSE,
    last_active_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL
);

-- Create command_logs table
CREATE TABLE IF NOT EXISTS command_logs (
    id VARCHAR(50) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    command VARCHAR(100) NOT NULL,
    args TEXT,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    created_at TIMESTAMP NOT NULL
);

-- Create indexes for better query performance
CREATE INDEX idx_trades_user_id ON trades(user_id);
CREATE INDEX idx_trades_status ON trades(status);
CREATE INDEX idx_positions_user_id ON positions(user_id);
CREATE INDEX idx_positions_symbol ON positions(symbol);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_risk_settings_user_id ON risk_settings(user_id);
CREATE INDEX idx_daily_stats_user_id_date ON daily_stats(user_id, trading_date);
CREATE INDEX idx_bot_states_user_id ON bot_states(user_id);
CREATE INDEX idx_command_logs_user_id ON command_logs(user_id);
