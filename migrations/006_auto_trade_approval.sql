-- Admin approval gate for /autostart (when AUTO_TRADE_REQUIRES_APPROVAL=true)
ALTER TABLE bot_states
    ADD COLUMN IF NOT EXISTS auto_trade_approved BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_bot_states_auto_approved
    ON bot_states (auto_trade_approved)
    WHERE auto_trade_approved = TRUE;
