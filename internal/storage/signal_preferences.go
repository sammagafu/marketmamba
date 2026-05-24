package storage

import (
	"database/sql"
	"time"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) GetUserSignalPreferences(userID int64) (models.SignalTypePreferences, bool, error) {
	var row models.UserSignalPreferencesRow
	err := ps.db.QueryRow(
		`SELECT user_id, forex, indexes, crypto, updated_at FROM user_signal_preferences WHERE user_id = $1`,
		userID,
	).Scan(&row.UserID, &row.Forex, &row.Indexes, &row.Crypto, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.DefaultSignalTypes(), false, nil
	}
	if err != nil {
		return models.SignalTypePreferences{}, false, err
	}
	return models.SignalTypePreferences{
		Forex: row.Forex, Indexes: row.Indexes, Crypto: row.Crypto,
	}, true, nil
}

func (ps *PostgresStorage) UpsertUserSignalPreferences(userID int64, prefs models.SignalTypePreferences) error {
	now := time.Now()
	_, err := ps.db.Exec(
		`INSERT INTO user_signal_preferences (user_id, forex, indexes, crypto, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id) DO UPDATE SET
		   forex = EXCLUDED.forex,
		   indexes = EXCLUDED.indexes,
		   crypto = EXCLUDED.crypto,
		   updated_at = EXCLUDED.updated_at`,
		userID, prefs.Forex, prefs.Indexes, prefs.Crypto, now,
	)
	return err
}
