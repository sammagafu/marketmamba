package pairs

import (
	"fmt"
	"strings"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/utils"
)

// Service manages per-user trading pair preferences.
type Service struct {
	store *storage.PostgresStorage
	cfg   *config.Config
}

func NewService(store *storage.PostgresStorage, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
}

func (s *Service) AvailableSymbols() []string {
	if s.cfg != nil {
		return s.cfg.SignalSymbols()
	}
	return []string{"EURUSD", "BTCUSD"}
}

func (s *Service) SeedDefaults(userID int64) error {
	return s.SetPreferences(userID, defaultPrefs(s.AvailableSymbols()))
}

func defaultPrefs(symbols []string) []models.UserTradingPair {
	out := make([]models.UserTradingPair, 0, len(symbols))
	for _, sym := range symbols {
		out = append(out, models.UserTradingPair{
			Symbol:         sym,
			ReceiveSignals: true,
			AutoTrade:      true,
		})
	}
	return out
}

func (s *Service) GetResponse(userID int64) (*models.TradingPairsResponse, error) {
	available := s.AvailableSymbols()
	customized, err := s.store.HasUserTradingPairs(userID)
	if err != nil {
		return nil, err
	}
	var pairs []models.UserTradingPair
	if customized {
		pairs, err = s.store.ListUserTradingPairs(userID)
		if err != nil {
			return nil, err
		}
	} else {
		pairs = defaultPrefs(available)
	}
	return &models.TradingPairsResponse{
		AvailableSymbols: available,
		Pairs:            mergeAvailable(available, pairs),
		Customized:       customized,
		SignalSymbols:    symbolsWithFlag(pairs, true, false),
		AutoTradeSymbols: symbolsWithFlag(pairs, false, true),
	}, nil
}

func mergeAvailable(available []string, pairs []models.UserTradingPair) []models.UserTradingPair {
	bySym := make(map[string]models.UserTradingPair)
	for _, p := range pairs {
		bySym[strings.ToUpper(p.Symbol)] = p
	}
	out := make([]models.UserTradingPair, 0, len(available))
	for _, sym := range available {
		if p, ok := bySym[sym]; ok {
			p.Symbol = sym
			out = append(out, p)
		} else {
			out = append(out, models.UserTradingPair{
				Symbol: sym, ReceiveSignals: false, AutoTrade: false,
			})
		}
	}
	return out
}

func symbolsWithFlag(pairs []models.UserTradingPair, signal, auto bool) []string {
	var out []string
	for _, p := range pairs {
		if signal && p.ReceiveSignals {
			out = append(out, p.Symbol)
		}
		if auto && p.AutoTrade {
			out = append(out, p.Symbol)
		}
	}
	return out
}

// SetPreferences replaces user pair rows. At least one symbol must receive signals.
func (s *Service) SetPreferences(userID int64, pairs []models.UserTradingPair) error {
	allowed := make(map[string]bool)
	for _, sym := range s.AvailableSymbols() {
		allowed[sym] = true
	}
	var normalized []models.UserTradingPair
	signalCount := 0
	for _, p := range pairs {
		sym := strings.ToUpper(strings.TrimSpace(p.Symbol))
		if !allowed[sym] {
			return fmt.Errorf("symbol %s is not available on this platform", sym)
		}
		if !utils.IsValidSymbol(sym) {
			return fmt.Errorf("invalid symbol: %s", sym)
		}
		if p.ReceiveSignals {
			signalCount++
		}
		normalized = append(normalized, models.UserTradingPair{
			UserID:         userID,
			Symbol:         sym,
			ReceiveSignals: p.ReceiveSignals,
			AutoTrade:      p.AutoTrade,
		})
	}
	if signalCount == 0 {
		return fmt.Errorf("enable signals for at least one pair")
	}
	return s.store.ReplaceUserTradingPairs(userID, normalized)
}

// SetSymbolsQuick enables listed symbols (signals + auto); others off.
func (s *Service) SetSymbolsQuick(userID int64, symbols []string) error {
	allowed := s.AvailableSymbols()
	want := make(map[string]bool)
	for _, sym := range symbols {
		sym = strings.ToUpper(strings.TrimSpace(sym))
		if sym == "" {
			continue
		}
		want[sym] = true
	}
	if len(want) == 0 {
		return fmt.Errorf("provide at least one symbol, e.g. EURUSD BTCUSD")
	}
	var pairs []models.UserTradingPair
	for _, sym := range allowed {
		pairs = append(pairs, models.UserTradingPair{
			Symbol:         sym,
			ReceiveSignals: want[sym],
			AutoTrade:      want[sym],
		})
	}
	return s.SetPreferences(userID, pairs)
}

func (s *Service) SignalSymbols(userID int64) ([]string, error) {
	resp, err := s.GetResponse(userID)
	if err != nil {
		return nil, err
	}
	return resp.SignalSymbols, nil
}

func (s *Service) AutoTradeSymbols(userID int64) ([]string, error) {
	resp, err := s.GetResponse(userID)
	if err != nil {
		return nil, err
	}
	return resp.AutoTradeSymbols, nil
}
