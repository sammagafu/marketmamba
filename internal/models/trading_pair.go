package models

import "time"

// UserTradingPair is one symbol preference for a trader (user_id = telegram_id).
type UserTradingPair struct {
	UserID          int64     `db:"user_id" json:"user_id"`
	Symbol          string    `db:"symbol" json:"symbol"`
	ReceiveSignals  bool      `db:"receive_signals" json:"receive_signals"`
	AutoTrade       bool      `db:"auto_trade" json:"auto_trade"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// TradingPairsResponse is returned by GET /api/v1/trading-pairs.
type TradingPairsResponse struct {
	AvailableSymbols []string              `json:"available_symbols"`
	Pairs            []UserTradingPair     `json:"pairs"`
	Customized       bool                  `json:"customized"`
	SignalSymbols    []string              `json:"signal_symbols"`
	AutoTradeSymbols []string              `json:"auto_trade_symbols"`
	SignalTypes      SignalTypePreferences `json:"signal_types"`
	AssetGroups      []SignalAssetGroup    `json:"asset_groups"`
}
