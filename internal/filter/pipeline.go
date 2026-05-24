package filter

import (
	"fmt"
	"math"

	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/signalgen"
)

// RunTechnical evaluates market + technical + setup filters and optionally builds a signal.
func RunTechnical(in Input, minBars int) (*Report, *models.TradeSignal) {
	r := &Report{
		Symbol:      in.Symbol,
		DataSource:  in.Source,
		BarCount:    in.BarCount,
		MinBars:     minBars,
		LiveReady:   in.BarCount >= minBars && in.Price > 0,
	}
	if in.Source == "" {
		r.DataSource = "simulated"
	}

	market := Layer{ID: CategoryMarket, Title: "Market quality"}
	technical := Layer{ID: CategoryTechnical, Title: "Technical filters"}
	setup := Layer{ID: CategorySetup, Title: "Setup & signal"}

	// Live history (informational for simulated)
	if in.BarCount > 0 && minBars > 0 {
		st := StatusPass
		msg := fmt.Sprintf("%d bars available (need %d)", in.BarCount, minBars)
		if in.BarCount < minBars {
			st = StatusWarn
			msg = fmt.Sprintf("Warming up: %d/%d bars — sniper may WAIT", in.BarCount, minBars)
		}
		market.Steps = append(market.Steps, Step{
			ID: "live_history", Name: "Live bar history", Category: CategoryMarket,
			Status: st, Message: msg,
			Metrics: map[string]interface{}{"bars": in.BarCount, "required": minBars},
		})
	}

	spread := math.Abs(in.Ask - in.Bid)
	maxSp := MaxSpread(in.Symbol, in.Price)
	spreadStep := Step{
		ID: "spread", Name: "Spread gate", Category: CategoryMarket,
		Description: "Tight spread required before entry",
		Metrics: map[string]interface{}{
			"spread": spread, "max": maxSp, "units": SpreadUnits(in.Symbol, spread),
			"bid": in.Bid, "ask": in.Ask,
		},
	}
	if in.Price <= 0 || in.Bid <= 0 || in.Ask <= 0 {
		spreadStep.Status = StatusFail
		spreadStep.Message = "Missing quote data"
	} else if spread > maxSp {
		spreadStep.Status = StatusFail
		spreadStep.Message = fmt.Sprintf("Spread %.5f exceeds max %.5f", spread, maxSp)
	} else {
		spreadStep.Status = StatusPass
		spreadStep.Message = fmt.Sprintf("Spread OK (%.1f units)", SpreadUnits(in.Symbol, spread))
	}
	market.Steps = append(market.Steps, spreadStep)

	// Volatility
	volStep := Step{ID: "volatility", Name: "ATR floor", Category: CategoryTechnical}
	minVol := in.Price * 0.0005
	volStep.Metrics = map[string]interface{}{"atr": in.ATR, "min": minVol, "price": in.Price}
	if in.ATR <= 0 || in.Price <= 0 || in.ATR < minVol {
		volStep.Status = StatusFail
		volStep.Message = "Volatility too low for scalping"
	} else {
		volStep.Status = StatusPass
		volStep.Message = fmt.Sprintf("ATR %.5f above floor", in.ATR)
	}
	technical.Steps = append(technical.Steps, volStep)

	// RSI
	rsiStep := Step{ID: "rsi_band", Name: "RSI band", Category: CategoryTechnical, Metrics: map[string]interface{}{"rsi": in.RSI}}
	if in.RSI <= 20 || in.RSI >= 80 {
		rsiStep.Status = StatusFail
		rsiStep.Message = fmt.Sprintf("RSI %.1f outside 20–80 band", in.RSI)
	} else {
		rsiStep.Status = StatusPass
		rsiStep.Message = fmt.Sprintf("RSI %.1f in range", in.RSI)
	}
	technical.Steps = append(technical.Steps, rsiStep)

	trend := TrendLabel(in.Price, in.EMA20, in.EMA50, in.EMA200)
	r.Trend = trend
	technical.Steps = append(technical.Steps, Step{
		ID: "ema_trend", Name: "EMA stack", Category: CategoryTechnical,
		Status: StatusPass, Message: trend,
		Metrics: map[string]interface{}{
			"ema20": in.EMA20, "ema50": in.EMA50, "ema200": in.EMA200,
		},
	})

	// Early exit if hard market/technical fails
	if spreadStep.Status == StatusFail || volStep.Status == StatusFail || rsiStep.Status == StatusFail {
		r.Layers = []Layer{market, technical}
		r.Verdict = StatusFail
		r.Summary = "Market or technical gates failed — no setup evaluated"
		return r, nil
	}

	// Setup via signal generator (same logic as production)
	minStr := in.MinStrength
	if minStr <= 0 {
		minStr = 0.7
	}
	rr := in.MinRR
	if rr <= 0 {
		rr = 1.0
	}
	sg := signalgen.NewSignalGenerator(in.Symbol, minStr, rr)
	sig := sg.GenerateSignal(in.Price, in.ATR, in.EMA20, in.EMA50, in.EMA200, in.RSI, in.Bid, in.Ask)

	setupStep := Step{ID: "setup", Name: "Setup pattern", Category: CategorySetup}
	if sig == nil {
		setupStep.Status = StatusFail
		setupStep.Message = fmt.Sprintf("No entry pattern for %s regime", trend)
		setup.Steps = append(setup.Steps, setupStep)
		r.Layers = []Layer{market, technical, setup}
		r.Verdict = StatusFail
		r.Summary = "Technical context OK but no qualified setup"
		return r, nil
	}
	setupStep.Status = StatusPass
	setupStep.Message = sig.Reason
	setup.Steps = append(setup.Steps, setupStep)

	strStep := Step{
		ID: "strength", Name: "Signal strength", Category: CategorySetup,
		Metrics: map[string]interface{}{"strength": sig.Strength, "min": minStr},
	}
	if sig.Strength < minStr {
		strStep.Status = StatusFail
		strStep.Message = fmt.Sprintf("Strength %.0f%% below %.0f%%", sig.Strength*100, minStr*100)
	} else {
		strStep.Status = StatusPass
		strStep.Message = fmt.Sprintf("Strength %.0f%%", sig.Strength*100)
	}
	setup.Steps = append(setup.Steps, strStep)

	rrStep := Step{
		ID: "risk_reward", Name: "Risk–reward", Category: CategoryRisk,
		Metrics: map[string]interface{}{
			"rr": sig.RiskRewardRatio, "min": rr, "sl": sig.StopLoss, "tp": sig.TakeProfit,
		},
	}
	if sig.RiskRewardRatio < rr {
		rrStep.Status = StatusFail
		rrStep.Message = fmt.Sprintf("R:R %.2f below minimum %.2f", sig.RiskRewardRatio, rr)
	} else {
		rrStep.Status = StatusPass
		rrStep.Message = fmt.Sprintf("R:R %.2f", sig.RiskRewardRatio)
	}
	setup.Steps = append(setup.Steps, rrStep)

	r.Layers = []Layer{market, technical, setup}
	r.SignalSide = sig.Type
	r.Strength = sig.Strength
	r.RiskReward = sig.RiskRewardRatio
	r.SetupReason = sig.Reason

	failed := strStep.Status == StatusFail || rrStep.Status == StatusFail
	if failed {
		r.Verdict = StatusFail
		r.Summary = "Setup formed but did not meet strength or R:R gates"
		return r, sig
	}
	r.Verdict = StatusPass
	r.Qualified = true
	r.Summary = fmt.Sprintf("Qualified %s setup — %s", sig.Type, trend)
	return r, sig
}

// AppendRisk runs structural risk checks on a candidate signal.
func AppendRisk(r *Report, sig *models.TradeSignal, validator *risk.RiskValidator, minStrength float64) {
	if r == nil || validator == nil {
		return
	}
	layer := Layer{ID: CategoryRisk, Title: "Risk envelope"}
	const refBalance = 10000.0

	step := Step{ID: "risk_rules", Name: "Structural risk", Category: CategoryRisk}
	if sig == nil {
		step.Status = StatusSkip
		step.Message = "No signal to validate"
		layer.Steps = append(layer.Steps, step)
		r.Layers = append(r.Layers, layer)
		return
	}
	if err := validator.ValidateTradeSignal(sig, refBalance, 0, 0, 0, false); err != nil {
		step.Status = StatusFail
		step.Message = err.Error()
		r.Qualified = false
		r.Verdict = StatusFail
		r.Summary = "Setup failed structural risk: " + err.Error()
	} else {
		step.Status = StatusPass
		step.Message = "SL/TP, R:R, and platform limits OK (reference book)"
	}
	layer.Steps = append(layer.Steps, step)

	minStep := Step{ID: "min_strength", Name: "Broadcast floor", Category: CategoryPlatform}
	if minStrength <= 0 {
		minStrength = 0.7
	}
	minStep.Metrics = map[string]interface{}{"strength": sig.Strength, "floor": minStrength}
	if sig.Strength < minStrength {
		minStep.Status = StatusFail
		minStep.Message = "Below broadcast minimum"
	} else {
		minStep.Status = StatusPass
		minStep.Message = "Meets broadcast minimum"
	}
	layer.Steps = append(layer.Steps, minStep)

	r.Layers = append(r.Layers, layer)
	if minStep.Status == StatusFail {
		r.Qualified = false
		r.Verdict = StatusFail
	}
}
