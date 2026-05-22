package telegram

import (
	"forex-bot/internal/models"
	"forex-bot/internal/signals"
)

// NotifySignal implements signals.Notifier for Telegram subscribers.
func (tb *TelegramBot) NotifySignal(telegramID int64, signal *models.TradeSignal) error {
	tb.sendMessage(telegramID, signals.FormatMessage(signal))
	return nil
}
