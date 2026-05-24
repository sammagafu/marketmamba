package filter

import (
	"strings"

	"forex-bot/internal/marketdata"
)

// Input is normalized market context for the technical filter chain.
type Input struct {
	Symbol      string
	Price       float64
	ATR         float64
	EMA20       float64
	EMA50       float64
	EMA200      float64
	RSI         float64
	Bid         float64
	Ask         float64
	MinStrength float64
	MinRR       float64
	BarCount    int
	Source      string
}

func MaxSpread(symbol string, price float64) float64 {
	s := strings.ToUpper(symbol)
	if strings.Contains(s, "BTC") {
		return price * 0.00025
	}
	if strings.Contains(s, "XAU") || strings.Contains(s, "GOLD") {
		return price * 0.0002
	}
	return price * 0.0003
}

func SpreadUnits(symbol string, spread float64) float64 {
	if strings.Contains(strings.ToUpper(symbol), "BTC") {
		return spread
	}
	return spread / 0.0001
}

func TrendLabel(price, ema20, ema50, ema200 float64) string {
	if ema20 > ema50 && ema50 > ema200 {
		return "STRONG_UPTREND"
	}
	if ema20 < ema50 && ema50 < ema200 {
		return "STRONG_DOWNTREND"
	}
	if ema20 > ema50 {
		return "UPTREND"
	}
	if ema20 < ema50 {
		return "DOWNTREND"
	}
	return "SIDEWAY"
}

func InputFromSnapshot(snap *marketdata.Snapshot, minStrength, minRR float64) Input {
	if snap == nil {
		return Input{}
	}
	return Input{
		Symbol:      snap.Symbol,
		Price:       snap.Mid,
		ATR:         snap.ATR,
		EMA20:       snap.EMA20,
		EMA50:       snap.EMA50,
		EMA200:      snap.EMA200,
		RSI:         snap.RSI,
		Bid:         snap.Bid,
		Ask:         snap.Ask,
		MinStrength: minStrength,
		MinRR:       minRR,
		BarCount:    snap.BarCount,
		Source:      snap.Source,
	}
}

// Catalog returns operator-facing documentation for all gates.
func Catalog() []CatalogEntry {
	return []CatalogEntry{
		{ID: "live_history", Name: "Live bar history", Category: CategoryMarket, Description: "Enough ticks to compute EMA/ATR/RSI reliably.", Threshold: "≥35 bars"},
		{ID: "spread", Name: "Spread gate", Category: CategoryMarket, Description: "Rejects when bid/ask width eats the edge.", Threshold: "~3 pips majors, wider on BTC/XAU"},
		{ID: "volatility", Name: "ATR floor", Category: CategoryTechnical, Description: "Skips dead markets where stops cannot breathe.", Threshold: "ATR ≥ 0.05% of price"},
		{ID: "rsi_band", Name: "RSI band", Category: CategoryTechnical, Description: "Avoids exhausted momentum zones for scalps.", Threshold: "20 < RSI < 80"},
		{ID: "ema_trend", Name: "EMA stack", Category: CategoryTechnical, Description: "Classifies trend from EMA 20/50/200 alignment.", Threshold: "Informational + setup routing"},
		{ID: "setup", Name: "Setup pattern", Category: CategorySetup, Description: "Price action near EMA with trend-aligned side.", Threshold: "UPTREND / DOWNTREND / STRONG variants"},
		{ID: "strength", Name: "Signal strength", Category: CategorySetup, Description: "Model confidence after pattern match.", Threshold: "≥ SIGNAL_MIN_STRENGTH (default 0.7)"},
		{ID: "risk_reward", Name: "Risk–reward", Category: CategoryRisk, Description: "Take-profit distance vs stop distance.", Threshold: "≥ configured R:R (default 1:1)"},
		{ID: "risk_rules", Name: "Structural risk", Category: CategoryRisk, Description: "SL/TP validity and platform risk envelope.", Threshold: "Same checks as auto-trade"},
		{ID: "min_strength", Name: "Broadcast floor", Category: CategoryPlatform, Description: "Minimum strength for signals and sniper TAKE.", Threshold: "env SIGNAL_MIN_STRENGTH"},
	}
}
