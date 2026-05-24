package filter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"forex-bot/internal/marketdata"
	"forex-bot/internal/risk"
	"forex-bot/internal/signalgen"
)

// Service produces filter audit reports for API and UI.
type Service struct {
	market      *marketdata.Service
	validator   *risk.RiskValidator
	minStrength float64
	minRR       float64
	minBars     int
}

func NewService(market *marketdata.Service, validator *risk.RiskValidator, minStrength, minRR float64) *Service {
	if minStrength <= 0 {
		minStrength = 0.7
	}
	if minRR <= 0 {
		minRR = 1.0
	}
	return &Service{
		market:      market,
		validator:   validator,
		minStrength: minStrength,
		minRR:       minRR,
		minBars:     marketdata.MinBarsForDecision,
	}
}

// Report builds a full filter stack for a symbol (live data when available).
func (s *Service) Report(ctx context.Context, symbol string) (*Report, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if sym == "" {
		return nil, fmt.Errorf("symbol required")
	}

	var in Input
	if s.market != nil {
		snap, err := s.market.Refresh(ctx, sym)
		if err == nil && snap != nil {
			in = InputFromSnapshot(snap, s.minStrength, s.minRR)
		}
	}
	if in.Price <= 0 {
		in = s.simulatedInput(sym)
	}

	report, sig := RunTechnical(in, s.minBars)
	report.GeneratedAt = time.Now().UTC()
	AppendRisk(report, sig, s.validator, s.minStrength)
	return report, nil
}

func (s *Service) simulatedInput(symbol string) Input {
	// Representative demo snapshot (same family as signalgen.SimulateScalpingOpportunity).
	price, atr, ema20, ema50, ema200, rsi, bid, ask, ok := demoSnapshot(symbol, "UPTREND_SCALP")
	if !ok {
		return Input{Symbol: symbol, Source: "simulated"}
	}
	return Input{
		Symbol: symbol, Price: price, ATR: atr,
		EMA20: ema20, EMA50: ema50, EMA200: ema200, RSI: rsi,
		Bid: bid, Ask: ask,
		MinStrength: s.minStrength, MinRR: s.minRR,
		Source: "simulated",
	}
}

func demoSnapshot(symbol, opportunityType string) (price, atr, ema20, ema50, ema200, rsi, bid, ask float64, ok bool) {
	sig := signalgen.SimulateScalpingOpportunity(symbol, opportunityType)
	if sig == nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, false
	}
	// Reverse-engineer approximate inputs from generator defaults for display only.
	if strings.Contains(strings.ToUpper(symbol), "BTC") {
		return 65000, 180, 64920, 64600, 63800, 55, 64999, 65001, true
	}
	return 1.1050, 0.0035, 1.1045, 1.1030, 1.1000, 55, 1.10495, 1.10505, true
}
