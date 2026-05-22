package storage

import (
	"database/sql"
	"time"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) UpsertBrokerConnection(conn *models.BrokerConnection) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE broker_connections SET is_active = FALSE, updated_at = $1 WHERE user_id = $2 AND is_active = TRUE`,
		time.Now(), conn.UserID,
	); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO broker_connections (id, user_id, provider, label, credentials_enc, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7)`,
		conn.ID, conn.UserID, conn.Provider, conn.Label, conn.CredentialsEnc, conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (ps *PostgresStorage) GetActiveBrokerConnection(userID int64) (*models.BrokerConnection, error) {
	row := ps.db.QueryRow(
		`SELECT id, user_id, provider, label, credentials_enc, is_active, created_at, updated_at
		 FROM broker_connections WHERE user_id = $1 AND is_active = TRUE LIMIT 1`,
		userID,
	)
	return scanBrokerConnection(row)
}

func scanBrokerConnection(row *sql.Row) (*models.BrokerConnection, error) {
	var c models.BrokerConnection
	err := row.Scan(
		&c.ID, &c.UserID, &c.Provider, &c.Label, &c.CredentialsEnc,
		&c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
