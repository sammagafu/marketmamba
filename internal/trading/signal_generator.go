package trading

import (
	"math"
	"time"

	"forex-bot/internal/logger"
	"forex-bot/internal/models"
)

// SignalGenerator produces trading signals based on technical analysis
type SignalGenerator struct {
	symbol          string
	minStrength     float64 // Minimum signal strength (0-1)
	riskRewardRatio float64
}

func NewSignalGenerator(symbol string, minStrength, riskRewardRatio float64) *SignalGenerator {
	return &SignalGenerator{
		symbol:          symbol,
		minStrength:     minStrength,
		riskRewardRatio: riskRewardRatio,
	}
}

// GenerateSignal creates a trading signal based on market data
// In production, this would use real price data
func (sg *SignalGenerator) GenerateSignal(
	currentPrice float64,
	atr float64,
	ema20 float64,
	ema50 float64,
	ema200 float64,
	rsi float64,
	bid float64,
	ask float64,
) *models.TradeSignal {

	// Calculate spread
	spread := ask - bid
	if spread < 0 {
		spread = -spread
	}
	maxSpread := currentPrice * 0.0003 // ~3 pips on EURUSD

	// Spread filter: reject if spread is too large (more than 3 pips)
	if spread > maxSpread {
		logger.Debug(
			"[%s] Signal rejected: spread %.5f (%.1f pips) > max %.5f (%.1f pips) | bid=%.5f ask=%.5f price=%.5f",
			sg.symbol, spread, spreadToPips(spread), maxSpread, spreadToPips(maxSpread), bid, ask, currentPrice,
		)
		return nil
	}

	// ATR volatility filter: volatility must be reasonable
	if !sg.checkVolatilityFilter(atr, currentPrice) {
		logger.Debug("[%s] Signal rejected: volatility too low (ATR=%.5f, price=%.5f)", sg.symbol, atr, currentPrice)
		return nil
	}

	// Trend filter: price must be in a reasonable position
	trend := sg.analyzeTrend(currentPrice, ema20, ema50, ema200)

	// RSI filter: avoid overbought/oversold extremes
	if !sg.checkRSIFilter(rsi) {
		logger.Debug("[%s] Signal rejected: RSI filter failed (RSI=%.2f)", sg.symbol, rsi)
		return nil
	}

	// Generate signal based on trend and technical setup
	signal := sg.generateTrendSignal(currentPrice, ema20, ema50, ema200, atr, rsi, trend)

	if signal == nil {
		return nil
	}

	// Calculate stop loss and take profit with ATR
	signal = sg.calculatePriceTargets(signal, currentPrice, atr)

	// Final strength validation
	if signal == nil || signal.Strength < sg.minStrength {
		if signal != nil {
			logger.Debug("[%s] Signal rejected: strength too low (%.2f < %.2f)", sg.symbol, signal.Strength, sg.minStrength)
		}
		return nil
	}

	logger.Info(
		"[%s] Signal passed filters | %s spread=%.5f (%.1f pips, max %.1f) bid=%.5f ask=%.5f strength=%.2f",
		sg.symbol, signal.Type, spread, spreadToPips(spread), spreadToPips(maxSpread), bid, ask, signal.Strength,
	)
	return signal
}

func (sg *SignalGenerator) checkVolatilityFilter(atr, currentPrice float64) bool {
	if atr <= 0 || currentPrice <= 0 {
		return false
	}

	// ATR must be at least 0.05% of current price
	minVolatility := currentPrice * 0.0005
	return atr >= minVolatility
}

func (sg *SignalGenerator) analyzeTrend(currentPrice, ema20, ema50, ema200 float64) string {
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

func (sg *SignalGenerator) checkRSIFilter(rsi float64) bool {
	// Avoid trading in extreme overbought (>75) or oversold (<25) conditions
	// These can be risky for scalping
	return rsi > 20 && rsi < 80
}

func (sg *SignalGenerator) generateTrendSignal(
	currentPrice, ema20, ema50, ema200, atr, rsi float64,
	trend string,
) *models.TradeSignal {

	switch trend {
	case "STRONG_UPTREND":
		// Buy signal: price at/near EMA20, RSI not overbought
		if currentPrice >= ema20*0.99 && currentPrice <= ema20*1.01 && rsi < 70 {
			return &models.TradeSignal{
				Symbol:      sg.symbol,
				Type:        "BUY",
				Strength:    0.9,
				TriggeredAt: time.Now(),
			}
		}

	case "STRONG_DOWNTREND":
		// Sell signal: price at/near EMA20, RSI not oversold
		if currentPrice >= ema20*0.99 && currentPrice <= ema20*1.01 && rsi > 30 {
			return &models.TradeSignal{
				Symbol:      sg.symbol,
				Type:        "SELL",
				Strength:    0.9,
				TriggeredAt: time.Now(),
			}
		}

	case "UPTREND":
		// Buy signal with lower strength
		if currentPrice >= ema50*0.995 && rsi < 70 {
			return &models.TradeSignal{
				Symbol:      sg.symbol,
				Type:        "BUY",
				Strength:    0.7,
				TriggeredAt: time.Now(),
			}
		}

	case "DOWNTREND":
		// Sell signal with lower strength
		if currentPrice <= ema50*1.005 && rsi > 30 {
			return &models.TradeSignal{
				Symbol:      sg.symbol,
				Type:        "SELL",
				Strength:    0.7,
				TriggeredAt: time.Now(),
			}
		}
	}

	return nil
}

func (sg *SignalGenerator) calculatePriceTargets(signal *models.TradeSignal, currentPrice, atr float64) *models.TradeSignal {
	if signal == nil || atr <= 0 {
		return nil
	}

	// Use ATR to set stop loss and take profit
	// Stop loss at 1x ATR away
	// Take profit at 2x ATR away (for 1:2 risk-reward)

	if signal.Type == "BUY" {
		signal.StopLoss = currentPrice - (atr * 1.0)
		signal.TakeProfit = currentPrice + (atr * sg.riskRewardRatio)
	} else { // SELL
		signal.StopLoss = currentPrice + (atr * 1.0)
		signal.TakeProfit = currentPrice - (atr * sg.riskRewardRatio)
	}

	// Validate targets
	if signal.StopLoss <= 0 || signal.TakeProfit <= 0 {
		logger.Debug("Invalid price targets: SL=%.5f TP=%.5f", signal.StopLoss, signal.TakeProfit)
		return nil
	}

	// Calculate actual risk-reward ratio
	var risk, reward float64
	if signal.Type == "BUY" {
		risk = currentPrice - signal.StopLoss
		reward = signal.TakeProfit - currentPrice
	} else {
		risk = signal.StopLoss - currentPrice
		reward = currentPrice - signal.TakeProfit
	}

	if risk > 0 {
		signal.RiskRewardRatio = reward / risk
	}

	return signal
}

// MockPriceData simulates market data for testing
type MockPriceData struct {
	CurrentPrice float64
	ATR          float64
	EMA20        float64
	EMA50        float64
	EMA200       float64
	RSI          float64
	Bid          float64
	Ask          float64
}

// GenerateRealisticSignal creates a signal from simulated market conditions
func (sg *SignalGenerator) GenerateRealisticSignal(data MockPriceData) *models.TradeSignal {
	return sg.GenerateSignal(
		data.CurrentPrice,
		data.ATR,
		data.EMA20,
		data.EMA50,
		data.EMA200,
		data.RSI,
		data.Bid,
		data.Ask,
	)
}

// CalculateATR computes Average True Range
func CalculateATR(high, low, close, prevClose float64, period int) float64 {
	if high <= 0 || low <= 0 || close <= 0 {
		return 0
	}

	tr1 := high - low
	tr2 := math.Abs(high - prevClose)
	tr3 := math.Abs(low - prevClose)

	tr := math.Max(tr1, math.Max(tr2, tr3))
	return tr / float64(period)
}

// CalculateEMA computes Exponential Moving Average
func CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	// Multiplier for EMA
	multiplier := 2.0 / float64(period+1)

	// Calculate SMA for first value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema := sum / float64(period)

	// Calculate EMA for remaining prices
	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// CalculateRSI computes Relative Strength Index
func CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50
	}

	gains := 0.0
	losses := 0.0

	for i := 1; i <= period; i++ {
		change := prices[len(prices)-i] - prices[len(prices)-i-1]
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// SimulateScalpingOpportunity creates a realistic scalping signal for testing
func SimulateScalpingOpportunity(symbol string, opportunityType string) *models.TradeSignal {
	sg := NewSignalGenerator(symbol, 0.7, 2.0)

	var currentPrice, atr, ema20, ema50, ema200, rsi, bid, ask float64

	// Simulate different market conditions
	switch opportunityType {
	case "UPTREND_SCALP":
		// Strong uptrend with pullback
		currentPrice = 1.1050
		atr = 0.0035
		ema20 = 1.1045
		ema50 = 1.1030
		ema200 = 1.1000
		rsi = 55
		bid = 1.10495
		ask = 1.10505

	case "DOWNTREND_SCALP":
		// Strong downtrend with pullback
		currentPrice = 1.0950
		atr = 0.0035
		ema20 = 1.0955
		ema50 = 1.0970
		ema200 = 1.1000
		rsi = 45
		bid = 1.09495
		ask = 1.09505

	case "TREND_CONFIRMATION":
		// Clean trend following setup
		currentPrice = 1.1055
		atr = 0.0040
		ema20 = 1.1050
		ema50 = 1.1040
		ema200 = 1.1020
		rsi = 60
		bid = 1.10545
		ask = 1.10555

	default:
		return nil
	}

	return sg.GenerateSignal(currentPrice, atr, ema20, ema50, ema200, rsi, bid, ask)
}

// spreadToPips converts price spread to approximate pips (5-digit FX quote).
func spreadToPips(spread float64) float64 {
	return spread / 0.0001
}
