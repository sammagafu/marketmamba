package telegram

import (
	"context"
	"strings"

	"forex-bot/internal/decision"
)

func (tb *TelegramBot) handleAnalyze(chatID, userID int64, args []string) {
	if !tb.requireTrading(chatID, userID) {
		return
	}
	if tb.decisionEngine == nil {
		tb.sendMessage(chatID, "❌ Live sniper analysis is not enabled on this server.")
		return
	}
	symbol := "EURUSD"
	if len(args) > 0 {
		symbol = strings.ToUpper(strings.TrimSpace(args[0]))
	}
	d, err := tb.decisionEngine.Evaluate(context.Background(), symbol)
	if err != nil {
		tb.sendMessage(chatID, "❌ Analysis failed: "+err.Error())
		return
	}
	tb.sendMessage(chatID, decision.FormatTelegram(d))
}
