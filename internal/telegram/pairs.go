package telegram

import (
	"fmt"
	"strings"

	"forex-bot/internal/pairs"
	"forex-bot/internal/storage"
)

func (tb *TelegramBot) pairService() *pairs.Service {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		return nil
	}
	return pairs.NewService(ps, tb.cfg)
}

func (tb *TelegramBot) handlePairs(chatID, userID int64, args []string) {
	svc := tb.pairService()
	if svc == nil {
		tb.sendMessage(chatID, "❌ Pair preferences unavailable")
		return
	}
	if len(args) == 0 {
		resp, err := svc.GetResponse(userID)
		if err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		var b strings.Builder
		b.WriteString("*Your trading pairs*\n\n")
		b.WriteString(fmt.Sprintf("Available: %s\n\n", strings.Join(resp.AvailableSymbols, ", ")))
		for _, p := range resp.Pairs {
			sig := "—"
			if p.ReceiveSignals {
				sig = "📡"
			}
			auto := "—"
			if p.AutoTrade {
				auto = "🤖"
			}
			b.WriteString(fmt.Sprintf("• *%s* signals %s · auto %s\n", p.Symbol, sig, auto))
		}
		b.WriteString("\nSet pairs:\n`/pairs EURUSD BTCUSD`\n")
		b.WriteString("_📡 = Telegram signals · 🤖 = auto-trade with /autostart_")
		tb.sendMessage(chatID, b.String())
		return
	}

	sub := args[0]
	if sub == "all" {
		if err := svc.SeedDefaults(userID); err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		tb.sendMessage(chatID, "✅ All platform pairs enabled for signals and auto-trade")
		return
	}

	if err := svc.SetSymbolsQuick(userID, args); err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	resp, _ := svc.GetResponse(userID)
	tb.sendMessage(chatID, fmt.Sprintf(
		"✅ Pairs updated\n\nSignals: %s\nAuto-trade: %s\n\nUse /autostart to run automation on your auto pairs.",
		strings.Join(resp.SignalSymbols, ", "),
		strings.Join(resp.AutoTradeSymbols, ", "),
	))
}
