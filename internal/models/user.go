package models

import "time"

type User struct {
	ID         string    `db:"id" json:"id"`
	TelegramID int64     `db:"telegram_id" json:"telegram_id"`
	Username   string    `db:"username" json:"username"`
	FirstName  string    `db:"first_name" json:"first_name"`
	LastName   string    `db:"last_name" json:"last_name"`
	IsBlocked  bool      `db:"is_blocked" json:"is_blocked"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	LastSeenAt time.Time `db:"last_seen_at" json:"last_seen_at"`
}

type Subscription struct {
	ID          string     `db:"id" json:"id"`
	UserID      int64      `db:"user_id" json:"user_id"`
	Plan        string     `db:"plan" json:"plan"`
	Status      string     `db:"status" json:"status"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	Notes       string     `db:"notes" json:"notes"`
	ActivatedBy string     `db:"activated_by" json:"activated_by"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type PaymentRecord struct {
	ID              string    `db:"id" json:"id"`
	UserID          int64     `db:"user_id" json:"user_id"`
	Amount          *float64  `db:"amount" json:"amount,omitempty"`
	Currency        string    `db:"currency" json:"currency"`
	Method          string    `db:"method" json:"method"`
	Reference       string    `db:"reference" json:"reference"`
	Notes           string    `db:"notes" json:"notes"`
	CreatedByAdmin  int64     `db:"created_by_admin" json:"created_by_admin"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type UserStats struct {
	TotalUsers            int `json:"total_users"`
	ActiveSubscriptions   int `json:"active_subscriptions"`
	AutoTradingUsers      int `json:"auto_trading_users"`
	NewUsersLast7Days     int `json:"new_users_last_7_days"`
}
