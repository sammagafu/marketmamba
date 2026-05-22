package decision

import (
	"context"
	"fmt"
	"strings"
	"time"

	"forex-bot/internal/marketdata"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/signalgen"
)

// Engine provides real-time TAKE / SKIP / WAIT decision support from live prices.
type Engine struct {
	market        *marketdata.Service
	validator     *risk.RiskValidator
	cooldown      *CooldownTracker
	minStrength   float64
	sniperMinConf float64
	rrRatio       float64
	minBars       int
}

func NewEngine(
	market *marketdata.Service,
	validator *risk.RiskValidator,
	cooldown *CooldownTracker,
	minStrength, sniperMinConf, rrRatio float64,
) *Engine {
	if sniperMinConf <= 0 {
		sniperMinConf = 0.75
	}
	if minStrength <= 0 {
		minStrength = 0.7
	}
	return &Engine{
		market:        market,
		validator:     validator,
		cooldown:      cooldown,
		minStrength:   minStrength,
		sniperMinConf: sniperMinConf,
		rrRatio:       rrRatio,
		minBars:       marketdata.MinBarsForDecision,
	}
}

// Evaluate analyzes live market data for one symbol.
func (e *Engine) Evaluate(ctx context.Context, symbol string) (*Decision, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	now := time.Now().UTC()

	snap, err := e.market.Refresh(ctx, sym)
	if err != nil {
		return &Decision{
			Action: ActionSkip,
			Reason: fmt.Sprintf("live price unavailable: %v", err),
			Symbol: sym,
			At:     now,
			Checks: []string{"price ✗"},
		}, nil
	}

	checks := []string{
		fmt.Sprintf("price ✓ (%s)", snap.Source),
		fmt.Sprintf("bars %d/%d", snap.BarCount, e.minBars),
	}

	if !snap.Ready(e.minBars) {
		return &Decision{
			Action:     ActionWait,
			Confidence: float64(snap.BarCount) / float64(e.minBars),
			Reason:     fmt.Sprintf("Building live history (%d/%d samples) — wait for sniper context", snap.BarCount, e.minBars),
			Symbol:     sym,
			At:         now,
			Checks:     checks,
		}, nil
	}

	types := []string{"UPTREND_SCALP", "DOWNTREND_SCALP", "TREND_CONFIRMATION"}
	var lastReason string
	for _, opp := range types {
		sig := e.signalFromSnapshot(snap, opp)
		if sig == nil {
			lastReason = fmt.Sprintf("no %s setup on live data", opp)
			continue
		}
		if err := meetsRequirements(sig, e.validator, e.minStrength); err != nil {
			lastReason = err.Error()
			continue
		}

		confidence := sig.Strength
		checks = append(checks,
			fmt.Sprintf("spread ✓"),
			fmt.Sprintf("trend ✓"),
			fmt.Sprintf("strength %.0f%%", confidence*100),
		)

		ok, remaining := e.cooldown.CanTake(sym)
		if !ok {
			return &Decision{
				Action:     ActionWait,
				Confidence: confidence,
				Reason:     fmt.Sprintf("Sniper cooldown active (%s remaining) — setup valid but wait", formatDuration(remaining)),
				Signal:     sig,
				Symbol:     sym,
				At:         now,
				Checks:     append(checks, "cooldown ⏳"),
			}, nil
		}

		return &Decision{
			Action:     ActionTake,
			Confidence: confidence,
			Reason:     sig.Reason,
			Signal:     sig,
			Symbol:     sym,
			At:         now,
			Checks:     append(checks, "cooldown ✓", "risk ✓"),
		}, nil
	}

	trend := analyzeTrendLabel(snap)
	if trend == "UPTREND" || trend == "DOWNTREND" || trend == "STRONG_UPTREND" || trend == "STRONG_DOWNTREND" {
		return &Decision{
			Action:     ActionWait,
			Confidence: 0.45,
			Reason:     fmt.Sprintf("Live %s regime — almost sniper-ready: %s", sym, lastReason),
			Symbol:     sym,
			At:         now,
			Checks:     append(checks, fmt.Sprintf("trend %s", trend)),
		}, nil
	}

	return &Decision{
		Action: ActionSkip,
		Reason: fmt.Sprintf("No sniper setup: %s", lastReason),
		Symbol: sym,
		At:     now,
		Checks: append(checks, "setup ✗"),
	}, nil
}

// SniperMinConfidence returns the threshold for assisted auto execution.
func (e *Engine) SniperMinConfidence() float64 {
	return e.sniperMinConf
}

// MarkTaken starts sniper cooldown after a TAKE was sent or executed.
func (e *Engine) MarkTaken(symbol string) {
	if e.cooldown != nil {
		e.cooldown.RecordTake(symbol)
	}
}

// AutoExecuteAllowed returns true when assisted auto may fire on this decision.
func (e *Engine) AutoExecuteAllowed(d *Decision) bool {
	if d == nil || d.Action != ActionTake || d.Signal == nil {
		return false
	}
	return d.Confidence >= e.sniperMinConf
}

func (e *Engine) signalFromSnapshot(snap *marketdata.Snapshot, opportunityType string) *models.TradeSignal {
	sg := signalgen.NewSignalGenerator(snap.Symbol, e.minStrength, e.rrRatio)
	sig := sg.GenerateSignal(snap.Mid, snap.ATR, snap.EMA20, snap.EMA50, snap.EMA200, snap.RSI, snap.Bid, snap.Ask)
	if sig == nil {
		return nil
	}
	if opportunityType != "" {
		if sig.Reason != "" {
			sig.Reason = opportunityType + " — " + sig.Reason
		} else {
			sig.Reason = opportunityType
		}
	}
	return sig
}

func analyzeTrendLabel(snap *marketdata.Snapshot) string {
	if snap.EMA20 > snap.EMA50 && snap.EMA50 > snap.EMA200 {
		return "STRONG_UPTREND"
	}
	if snap.EMA20 < snap.EMA50 && snap.EMA50 < snap.EMA200 {
		return "STRONG_DOWNTREND"
	}
	if snap.EMA20 > snap.EMA50 {
		return "UPTREND"
	}
	if snap.EMA20 < snap.EMA50 {
		return "DOWNTREND"
	}
	return "SIDEWAY"
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
