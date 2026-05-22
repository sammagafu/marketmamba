package storage

import (
	"database/sql"
	"time"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) UpsertUser(u *models.User) error {
	_, err := ps.db.Exec(
		`INSERT INTO users (id, telegram_id, username, first_name, last_name, is_blocked, created_at, last_seen_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (telegram_id) DO UPDATE SET
		   username = EXCLUDED.username,
		   first_name = EXCLUDED.first_name,
		   last_name = EXCLUDED.last_name,
		   last_seen_at = EXCLUDED.last_seen_at`,
		u.ID, u.TelegramID, u.Username, u.FirstName, u.LastName, u.IsBlocked, u.CreatedAt, u.LastSeenAt,
	)
	return err
}

func (ps *PostgresStorage) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	row := ps.db.QueryRow(
		`SELECT id, telegram_id, username, first_name, last_name, is_blocked, created_at, last_seen_at
		 FROM users WHERE telegram_id = $1`, telegramID,
	)
	var u models.User
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.IsBlocked, &u.CreatedAt, &u.LastSeenAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (ps *PostgresStorage) UpdateUserLastSeen(telegramID int64, t time.Time) error {
	_, err := ps.db.Exec(`UPDATE users SET last_seen_at = $1 WHERE telegram_id = $2`, t, telegramID)
	return err
}

func (ps *PostgresStorage) CreateSubscription(sub *models.Subscription) error {
	_, err := ps.db.Exec(
		`INSERT INTO subscriptions (id, user_id, plan, status, expires_at, notes, activated_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		sub.ID, sub.UserID, sub.Plan, sub.Status, sub.ExpiresAt, sub.Notes, sub.ActivatedBy, sub.CreatedAt, sub.UpdatedAt,
	)
	return err
}

func (ps *PostgresStorage) DeactivateSubscriptions(userID int64) error {
	_, err := ps.db.Exec(
		`UPDATE subscriptions SET status = 'replaced', updated_at = $1 WHERE user_id = $2 AND status = 'active'`,
		time.Now(), userID,
	)
	return err
}

func (ps *PostgresStorage) GetActiveSubscription(userID int64) (*models.Subscription, error) {
	row := ps.db.QueryRow(
		`SELECT id, user_id, plan, status, expires_at, notes, activated_by, created_at, updated_at
		 FROM subscriptions WHERE user_id = $1 AND status = 'active'
		 ORDER BY created_at DESC LIMIT 1`, userID,
	)
	var sub models.Subscription
	var exp sql.NullTime
	err := row.Scan(&sub.ID, &sub.UserID, &sub.Plan, &sub.Status, &exp, &sub.Notes, &sub.ActivatedBy, &sub.CreatedAt, &sub.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if exp.Valid {
		sub.ExpiresAt = &exp.Time
	}
	return &sub, nil
}

func (ps *PostgresStorage) GetUserStats() (*models.UserStats, error) {
	stats := &models.UserStats{}
	_ = ps.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	_ = ps.db.QueryRow(
		`SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'active'
		 AND (expires_at IS NULL OR expires_at > NOW())`,
	).Scan(&stats.ActiveSubscriptions)
	_ = ps.db.QueryRow(
		`SELECT COUNT(*) FROM bot_states WHERE auto_trading_active = TRUE`,
	).Scan(&stats.AutoTradingUsers)
	_ = ps.db.QueryRow(
		`SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '7 days'`,
	).Scan(&stats.NewUsersLast7Days)
	_ = ps.db.QueryRow(`SELECT COUNT(*) FROM trades`).Scan(&stats.TotalTrades)
	_ = ps.db.QueryRow(`SELECT COUNT(*) FROM trades WHERE status = 'OPEN'`).Scan(&stats.OpenTrades)
	_ = ps.db.QueryRow(`SELECT COUNT(*) FROM trades WHERE status = 'CLOSED'`).Scan(&stats.ClosedTrades)
	_ = ps.db.QueryRow(
		`SELECT COUNT(*) FROM trades WHERE created_at > NOW() - INTERVAL '24 hours'`,
	).Scan(&stats.TradesLast24h)
	_ = ps.db.QueryRow(
		`SELECT COALESCE(SUM(profit), 0) FROM trades WHERE status = 'CLOSED' AND profit IS NOT NULL`,
	).Scan(&stats.NetProfitClosed)
	return stats, nil
}

func (ps *PostgresStorage) ListAutoTradingUserIDs() ([]int64, error) {
	rows, err := ps.db.Query(
		`SELECT user_id FROM bot_states WHERE auto_trading_active = TRUE AND is_paused = FALSE AND daily_loss_hit = FALSE`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func (ps *PostgresStorage) CreatePaymentRecord(p *models.PaymentRecord) error {
	_, err := ps.db.Exec(
		`INSERT INTO payment_records (id, user_id, amount, currency, method, reference, notes, created_by_admin, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.UserID, p.Amount, p.Currency, p.Method, p.Reference, p.Notes, p.CreatedByAdmin, p.CreatedAt,
	)
	return err
}

// ListSignalSubscriberTelegramIDs returns non-blocked users who may receive signal alerts.
func (ps *PostgresStorage) ListSignalSubscriberTelegramIDs() ([]int64, error) {
	rows, err := ps.db.Query(`SELECT telegram_id FROM users WHERE is_blocked = FALSE ORDER BY telegram_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func (ps *PostgresStorage) ListRecentUsers(limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := ps.db.Query(
		`SELECT id, telegram_id, username, first_name, last_name, is_blocked, created_at, last_seen_at
		 FROM users ORDER BY last_seen_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.IsBlocked, &u.CreatedAt, &u.LastSeenAt); err != nil {
			return nil, err
		}
		list = append(list, &u)
	}
	return list, rows.Err()
}
