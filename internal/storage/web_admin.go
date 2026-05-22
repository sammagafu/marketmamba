package storage

import (
	"database/sql"
	"strings"
	"time"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) UpsertWebAdmin(a *models.WebAdmin) error {
	_, err := ps.db.Exec(
		`INSERT INTO web_admins (id, email, password_hash, telegram_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (email) DO UPDATE SET
		   password_hash = EXCLUDED.password_hash,
		   telegram_id = EXCLUDED.telegram_id,
		   updated_at = EXCLUDED.updated_at`,
		a.ID, strings.ToLower(strings.TrimSpace(a.Email)), a.PasswordHash, a.TelegramID, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (ps *PostgresStorage) GetWebAdminByEmail(email string) (*models.WebAdmin, error) {
	row := ps.db.QueryRow(
		`SELECT id, email, password_hash, telegram_id, created_at, updated_at
		 FROM web_admins WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	)
	var a models.WebAdmin
	err := row.Scan(&a.ID, &a.Email, &a.PasswordHash, &a.TelegramID, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (ps *PostgresStorage) SetUserBlocked(telegramID int64, blocked bool) error {
	res, err := ps.db.Exec(`UPDATE users SET is_blocked = $1 WHERE telegram_id = $2`, blocked, telegramID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (ps *PostgresStorage) RevokeActiveSubscription(telegramID int64) error {
	_, err := ps.db.Exec(
		`UPDATE subscriptions SET status = 'revoked', updated_at = $1
		 WHERE user_id = $2 AND status = 'active'`,
		time.Now(), telegramID,
	)
	return err
}
