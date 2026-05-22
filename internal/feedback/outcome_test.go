package feedback

import (
	"strings"
	"testing"

	"forex-bot/internal/models"
)

func TestFormatOutcomeMessage_TP(t *testing.T) {
	p := 42.5
	ep := 1.09
	trade := &models.Trade{
		Symbol: "EURUSD", Type: "BUY", EntryPrice: 1.085,
		Quantity: 0.1, StopLoss: 1.08, TakeProfit: 1.09,
		Profit: &p, ExitPrice: &ep,
	}
	msg := FormatOutcomeMessage(trade, "TP")
	if !strings.Contains(msg, "Take profit hit") || !strings.Contains(msg, "+$42.50") {
		t.Fatal(msg)
	}
}

func TestFormatOutcomeMessage_SL(t *testing.T) {
	p := -20.0
	ep := 1.08
	trade := &models.Trade{
		Symbol: "GBPUSD", Type: "SELL", EntryPrice: 1.27,
		Quantity: 0.2, StopLoss: 1.275, TakeProfit: 1.26,
		Profit: &p, ExitPrice: &ep,
	}
	msg := FormatOutcomeMessage(trade, "SL")
	if !strings.Contains(msg, "Stop loss hit") || !strings.Contains(msg, "-$20.00") {
		t.Fatal(msg)
	}
}
