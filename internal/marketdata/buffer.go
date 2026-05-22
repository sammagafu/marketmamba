package marketdata

import (
	"sync"
	"time"
)

const defaultMaxBars = 250

// PriceBuffer accumulates mid prices into synthetic OHLC bars for indicators.
type PriceBuffer struct {
	mu       sync.Mutex
	closes   []float64
	highs    []float64
	lows     []float64
	max      int
	lastTick time.Time
}

func NewPriceBuffer(max int) *PriceBuffer {
	if max <= 0 {
		max = defaultMaxBars
	}
	return &PriceBuffer{max: max}
}

func (b *PriceBuffer) Append(mid float64, t time.Time) {
	if mid <= 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closes = append(b.closes, mid)
	// Synthetic micro-range for ATR (≈0.01% of price).
	eps := mid * 0.0001
	if eps < 1e-8 {
		eps = 1e-8
	}
	b.highs = append(b.highs, mid+eps)
	b.lows = append(b.lows, mid-eps)
	if len(b.closes) > b.max {
		b.closes = b.closes[len(b.closes)-b.max:]
		b.highs = b.highs[len(b.highs)-b.max:]
		b.lows = b.lows[len(b.lows)-b.max:]
	}
	b.lastTick = t
}

func (b *PriceBuffer) Closes() []float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]float64, len(b.closes))
	copy(out, b.closes)
	return out
}

func (b *PriceBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.closes)
}

func (b *PriceBuffer) LastClose() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.closes) == 0 {
		return 0
	}
	return b.closes[len(b.closes)-1]
}

func (b *PriceBuffer) ATR(period int) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.closes) < 2 || period <= 0 {
		return 0
	}
	n := len(b.closes)
	if n > period+1 {
		n = period + 1
	}
	start := len(b.closes) - n
	var sum float64
	for i := start + 1; i < len(b.closes); i++ {
		h := b.highs[i]
		l := b.lows[i]
		pc := b.closes[i-1]
		tr1 := h - l
		tr2 := abs(h - pc)
		tr3 := abs(l - pc)
		sum += max3(tr1, tr2, tr3)
	}
	return sum / float64(n-1)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max3(a, b, c float64) float64 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}
