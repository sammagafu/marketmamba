package marketdata

import "math"

// CalculateEMA computes exponential moving average for the last window.
func CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	multiplier := 2.0 / float64(period+1)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema := sum / float64(period)
	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}
	return ema
}

// CalculateRSI computes RSI for the trailing window.
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

// CalculateATR approximates ATR from OHLC slices.
func CalculateATR(highs, lows, closes []float64, period int) float64 {
	if len(closes) < 2 || period <= 0 {
		return 0
	}
	n := len(closes)
	if n > period+1 {
		n = period + 1
	}
	start := len(closes) - n
	var sum float64
	for i := start + 1; i < len(closes); i++ {
		h := highs[i]
		l := lows[i]
		pc := closes[i-1]
		tr1 := h - l
		tr2 := math.Abs(h - pc)
		tr3 := math.Abs(l - pc)
		sum += max3(tr1, tr2, tr3)
	}
	return sum / float64(n-1)
}
