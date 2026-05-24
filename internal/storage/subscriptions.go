package storage

import "context"

// CountPaidSubscribers returns distinct users with an active paid plan (monthly, pro, manual).
// Trial and revoked subscriptions are excluded. Used internally for community asset unlock.
func (ps *PostgresStorage) CountPaidSubscribers(ctx context.Context) (int, error) {
	var n int
	err := ps.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM subscriptions
		WHERE status = 'active'
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND plan IN ('monthly', 'pro', 'manual')
	`).Scan(&n)
	return n, err
}
