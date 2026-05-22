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

func (c *Config) SignalSymbols() []string {
	if c == nil || len(c.App.SignalSymbols) == 0 {
		return []string{"EURUSD", "BTCUSD"}
	}
	return c.App.SignalSymbols
}
