package pairs

import (
	"strings"

	"forex-bot/internal/models"
)

// Asset class ids (stored in user_signal_preferences).
const (
	AssetForex   = "forex"
	AssetIndexes = "indexes"
	AssetCrypto  = "crypto"
)

// PlatformCatalog holds symbols grouped by asset class (from server config).
type PlatformCatalog struct {
	Forex   []string
	Indexes []string
	Crypto  []string
}

func (c PlatformCatalog) All() []string {
	seen := make(map[string]bool)
	var out []string
	add := func(list []string) {
		for _, s := range list {
			s = strings.ToUpper(strings.TrimSpace(s))
			if s == "" || seen[s] {
				continue
			}
			seen[s] = true
			out = append(out, s)
		}
	}
	add(c.Forex)
	add(c.Indexes)
	add(c.Crypto)
	return out
}

// ClassOf returns the asset class for a symbol, or empty if unknown.
func (c PlatformCatalog) ClassOf(symbol string) string {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if contains(c.Forex, sym) {
		return AssetForex
	}
	if contains(c.Indexes, sym) {
		return AssetIndexes
	}
	if contains(c.Crypto, sym) {
		return AssetCrypto
	}
	return ""
}

func contains(list []string, sym string) bool {
	for _, s := range list {
		if strings.EqualFold(s, sym) {
			return true
		}
	}
	return false
}

// FilterByTypes returns symbols from catalog limited to enabled preference flags.
func (c PlatformCatalog) FilterByTypes(prefs models.SignalTypePreferences) []string {
	var out []string
	if prefs.Forex {
		out = append(out, c.Forex...)
	}
	if prefs.Indexes {
		out = append(out, c.Indexes...)
	}
	if prefs.Crypto {
		out = append(out, c.Crypto...)
	}
	seen := make(map[string]bool)
	var uniq []string
	for _, s := range out {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		uniq = append(uniq, s)
	}
	return uniq
}

// AssetGroups builds UI/API groups with enabled flags from prefs.
func (c PlatformCatalog) AssetGroups(prefs models.SignalTypePreferences) []models.SignalAssetGroup {
	return []models.SignalAssetGroup{
		{
			ID:          AssetForex,
			Label:       "Forex",
			Description: "Major and cross currency pairs (EUR/USD, GBP/USD, etc.).",
			Symbols:     append([]string(nil), c.Forex...),
			Enabled:     prefs.Forex,
		},
		{
			ID:          AssetIndexes,
			Label:       "Indexes",
			Description: "Stock indices and volatility synthetics (US 500, NAS 100, Volatility 75).",
			Symbols:     append([]string(nil), c.Indexes...),
			Enabled:     prefs.Indexes,
		},
		{
			ID:          AssetCrypto,
			Label:       "Bitcoin & crypto",
			Description: "Crypto vs USD (BTC/USD, ETH/USD).",
			Symbols:     append([]string(nil), c.Crypto...),
			Enabled:     prefs.Crypto,
		},
	}
}

// ParseSignalTypesFromArgs maps telegram/web keywords to preference toggles.
func ParseSignalTypesFromArgs(args []string) (models.SignalTypePreferences, bool) {
	if len(args) == 0 {
		return models.SignalTypePreferences{}, false
	}
	prefs := models.SignalTypePreferences{}
	any := false
	for _, a := range args {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "forex", "fx", "currencies":
			prefs.Forex = true
			any = true
		case "indexes", "index", "indices", "synthetics", "volatility":
			prefs.Indexes = true
			any = true
		case "crypto", "bitcoin", "btc", "coin", "coins":
			prefs.Crypto = true
			any = true
		case "all":
			return models.DefaultSignalTypes(), true
		}
	}
	return prefs, any
}

// AllowsClass reports whether prefs include the given asset class id.
func AllowsClass(prefs models.SignalTypePreferences, classID string) bool {
	switch classID {
	case AssetForex:
		return prefs.Forex
	case AssetIndexes:
		return prefs.Indexes
	case AssetCrypto:
		return prefs.Crypto
	default:
		return true
	}
}

func AtLeastOneType(prefs models.SignalTypePreferences) bool {
	return prefs.Forex || prefs.Indexes || prefs.Crypto
}
