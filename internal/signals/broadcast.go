package signals

import (
	"context"
	"fmt"
	"time"

	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
)

// Notifier delivers a trade signal alert to one Telegram user (chat_id = telegram_id).
type Notifier interface {
	NotifySignal(telegramID int64, signal *models.TradeSignal) error
}

// FormatMessage builds a Telegram-friendly signal alert.
func FormatMessage(signal *models.TradeSignal) string {
	if signal == nil {
		return ""
	}
	return fmt.Sprintf(
		"📡 *Market Mamba signal*\n\n"+
			"*%s %s*\n"+
			"Strength: %.0f%%\n"+
			"Stop loss: %.5f\n"+
			"Take profit: %.5f\n"+
			"R:R %.2f\n\n"+
			"_Not financial advice. Use /autostart for auto-execution._",
		signal.Symbol,
		signal.Type,
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

// Publisher periodically generates and broadcasts signals that pass technical + risk filters.
type Publisher struct {
	store       *storage.PostgresStorage
	subs        *subscription.Service
	notifier    Notifier
	validator   *risk.RiskValidator
	symbols     []string
	interval    time.Duration
	minStrength float64
}

func NewPublisher(
	store *storage.PostgresStorage,
	subs *subscription.Service,
	notifier Notifier,
	validator *risk.RiskValidator,
	symbols []string,
	interval time.Duration,
	minStrength float64,
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
		symbols: symbols, interval: interval, minStrength: minStrength,
	}
}

func (p *Publisher) Start(ctx context.Context) {
	if p.notifier == nil || p.validator == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()
		logger.Info("Signal broadcast publisher started (interval %v, symbols %v, min strength %.2f)", p.interval, p.symbols, p.minStrength)
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
	for _, symbol := range p.symbols {
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
