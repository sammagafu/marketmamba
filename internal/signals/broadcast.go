package signals

import (
	"context"
	"fmt"
	"time"

	"forex-bot/internal/decision"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
)

// Notifier delivers trade alerts to one Telegram user (chat_id = telegram_id).
type Notifier interface {
	NotifySignal(telegramID int64, signal *models.TradeSignal) error
	NotifyDecision(telegramID int64, d *decision.Decision) error
}

// FormatMessage builds a Telegram-friendly signal alert.
func FormatMessage(signal *models.TradeSignal) string {
	if signal == nil {
		return ""
	}
	reasonBlock := ""
	if signal.Reason != "" {
		reasonBlock = fmt.Sprintf("Setup: %s\n", signal.Reason)
	}
	return fmt.Sprintf(
		"📡 *Market Mamba signal*\n\n"+
			"*%s %s*\n"+
			"%s"+
			"Strength: %.0f%%\n"+
			"Stop loss: %.5f\n"+
			"Take profit: %.5f\n"+
			"R:R %.2f\n\n"+
			"_Not financial advice. Use /autostart for auto-execution._",
		signal.Symbol,
		signal.Type,
		reasonBlock,
		signal.Strength*100,
		signal.StopLoss,
		signal.TakeProfit,
		signal.RiskRewardRatio,
	)
}

// Broadcast sends a signal to all eligible subscribers (signal must already qualify).
func Broadcast(store *storage.PostgresStorage, subs *subscription.Service, notifier Notifier, signal *models.TradeSignal) (int, error) {
	if signal == nil || notifier == nil {
		return 0, nil
	}
	ids, err := store.ListSignalSubscriberTelegramIDsForSymbol(signal.Symbol)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, id := range ids {
		ok, _ := subs.CanTrade(id)
		if !ok {
			continue
		}
		if err := notifier.NotifySignal(id, signal); err != nil {
			logger.Error("Signal notify user %d: %v", id, err)
			continue
		}
		sent++
	}
	return sent, nil
}

// BroadcastDecision sends a sniper decision to eligible subscribers (TAKE only).
func BroadcastDecision(store *storage.PostgresStorage, subs *subscription.Service, notifier Notifier, d *decision.Decision) (int, error) {
	if d == nil || d.Action != decision.ActionTake || d.Signal == nil || notifier == nil {
		return 0, nil
	}
	ids, err := store.ListSignalSubscriberTelegramIDsForSymbol(d.Symbol)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, id := range ids {
		ok, _ := subs.CanTrade(id)
		if !ok {
			continue
		}
		if err := notifier.NotifyDecision(id, d); err != nil {
			logger.Error("Sniper notify user %d: %v", id, err)
			continue
		}
		sent++
	}
	return sent, nil
}

// Publisher periodically evaluates live markets and broadcasts sniper TAKE alerts.
type Publisher struct {
	store       *storage.PostgresStorage
	subs        *subscription.Service
	notifier    Notifier
	validator   *risk.RiskValidator
	engine      *decision.Engine
	symbols     []string
	interval    time.Duration
	minStrength float64
	useDecision bool
}

func NewPublisher(
	store *storage.PostgresStorage,
	subs *subscription.Service,
	notifier Notifier,
	validator *risk.RiskValidator,
	engine *decision.Engine,
	symbols []string,
	interval time.Duration,
	minStrength float64,
	useDecision bool,
) *Publisher {
	if len(symbols) == 0 {
		symbols = []string{"EURUSD", "BTCUSD"}
	}
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	if minStrength <= 0 {
		minStrength = 0.7
	}
	return &Publisher{
		store: store, subs: subs, notifier: notifier, validator: validator,
		engine: engine, symbols: symbols, interval: interval, minStrength: minStrength,
		useDecision: useDecision,
	}
}

func (p *Publisher) Start(ctx context.Context) {
	if p.notifier == nil || p.validator == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		mode := "legacy-mock"
		if p.useDecision && p.engine != nil {
			mode = "live-sniper"
		}
		logger.Info("Signal broadcast publisher started (%s, interval %v, symbols %v)", mode, p.interval, p.symbols)
		for {
			select {
			case <-ctx.Done():
				logger.Info("Signal broadcast publisher stopped")
				return
			case <-ticker.C:
				p.publishOnce()
			}
		}
	}()
}

func (p *Publisher) publishOnce() {
	ctx := context.Background()
	for _, symbol := range p.symbols {
		if p.useDecision && p.engine != nil {
			d, err := p.engine.Evaluate(ctx, symbol)
			if err != nil {
				logger.Error("Sniper evaluate %s: %v", symbol, err)
				continue
			}
			if d.Action != decision.ActionTake {
				logger.Debug("Sniper broadcast skip %s: %s — %s", symbol, d.Action, d.Reason)
				continue
			}
			n, err := BroadcastDecision(p.store, p.subs, p.notifier, d)
			if err != nil {
				logger.Error("Sniper broadcast %s: %v", symbol, err)
				continue
			}
			if n > 0 {
				p.engine.MarkTaken(symbol)
				logger.Info("Sniper broadcast %s %s (%.0f%%) → %d subscribers | %s",
					d.Symbol, d.Signal.Type, d.Confidence*100, n, d.Reason)
			}
			continue
		}
		signal, err := GenerateQualified(symbol, p.minStrength, 0, p.validator)
		if err != nil {
			logger.Debug("Signal broadcast skipped %s: %v", symbol, err)
			continue
		}
		n, err := Broadcast(p.store, p.subs, p.notifier, signal)
		if err != nil {
			logger.Error("Signal broadcast %s: %v", symbol, err)
			continue
		}
		if n > 0 {
			logger.Info("Signal broadcast %s %s (strength %.2f, R:R %.2f) → %d subscribers",
				signal.Symbol, signal.Type, signal.Strength, signal.RiskRewardRatio, n)
		}
	}
}

// PublishManual broadcasts only if the signal meets requirements (unless force is true).
func PublishManual(
	store *storage.PostgresStorage,
	subs *subscription.Service,
	notifier Notifier,
	validator *risk.RiskValidator,
	minStrength float64,
	signal *models.TradeSignal,
	force bool,
) (int, error) {
	if signal == nil {
		return 0, fmt.Errorf("no signal")
	}
	if !force {
		if err := MeetsRequirements(signal, validator, minStrength); err != nil {
			return 0, err
		}
	}
	return Broadcast(store, subs, notifier, signal)
}
