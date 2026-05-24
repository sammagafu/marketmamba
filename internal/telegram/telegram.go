package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/decision"
	"forex-bot/internal/feedback"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/positions"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/tier"
	"forex-bot/internal/users"
)

type BrokerResolver func(userID int64) (broker.Broker, error)

type TelegramBot struct {
	api              *tgbotapi.BotAPI
	cfg              *config.Config
	storage          storage.Storage
	validator        *risk.RiskValidator
	users            *users.Service
	subs             *subscription.Service
	tier             *tier.Service
	resolveBroker    BrokerResolver
	outcomeNotifier  feedback.OutcomeNotifier
	decisionEngine   *decision.Engine
}

// SetDecisionEngine enables /analyze and sniper advisory messages.
func (tb *TelegramBot) SetDecisionEngine(e *decision.Engine) {
	tb.decisionEngine = e
}

func NewTelegramBot(
	cfg *config.Config,
	brokerResolver BrokerResolver,
	s storage.Storage,
	v *risk.RiskValidator,
	u *users.Service,
	sub *subscription.Service,
	tierSvc *tier.Service,
) (*TelegramBot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, err
	}
	return &TelegramBot{
		api:           api,
		cfg:           cfg,
		storage:       s,
		validator:     v,
		users:         u,
		subs:          sub,
		tier:          tierSvc,
		resolveBroker: brokerResolver,
	}, nil
}

// SetOutcomeNotifier wires TP/SL feedback (trader + signal subscribers).
func (tb *TelegramBot) SetOutcomeNotifier(n feedback.OutcomeNotifier) {
	tb.outcomeNotifier = n
}

func (tb *TelegramBot) outcomes() feedback.OutcomeNotifier {
	if tb.outcomeNotifier != nil {
		return tb.outcomeNotifier
	}
	return tb
}

func (tb *TelegramBot) Start() error {
	logger.Info("Telegram bot started: @%s (public=%v)", tb.api.Self.UserName, tb.cfg.App.PublicMode)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tb.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		tb.processMessage(update.Message)
	}
	return nil
}

func (tb *TelegramBot) processMessage(msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	if tb.cfg.App.PublicMode {
		if _, err := tb.users.RegisterFromTelegram(msg.From); err != nil {
			logger.Error("Register user %d: %v", userID, err)
		}
		_ = tb.users.Touch(userID)
	} else if !tb.isLegacyAllowed(userID) {
		tb.sendMessage(chatID, "❌ This bot is private. Contact the administrator.")
		logger.Warn("Private mode reject user %d", userID)
		return
	}

	if tb.isUserBlocked(userID) {
		tb.sendMessage(chatID, "❌ Your account is blocked. Contact support.")
		return
	}

	text := strings.TrimSpace(msg.Text)
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}
	command := parts[0]
	logger.Info("User %d executed command: %s", userID, command)

	if strings.HasPrefix(command, "/admin") {
		tb.handleAdmin(chatID, userID, parts)
		return
	}

	switch command {
	case "/start":
		tb.handleStart(chatID, userID)
	case "/subscribe":
		tb.handleSubscribe(chatID, userID)
	case "/myplan":
		tb.handleMyPlan(chatID, userID)
	case "/status":
		tb.handleStatus(chatID, userID)
	case "/broker":
		tb.handleBroker(chatID, userID, parts[1:])
	case "/balance":
		tb.handleBalance(chatID, userID)
	case "/positions":
		tb.handlePositions(chatID, userID)
	case "/trades":
		tb.handleTrades(chatID, userID)
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
	case "/pairs":
		tb.handlePairs(chatID, userID, parts[1:])
	case "/signaltypes":
		tb.handleSignalTypes(chatID, userID, parts[1:])
	case "/autostart":
		tb.handleAutoStart(chatID, userID)
	case "/approveauto":
		tb.handleApproveAuto(chatID, userID, parts[1:])
	case "/revokeauto":
		tb.handleRevokeAuto(chatID, userID, parts[1:])
	case "/autostop":
		tb.handleAutoStop(chatID, userID)
	case "/autostatus":
		tb.handleAutoStatus(chatID, userID)
	case "/analyze":
		tb.handleAnalyze(chatID, userID, parts[1:])
	default:
		tb.sendMessage(chatID, "❓ Unknown command. Use /start for help.")
	}
}

func (tb *TelegramBot) isLegacyAllowed(userID int64) bool {
	for _, id := range tb.cfg.Telegram.AllowedUserIDs {
		if id == userID {
			return true
		}
	}
	return tb.cfg.IsAdmin(userID)
}

func (tb *TelegramBot) isUserBlocked(userID int64) bool {
	if tb.cfg.IsAdmin(userID) {
		return false
	}
	u, err := tb.storage.GetUserByTelegramID(userID)
	return err == nil && u != nil && u.IsBlocked
}

func (tb *TelegramBot) requireTrading(chatID, userID int64) bool {
	ok, msg := tb.subs.CanTrade(userID)
	if !ok {
		tb.sendMessage(chatID, "🔒 "+msg)
		return false
	}
	return true
}

func (tb *TelegramBot) brokerFor(userID int64) (broker.Broker, error) {
	return tb.resolveBroker(userID)
}

func (tb *TelegramBot) handleStart(chatID, userID int64) {
	sub, _ := tb.subs.GetForUser(userID)
	planLine := fmt.Sprintf(
		"*%d-day free trial* — then *%.0f USDT/month* (Binance USDT only, no cards).\n%s",
		tb.cfg.App.FreeTrialDays,
		tb.cfg.Payments.SubscriptionPriceUSDT,
		tb.cfg.App.ValueProposition,
	)
	if sub != nil && sub.ExpiresAt != nil {
		planLine = fmt.Sprintf("Plan: *%s* until %s", sub.Plan, sub.ExpiresAt.Format("2006-01-02 15:04"))
	}
	planLine += "\nOpen *📊 Dashboard* (menu) for trades & subscription."
	msg := fmt.Sprintf(`🐍 *Market Mamba*

Welcome! Your Telegram ID: `+"`%d`"+`

%s

*Account:*
/subscribe — plans & payment info
/myplan — your subscription
/status — bot status
/broker — connect broker (demo or MetaAPI MT wizard link)

*Trading:*
/open /close /positions /trades /balance
/autostart /autostop — automation
/analyze [SYMBOL] — live sniper decision (TAKE/SKIP/WAIT)

_Admins:_ /approveauto [user_id] · /revokeauto [user_id]

*Web dashboard:*
https://marketmamba.kkooapp.co.tz

*Signals & pairs:*
/signaltypes — forex, indexes, or crypto (bitcoin)
/signaltypes forex crypto — example
/pairs — per-symbol signals & auto-trade
/pairs EURUSD BTCUSD — example
Active subscribers receive alerts for *your* selections only.

⚠️ Forex trading is high risk.`, userID, planLine)
	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleSubscribe(chatID, userID int64) {
	sub, _ := tb.subs.GetForUser(userID)
	exp := "after trial"
	if sub != nil && sub.ExpiresAt != nil {
		exp = sub.ExpiresAt.Format("2006-01-02")
	}
	price := tb.cfg.Payments.SubscriptionPriceUSDT
	if price <= 0 {
		price = 10
	}
	contact := tb.cfg.App.ContactUsURL
	if contact == "" && tb.api.Self.UserName != "" {
		contact = "https://t.me/" + tb.api.Self.UserName
	}
	contactLine := ""
	if contact != "" {
		contactLine = fmt.Sprintf("\n❓ Pro / teams: [%s](%s)", tb.cfg.App.ContactUsLabel, contact)
	}
	msg := fmt.Sprintf(`*Market Mamba*

%s

🎁 *%d-day free trial* (automatic on /start)
💳 *%.0f USDT / month* — Binance USDT only (no Stripe)

📊 *Dashboard* menu → trades, subscribe, connect broker

Current access until: %s
Your Telegram ID: `+"`%d`"+`
%s%s`,
		tb.cfg.App.ValueProposition,
		tb.cfg.App.FreeTrialDays,
		price,
		exp,
		userID,
		tb.cfg.App.SubscriptionContactMessage,
		contactLine,
	)
	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleMyPlan(chatID, userID int64) {
	sub, err := tb.subs.GetForUser(userID)
	if err != nil || sub == nil {
		tb.sendMessage(chatID, "No active plan found. Use /start to register.")
		return
	}
	exp := "no expiry"
	if sub.ExpiresAt != nil {
		exp = sub.ExpiresAt.Format("2006-01-02 15:04")
	}
	tb.sendMessage(chatID, fmt.Sprintf(`*Your plan*
Plan: %s
Status: %s
Expires: %s
Notes: %s`, sub.Plan, sub.Status, exp, sub.Notes))
}

func (tb *TelegramBot) handleAdmin(chatID, adminID int64, parts []string) {
	if !tb.cfg.IsAdmin(adminID) {
		tb.sendMessage(chatID, "❌ Admin only")
		return
	}
	if len(parts) < 2 {
		tb.sendMessage(chatID, "Admin: /admin stats | /admin activate <id> <days> | /admin signal")
		return
	}
	switch parts[1] {
	case "stats":
		tb.handleAdminStats(chatID)
	case "signal":
		tb.handleAdminSignal(chatID)
	case "activate":
		if len(parts) < 4 {
			tb.sendMessage(chatID, "Usage: /admin activate <telegram_id> <days>")
			return
		}
		targetID, _ := strconv.ParseInt(parts[2], 10, 64)
		days, _ := strconv.Atoi(parts[3])
		sub, err := tb.subs.ActivateManual(targetID, days, "manual", "Activated via Telegram admin", adminID)
		if err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		exp := "never"
		if sub.ExpiresAt != nil {
			exp = sub.ExpiresAt.Format("2006-01-02")
		}
		tb.sendMessage(chatID, fmt.Sprintf("✅ Activated user %d until %s", targetID, exp))
	default:
		tb.sendMessage(chatID, "Unknown admin command")
	}
}

func (tb *TelegramBot) handleAdminStats(chatID int64) {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ internal error")
		return
	}
	stats, err := ps.GetUserStats()
	if err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf(`*Market Mamba stats*
Total users: %d
Active subscriptions: %d
Auto-trading users: %d
New users (7d): %d`,
		stats.TotalUsers, stats.ActiveSubscriptions, stats.AutoTradingUsers, stats.NewUsersLast7Days))
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
	ok, _ := tb.subs.CanTrade(userID)
	subLine := "inactive"
	if ok {
		subLine = "active"
	}
	msg := fmt.Sprintf(`*Bot Status*
Status: %s
Subscription: %s
Auto Trading: %v
Daily Loss Hit: %v
Last Active: %s`,
		status, subLine, botState.AutoTradingActive, botState.DailyLossHit,
		botState.LastActiveAt.Format("2006-01-02 15:04:05"))
	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleBalance(chatID int64, userID int64) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Broker not configured. Use /broker connect or the web dashboard.")
		return
	}
	balance, _ := b.GetBalance()
	equity, _ := b.GetEquity()
	tb.sendMessage(chatID, fmt.Sprintf("*Account Balance*\nBalance: $%.2f\nEquity: $%.2f", balance, equity))
}

func (tb *TelegramBot) handlePositions(chatID int64, userID int64) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Broker not configured")
		return
	}
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ Trade history unavailable")
		return
	}
	userPos, err := positions.ListOpenForUser(ps, userID, b)
	if err != nil || len(userPos) == 0 {
		tb.sendMessage(chatID, "No open positions")
		return
	}
	msg := "*Your open positions*\n\n"
	for i, pos := range userPos {
		msg += fmt.Sprintf("%d. %s %s\n   Entry: %.5f | P/L: %.2f\n", i+1, pos.Symbol, pos.Type, pos.EntryPrice, pos.Profit)
	}
	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleOpen(chatID int64, userID int64, args []string) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	if len(args) < 5 {
		tb.sendMessage(chatID, "❌ Usage: /open <symbol> <BUY|SELL> <quantity> <stopLoss> <takeProfit>")
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Broker not configured")
		return
	}
	symbol := strings.ToUpper(args[0])
	orderType := strings.ToUpper(args[1])
	qty, _ := strconv.ParseFloat(args[2], 64)
	sl, _ := strconv.ParseFloat(args[3], 64)
	tp, _ := strconv.ParseFloat(args[4], 64)
	if tb.tier != nil {
		if err := tb.tier.CanExecuteTrade(userID, orderType); err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
	}
	signal := &models.TradeSignal{Symbol: symbol, Type: orderType, StopLoss: sl, TakeProfit: tp, Strength: 1.0, TriggeredAt: time.Now()}
	botState, _ := tb.storage.GetBotState(userID)
	if err := tb.validator.ValidateTradeSignal(signal, 10000, 0, 0, 0, botState.IsPaused); err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ Validation failed: %v", err))
		return
	}
	pos, err := b.OpenMarketOrder(symbol, orderType, qty, sl, tp)
	if err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ Failed: %v", err))
		return
	}
	pos.UserID = userID
	if tb.tier != nil {
		_ = tb.tier.RecordTrade(userID, orderType)
	}
	if err := tb.logTradeOpen(userID, pos, "MANUAL"); err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("✅ Opened %s %s — ID %s\n⚠️ Log failed: %v", symbol, orderType, pos.ID, err))
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf("✅ Trade opened & logged: %s %s — ID %s", symbol, orderType, pos.ID))
}

func (tb *TelegramBot) handleClose(chatID int64, userID int64, args []string) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	if len(args) < 1 {
		tb.sendMessage(chatID, "❌ Usage: /close <positionID>")
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Broker not configured")
		return
	}
	posID := args[0]
	exitPrice := 0.0
	if pos, err := b.GetPositionByID(posID); err == nil && pos != nil {
		exitPrice = pos.CurrentPrice
		if exitPrice <= 0 {
			exitPrice = pos.EntryPrice
		}
	}
	if err := b.ClosePosition(posID); err != nil {
		tb.sendMessage(chatID, fmt.Sprintf("❌ %v", err))
		return
	}
	if _, err := tb.logTradeClose(userID, posID, exitPrice, "MANUAL"); err != nil {
		tb.sendMessage(chatID, "✅ Closed on broker\n⚠️ Log failed: "+err.Error())
		return
	}
	tb.sendMessage(chatID, "✅ Position closed & logged")
}

func (tb *TelegramBot) handleCloseAll(chatID int64, userID int64) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Broker not configured")
		return
	}
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ Trade history unavailable")
		return
	}
	userPos, err := positions.ListOpenForUser(ps, userID, b)
	if err != nil || len(userPos) == 0 {
		tb.sendMessage(chatID, "No open positions")
		return
	}
	for _, pos := range userPos {
		if err := b.ClosePosition(pos.ID); err != nil {
			logger.Error("close position %s user %d: %v", pos.ID, userID, err)
			continue
		}
		exit := pos.CurrentPrice
		if exit <= 0 {
			exit = pos.EntryPrice
		}
		if _, err := tb.logTradeClose(userID, pos.ID, exit, "MANUAL"); err != nil {
			logger.Error("logTradeClose %s: %v", pos.ID, err)
		}
	}
	tb.sendMessage(chatID, "✅ Your positions closed & logged")
}

func (tb *TelegramBot) handlePause(chatID int64, userID int64) {
	_ = tb.storage.UpdateBotState(userID, true, false, false)
	tb.sendMessage(chatID, "⏸️ Trading paused")
}

func (tb *TelegramBot) handleResume(chatID int64, userID int64) {
	_ = tb.storage.UpdateBotState(userID, false, false, false)
	tb.sendMessage(chatID, "▶️ Trading resumed")
}

func (tb *TelegramBot) handleRisk(chatID int64, userID int64, _ []string) {
	settings, err := tb.storage.GetRiskSettings(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching risk settings")
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf(`*Risk Settings*
Max risk/trade: %.2f%%
Max daily loss: %.2f%%
Max open trades: %d`,
		settings.MaxRiskPerTrade*100, settings.MaxDailyLoss*100, settings.MaxOpenTrades))
}

func (tb *TelegramBot) handleDailyReport(chatID int64, userID int64) {
	stats, err := tb.storage.GetDailyStats(userID, time.Now())
	if err != nil {
		tb.sendMessage(chatID, "❌ Error fetching daily stats")
		return
	}
	msg := fmt.Sprintf(`*Daily Report*
Trades: %d (W:%d L:%d)
Net P/L: $%.2f
Profit: $%.2f | Loss: $%.2f`,
		stats.TradeCount, stats.WinCount, stats.LossCount,
		stats.NetProfit, stats.TotalProfit, stats.TotalLoss)
	if stats.TradeCount == 0 {
		msg += "\n\n_No trades recorded today yet._"
	}
	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleAutoStart(chatID int64, userID int64) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	botState, _ := tb.storage.GetBotState(userID)
	if botState.AutoTradingActive {
		tb.sendMessage(chatID, "⚠️ Already active")
		return
	}
	_ = tb.storage.UpdateBotState(userID, false, true, false)
	if !tb.cfg.App.AutoTradeRequiresApproval || tb.cfg.IsAdmin(userID) {
		_ = tb.storage.SetAutoTradeApproved(userID, true)
	}
	msg := "🤖 Automated trading enabled."
	if tb.cfg.App.AutoTradeRequiresApproval && !tb.cfg.IsAdmin(userID) {
		state, _ := tb.storage.GetBotState(userID)
		if state != nil && !state.AutoTradeApproved {
			msg += "\n\n⏳ *Pending admin approval* for assisted auto-trade. Signals and /analyze still work."
		}
	}
	msg += "\n\nConnect broker: /broker connect (mock demo) or OANDA practice on the dashboard."
	tb.sendMessage(chatID, msg)
}

func (tb *TelegramBot) handleAutoStop(chatID int64, userID int64) {
	_ = tb.storage.UpdateBotState(userID, false, false, false)
	tb.sendMessage(chatID, "⏹️ Automated trading stopped")
}

func (tb *TelegramBot) handleAutoStatus(chatID int64, userID int64) {
	botState, _ := tb.storage.GetBotState(userID)
	auto := "❌ Off"
	if botState.AutoTradingActive {
		auto = "✅ On"
	}
	approved := "—"
	if tb.cfg.App.AutoTradeRequiresApproval {
		if botState.AutoTradeApproved {
			approved = "✅ Approved"
		} else {
			approved = "⏳ Pending"
		}
	}
	tb.sendMessage(chatID, fmt.Sprintf("*Automation:* %s\n*Auto-trade approval:* %s", auto, approved))
}

func (tb *TelegramBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := tb.api.Send(msg); err != nil {
		logger.Error("Failed to send message: %v", err)
	}
}
