package signals

import (
	"testing"
	"time"

	"forex-bot/internal/models"
	"forex-bot/internal/risk"
)

func testValidator() *risk.RiskValidator {
	return risk.NewRiskValidator(&models.RiskSettings{
		MaxRiskPerTrade: 0.005,
		MaxDailyLoss:    0.02,
		MaxOpenTrades:   2,
		MaxTradesPerDay: 10,
		RiskRewardRatio: 1.0,
	})
}

func TestMeetsRequirements_rejectsWeakSignal(t *testing.T) {
	sig := &models.TradeSignal{
		Symbol: "EURUSD", Type: "BUY",
		StopLoss: 1.08, TakeProfit: 1.10, Strength: 0.3,
		RiskRewardRatio: 2.0, TriggeredAt: time.Now(),
	}
	if err := MeetsRequirements(sig, testValidator(), 0.7); err == nil {
		t.Fatal("expected strength rejection")
	}
}

func TestMeetsRequirements_acceptsStrongSignal(t *testing.T) {
	sig := &models.TradeSignal{
		Symbol: "EURUSD", Type: "BUY",
		StopLoss: 1.085, TakeProfit: 1.095, Strength: 0.85,
		RiskRewardRatio: 2.0, TriggeredAt: time.Now(),
	}
	if err := MeetsRequirements(sig, testValidator(), 0.7); err != nil {
		t.Fatalf("expected pass: %v", err)
	}
}
