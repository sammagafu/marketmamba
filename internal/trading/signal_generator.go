package trading

import (
	"forex-bot/internal/models"
	"forex-bot/internal/signalgen"
)

// Re-exports for backward compatibility.
type SignalGenerator = signalgen.SignalGenerator

func NewSignalGenerator(symbol string, minStrength, riskRewardRatio float64) *SignalGenerator {
	return signalgen.NewSignalGenerator(symbol, minStrength, riskRewardRatio)
}

func SimulateScalpingOpportunity(symbol string, opportunityType string) *models.TradeSignal {
	return signalgen.SimulateScalpingOpportunity(symbol, opportunityType)
}

func CalculateATR(high, low, close, prevClose float64, period int) float64 {
	return signalgen.CalculateATR(high, low, close, prevClose, period)
}

func CalculateEMA(prices []float64, period int) float64 {
	return signalgen.CalculateEMA(prices, period)
}

func CalculateRSI(prices []float64, period int) float64 {
	return signalgen.CalculateRSI(prices, period)
}
