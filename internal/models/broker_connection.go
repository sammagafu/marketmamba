package models

import "time"

type BrokerConnection struct {
	ID               string    `db:"id"`
	UserID           int64     `db:"user_id"`
	Provider         string    `db:"provider"`
	Label            string    `db:"label"`
	CredentialsEnc   string    `db:"credentials_enc"`
	IsActive         bool      `db:"is_active"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}
