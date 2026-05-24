package config

import "strings"

// ParseSignalSymbols returns symbols to broadcast and auto-monitor.
// Uses SIGNAL_BROADCAST_SYMBOLS (CSV) when set, else SIGNAL_BROADCAST_SYMBOL, else defaults.
func ParseSignalSymbols(csv, single string) []string {
	var out []string
	seen := make(map[string]bool)
	add := func(s string) {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	if strings.TrimSpace(csv) != "" {
		for _, part := range strings.Split(csv, ",") {
			add(part)
		}
	} else {
		add(single)
	}
	if len(out) == 0 {
		return []string{"EURUSD", "BTCUSD"}
	}
	return out
}

func parseSymbolCSV(csv string, defaults []string) []string {
	if strings.TrimSpace(csv) == "" {
		return append([]string(nil), defaults...)
	}
	return ParseSignalSymbols(csv, "")
}

// SignalCatalog returns platform symbols grouped by asset class.
func (c *Config) SignalCatalog() (forex, indexes, crypto []string) {
	forexDef := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "EURJPY"}
	indexDef := []string{"US500", "USTEC", "GER40", "UK100", "VOL75"}
	cryptoDef := []string{"BTCUSD", "ETHUSD"}
	if c == nil {
		return forexDef, indexDef, cryptoDef
	}
	forex = parseSymbolCSV(getEnv("SIGNAL_FOREX_SYMBOLS", ""), forexDef)
	indexes = parseSymbolCSV(getEnv("SIGNAL_INDEX_SYMBOLS", ""), indexDef)
	crypto = parseSymbolCSV(getEnv("SIGNAL_CRYPTO_SYMBOLS", ""), cryptoDef)
	// Legacy flat list augments forex+crypto when per-class envs are unset
	if getEnv("SIGNAL_FOREX_SYMBOLS", "") == "" && getEnv("SIGNAL_INDEX_SYMBOLS", "") == "" &&
		getEnv("SIGNAL_CRYPTO_SYMBOLS", "") == "" && len(c.App.SignalSymbols) > 0 {
		for _, sym := range c.App.SignalSymbols {
			sym = strings.ToUpper(strings.TrimSpace(sym))
			switch sym {
			case "BTCUSD", "ETHUSD":
				if !containsSym(crypto, sym) {
					crypto = append(crypto, sym)
				}
			case "US500", "USTEC", "GER40", "UK100", "VOL75":
				if !containsSym(indexes, sym) {
					indexes = append(indexes, sym)
				}
			default:
				if !containsSym(forex, sym) {
					forex = append(forex, sym)
				}
			}
		}
	}
	return forex, indexes, crypto
}

func containsSym(list []string, sym string) bool {
	for _, s := range list {
		if strings.EqualFold(s, sym) {
			return true
		}
	}
	return false
}

func (c *Config) SignalSymbols() []string {
	if c == nil {
		return []string{"EURUSD", "BTCUSD"}
	}
	fx, idx, cry := c.SignalCatalog()
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
	add(fx)
	add(idx)
	add(cry)
	if len(out) == 0 {
		return []string{"EURUSD", "BTCUSD"}
	}
	return out
}

// getEnv is duplicated call - SignalCatalog uses getEnv but it's in same package config.go - good
