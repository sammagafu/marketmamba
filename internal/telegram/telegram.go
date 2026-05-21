package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"forex-bot/internal/broker"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
)

type TelegramBot struct {
	api          *tgbotapi.BotAPI
	allowedUsers []int64
	broker       broker.Broker
	storage      storage.Storage
	validator    *risk.RiskValidator
}

func NewTelegramBot(token string, allowedUsers []int64, b broker.Broker, s storage.Storage, v *risk.RiskValidator) (*TelegramBot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		api:          api,
		allowedUsers: allowedUsers,
		broker:       b,
		storage:      s,
		validator:    v,
	}, nil
}

func (tb *TelegramBot) isAllowed(userID int64) bool {
	for _, id := range tb.allowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}

func (tb *TelegramBot) Start() error {
	logger.Info("Telegram bot started: @%s", tb.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !tb.isAllowed(update.Message.From.ID) {
			tb.sendMessage(update.Message.Chat.ID, "❌ Unauthorized access")
			logger.Warn("Unauthorized access attempt from user %d", update.Message.From.ID)
			continue
		}

		tb.handleMessage(update.Message)
	}

	return nil
}

func (tb *TelegramBot) handleMessage(msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)

	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	command := parts[0]

	logger.Info("User %d executed command: %s", userID, command)

	switch command {
	case "/start":
		tb.handleStart(chatID)
	case "/status":
		tb.handleStatus(chatID, userID)
	case "/balance":
		tb.handleBalance(chatID, userID)
	case "/positions":
		tb.handlePositions(chatID, userID)
	case "/open":
		tb.handleOpen(chatID, userID, parts[1:])
	case "/close":
		tb.handleClose(chatID, userID, parts[1:])
	case "/closeall":
		tb.handleCloseAll(chatID, userID)
	case "/pause":
		tb.handlePause(chatID, userID)
	case "/resume":
		tb.handleResume(chatID, userID)
	case "/risk":
		tb.handleRisk(chatID, userID, parts[1:])
	case "/dailyreport":
		tb.handleDailyReport(chatID, userID)
	default:
		tb.sendMessage(chatID, "❓ Unknown command. Use /start for help.")
	}
}

func (tb *TelegramBot) handleStart(chatID int64) {
	msg := `🤖 *Forex Scalping Bot*

*Available Commands:*
/status - Bot and trading status
/balance - Account balance
/positions - Open positions
/open <symbol> <type> <qty> <sl> <tp> - Open trade
/close <positionID> - Close position
/closeall - Close all positions
/pause - Pause trading
/resume - Resume trading
/risk - View risk settings
/dailyreport - Daily trading report

*Example:*
/open EURUSD BUY 1.0 1.0900 1.1000

⚠️ *DISCLAIMER*: Forex trading carries high risk. Use this bot responsibly.`

	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleStatus(chatID int64, userID int64) {
	botState, err := tb.storage.GetBotState(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching bot state")
		return
	}

	status := "✅ Active"
	if botState.IsPaused {
		status = "⏸️ Paused"
	}

	msg := fmt.Sprintf(`*Bot Status*
Status: %s
Auto Trading: %v
Daily Loss Hit: %v
Last Active: %s`,
		status,
		botState.AutoTradingActive,
		botState.DailyLossHit,
		botState.LastActiveAt.Format("2006-01-02 15:04:05"))

	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleBalance(chatID int64, userID int64) {
	balance, err := tb.broker.GetBalance()
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching balance")
		return
	}

	equity, err := tb.broker.GetEquity()
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching equity")
		return
	}

	msg := fmt.Sprintf(`*Account Balance*
Balance: $%.2f
Equity: $%.2f`, balance, equity)

	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handlePositions(chatID int64, userID int64) {
	positions, err := tb.broker.GetOpenPositions()
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching positions")
		return
	}

	if len(positions) == 0 {
		tb.sendMessage(chatID, "No open positions")
		return
	}

	msg := "*Open Positions*\n\n"
	for i, pos := range positions {
		msg += fmt.Sprintf("%d. %s %s\n", i+1, pos.Symbol, pos.Type)
		msg += fmt.Sprintf("   Entry: %.5f | SL: %.5f | TP: %.5f\n", pos.EntryPrice, pos.StopLoss, pos.TakeProfit)
		msg += fmt.Sprintf("   Profit: %.2f (%.2f%%)\n\n", pos.Profit, pos.ProfitPct)
	}

	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleOpen(chatID int64, userID int64, args []string) {
	if len(args) < 5 {
		tb.sendMessage(chatID, "❌ Usage: /open <symbol> <BUY|SELL> <quantity> <stopLoss> <takeProfit>")
		return
	}

	symbol := strings.ToUpper(args[0])
	orderType := strings.ToUpper(args[1])
	qty, _ := strconv.ParseFloat(args[2], 64)
	sl, _ := strconv.ParseFloat(args[3], 64)
	tp, _ := strconv.ParseFloat(args[4], 64)

	if qty <= 0 || sl <= 0 || tp <= 0 {
		tb.sendMessage(chatID, "❌ Invalid parameters")
		return
	}

	// Create signal for validation
	signal := &models.TradeSignal{
		Symbol:      symbol,
		Type:        orderType,
		StopLoss:    sl,
		TakeProfit:  tp,
		Strength:    1.0,
		TriggeredAt: time.Now(),
	}

	botState, _ := tb.storage.GetBotState(userID)
	if err := tb.validator.ValidateTradeSignal(signal, 10000, 0, 0, 0, botState.IsPaused); err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ Validation failed: %v", err))
		return
	}

	pos, err := tb.broker.OpenMarketOrder(symbol, orderType, qty, sl, tp)
	if err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ Failed to open trade: %v", err))
		return
	}

	msg := fmt.Sprintf(`✅ Trade Opened
Symbol: %s
Type: %s
Entry: %.5f
Stop Loss: %.5f
Take Profit: %.5f
Position ID: %s`, symbol, orderType, pos.EntryPrice, sl, tp, pos.ID)

	tb.sendMessage(chatID, msg)
	logger.Info("Trade opened for user %d: %s %s", userID, symbol, orderType)
}

func (tb *TelegramBot) handleClose(chatID int64, userID int64, args []string) {
	if len(args) < 1 {
		tb.sendMessage(chatID, "❌ Usage: /close <positionID>")
		return
	}

	positionID := args[0]
	if err := tb.broker.ClosePosition(positionID); err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ Failed to close position: %v", err))
		return
	}

	tb.sendMessage(chatID, "✅ Position closed")
	logger.Info("Position closed for user %d: %s", userID, positionID)
}

func (tb *TelegramBot) handleCloseAll(chatID int64, userID int64) {
	if err := tb.broker.CloseAllPositions(); err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ Failed to close all positions: %v", err))
		return
	}

	tb.sendMessage(chatID, "✅ All positions closed")
	logger.Info("All positions closed for user %d", userID)
}

func (tb *TelegramBot) handlePause(chatID int64, userID int64) {
	if err := tb.storage.UpdateBotState(userID, true, false, false); err != nil {
		tb.sendMessage(chatID, "❌ Failed to pause bot")
		return
	}

	tb.sendMessage(chatID, "⏸️ Trading paused")
	logger.Info("Trading paused for user %d", userID)
}

func (tb *TelegramBot) handleResume(chatID int64, userID int64) {
	if err := tb.storage.UpdateBotState(userID, false, false, false); err != nil {
		tb.sendMessage(chatID, "❌ Failed to resume bot")
		return
	}

	tb.sendMessage(chatID, "▶️ Trading resumed")
	logger.Info("Trading resumed for user %d", userID)
}

func (tb *TelegramBot) handleRisk(chatID int64, userID int64, args []string) {
	settings, err := tb.storage.GetRiskSettings(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching risk settings")
		return
	}

	msg := fmt.Sprintf(`*Risk Settings*
Max Risk Per Trade: %.2f%%
Max Daily Loss: %.2f%%
Max Open Trades: %d
Max Trades/Day: %d
Risk-Reward Ratio: %.2f`,
		settings.MaxRiskPerTrade*100,
		settings.MaxDailyLoss*100,
		settings.MaxOpenTrades,
		settings.MaxTradesPerDay,
		settings.RiskRewardRatio)

	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleDailyReport(chatID int64, userID int64) {
	stats, err := tb.storage.GetDailyStats(userID, time.Now())
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching daily stats")
		return
	}

	msg := fmt.Sprintf(`*Daily Report - %s*
Trades: %d (W: %d | L: %d)
Win Rate: %.2f%%
Net Profit: $%.2f
Max Drawdown: %.2f%%`,
		time.Now().Format("2006-01-02"),
		stats.TradeCount,
		stats.WinCount,
		stats.LossCount,
		stats.WinRate,
		stats.NetProfit,
		stats.MaxDrawdown)

	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err := tb.api.Send(msg); err != nil {
		logger.Error("Failed to send message: %v", err)
	}
}
