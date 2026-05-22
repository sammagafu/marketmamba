package config

import "testing"

func TestParseSignalSymbols(t *testing.T) {
	got := ParseSignalSymbols("EURUSD,BTCUSD", "")
	if len(got) != 2 || got[0] != "EURUSD" || got[1] != "BTCUSD" {
		t.Fatalf("got %v", got)
	}
	got = ParseSignalSymbols("", "GBPUSD")
	if len(got) != 1 || got[0] != "GBPUSD" {
		t.Fatalf("got %v", got)
	}
}
