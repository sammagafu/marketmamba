package decision

import (
	"fmt"

	"forex-bot/internal/models"
	"forex-bot/internal/risk"
)

const referenceBalance = 10000.0

func meetsRequirements(signal *models.TradeSignal, validator *risk.RiskValidator, minStrength float64) error {
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
	if err := validator.ValidateTradeSignal(signal, referenceBalance, 0, 0, 0, false); err != nil {
		return fmt.Errorf("risk requirements: %w", err)
	}
	return nil
}
