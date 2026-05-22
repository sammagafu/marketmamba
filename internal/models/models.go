package models

import (
	"time"
)

// Trade represents a completed or ongoing trade
type Trade struct {
	ID              string     `db:"id" json:"id"`
	UserID          int64      `db:"user_id" json:"user_id"`
	Symbol          string     `db:"symbol" json:"symbol"`
	Type            string     `db:"type" json:"type"` // BUY or SELL
	EntryPrice      float64    `db:"entry_price" json:"entry_price"`
	Quantity        float64    `db:"quantity" json:"quantity"`
	StopLoss        float64    `db:"stop_loss" json:"stop_loss"`
	TakeProfit      float64    `db:"take_profit" json:"take_profit"`
	RiskAmount      float64    `db:"risk_amount" json:"risk_amount"`
	RewardAmount    float64    `db:"reward_amount" json:"reward_amount"`
	RiskRewardRatio float64    `db:"risk_reward_ratio" json:"risk_reward_ratio"`
	Status          string     `db:"status" json:"status"` // OPEN, CLOSED, CANCELLED
	ExitPrice       *float64   `db:"exit_price" json:"exit_price,omitempty"`
	ExitTime        *time.Time `db:"exit_time" json:"exit_time,omitempty"`
	Profit          *float64   `db:"profit" json:"profit,omitempty"`
	ClosureReason   *string    `db:"closure_reason" json:"closure_reason,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// Position represents an open trading position
type Position struct {
	ID           string    `db:"id" json:"id"`
	TradeID      string    `db:"trade_id" json:"trade_id"`
	BrokerID     string    `db:"broker_id" json:"broker_id"`
	UserID       int64     `db:"user_id" json:"user_id"`
	Symbol       string    `db:"symbol" json:"symbol"`
	Type         string    `db:"type" json:"type"`
	Quantity     float64   `db:"quantity" json:"quantity"`
	EntryPrice   float64   `db:"entry_price" json:"entry_price"`
	CurrentPrice float64   `db:"current_price" json:"current_price"`
	StopLoss     float64   `db:"stop_loss" json:"stop_loss"`
	TakeProfit   float64   `db:"take_profit" json:"take_profit"`
	Profit       float64   `db:"profit" json:"profit"`
	ProfitPct    float64   `db:"profit_pct" json:"profit_pct"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Account represents user account information
type Account struct {
	ID              string    `db:"id"`
	UserID          int64     `db:"user_id"`
	BrokerProvider  string    `db:"broker_provider"`
	Balance         float64   `db:"balance"`
	Equity          float64   `db:"equity"`
	UsedMargin      float64   `db:"used_margin"`
	FreeMargin      float64   `db:"free_margin"`
	Leverage        int       `db:"leverage"`
	LastSyncedAt    time.Time `db:"last_synced_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// RiskSettings holds risk management configuration
type RiskSettings struct {
	ID              string    `db:"id"`
	UserID          int64     `db:"user_id"`
	MaxRiskPerTrade float64   `db:"max_risk_per_trade"` // percentage
	MaxDailyLoss    float64   `db:"max_daily_loss"`     // percentage
	MaxOpenTrades   int       `db:"max_open_trades"`
	MaxTradesPerDay int       `db:"max_trades_per_day"`
	RiskRewardRatio float64   `db:"risk_reward_ratio"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// DailyStats tracks daily trading statistics
type DailyStats struct {
	ID                string    `db:"id"`
	UserID            int64     `db:"user_id"`
	TradingDate       time.Time `db:"trading_date"`
	TradeCount        int       `db:"trade_count"`
	WinCount          int       `db:"win_count"`
	LossCount         int       `db:"loss_count"`
	TotalProfit       float64   `db:"total_profit"`
	TotalLoss         float64   `db:"total_loss"`
	NetProfit         float64   `db:"net_profit"`
	WinRate           float64   `db:"win_rate"`
	MaxDrawdown       float64   `db:"max_drawdown"`
	UpdatedAt         time.Time `db:"updated_at"`
}

// BotState tracks the trading bot state
type BotState struct {
	ID                string    `db:"id"`
	UserID            int64     `db:"user_id"`
	IsPaused          bool      `db:"is_paused"`
	AutoTradingActive bool      `db:"auto_trading_active"`
	DailyLossHit      bool      `db:"daily_loss_hit"`
	LastActiveAt      time.Time `db:"last_active_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

// TradeSignal represents a trading signal
type TradeSignal struct {
	Symbol           string
	Type             string  // BUY or SELL
	Strength         float64 // 0-1
	StopLoss         float64
	TakeProfit       float64
	RiskRewardRatio  float64
	TriggeredAt      time.Time
}

// CommandLog tracks command execution for audit
type CommandLog struct {
	ID        string    `db:"id"`
	UserID    int64     `db:"user_id"`
	Command   string    `db:"command"`
	Args      string    `db:"args"`
	Status    string    `db:"status"` // SUCCESS, FAILED
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}
