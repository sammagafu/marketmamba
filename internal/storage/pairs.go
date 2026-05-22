package storage

import (
	"database/sql"
	"time"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) HasUserTradingPairs(userID int64) (bool, error) {
	var n int
	err := ps.db.QueryRow(
		`SELECT COUNT(*) FROM user_trading_pairs WHERE user_id = $1`, userID,
	).Scan(&n)
	return n > 0, err
}

func (ps *PostgresStorage) ListUserTradingPairs(userID int64) ([]models.UserTradingPair, error) {
	rows, err := ps.db.Query(
		`SELECT user_id, symbol, receive_signals, auto_trade, created_at, updated_at
		 FROM user_trading_pairs WHERE user_id = $1 ORDER BY symbol`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.UserTradingPair
	for rows.Next() {
		var p models.UserTradingPair
		if err := rows.Scan(&p.UserID, &p.Symbol, &p.ReceiveSignals, &p.AutoTrade, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (ps *PostgresStorage) ReplaceUserTradingPairs(userID int64, pairs []models.UserTradingPair) error {
	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM user_trading_pairs WHERE user_id = $1`, userID); err != nil {
		return err
	}
	now := time.Now()
	for _, p := range pairs {
		_, err := tx.Exec(
			`INSERT INTO user_trading_pairs (user_id, symbol, receive_signals, auto_trade, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			userID, p.Symbol, p.ReceiveSignals, p.AutoTrade, now, now,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListSignalSubscriberTelegramIDsForSymbol returns users who want signals for this pair.
func (ps *PostgresStorage) ListSignalSubscriberTelegramIDsForSymbol(symbol string) ([]int64, error) {
	rows, err := ps.db.Query(
		`SELECT u.telegram_id
		 FROM users u
		 WHERE u.is_blocked = FALSE
		   AND (
		     NOT EXISTS (SELECT 1 FROM user_trading_pairs p WHERE p.user_id = u.telegram_id)
		     OR EXISTS (
		       SELECT 1 FROM user_trading_pairs p
		       WHERE p.user_id = u.telegram_id AND p.symbol = $1 AND p.receive_signals = TRUE
		     )
		   )
		 ORDER BY u.telegram_id`,
		symbol,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTelegramIDs(rows)
}

func (ps *PostgresStorage) UserReceivesSignalForSymbol(userID int64, symbol string) (bool, error) {
	var receives bool
	err := ps.db.QueryRow(
		`SELECT COALESCE(
		   (SELECT p.receive_signals FROM user_trading_pairs p
		    WHERE p.user_id = $1 AND p.symbol = $2 LIMIT 1),
		   TRUE
		 )`,
		userID, symbol,
	).Scan(&receives)
	if err == sql.ErrNoRows {
		return true, nil
	}
	return receives, err
}

func (ps *PostgresStorage) UserAutoTradesSymbol(userID int64, symbol string) (bool, error) {
	var hasPrefs bool
	if err := ps.db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM user_trading_pairs WHERE user_id = $1)`, userID,
	).Scan(&hasPrefs); err != nil {
		return false, err
	}
	if !hasPrefs {
		return true, nil
	}
	var auto bool
	err := ps.db.QueryRow(
		`SELECT auto_trade FROM user_trading_pairs WHERE user_id = $1 AND symbol = $2`,
		userID, symbol,
	).Scan(&auto)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return auto, err
}

func scanTelegramIDs(rows *sql.Rows) ([]int64, error) {
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
