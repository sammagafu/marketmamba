package risk

import (
	"fmt"

	"forex-bot/internal/models"
)

// RiskValidator checks if trades meet risk management criteria
type RiskValidator struct {
	settings *models.RiskSettings
}

func NewRiskValidator(settings *models.RiskSettings) *RiskValidator {
	return &RiskValidator{settings: settings}
}

// ValidateTradeSignal checks if a trade signal is safe to execute
func (v *RiskValidator) ValidateTradeSignal(
	signal *models.TradeSignal,
	balance float64,
	openTradeCount int,
	todayTradeCount int,
	dailyLoss float64,
	isPaused bool,
) error {
	// Check if bot is paused
	if isPaused {
		return fmt.Errorf("trading is paused")
	}

	// Check daily loss limit
	if dailyLoss <= 0 && dailyLoss < -v.settings.MaxDailyLoss*balance {
		return fmt.Errorf("daily loss limit hit: %.2f%%", (dailyLoss/balance)*100)
	}

	// Check max open trades
	if openTradeCount >= v.settings.MaxOpenTrades {
		return fmt.Errorf("max open trades (%d) reached", v.settings.MaxOpenTrades)
	}

	// Check max trades per day
	if todayTradeCount >= v.settings.MaxTradesPerDay {
		return fmt.Errorf("max trades per day (%d) reached", v.settings.MaxTradesPerDay)
	}

	// Validate signal parameters
	if err := v.validateSignal(signal); err != nil {
		return err
	}

	// Validate risk-reward ratio
	if err := v.validateRiskRewardRatio(signal); err != nil {
		return err
	}

	return nil
}

// CalculateLotSize calculates safe lot size based on risk per trade
func (v *RiskValidator) CalculateLotSize(balance, entryPrice, stopLoss float64) (float64, error) {
	if balance <= 0 || entryPrice <= 0 || stopLoss <= 0 {
		return 0, fmt.Errorf("invalid parameters for lot size calculation")
	}

	riskAmount := balance * v.settings.MaxRiskPerTrade
	pipsRisk := 0.0

	if entryPrice > stopLoss {
		pipsRisk = entryPrice - stopLoss
	} else {
		pipsRisk = stopLoss - entryPrice
	}

	if pipsRisk <= 0 {
		return 0, fmt.Errorf("invalid stop loss")
	}

	lotSize := riskAmount / pipsRisk
	return lotSize, nil
}

// validateSignal checks signal validity
func (v *RiskValidator) validateSignal(signal *models.TradeSignal) error {
	if signal.Type != "BUY" && signal.Type != "SELL" {
		return fmt.Errorf("invalid order type: %s", signal.Type)
	}

	if signal.StopLoss <= 0 || signal.TakeProfit <= 0 {
		return fmt.Errorf("stop loss and take profit must be positive")
	}

	if signal.Type == "BUY" && signal.StopLoss >= signal.TakeProfit {
		return fmt.Errorf("for BUY: stop loss must be below take profit")
	}

	if signal.Type == "SELL" && signal.StopLoss <= signal.TakeProfit {
		return fmt.Errorf("for SELL: stop loss must be above take profit")
	}

	if signal.Strength < 0 || signal.Strength > 1 {
		return fmt.Errorf("signal strength must be between 0 and 1")
	}

	return nil
}

// validateRiskRewardRatio checks if signal meets minimum risk-reward ratio.
func (v *RiskValidator) validateRiskRewardRatio(signal *models.TradeSignal) error {
	if signal.RiskRewardRatio > 0 {
		if signal.RiskRewardRatio < v.settings.RiskRewardRatio {
			return fmt.Errorf("risk-reward ratio %.2f is below minimum %.2f", signal.RiskRewardRatio, v.settings.RiskRewardRatio)
		}
		return nil
	}

	entry := (signal.StopLoss + signal.TakeProfit) / 2
	var riskDist, rewardDist float64
	if signal.Type == "BUY" {
		riskDist = entry - signal.StopLoss
		rewardDist = signal.TakeProfit - entry
	} else {
		riskDist = signal.StopLoss - entry
		rewardDist = entry - signal.TakeProfit
	}
	if riskDist <= 0 || rewardDist <= 0 {
		return fmt.Errorf("invalid stop loss / take profit levels")
	}
	ratio := rewardDist / riskDist
	if ratio < v.settings.RiskRewardRatio {
		return fmt.Errorf("risk-reward ratio %.2f is below minimum %.2f", ratio, v.settings.RiskRewardRatio)
	}
	return nil
}

// GetMaxPositionRisk calculates maximum risk for a position
func (v *RiskValidator) GetMaxPositionRisk(balance float64) float64 {
	return balance * v.settings.MaxRiskPerTrade
}

// CanOpenMoreTrades checks if more trades can be opened
func (v *RiskValidator) CanOpenMoreTrades(openCount, todayCount int) bool {
	return openCount < v.settings.MaxOpenTrades && todayCount < v.settings.MaxTradesPerDay
}
