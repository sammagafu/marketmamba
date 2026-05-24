package filter

import (
	"testing"

	"forex-bot/internal/models"
	"forex-bot/internal/risk"
)

func TestRunTechnical_passesCleanFX(t *testing.T) {
	in := Input{
		Symbol: "EURUSD", Price: 1.1050, ATR: 0.0035,
		EMA20: 1.1045, EMA50: 1.1030, EMA200: 1.1000, RSI: 55,
		Bid: 1.10495, Ask: 1.10505, MinStrength: 0.7, MinRR: 1.0, Source: "test",
	}
	report, sig := RunTechnical(in, 35)
	if report == nil {
		t.Fatal("nil report")
	}
	if sig == nil {
		t.Fatal("expected signal")
	}
	if report.Verdict != StatusPass || !report.Qualified {
		t.Fatalf("verdict=%s qualified=%v", report.Verdict, report.Qualified)
	}
}

func TestRunTechnical_spreadFail(t *testing.T) {
	in := Input{
		Symbol: "EURUSD", Price: 1.1050, ATR: 0.0035,
		EMA20: 1.1045, EMA50: 1.1030, EMA200: 1.1000, RSI: 55,
		Bid: 1.1000, Ask: 1.1200, MinStrength: 0.7, MinRR: 1.0,
	}
	report, sig := RunTechnical(in, 35)
	if sig != nil {
		t.Fatal("expected no signal")
	}
	if report.Verdict != StatusFail {
		t.Fatalf("verdict=%s", report.Verdict)
	}
}

func TestAppendRisk(t *testing.T) {
	v := risk.NewRiskValidator(&models.RiskSettings{
		MaxRiskPerTrade: 0.005, MaxDailyLoss: 0.02, MaxOpenTrades: 2,
		MaxTradesPerDay: 10, RiskRewardRatio: 1.0,
	})
	in := Input{
		Symbol: "EURUSD", Price: 1.1050, ATR: 0.0035,
		EMA20: 1.1045, EMA50: 1.1030, EMA200: 1.1000, RSI: 55,
		Bid: 1.10495, Ask: 1.10505, MinStrength: 0.7, MinRR: 1.0,
	}
	report, sig := RunTechnical(in, 35)
	AppendRisk(report, sig, v, 0.7)
	if len(report.Layers) < 3 {
		t.Fatalf("layers=%d", len(report.Layers))
	}
}

func TestCatalog(t *testing.T) {
	if len(Catalog()) < 8 {
		t.Fatal("expected catalog entries")
	}
}
