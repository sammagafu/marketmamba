package signals

import (
	"fmt"

	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/signalgen"
)

// referenceBalance is used only for risk-rule structure checks (not per-user limits).
const referenceBalance = 10000.0

// MeetsRequirements returns nil if a signal passes generator-quality and risk rules for broadcast.
func MeetsRequirements(signal *models.TradeSignal, validator *risk.RiskValidator, minStrength float64) error {
	if signal == nil {
		return fmt.Errorf("no signal")
	}
	if minStrength <= 0 {
		minStrength = 0.7
	}
	if signal.Strength < minStrength {
		return fmt.Errorf("signal strength %.2f below minimum %.2f", signal.Strength, minStrength)
	}
	if signal.Symbol == "" || (signal.Type != "BUY" && signal.Type != "SELL") {
		return fmt.Errorf("invalid signal symbol or side")
	}
	// Same checks as auto-trade execution (paused/open-trade limits use neutral defaults).
	if err := validator.ValidateTradeSignal(signal, referenceBalance, 0, 0, 0, false); err != nil {
		return fmt.Errorf("risk requirements: %w", err)
	}
	return nil
}

// GenerateQualified tries each setup type until one passes technical + risk filters.
func GenerateQualified(symbol string, minStrength, riskRewardRatio float64, validator *risk.RiskValidator) (*models.TradeSignal, error) {
	types := []string{"UPTREND_SCALP", "DOWNTREND_SCALP", "TREND_CONFIRMATION"}
	var lastErr error
	for _, opp := range types {
		sig := signalgen.SimulateScalpingOpportunity(symbol, opp)
		if sig == nil {
			lastErr = fmt.Errorf("technical filters rejected %s setup", opp)
			continue
		}
		if err := MeetsRequirements(sig, validator, minStrength); err != nil {
			lastErr = err
			continue
		}
		return sig, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no setup met broadcast requirements")
}
