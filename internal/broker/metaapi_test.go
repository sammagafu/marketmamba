package broker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func derivDemoCreds() Credentials {
	return Credentials{
		"metaapi_token": "test-metaapi-token",
		"login":         "201620473",
		"password":      "secret-mt5",
		"server":        "Deriv-Demo",
		"platform":      "mt5",
		"keywords":      "Deriv.com Limited",
	}
}

func TestValidateMetaAPICredentials(t *testing.T) {
	tests := []struct {
		name    string
		creds   Credentials
		wantErr string
	}{
		{"nil creds", nil, "required"},
		{"missing token", Credentials{"login": "1", "password": "p", "server": "S"}, "token"},
		{"cloud id only", Credentials{"metaapi_token": "tok", "metaapi_account_id": "865d3a4d-3803-486d-bdf3-a85679d9fad2"}, ""},
		{"missing login", Credentials{"metaapi_token": "tok", "password": "p", "server": "S"}, "login"},
		{"missing password", Credentials{"metaapi_token": "tok", "login": "1", "server": "S"}, "password"},
		{"missing server", Credentials{"metaapi_token": "tok", "login": "1", "password": "p"}, "server"},
		{"deriv demo ok", derivDemoCreds(), ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMetaAPICredentials(tc.creds)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.wantErr)) {
				t.Fatalf("error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestIsMetaAPICloudID(t *testing.T) {
	if !isMetaAPICloudID("865d3a4d-3803-486d-bdf3-a85679d9fad2") {
		t.Fatal("expected uuid to be cloud id")
	}
	if isMetaAPICloudID("201620473") {
		t.Fatal("numeric login should not be cloud id")
	}
}

func TestMetaAPISymbolCandidates(t *testing.T) {
	eur := metaAPISymbolCandidates("EURUSD")
	if len(eur) == 0 || eur[0] != "frxEURUSD" {
		t.Fatalf("EURUSD candidates = %v", eur)
	}
	if metaAPIToSymbol("frxEURUSD") != "EURUSD" {
		t.Fatal("frxEURUSD mapping")
	}
	if metaAPIToSymbol("cryBTCUSD") != "BTCUSD" {
		t.Fatal("cryBTCUSD mapping")
	}
}

func TestMetaAPITokenAliases(t *testing.T) {
	c := Credentials{"token": "a"}
	if metaAPIToken(c) != "a" {
		t.Fatal("token alias")
	}
	c = Credentials{"metaapi_token": "b", "token": "a"}
	if metaAPIToken(c) != "b" {
		t.Fatal("metaapi_token preferred")
	}
}

func TestSaveConnectionMetaAPI(t *testing.T) {
	store := &memConnStore{}
	err := SaveConnection(store, "test-encryption-key-32bytes!!", 123, "metaapi", "", derivDemoCreds())
	if err != nil {
		t.Fatal(err)
	}
	if store.conn.Provider != "metaapi" {
		t.Fatalf("expected metaapi, got %s", store.conn.Provider)
	}
}

func TestIsLiveProviderMetaAPI(t *testing.T) {
	if !IsLiveProvider("metaapi") {
		t.Fatal("metaapi should be live")
	}
}

func newTestMetaAPIBroker(t *testing.T, provisionURL, clientURL string, creds Credentials) *MetaAPIBroker {
	t.Helper()
	b, err := NewMetaAPIBroker(creds)
	if err != nil {
		t.Fatal(err)
	}
	b.provisionBase = provisionURL
	b.clientBase = clientURL
	return b
}

func TestMetaAPIBroker_ExistingCloudAccount(t *testing.T) {
	const accountID = "865d3a4d-3803-486d-bdf3-a85679d9fad2"
	provision := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, accountID) {
			_ = json.NewEncoder(w).Encode(metaAPIAccount{ID: accountID, State: "DEPLOYED"})
			return
		}
		http.NotFound(w, r)
	}))
	defer provision.Close()

	client := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "account-information"):
			_ = json.NewEncoder(w).Encode(metaAPIAccountInfo{Balance: 10000, Equity: 10050})
		case strings.Contains(r.URL.Path, "positions"):
			_ = json.NewEncoder(w).Encode([]metaAPIPosition{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer client.Close()

	creds := Credentials{
		"metaapi_token":        "tok",
		"metaapi_account_id": accountID,
	}
	b := newTestMetaAPIBroker(t, provision.URL, client.URL, creds)

	bal, err := b.GetBalance()
	if err != nil {
		t.Fatal(err)
	}
	if bal != 10000 {
		t.Fatalf("balance = %v", bal)
	}
}

func TestMetaAPIBroker_FindExistingByLogin(t *testing.T) {
	const accountID = "acc-existing-uuid-1234567890ab"
	creds := derivDemoCreds()

	provision := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/users/current/accounts":
			_ = json.NewEncoder(w).Encode([]metaAPIAccount{{
				ID:     accountID,
				Login:  creds["login"],
				Server: creds["server"],
				State:  "DEPLOYED",
			}})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, accountID):
			_ = json.NewEncoder(w).Encode(metaAPIAccount{ID: accountID, State: "DEPLOYED"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer provision.Close()

	client := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "account-information") {
			_ = json.NewEncoder(w).Encode(metaAPIAccountInfo{Balance: 5000, Equity: 5000})
			return
		}
		http.NotFound(w, r)
	}))
	defer client.Close()

	b := newTestMetaAPIBroker(t, provision.URL, client.URL, creds)
	bal, err := b.GetBalance()
	if err != nil {
		t.Fatal(err)
	}
	if bal != 5000 {
		t.Fatalf("balance = %v", bal)
	}
}

func TestMetaAPIBroker_CreateAccountAndTrade(t *testing.T) {
	const accountID = "new-acc-uuid-1234567890123456"
	creds := derivDemoCreds()
	var created bool

	provision := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/users/current/accounts":
			if !created {
				_ = json.NewEncoder(w).Encode([]metaAPIAccount{})
				return
			}
			_ = json.NewEncoder(w).Encode([]metaAPIAccount{{ID: accountID, Login: creds["login"], Server: creds["server"], State: "DEPLOYED"}})
		case r.Method == http.MethodPost && r.URL.Path == "/users/current/accounts":
			created = true
			_ = json.NewEncoder(w).Encode(metaAPIAccount{ID: accountID, State: "DEPLOYING"})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, accountID):
			_ = json.NewEncoder(w).Encode(metaAPIAccount{ID: accountID, State: "DEPLOYED"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer provision.Close()

	client := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "account-information"):
			_ = json.NewEncoder(w).Encode(metaAPIAccountInfo{Balance: 10000, Equity: 10000})
		case strings.Contains(r.URL.Path, "positions"):
			_ = json.NewEncoder(w).Encode([]metaAPIPosition{{
				ID: "pos-1", Symbol: "frxEURUSD", Type: "POSITION_TYPE_BUY",
				Volume: 0.01, OpenPrice: 1.08, CurrentPrice: 1.081, Profit: 1,
			}})
		case strings.Contains(r.URL.Path, "trade") && r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(metaAPITradeResponse{
				NumericCode: 10009, StringCode: "TRADE_RETCODE_DONE", PositionID: "pos-2",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer client.Close()

	b := newTestMetaAPIBroker(t, provision.URL, client.URL, creds)
	b.waitDeployedFn = func(string) error { return nil }

	bal, err := b.GetBalance()
	if err != nil {
		t.Fatal(err)
	}
	if bal != 10000 {
		t.Fatalf("balance = %v", bal)
	}

	positions, err := b.GetOpenPositions()
	if err != nil || len(positions) != 1 || positions[0].Symbol != "EURUSD" {
		t.Fatalf("positions = %+v err=%v", positions, err)
	}

	pos, err := b.OpenMarketOrder("EURUSD", "BUY", 0.01, 1.07, 1.09)
	if err != nil {
		t.Fatal(err)
	}
	if pos.Symbol != "EURUSD" {
		t.Fatalf("position symbol = %s", pos.Symbol)
	}

	if err := b.ClosePosition("pos-1"); err != nil {
		t.Fatal(err)
	}
}

func TestMetaAPIBroker_OpenMarketOrderSymbolFallback(t *testing.T) {
	const accountID = "865d3a4d-3803-486d-bdf3-a85679d9fad2"
	var tradeBody map[string]interface{}

	provision := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(metaAPIAccount{ID: accountID, State: "DEPLOYED"})
	}))
	defer provision.Close()

	client := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "trade") && r.Method == http.MethodPost {
			_ = json.NewDecoder(r.Body).Decode(&tradeBody)
			if tradeBody["symbol"] == "frxEURUSD" {
				_ = json.NewEncoder(w).Encode(metaAPITradeResponse{NumericCode: 10006, Message: "reject"})
				return
			}
			_ = json.NewEncoder(w).Encode(metaAPITradeResponse{NumericCode: 10009, StringCode: "TRADE_RETCODE_DONE", PositionID: "p1"})
			return
		}
		if strings.Contains(r.URL.Path, "positions") {
			_ = json.NewEncoder(w).Encode([]metaAPIPosition{})
			return
		}
		if strings.Contains(r.URL.Path, "account-information") {
			_ = json.NewEncoder(w).Encode(metaAPIAccountInfo{Balance: 1000})
			return
		}
		http.NotFound(w, r)
	}))
	defer client.Close()

	creds := Credentials{
		"metaapi_token":        "tok",
		"metaapi_account_id": accountID,
	}
	b := newTestMetaAPIBroker(t, provision.URL, client.URL, creds)
	b.waitDeployedFn = func(string) error { return nil }

	_, err := b.OpenMarketOrder("EURUSD", "BUY", 0.01, 1.07, 1.09)
	if err != nil {
		t.Fatal(err)
	}
	if tradeBody["symbol"] != "EURUSD" {
		t.Fatalf("expected fallback to EURUSD, got %v", tradeBody["symbol"])
	}
}
