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
func (c PlatformCatalog) AssetGroups(prefs models.SignalTypePreferences, fullCatalog bool, lockedHint string) []models.SignalAssetGroup {
	forexLocked := !fullCatalog && len(c.Forex) > 0
	indexLocked := !fullCatalog && len(c.Indexes) > 0
	forexDesc := "Major and cross currency pairs (EUR/USD, GBP/USD, etc.)."
	indexDesc := "Stock indices and volatility synthetics (US 500, NAS 100, Volatility 75)."
	if forexLocked && lockedHint != "" {
		forexDesc = lockedHint
	}
	if indexLocked && lockedHint != "" {
		indexDesc = lockedHint
	}
	return []models.SignalAssetGroup{
		{
			ID:          AssetForex,
			Label:       "Forex",
			Description: forexDesc,
			Symbols:     append([]string(nil), c.Forex...),
			Enabled:     prefs.Forex && !forexLocked,
			Locked:      forexLocked,
			ComingSoon:  forexLocked,
		},
		{
			ID:          AssetIndexes,
			Label:       "Indexes",
			Description: indexDesc,
			Symbols:     append([]string(nil), c.Indexes...),
			Enabled:     prefs.Indexes && !indexLocked,
			Locked:      indexLocked,
			ComingSoon:  indexLocked,
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

// FullCatalog returns the unrestricted platform catalog (for locked-group symbol lists).
func (c PlatformCatalog) FullCatalog(full PlatformCatalog) PlatformCatalog {
	out := c
	if len(full.Forex) > 0 {
		out.Forex = full.Forex
	}
	if len(full.Indexes) > 0 {
		out.Indexes = full.Indexes
	}
	return out
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
