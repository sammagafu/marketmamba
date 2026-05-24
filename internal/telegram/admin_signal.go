package telegram

import (
	"fmt"
	"strings"

	"forex-bot/internal/models"
	"forex-bot/internal/signals"
	"forex-bot/internal/storage"
)

func (tb *TelegramBot) handleAdminSignal(chatID int64) {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ internal error")
		return
	}
	symbols := tb.cfg.SignalSymbols()
	var sig *models.TradeSignal
	var lastErr error
	for _, symbol := range symbols {
		s, err := signals.GenerateQualified(symbol, tb.cfg.App.SignalMinStrength, 0, tb.validator)
		if err != nil {
			lastErr = err
			continue
		}
		sig = s
		break
	}
	if sig == nil {
		msg := "⏭️ No broadcast — no symbol met requirements"
		if lastErr != nil {
			msg += ":\n" + lastErr.Error()
		}
		msg += "\nChecked: " + strings.Join(symbols, ", ")
		tb.sendMessage(chatID, msg)
		return
	}
	n, err := signals.PublishManual(ps, tb.subs, tb.tier, tb, tb.validator, tb.cfg.App.SignalMinStrength, sig, false)
	if err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	if n == 0 {
		tb.sendMessage(chatID, "✅ Signal qualified but no eligible subscribers")
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf("✅ Signal sent to *%d* subscribers\n%s %s (strength %.0f%%)",
		n, sig.Symbol, sig.Type, sig.Strength*100))
}
