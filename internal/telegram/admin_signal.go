package telegram

import (
	"fmt"

	"forex-bot/internal/signals"
	"forex-bot/internal/storage"
)

func (tb *TelegramBot) handleAdminSignal(chatID int64) {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ internal error")
		return
	}
	symbol := tb.cfg.App.SignalBroadcastSymbol
	if symbol == "" {
		symbol = "EURUSD"
	}
	sig, err := signals.GenerateQualified(symbol, tb.cfg.App.SignalMinStrength, 0, tb.validator)
	if err != nil {
		tb.sendMessage(chatID, "⏭️ No broadcast — signal did not meet requirements:\n"+err.Error())
		return
	}
	n, err := signals.PublishManual(ps, tb.subs, tb, tb.validator, tb.cfg.App.SignalMinStrength, sig, false)
	if err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	if n == 0 {
		tb.sendMessage(chatID, "✅ Signal qualified but no eligible subscribers online")
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf("✅ Signal sent to *%d* subscribers\n%s %s (strength %.0f%%)",
		n, sig.Symbol, sig.Type, sig.Strength*100))
}
