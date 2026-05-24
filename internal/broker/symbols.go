package broker

import (
	"math"
	"strings"
)

// MetaAPISymbolCandidates returns broker symbol names to try for a canonical pair (e.g. EURUSD).
func MetaAPISymbolCandidates(symbol string) []string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	switch s {
	case "EURUSD":
		return []string{"frxEURUSD", "EURUSD", "EURUSDm"}
	case "GBPUSD":
		return []string{"frxGBPUSD", "GBPUSD"}
	case "USDJPY":
		return []string{"frxUSDJPY", "USDJPY"}
	case "BTCUSD":
		return []string{"cryBTCUSD", "BTCUSD"}
	default:
		if strings.HasPrefix(s, "FRX") || strings.HasPrefix(s, "CRY") {
			return []string{s}
		}
		return []string{"frx" + s, s}
	}
}

// MetaAPIToCanonical maps a broker symbol back to canonical form (EURUSD).
func MetaAPIToCanonical(sym string) string {
	s := strings.TrimSpace(sym)
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "frx") && len(s) > 3 {
		return strings.ToUpper(s[3:])
	}
	if strings.HasPrefix(low, "cry") && len(s) > 3 {
		return strings.ToUpper(s[3:])
	}
	return strings.ToUpper(s)
}

// NormalizeSymbolForProvider maps canonical symbol to broker-native form for order placement.
func NormalizeSymbolForProvider(provider, canonical string) string {
	switch provider {
	case "oanda":
		return symbolToOANDA(canonical)
	case "metaapi":
		cands := MetaAPISymbolCandidates(canonical)
		if len(cands) > 0 {
			return cands[0]
		}
		return canonical
	default:
		return strings.ToUpper(strings.TrimSpace(canonical))
	}
}

// NormalizeLots rounds quantity to adapter min/step.
func NormalizeLots(caps BrokerCapabilities, lots float64) float64 {
	min := caps.MinLot
	if min <= 0 {
		min = 0.01
	}
	step := caps.LotStep
	if step <= 0 {
		step = 0.01
	}
	if lots < min {
		return min
	}
	steps := math.Floor((lots - min) / step)
	return min + steps*step
}
