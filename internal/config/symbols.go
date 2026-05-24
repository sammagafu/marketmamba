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

// SignalCatalog returns platform symbols grouped by asset class (respects community launch phase).
func (c *Config) SignalCatalog() (forex, indexes, crypto []string) {
	return c.PhasedSignalCatalog()
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
