package models

import "time"

type PaymentOrder struct {
	ID              string     `db:"id" json:"id"`
	UserID          int64      `db:"user_id" json:"user_id"`
	AmountUSDT      float64    `db:"amount_usdt" json:"amount_usdt"`
	Plan            string     `db:"plan" json:"plan"`
	Status          string     `db:"status" json:"status"`
	MerchantTradeNo string     `db:"merchant_trade_no" json:"merchant_trade_no"`
	BinancePrepayID string     `db:"binance_prepay_id" json:"binance_prepay_id,omitempty"`
	CheckoutURL     string     `db:"checkout_url" json:"checkout_url,omitempty"`
	PayMethod       string     `db:"pay_method" json:"pay_method"`
	TxReference     string     `db:"tx_reference" json:"tx_reference,omitempty"`
	ExpiresAt       time.Time  `db:"expires_at" json:"expires_at"`
	PaidAt          *time.Time `db:"paid_at" json:"paid_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}
