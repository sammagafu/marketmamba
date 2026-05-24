package pairs

import (
	"testing"

	"forex-bot/internal/models"
)

func TestFilterByTypes(t *testing.T) {
	c := PlatformCatalog{
		Forex:   []string{"EURUSD"},
		Indexes: []string{"US500"},
		Crypto:  []string{"BTCUSD"},
	}
	out := c.FilterByTypes(models.SignalTypePreferences{Forex: true, Crypto: true})
	if len(out) != 2 {
		t.Fatalf("got %v", out)
	}
}

func TestParseSignalTypesFromArgs(t *testing.T) {
	p, ok := ParseSignalTypesFromArgs([]string{"forex", "crypto"})
	if !ok || !p.Forex || !p.Crypto || p.Indexes {
		t.Fatalf("%+v", p)
	}
}

func TestAtLeastOneType(t *testing.T) {
	if AtLeastOneType(models.SignalTypePreferences{}) {
		t.Fatal("expected false")
	}
}
