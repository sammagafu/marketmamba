package telegram

import (
	"forex-bot/internal/decision"
	"forex-bot/internal/models"
	"forex-bot/internal/signals"
)

// NotifySignal implements signals.Notifier for Telegram subscribers.
func (tb *TelegramBot) NotifySignal(telegramID int64, signal *models.TradeSignal) error {
	footer := ""
	if tb.cfg != nil && !tb.cfg.IsFullAssetCatalog() {
		footer = "Launch phase · BTC & ETH"
	}
	tb.sendMessage(telegramID, signals.FormatMessageWithFooter(signal, footer))
	return nil
}

// NotifyDecision implements signals.Notifier for sniper advisory broadcasts.
func (tb *TelegramBot) NotifyDecision(telegramID int64, d *decision.Decision) error {
	tb.sendMessage(telegramID, decision.FormatTelegram(d))
	return nil
}

// NotifySniper sends a real-time decision to one user (autostart / /analyze).
func (tb *TelegramBot) NotifySniper(telegramID int64, d *decision.Decision) error {
	return tb.NotifyDecision(telegramID, d)
}
