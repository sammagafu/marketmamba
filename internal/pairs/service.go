package pairs

import (
	"fmt"
	"strings"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/utils"
)

// Service manages per-user trading pair preferences and signal asset classes.
type Service struct {
	store *storage.PostgresStorage
	cfg   *config.Config
}

func NewService(store *storage.PostgresStorage, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
}

func (s *Service) catalog() PlatformCatalog {
	if s.cfg == nil {
		return PlatformCatalog{
			Forex:   []string{"EURUSD", "GBPUSD", "USDJPY"},
			Indexes: []string{"US500", "USTEC", "VOL75"},
			Crypto:  []string{"BTCUSD", "ETHUSD"},
		}
	}
	fx, idx, cry := s.cfg.SignalCatalog()
	return PlatformCatalog{Forex: fx, Indexes: idx, Crypto: cry}
}

func (s *Service) AvailableSymbols() []string {
	return s.catalog().All()
}

func (s *Service) availableForUser(userID int64) ([]string, error) {
	prefs, err := s.GetSignalTypes(userID)
	if err != nil {
		return nil, err
	}
	return s.catalog().FilterByTypes(prefs), nil
}

func (s *Service) GetSignalTypes(userID int64) (models.SignalTypePreferences, error) {
	prefs, _, err := s.store.GetUserSignalPreferences(userID)
	return prefs, err
}

// SetSignalTypes enables asset classes; disables pairs outside selected types.
func (s *Service) SetSignalTypes(userID int64, prefs models.SignalTypePreferences) error {
	if !AtLeastOneType(prefs) {
		return fmt.Errorf("enable at least one signal type: forex, indexes, or crypto")
	}
	if err := s.store.UpsertUserSignalPreferences(userID, prefs); err != nil {
		return err
	}
	allowed := s.catalog().FilterByTypes(prefs)
	customized, err := s.store.HasUserTradingPairs(userID)
	if err != nil {
		return err
	}
	if !customized {
		return s.SetPreferences(userID, defaultPrefs(allowed))
	}
	pairs, err := s.store.ListUserTradingPairs(userID)
	if err != nil {
		return err
	}
	allowedSet := make(map[string]bool)
	for _, sym := range allowed {
		allowedSet[sym] = true
	}
	for i := range pairs {
		if !allowedSet[strings.ToUpper(pairs[i].Symbol)] {
			pairs[i].ReceiveSignals = false
			pairs[i].AutoTrade = false
		}
	}
	return s.SetPreferences(userID, mergeAvailable(allowed, pairs))
}

func (s *Service) SeedDefaults(userID int64) error {
	prefs := models.DefaultSignalTypes()
	_ = s.store.UpsertUserSignalPreferences(userID, prefs)
	return s.SetPreferences(userID, defaultPrefs(s.catalog().FilterByTypes(prefs)))
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
	prefs, err := s.GetSignalTypes(userID)
	if err != nil {
		return nil, err
	}
	cat := s.catalog()
	available, err := s.availableForUser(userID)
	if err != nil {
		return nil, err
	}
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
	merged := mergeAvailable(available, pairs)
	return &models.TradingPairsResponse{
		AvailableSymbols: available,
		Pairs:            merged,
		Customized:       customized,
		SignalSymbols:    symbolsWithFlag(merged, true, false),
		AutoTradeSymbols: symbolsWithFlag(merged, false, true),
		SignalTypes:      prefs,
		AssetGroups:      cat.AssetGroups(prefs),
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
	available, err := s.availableForUser(userID)
	if err != nil {
		return err
	}
	allowed := make(map[string]bool)
	for _, sym := range available {
		allowed[sym] = true
	}
	var normalized []models.UserTradingPair
	signalCount := 0
	for _, p := range pairs {
		sym := strings.ToUpper(strings.TrimSpace(p.Symbol))
		if !allowed[sym] {
			return fmt.Errorf("symbol %s is not available for your selected signal types", sym)
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
		return fmt.Errorf("enable signals for at least one pair in your selected types")
	}
	return s.store.ReplaceUserTradingPairs(userID, normalized)
}

// SetSymbolsQuick enables listed symbols (signals + auto); others off.
func (s *Service) SetSymbolsQuick(userID int64, symbols []string) error {
	allowed, err := s.availableForUser(userID)
	if err != nil {
		return err
	}
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

// UserWantsSymbol checks asset-class prefs and per-pair flags.
func (s *Service) UserWantsSymbol(userID int64, symbol string) (bool, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	prefs, err := s.GetSignalTypes(userID)
	if err != nil {
		return false, err
	}
	class := s.catalog().ClassOf(sym)
	if class != "" && !AllowsClass(prefs, class) {
		return false, nil
	}
	return s.store.UserReceivesSignalForSymbol(userID, sym)
}
