package marketdata

import (
	"context"
	"fmt"
	"sync"
	"time"

)

const MinBarsForDecision = 35

// Service maintains live quotes and rolling price history per symbol.
type Service struct {
	provider Provider
	mu       sync.Mutex
	buffers  map[string]*PriceBuffer
	seeded   map[string]bool
	minBars  int
}

func NewService(twelveDataAPIKey string) *Service {
	return &Service{
		provider: NewCompositeProvider(twelveDataAPIKey),
		buffers:  make(map[string]*PriceBuffer),
		seeded:   make(map[string]bool),
		minBars:  MinBarsForDecision,
	}
}

func (s *Service) ProviderName() string {
	if s.provider == nil {
		return "none"
	}
	return s.provider.Name()
}

func (s *Service) Refresh(ctx context.Context, symbol string) (*Snapshot, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("market data provider not configured")
	}
	sym := normalizeForexSymbol(symbol)
	if err := s.ensureSeeded(ctx, sym); err != nil {
		// Non-fatal: continue accumulating live ticks.
		_ = err
	}

	q, err := s.provider.FetchQuote(ctx, sym)
	if err != nil {
		return nil, err
	}

	buf := s.buffer(sym)
	buf.Append(q.Mid, q.FetchedAt)

	return s.buildSnapshot(sym, q, buf), nil
}

func (s *Service) ensureSeeded(ctx context.Context, symbol string) error {
	s.mu.Lock()
	done := s.seeded[symbol]
	cp, ok := s.provider.(*CompositeProvider)
	s.mu.Unlock()
	if done || !ok || cp.twelveKey == "" {
		return nil
	}
	closes, err := cp.SeedBars(ctx, symbol, 120)
	if err != nil {
		return err
	}
	buf := s.buffer(symbol)
	for _, c := range closes {
		buf.Append(c, time.Now().UTC())
	}
	s.mu.Lock()
	s.seeded[symbol] = true
	s.mu.Unlock()
	return nil
}

func (s *Service) buffer(symbol string) *PriceBuffer {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.buffers[symbol] == nil {
		s.buffers[symbol] = NewPriceBuffer(defaultMaxBars)
	}
	return s.buffers[symbol]
}

func (s *Service) buildSnapshot(symbol string, q *Quote, buf *PriceBuffer) *Snapshot {
	closes := buf.Closes()
	ema20 := CalculateEMA(closes, 20)
	ema50 := CalculateEMA(closes, 50)
	ema200 := CalculateEMA(closes, 200)
	rsi := CalculateRSI(closes, 14)
	atr := buf.ATR(14)
	if atr <= 0 && len(closes) > 1 {
		atr = abs(closes[len(closes)-1]-closes[len(closes)-2]) * 2
	}
	return &Snapshot{
		Symbol:    symbol,
		Bid:       q.Bid,
		Ask:       q.Ask,
		Mid:       q.Mid,
		ATR:       atr,
		EMA20:     ema20,
		EMA50:     ema50,
		EMA200:    ema200,
		RSI:       rsi,
		BarCount:  len(closes),
		Source:    q.Source,
		FetchedAt: q.FetchedAt,
	}
}
