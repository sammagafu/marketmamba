package marketdata

import "time"

// Quote is a live bid/ask snapshot.
type Quote struct {
	Symbol    string
	Bid       float64
	Ask       float64
	Mid       float64
	Source    string
	FetchedAt time.Time
}

// Snapshot is the live context used for sniper decisions.
type Snapshot struct {
	Symbol    string
	Bid       float64
	Ask       float64
	Mid       float64
	ATR       float64
	EMA20     float64
	EMA50     float64
	EMA200    float64
	RSI       float64
	BarCount  int
	Source    string
	FetchedAt time.Time
}

// Ready returns true when enough live samples exist for indicators.
func (s *Snapshot) Ready(minBars int) bool {
	if s == nil {
		return false
	}
	return s.BarCount >= minBars && s.Mid > 0
}
