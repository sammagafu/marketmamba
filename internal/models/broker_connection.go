package models

import "time"

type BrokerConnection struct {
	ID             string    `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	Provider       string    `db:"provider" json:"provider"`
	Label          string    `db:"label" json:"label"`
	CredentialsEnc string    `db:"credentials_enc" json:"-"`
	IsActive       bool      `db:"is_active" json:"is_active"`
	IsPrimary      bool      `db:"is_primary" json:"is_primary"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
