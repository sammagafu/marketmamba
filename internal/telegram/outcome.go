package telegram

import (
	"forex-bot/internal/feedback"
	"forex-bot/internal/models"
)

// NotifyTradeOutcome implements feedback.OutcomeNotifier — sends one Telegram message.
func (tb *TelegramBot) NotifyTradeOutcome(telegramID int64, trade *models.Trade, reason string) error {
	msg := feedback.FormatOutcomeMessage(trade, reason)
	if msg == "" {
		return nil
	}
	tb.sendMessage(telegramID, msg)
	return nil
}

// NotifyCommunityOutcome sends the shared signal-result template (subscribers).
func (tb *TelegramBot) NotifyCommunityOutcome(telegramID int64, trade *models.Trade, reason string) error {
	msg := feedback.FormatCommunityOutcomeMessage(trade, reason)
	if msg == "" {
		return nil
	}
	tb.sendMessage(telegramID, msg)
	return nil
}
