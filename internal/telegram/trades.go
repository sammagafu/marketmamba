package telegram

import (
	"fmt"
	"strings"

	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/trading"
)

func (tb *TelegramBot) tradeLog() *trading.TradeLog {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		return nil
	}
	return trading.NewTradeLog(ps)
}

func (tb *TelegramBot) handleTrades(chatID, userID int64) {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ Trade history unavailable")
		return
	}
	trades, err := ps.ListTradesByUser(userID, 15)
	if err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	if len(trades) == 0 {
		tb.sendMessage(chatID, "No trades logged yet. Open one with /open or enable /autostart.")
		return
	}
	var b strings.Builder
	b.WriteString("*Recent trades*\n\n")
	for i, t := range trades {
		line := fmt.Sprintf("%d. %s %s %.2f @ %.5f — %s", i+1, t.Symbol, t.Type, t.Quantity, t.EntryPrice, t.Status)
		if t.Status == "CLOSED" && t.Profit != nil {
			line += fmt.Sprintf(" P/L $%.2f", *t.Profit)
			if t.ClosureReason != nil {
				line += " (" + *t.ClosureReason + ")"
			}
		}
		b.WriteString(line + "\n")
	}
	tb.sendMessage(chatID, b.String())
}

func (tb *TelegramBot) logTradeOpen(userID int64, pos *models.Position, source string) error {
	tl := tb.tradeLog()
	if tl == nil || pos == nil {
		return nil
	}
	_, err := tl.RecordOpen(userID, pos, source)
	return err
}

func (tb *TelegramBot) logTradeClose(userID int64, brokerPosID string, exitPrice float64, reason string) error {
	tl := tb.tradeLog()
	if tl == nil {
		return nil
	}
	_, err := tl.RecordClose(userID, brokerPosID, exitPrice, reason)
	return err
}
