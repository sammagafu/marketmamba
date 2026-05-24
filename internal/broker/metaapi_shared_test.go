package broker

import "testing"

func TestApplySharedMetaAPIToken(t *testing.T) {
	SetSharedMetaAPIToken("platform-token-xyz")
	defer SetSharedMetaAPIToken("")

	merged := ApplySharedMetaAPIToken(Credentials{
		"login":    "123",
		"password": "pw",
		"server":   "Deriv-Demo",
	})
	if merged["metaapi_token"] != "platform-token-xyz" {
		t.Fatalf("expected injected token, got %q", merged["metaapi_token"])
	}

	keep := ApplySharedMetaAPIToken(Credentials{"metaapi_token": "user-own"})
	if keep["metaapi_token"] != "user-own" {
		t.Fatal("user token should win when provided")
	}
}

func TestAnyMTBrandEnabled(t *testing.T) {
	SetEnabledBrands([]string{"any_mt", "mock"})
	b, ok := BrandByID("any_mt")
	if !ok {
		t.Fatal("any_mt brand missing")
	}
	if !b.UsesMetaAPI || b.AdapterID != "metaapi" {
		t.Fatal("any_mt should use metaapi")
	}
	brands := SupportedBrands()
	found := false
	for _, x := range brands {
		if x.ID == "any_mt" {
			found = true
		}
	}
	if !found {
		t.Fatal("any_mt not in supported brands")
	}
}
