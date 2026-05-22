package feedback

import (
	"fmt"
	"strings"

	"forex-bot/internal/models"
)

// OutcomeNotifier sends trade result alerts when TP, SL, or manual close fires.
type OutcomeNotifier interface {
	NotifyTradeOutcome(telegramID int64, trade *models.Trade, reason string) error
}

// FormatOutcomeMessage builds a Telegram-friendly close notification.
func FormatOutcomeMessage(trade *models.Trade, reason string) string {
	if trade == nil {
		return ""
	}
	reason = strings.ToUpper(strings.TrimSpace(reason))

	var headline, emoji string
	switch reason {
	case "TP":
		headline = "Take profit hit"
		emoji = "🎯"
	case "SL":
		headline = "Stop loss hit"
		emoji = "🛑"
	case "MANUAL":
		headline = "Trade closed manually"
		emoji = "📤"
	default:
		headline = "Trade closed"
		emoji = "📋"
		if reason != "" {
			headline = fmt.Sprintf("Trade closed (%s)", reason)
		}
	}

	exitStr := "—"
	if trade.ExitPrice != nil {
		exitStr = fmt.Sprintf("%.5f", *trade.ExitPrice)
	}

	plLine := ""
	if trade.Profit != nil {
		p := *trade.Profit
		if p > 0 {
			plLine = fmt.Sprintf("P/L: *+$%.2f* ✅\n", p)
		} else if p < 0 {
			plLine = fmt.Sprintf("P/L: *-$%.2f*\n", -p)
		} else {
			plLine = "P/L: *$0.00*\n"
		}
	}

	return fmt.Sprintf(
		"%s *%s*\n\n"+
			"*%s %s*\n"+
			"Entry: %.5f → Exit: %s\n"+
			"Qty: %.2f\n"+
			"%s"+
			"SL: %.5f · TP: %.5f\n\n"+
			"_Signal outcome logged in your trade book._",
		emoji,
		headline,
		trade.Symbol,
		trade.Type,
		trade.EntryPrice,
		exitStr,
		trade.Quantity,
		plLine,
		trade.StopLoss,
		trade.TakeProfit,
	)
}
