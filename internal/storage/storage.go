package storage

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"forex-bot/internal/models"
)

type Storage interface {
	// Trades
	CreateTrade(trade *models.Trade) error
	GetTradeByID(tradeID string) (*models.Trade, error)
	UpdateTrade(trade *models.Trade) error

	// Positions
	CreatePosition(position *models.Position) error
	GetPositionByID(positionID string) (*models.Position, error)
	GetOpenPositionsByUser(userID int64) ([]*models.Position, error)
	UpdatePosition(position *models.Position) error
	DeletePosition(positionID string) error

	// Account
	CreateAccount(account *models.Account) error
	GetAccountByUser(userID int64) (*models.Account, error)
	UpdateAccount(account *models.Account) error

	// Risk Settings
	CreateRiskSettings(settings *models.RiskSettings) error
	GetRiskSettings(userID int64) (*models.RiskSettings, error)
	UpdateRiskSettings(settings *models.RiskSettings) error

	// Daily Stats
	CreateDailyStats(stats *models.DailyStats) error
	GetDailyStats(userID int64, date time.Time) (*models.DailyStats, error)
	UpdateDailyStats(stats *models.DailyStats) error

	// Bot State
	CreateBotState(state *models.BotState) error
	GetBotState(userID int64) (*models.BotState, error)
	UpdateBotState(userID int64, isPaused, autoTrading, dailyLossHit bool) error
	SetAutoTradeApproved(userID int64, approved bool) error

	// Command Logs
	LogCommand(log *models.CommandLog) error

	// Users (ACL)
	GetUserByTelegramID(telegramID int64) (*models.User, error)

	// Health
	Health() error
	Close() error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(databaseURL string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

// Trades
func (ps *PostgresStorage) CreateTrade(trade *models.Trade) error {
	query := `INSERT INTO trades (id, user_id, symbol, type, entry_price, quantity, stop_loss, take_profit, risk_amount, reward_amount, risk_reward_ratio, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := ps.db.Exec(query,
		trade.ID, trade.UserID, trade.Symbol, trade.Type, trade.EntryPrice,
		trade.Quantity, trade.StopLoss, trade.TakeProfit, trade.RiskAmount,
		trade.RewardAmount, trade.RiskRewardRatio, trade.Status,
		trade.CreatedAt, trade.UpdatedAt)

	return err
}

func (ps *PostgresStorage) GetTradeByID(tradeID string) (*models.Trade, error) {
	trade := &models.Trade{}
	query := `SELECT id, user_id, symbol, type, entry_price, quantity, stop_loss, take_profit, risk_amount, reward_amount, risk_reward_ratio, status, exit_price, exit_time, profit, closure_reason, created_at, updated_at FROM trades WHERE id = $1`

	err := ps.db.QueryRow(query, tradeID).Scan(
		&trade.ID, &trade.UserID, &trade.Symbol, &trade.Type, &trade.EntryPrice,
		&trade.Quantity, &trade.StopLoss, &trade.TakeProfit, &trade.RiskAmount,
		&trade.RewardAmount, &trade.RiskRewardRatio, &trade.Status, &trade.ExitPrice,
		&trade.ExitTime, &trade.Profit, &trade.ClosureReason, &trade.CreatedAt,
		&trade.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return trade, nil
}

func (ps *PostgresStorage) UpdateTrade(trade *models.Trade) error {
	query := `UPDATE trades SET status=$1, exit_price=$2, exit_time=$3, profit=$4, closure_reason=$5, updated_at=$6 WHERE id=$7`

	_, err := ps.db.Exec(query, trade.Status, trade.ExitPrice, trade.ExitTime, trade.Profit, trade.ClosureReason, trade.UpdatedAt, trade.ID)
	return err
}

// Positions
func (ps *PostgresStorage) CreatePosition(position *models.Position) error {
	query := `INSERT INTO positions (id, trade_id, broker_id, user_id, symbol, type, quantity, entry_price, current_price, stop_loss, take_profit, profit, profit_pct, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := ps.db.Exec(query,
		position.ID, position.TradeID, position.BrokerID, position.UserID,
		position.Symbol, position.Type, position.Quantity, position.EntryPrice,
		position.CurrentPrice, position.StopLoss, position.TakeProfit,
		position.Profit, position.ProfitPct, position.UpdatedAt)

	return err
}

func (ps *PostgresStorage) GetPositionByID(positionID string) (*models.Position, error) {
	position := &models.Position{}
	query := `SELECT id, trade_id, broker_id, user_id, symbol, type, quantity, entry_price, current_price, stop_loss, take_profit, profit, profit_pct, updated_at FROM positions WHERE id = $1`

	err := ps.db.QueryRow(query, positionID).Scan(
		&position.ID, &position.TradeID, &position.BrokerID, &position.UserID,
		&position.Symbol, &position.Type, &position.Quantity, &position.EntryPrice,
		&position.CurrentPrice, &position.StopLoss, &position.TakeProfit,
		&position.Profit, &position.ProfitPct, &position.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return position, nil
}

func (ps *PostgresStorage) GetOpenPositionsByUser(userID int64) ([]*models.Position, error) {
	query := `SELECT id, trade_id, broker_id, user_id, symbol, type, quantity, entry_price, current_price, stop_loss, take_profit, profit, profit_pct, updated_at FROM positions WHERE user_id = $1`

	rows, err := ps.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []*models.Position
	for rows.Next() {
		position := &models.Position{}
		err := rows.Scan(
			&position.ID, &position.TradeID, &position.BrokerID, &position.UserID,
			&position.Symbol, &position.Type, &position.Quantity, &position.EntryPrice,
			&position.CurrentPrice, &position.StopLoss, &position.TakeProfit,
			&position.Profit, &position.ProfitPct, &position.UpdatedAt)

		if err != nil {
			return nil, err
		}

		positions = append(positions, position)
	}

	return positions, rows.Err()
}

func (ps *PostgresStorage) UpdatePosition(position *models.Position) error {
	query := `UPDATE positions SET current_price=$1, profit=$2, profit_pct=$3, updated_at=$4 WHERE id=$5`

	_, err := ps.db.Exec(query, position.CurrentPrice, position.Profit, position.ProfitPct, position.UpdatedAt, position.ID)
	return err
}

func (ps *PostgresStorage) DeletePosition(positionID string) error {
	query := `DELETE FROM positions WHERE id=$1`
	_, err := ps.db.Exec(query, positionID)
	return err
}

// Account
func (ps *PostgresStorage) CreateAccount(account *models.Account) error {
	query := `INSERT INTO accounts (id, user_id, broker_provider, balance, equity, used_margin, free_margin, leverage, last_synced_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := ps.db.Exec(query,
		account.ID, account.UserID, account.BrokerProvider, account.Balance, account.Equity,
		account.UsedMargin, account.FreeMargin, account.Leverage, account.LastSyncedAt, account.UpdatedAt)

	return err
}

func (ps *PostgresStorage) GetAccountByUser(userID int64) (*models.Account, error) {
	account := &models.Account{}
	query := `SELECT id, user_id, broker_provider, balance, equity, used_margin, free_margin, leverage, last_synced_at, updated_at FROM accounts WHERE user_id = $1`

	err := ps.db.QueryRow(query, userID).Scan(
		&account.ID, &account.UserID, &account.BrokerProvider, &account.Balance, &account.Equity,
		&account.UsedMargin, &account.FreeMargin, &account.Leverage, &account.LastSyncedAt, &account.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (ps *PostgresStorage) UpdateAccount(account *models.Account) error {
	query := `UPDATE accounts SET balance=$1, equity=$2, used_margin=$3, free_margin=$4, last_synced_at=$5, updated_at=$6 WHERE id=$7`

	_, err := ps.db.Exec(query, account.Balance, account.Equity, account.UsedMargin, account.FreeMargin, account.LastSyncedAt, account.UpdatedAt, account.ID)
	return err
}

// Risk Settings
func (ps *PostgresStorage) CreateRiskSettings(settings *models.RiskSettings) error {
	query := `INSERT INTO risk_settings (id, user_id, max_risk_per_trade, max_daily_loss, max_open_trades, max_trades_per_day, risk_reward_ratio, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := ps.db.Exec(query,
		settings.ID, settings.UserID, settings.MaxRiskPerTrade, settings.MaxDailyLoss,
		settings.MaxOpenTrades, settings.MaxTradesPerDay, settings.RiskRewardRatio,
		settings.CreatedAt, settings.UpdatedAt)

	return err
}

func (ps *PostgresStorage) GetRiskSettings(userID int64) (*models.RiskSettings, error) {
	settings := &models.RiskSettings{}
	query := `SELECT id, user_id, max_risk_per_trade, max_daily_loss, max_open_trades, max_trades_per_day, risk_reward_ratio, created_at, updated_at FROM risk_settings WHERE user_id = $1`

	err := ps.db.QueryRow(query, userID).Scan(
		&settings.ID, &settings.UserID, &settings.MaxRiskPerTrade, &settings.MaxDailyLoss,
		&settings.MaxOpenTrades, &settings.MaxTradesPerDay, &settings.RiskRewardRatio,
		&settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (ps *PostgresStorage) UpdateRiskSettings(settings *models.RiskSettings) error {
	query := `UPDATE risk_settings SET max_risk_per_trade=$1, max_daily_loss=$2, max_open_trades=$3, max_trades_per_day=$4, risk_reward_ratio=$5, updated_at=$6 WHERE id=$7`

	_, err := ps.db.Exec(query,
		settings.MaxRiskPerTrade, settings.MaxDailyLoss, settings.MaxOpenTrades,
		settings.MaxTradesPerDay, settings.RiskRewardRatio, settings.UpdatedAt, settings.ID)

	return err
}

// Daily Stats
func (ps *PostgresStorage) CreateDailyStats(stats *models.DailyStats) error {
	query := `INSERT INTO daily_stats (id, user_id, trading_date, trade_count, win_count, loss_count, total_profit, total_loss, net_profit, win_rate, max_drawdown, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := ps.db.Exec(query,
		stats.ID, stats.UserID, stats.TradingDate, stats.TradeCount, stats.WinCount, stats.LossCount,
		stats.TotalProfit, stats.TotalLoss, stats.NetProfit, stats.WinRate, stats.MaxDrawdown, stats.UpdatedAt)

	return err
}

func (ps *PostgresStorage) GetDailyStats(userID int64, date time.Time) (*models.DailyStats, error) {
	stats := &models.DailyStats{}
	query := `SELECT id, user_id, trading_date, trade_count, win_count, loss_count, total_profit, total_loss, net_profit, win_rate, max_drawdown, updated_at FROM daily_stats WHERE user_id = $1 AND DATE(trading_date) = $2`

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	err := ps.db.QueryRow(query, userID, startOfDay).Scan(
		&stats.ID, &stats.UserID, &stats.TradingDate, &stats.TradeCount, &stats.WinCount, &stats.LossCount,
		&stats.TotalProfit, &stats.TotalLoss, &stats.NetProfit, &stats.WinRate, &stats.MaxDrawdown, &stats.UpdatedAt)

	if err == sql.ErrNoRows {
		return &models.DailyStats{
			UserID:       userID,
			TradingDate:  startOfDay,
			TradeCount:   0,
			WinCount:     0,
			LossCount:    0,
			TotalProfit:  0,
			TotalLoss:    0,
			NetProfit:    0,
			WinRate:      0,
			MaxDrawdown:  0,
			UpdatedAt:    time.Now(),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (ps *PostgresStorage) UpdateDailyStats(stats *models.DailyStats) error {
	query := `UPDATE daily_stats SET trade_count=$1, win_count=$2, loss_count=$3, total_profit=$4, total_loss=$5, net_profit=$6, win_rate=$7, max_drawdown=$8, updated_at=$9 WHERE id=$10`

	_, err := ps.db.Exec(query,
		stats.TradeCount, stats.WinCount, stats.LossCount, stats.TotalProfit, stats.TotalLoss,
		stats.NetProfit, stats.WinRate, stats.MaxDrawdown, stats.UpdatedAt, stats.ID)

	return err
}

// Bot State
func (ps *PostgresStorage) CreateBotState(state *models.BotState) error {
	query := `INSERT INTO bot_states (id, user_id, is_paused, auto_trading_active, auto_trade_approved, daily_loss_hit, last_active_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := ps.db.Exec(query, state.ID, state.UserID, state.IsPaused, state.AutoTradingActive, state.AutoTradeApproved, state.DailyLossHit, state.LastActiveAt, state.UpdatedAt)
	return err
}

func (ps *PostgresStorage) GetBotState(userID int64) (*models.BotState, error) {
	state := &models.BotState{}
	query := `SELECT id, user_id, is_paused, auto_trading_active, auto_trade_approved, daily_loss_hit, last_active_at, updated_at FROM bot_states WHERE user_id = $1`

	err := ps.db.QueryRow(query, userID).Scan(
		&state.ID, &state.UserID, &state.IsPaused, &state.AutoTradingActive, &state.AutoTradeApproved, &state.DailyLossHit, &state.LastActiveAt, &state.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return state, nil
}

func (ps *PostgresStorage) UpdateBotState(userID int64, isPaused, autoTrading, dailyLossHit bool) error {
	query := `UPDATE bot_states SET is_paused=$1, auto_trading_active=$2, daily_loss_hit=$3, last_active_at=$4, updated_at=$5 WHERE user_id=$6`

	_, err := ps.db.Exec(query, isPaused, autoTrading, dailyLossHit, time.Now(), time.Now(), userID)
	return err
}

func (ps *PostgresStorage) SetAutoTradeApproved(userID int64, approved bool) error {
	_, err := ps.db.Exec(
		`UPDATE bot_states SET auto_trade_approved=$1, updated_at=$2 WHERE user_id=$3`,
		approved, time.Now(), userID,
	)
	return err
}

// Command Logs
func (ps *PostgresStorage) LogCommand(log *models.CommandLog) error {
	query := `INSERT INTO command_logs (id, user_id, command, args, status, message, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := ps.db.Exec(query, log.ID, log.UserID, log.Command, log.Args, log.Status, log.Message, log.CreatedAt)
	return err
}

// Health
func (ps *PostgresStorage) Health() error {
	return ps.db.Ping()
}

func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}
