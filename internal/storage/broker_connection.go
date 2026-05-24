package storage

import (
	"database/sql"
	"time"

	"forex-bot/internal/models"
)

// AddBrokerConnection inserts a new active connection and marks it primary.
func (ps *PostgresStorage) AddBrokerConnection(conn *models.BrokerConnection) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE broker_connections SET is_primary = FALSE, updated_at = $1 WHERE user_id = $2`,
		time.Now(), conn.UserID,
	); err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO broker_connections (
			id, user_id, provider, label, credentials_enc, is_active, is_primary, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, TRUE, TRUE, $6, $7)`,
		conn.ID, conn.UserID, conn.Provider, conn.Label, conn.CredentialsEnc, conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// UpsertBrokerConnection adds a new primary connection (legacy name for callers).
func (ps *PostgresStorage) UpsertBrokerConnection(conn *models.BrokerConnection) error {
	conn.IsActive = true
	conn.IsPrimary = true
	return ps.AddBrokerConnection(conn)
}

func (ps *PostgresStorage) GetPrimaryBrokerConnection(userID int64) (*models.BrokerConnection, error) {
	row := ps.db.QueryRow(
		`SELECT id, user_id, provider, label, credentials_enc, is_active, is_primary, created_at, updated_at
		 FROM broker_connections WHERE user_id = $1 AND is_primary = TRUE LIMIT 1`,
		userID,
	)
	c, err := scanBrokerConnection(row)
	if err != nil || c != nil {
		return c, err
	}
	row = ps.db.QueryRow(
		`SELECT id, user_id, provider, label, credentials_enc, is_active, is_primary, created_at, updated_at
		 FROM broker_connections WHERE user_id = $1 AND is_active = TRUE
		 ORDER BY updated_at DESC LIMIT 1`,
		userID,
	)
	return scanBrokerConnection(row)
}

// GetActiveBrokerConnection returns the primary trading connection.
func (ps *PostgresStorage) GetActiveBrokerConnection(userID int64) (*models.BrokerConnection, error) {
	return ps.GetPrimaryBrokerConnection(userID)
}

func (ps *PostgresStorage) ListBrokerConnections(userID int64) ([]*models.BrokerConnection, error) {
	rows, err := ps.db.Query(
		`SELECT id, user_id, provider, label, credentials_enc, is_active, is_primary, created_at, updated_at
		 FROM broker_connections WHERE user_id = $1 AND is_active = TRUE ORDER BY is_primary DESC, updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.BrokerConnection
	for rows.Next() {
		c, err := scanBrokerConnectionRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (ps *PostgresStorage) SetPrimaryBrokerConnection(userID int64, connectionID string) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now()
	if _, err := tx.Exec(
		`UPDATE broker_connections SET is_primary = FALSE, updated_at = $1 WHERE user_id = $2`,
		now, userID,
	); err != nil {
		return err
	}
	res, err := tx.Exec(
		`UPDATE broker_connections SET is_primary = TRUE, updated_at = $1
		 WHERE id = $2 AND user_id = $3 AND is_active = TRUE`,
		now, connectionID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

// CountBrokerConnectionsByProvider returns active connection counts grouped by provider.
func (ps *PostgresStorage) CountBrokerConnectionsByProvider() (map[string]int, error) {
	rows, err := ps.db.Query(
		`SELECT provider, COUNT(*) FROM broker_connections WHERE is_active = TRUE GROUP BY provider`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var provider string
		var n int
		if err := rows.Scan(&provider, &n); err != nil {
			return nil, err
		}
		out[provider] = n
	}
	return out, rows.Err()
}

func scanBrokerConnection(row *sql.Row) (*models.BrokerConnection, error) {
	return scanBrokerConnectionRow(row)
}

func scanBrokerConnectionRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.BrokerConnection, error) {
	var c models.BrokerConnection
	err := scanner.Scan(
		&c.ID, &c.UserID, &c.Provider, &c.Label, &c.CredentialsEnc,
		&c.IsActive, &c.IsPrimary, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
