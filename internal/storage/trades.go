package storage

import (
	"database/sql"

	"forex-bot/internal/models"
)

func (ps *PostgresStorage) ListRecentTrades(limit int) ([]*models.Trade, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := ps.db.Query(
		`SELECT id, user_id, symbol, type, entry_price, quantity, stop_loss, take_profit,
		        risk_amount, reward_amount, risk_reward_ratio, status, exit_price, exit_time,
		        profit, closure_reason, created_at, updated_at
		 FROM trades ORDER BY created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrades(rows)
}

func (ps *PostgresStorage) ListTradesByUser(userID int64, limit int) ([]*models.Trade, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := ps.db.Query(
		`SELECT id, user_id, symbol, type, entry_price, quantity, stop_loss, take_profit,
		        risk_amount, reward_amount, risk_reward_ratio, status, exit_price, exit_time,
		        profit, closure_reason, created_at, updated_at
		 FROM trades WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrades(rows)
}

func (ps *PostgresStorage) GetOpenTradesByUser(userID int64) ([]*models.Trade, error) {
	rows, err := ps.db.Query(
		`SELECT id, user_id, symbol, type, entry_price, quantity, stop_loss, take_profit,
		        risk_amount, reward_amount, risk_reward_ratio, status, exit_price, exit_time,
		        profit, closure_reason, created_at, updated_at
		 FROM trades WHERE user_id = $1 AND status = 'OPEN'
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTrades(rows)
}

func (ps *PostgresStorage) GetTradeByBrokerPositionID(userID int64, brokerPositionID string) (*models.Trade, error) {
	row := ps.db.QueryRow(
		`SELECT t.id, t.user_id, t.symbol, t.type, t.entry_price, t.quantity, t.stop_loss, t.take_profit,
		        t.risk_amount, t.reward_amount, t.risk_reward_ratio, t.status, t.exit_price, t.exit_time,
		        t.profit, t.closure_reason, t.created_at, t.updated_at
		 FROM trades t
		 INNER JOIN positions p ON p.trade_id = t.id
		 WHERE p.user_id = $1 AND (p.id = $2 OR p.broker_id = $2) AND t.status = 'OPEN'
		 LIMIT 1`,
		userID, brokerPositionID,
	)
	return scanTradeRow(row)
}

func scanTradeRow(row *sql.Row) (*models.Trade, error) {
	var t models.Trade
	var exitPrice, profit sql.NullFloat64
	var exitTime sql.NullTime
	var closureReason sql.NullString
	err := row.Scan(
		&t.ID, &t.UserID, &t.Symbol, &t.Type, &t.EntryPrice, &t.Quantity, &t.StopLoss, &t.TakeProfit,
		&t.RiskAmount, &t.RewardAmount, &t.RiskRewardRatio, &t.Status,
		&exitPrice, &exitTime, &profit, &closureReason, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if exitPrice.Valid {
		t.ExitPrice = &exitPrice.Float64
	}
	if exitTime.Valid {
		t.ExitTime = &exitTime.Time
	}
	if profit.Valid {
		t.Profit = &profit.Float64
	}
	if closureReason.Valid {
		t.ClosureReason = &closureReason.String
	}
	return &t, nil
}

func scanTrades(rows *sql.Rows) ([]*models.Trade, error) {
	var list []*models.Trade
	for rows.Next() {
		var t models.Trade
		var exitPrice, profit sql.NullFloat64
		var exitTime sql.NullTime
		var closureReason sql.NullString
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Symbol, &t.Type, &t.EntryPrice, &t.Quantity, &t.StopLoss, &t.TakeProfit,
			&t.RiskAmount, &t.RewardAmount, &t.RiskRewardRatio, &t.Status,
			&exitPrice, &exitTime, &profit, &closureReason, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if exitPrice.Valid {
			t.ExitPrice = &exitPrice.Float64
		}
		if exitTime.Valid {
			t.ExitTime = &exitTime.Time
		}
		if profit.Valid {
			t.Profit = &profit.Float64
		}
		if closureReason.Valid {
			t.ClosureReason = &closureReason.String
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}
