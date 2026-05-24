package storage

import (
	"database/sql"
	"time"
)

func periodStartUTC(t time.Time) time.Time {
	y, m, _ := t.UTC().Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

// GetOrCreateUsage returns usage counters for the current calendar month.
func (ps *PostgresStorage) GetOrCreateUsage(userID int64, now time.Time) (signals, longTrades, shortTrades int, period time.Time, err error) {
	period = periodStartUTC(now)
	row := ps.db.QueryRow(
		`SELECT signals_received, long_trades, short_trades FROM user_plan_usage
		 WHERE user_id = $1 AND period_start = $2`,
		userID, period,
	)
	err = row.Scan(&signals, &longTrades, &shortTrades)
	if err == sql.ErrNoRows {
		_, err = ps.db.Exec(
			`INSERT INTO user_plan_usage (user_id, period_start, signals_received, long_trades, short_trades, updated_at)
			 VALUES ($1, $2, 0, 0, 0, $3) ON CONFLICT DO NOTHING`,
			userID, period, now,
		)
		if err != nil {
			return 0, 0, 0, period, err
		}
		return 0, 0, 0, period, nil
	}
	return signals, longTrades, shortTrades, period, err
}

func (ps *PostgresStorage) IncrementSignalUsage(userID int64, now time.Time) error {
	period := periodStartUTC(now)
	_, err := ps.db.Exec(
		`INSERT INTO user_plan_usage (user_id, period_start, signals_received, long_trades, short_trades, updated_at)
		 VALUES ($1, $2, 1, 0, 0, $3)
		 ON CONFLICT (user_id, period_start) DO UPDATE SET
		   signals_received = user_plan_usage.signals_received + 1,
		   updated_at = EXCLUDED.updated_at`,
		userID, period, now,
	)
	return err
}

func (ps *PostgresStorage) IncrementTradeUsage(userID int64, tradeType string, now time.Time) error {
	period := periodStartUTC(now)
	isLong := tradeType == "BUY"
	_, err := ps.db.Exec(
		`INSERT INTO user_plan_usage (user_id, period_start, signals_received, long_trades, short_trades, updated_at)
		 VALUES ($1, $2, 0, $3, $4, $5)
		 ON CONFLICT (user_id, period_start) DO UPDATE SET
		   long_trades = user_plan_usage.long_trades + $3,
		   short_trades = user_plan_usage.short_trades + $4,
		   updated_at = EXCLUDED.updated_at`,
		userID, period, boolToInt(isLong), boolToInt(!isLong), now,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (ps *PostgresStorage) CountActiveBrokerConnections(userID int64) (int, error) {
	var n int
	err := ps.db.QueryRow(
		`SELECT COUNT(*) FROM broker_connections WHERE user_id = $1 AND is_active = TRUE`,
		userID,
	).Scan(&n)
	return n, err
}
