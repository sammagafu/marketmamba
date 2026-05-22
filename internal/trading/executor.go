package trading

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"forex-bot/internal/accounts"
	"forex-bot/internal/broker"
	"forex-bot/internal/feedback"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
	"forex-bot/internal/utils"
)

// TradeExecutor handles automated trade execution
type TradeExecutor struct {
	broker          broker.Broker
	storage         storage.Storage
	tradeLog        *TradeLog
	validator       *risk.RiskValidator
	outcomeNotifier feedback.OutcomeNotifier
	userID          int64
	maxRetries      int
}

func NewTradeExecutor(b broker.Broker, s storage.Storage, v *risk.RiskValidator, userID int64, notifier feedback.OutcomeNotifier) *TradeExecutor {
	te := &TradeExecutor{
		broker:          b,
		storage:         s,
		validator:       v,
		outcomeNotifier: notifier,
		userID:          userID,
		maxRetries:      3,
	}
	if ps, ok := s.(*storage.PostgresStorage); ok {
		te.tradeLog = NewTradeLog(ps)
	}
	return te
}

// ExecuteSignal executes a trade signal with proper validation
func (te *TradeExecutor) ExecuteSignal(signal *models.TradeSignal) (*models.Position, error) {
	logger.Info("Executing signal for user %d: %s %s", te.userID, signal.Symbol, signal.Type)

	// Get current state
	botState, err := te.storage.GetBotState(te.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot state: %w", err)
	}

	// Check if trading is paused
	if botState.IsPaused {
		return nil, fmt.Errorf("trading is paused")
	}

	// Check if daily loss limit was hit
	if botState.DailyLossHit {
		return nil, fmt.Errorf("daily loss limit already hit")
	}

	account, err := te.ensureAccount()
	if err != nil {
		return nil, err
	}

	// Get open positions
	openPositions, err := te.broker.GetOpenPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	// Get today's trade count
	todayStats, err := te.storage.GetDailyStats(te.userID, time.Now())
	if err != nil {
		todayStats = &models.DailyStats{
			ID:          utils.GenerateID("stats"),
			UserID:      te.userID,
			TradingDate: time.Now(),
		}
	}

	// Get daily loss
	dailyLoss := todayStats.TotalLoss - todayStats.TotalProfit

	// Validate signal against risk rules
	if err := te.validator.ValidateTradeSignal(signal, account.Balance, len(openPositions), todayStats.TradeCount, dailyLoss, botState.IsPaused); err != nil {
		logger.Warn("Signal validation failed for user %d: %v", te.userID, err)
		return nil, fmt.Errorf("signal validation failed: %w", err)
	}

	// Calculate lot size based on risk
	lotSize, err := te.validator.CalculateLotSize(account.Balance, signal.StopLoss, signal.TakeProfit)
	if err != nil {
		logger.Error("Failed to calculate lot size: %v", err)
		return nil, fmt.Errorf("failed to calculate lot size: %w", err)
	}

	// Execute order with retry logic
	var position *models.Position
	for attempt := 1; attempt <= te.maxRetries; attempt++ {
		position, err = te.broker.OpenMarketOrder(
			signal.Symbol,
			signal.Type,
			lotSize,
			signal.StopLoss,
			signal.TakeProfit,
		)

		if err == nil {
			break
		}

		logger.Warn("Trade execution failed (attempt %d/%d): %v", attempt, te.maxRetries, err)
		if attempt < te.maxRetries {
			time.Sleep(time.Second * time.Duration(attempt))
		}
	}

	if err != nil {
		logger.Error("Failed to execute trade after %d attempts: %v", te.maxRetries, err)
		return nil, fmt.Errorf("failed to execute trade: %w", err)
	}

	position.UserID = te.userID
	source := autoTradeSource(signal)
	if te.tradeLog != nil {
		if _, err := te.tradeLog.RecordOpen(te.userID, position, source); err != nil {
			logger.Error("Failed to log trade open: %v", err)
		}
	}

	reasonMsg := tradeReasonSummary(signal, lotSize, account.Balance)
	te.logCommand("AUTO_TRADE", fmt.Sprintf("%s %s %.2f", signal.Symbol, signal.Type, position.Quantity), "SUCCESS", reasonMsg)

	logger.Info(
		"Trade taken for user %d: %s %s @ %.5f | %s | qty=%.2f SL=%.5f TP=%.5f",
		te.userID, signal.Symbol, signal.Type, position.EntryPrice, reasonMsg,
		position.Quantity, signal.StopLoss, signal.TakeProfit,
	)

	return position, nil
}

// CheckAndClosePositions checks positions against TP/SL levels
func (te *TradeExecutor) CheckAndClosePositions() error {
	positions, err := te.broker.GetOpenPositions()
	if err != nil {
		logger.Error("Failed to get positions: %v", err)
		return err
	}

	for _, pos := range positions {
		// Check take profit
		if shouldCloseTakeProfit(pos) {
			if err := te.closePosition(pos, "TP"); err != nil {
				logger.Error("Failed to close position at TP: %v", err)
			}
			continue
		}

		// Check stop loss
		if shouldCloseStopLoss(pos) {
			if err := te.closePosition(pos, "SL"); err != nil {
				logger.Error("Failed to close position at SL: %v", err)
			}
		}
	}

	return nil
}

func (te *TradeExecutor) closePosition(pos *models.Position, reason string) error {
	if err := te.broker.ClosePosition(pos.ID); err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	exitPrice := pos.CurrentPrice
	if exitPrice <= 0 {
		exitPrice = pos.EntryPrice
	}
	var closed *models.Trade
	if te.tradeLog != nil {
		if trade, err := te.tradeLog.RecordClose(te.userID, pos.ID, exitPrice, reason); err != nil {
			logger.Warn("Trade close log failed for %s: %v", pos.ID, err)
		} else if trade != nil {
			closed = trade
			te.updateDailyStats(trade)
		}
	}

	te.notifyOutcome(closed, reason)
	te.logCommand("AUTO_CLOSE", pos.ID, "SUCCESS", fmt.Sprintf("Position closed at %s", reason))

	return nil
}

func (te *TradeExecutor) notifyOutcome(trade *models.Trade, reason string) {
	if te.outcomeNotifier == nil || trade == nil {
		return
	}
	if err := te.outcomeNotifier.NotifyTradeOutcome(te.userID, trade, reason); err != nil {
		logger.Warn("Trade outcome notify user %d: %v", te.userID, err)
	}
}

func (te *TradeExecutor) updateDailyStats(trade *models.Trade) {
	stats, _ := te.storage.GetDailyStats(te.userID, time.Now())
	if stats == nil {
		stats = &models.DailyStats{
			ID:          utils.GenerateID("stats"),
			UserID:      te.userID,
			TradingDate: time.Now(),
		}
	}

	stats.TradeCount++
	if trade.Profit != nil && *trade.Profit > 0 {
		stats.WinCount++
		stats.TotalProfit += *trade.Profit
	} else if trade.Profit != nil {
		stats.LossCount++
		stats.TotalLoss += -*trade.Profit
	}

	stats.NetProfit = stats.TotalProfit - stats.TotalLoss
	if stats.TradeCount > 0 {
		stats.WinRate = float64(stats.WinCount) / float64(stats.TradeCount) * 100
	}
	stats.UpdatedAt = time.Now()

	if err := te.storage.UpdateDailyStats(stats); err != nil {
		logger.Error("Failed to update daily stats: %v", err)
	}

	// Check if daily loss limit hit (uses MAX_DAILY_LOSS from risk settings)
	if account, accErr := te.storage.GetAccountByUser(te.userID); accErr == nil && account != nil {
		maxLoss := te.validator.MaxDailyLossAmount(account.Balance)
		if maxLoss > 0 && stats.TotalLoss-stats.TotalProfit > maxLoss {
			if err := te.storage.UpdateBotState(te.userID, true, false, true); err != nil {
				logger.Error("Failed to update bot state: %v", err)
			}
			logger.Warn("Daily loss limit hit for user %d (loss %.2f > max %.2f)", te.userID, stats.TotalLoss-stats.TotalProfit, maxLoss)
		}
	}
}

func (te *TradeExecutor) logCommand(command, args, status, message string) {
	log := &models.CommandLog{
		ID:        utils.GenerateID("log"),
		UserID:    te.userID,
		Command:   command,
		Args:      args,
		Status:    status,
		Message:   message,
		CreatedAt: time.Now(),
	}

	if err := te.storage.LogCommand(log); err != nil {
		logger.Error("Failed to log command: %v", err)
	}
}

// Helper functions
func (te *TradeExecutor) ensureAccount() (*models.Account, error) {
	account, err := te.storage.GetAccountByUser(te.userID)
	if err == nil {
		return account, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	provider := "mock"
	if ps, ok := te.storage.(*storage.PostgresStorage); ok {
		if conn, connErr := ps.GetActiveBrokerConnection(te.userID); connErr == nil && conn != nil {
			provider = conn.Provider
		}
	}
	acctStore := accountStore(te.storage)
	if acctStore == nil {
		return nil, fmt.Errorf("failed to sync account: storage unavailable")
	}
	if syncErr := accounts.SyncFromBroker(acctStore, te.userID, provider, te.broker); syncErr != nil {
		return nil, fmt.Errorf("failed to sync account: %w", syncErr)
	}
	account, err = te.storage.GetAccountByUser(te.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account after sync: %w", err)
	}
	logger.Info("Synced trading account for user %d (provider=%s balance=%.2f)", te.userID, provider, account.Balance)
	return account, nil
}

func accountStore(s storage.Storage) accounts.AccountStore {
	if a, ok := s.(accounts.AccountStore); ok {
		return a
	}
	return nil
}

func autoTradeSource(signal *models.TradeSignal) string {
	if signal == nil || signal.Reason == "" {
		return "AUTO"
	}
	return "AUTO: " + signal.Reason
}

func tradeReasonSummary(signal *models.TradeSignal, lotSize, balance float64) string {
	if signal == nil {
		return "auto-trade"
	}
	reason := signal.Reason
	if reason == "" {
		reason = "signal filters passed"
	}
	return fmt.Sprintf(
		"%s | lot=%.2f balance=%.2f strength=%.0f%% R:R=%.2f",
		reason, lotSize, balance, signal.Strength*100, signal.RiskRewardRatio,
	)
}

func shouldCloseTakeProfit(pos *models.Position) bool {
	if pos.CurrentPrice <= 0 || pos.TakeProfit <= 0 {
		return false
	}

	if pos.Type == "BUY" {
		return pos.CurrentPrice >= pos.TakeProfit
	}
	return pos.CurrentPrice <= pos.TakeProfit
}

func shouldCloseStopLoss(pos *models.Position) bool {
	if pos.CurrentPrice <= 0 || pos.StopLoss <= 0 {
		return false
	}

	if pos.Type == "BUY" {
		return pos.CurrentPrice <= pos.StopLoss
	}
	return pos.CurrentPrice >= pos.StopLoss
}
