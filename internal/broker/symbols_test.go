package broker

import "testing"

func TestMetaAPISymbolCandidates(t *testing.T) {
	eur := MetaAPISymbolCandidates("EURUSD")
	if len(eur) == 0 || eur[0] != "frxEURUSD" {
		t.Fatalf("EURUSD candidates = %v", eur)
	}
	if MetaAPIToCanonical("frxEURUSD") != "EURUSD" {
		t.Fatal("frxEURUSD mapping")
	}
	if MetaAPIToCanonical("cryBTCUSD") != "BTCUSD" {
		t.Fatal("cryBTCUSD mapping")
	}
}

func TestNormalizeLots(t *testing.T) {
	caps := BrokerCapabilities{MinLot: 0.01, LotStep: 0.01}
	if NormalizeLots(caps, 0.005) != 0.01 {
		t.Fatal("min lot")
	}
	if NormalizeLots(caps, 0.156) != 0.15 {
		t.Fatalf("step lot got %v", NormalizeLots(caps, 0.156))
	}
}

func TestResolveBrandConnection(t *testing.T) {
	SetEnabledBrands([]string{"deriv", "mock"})
	provider, merged, label, err := ResolveBrandConnection("deriv", "", Credentials{
		"metaapi_token": "t",
		"login":         "1",
		"password":      "p",
		"server":        "Deriv-Demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if provider != "metaapi" {
		t.Fatalf("provider=%s", provider)
	}
	if merged["platform"] != "mt5" {
		t.Fatalf("preset platform=%s", merged["platform"])
	}
	if merged["brand_id"] != "deriv" {
		t.Fatal("brand_id not set")
	}
	if label == "" {
		t.Fatal("empty label")
	}
}
