package broker

import (
	"fmt"
	"strings"
)

// Brand is a user-facing broker choice (Deriv, Exness) mapped to a technical adapter.
type Brand struct {
	ID               string            `json:"id"`
	DisplayName      string            `json:"display_name"`
	AdapterID        string            `json:"adapter_id"`
	Status           string            `json:"status"`
	Description      string            `json:"description"`
	Fields           []Field           `json:"fields"`
	CredentialPreset map[string]string `json:"credential_preset,omitempty"`
	ServerExamples   []string          `json:"server_examples,omitempty"`
	Warnings         []string          `json:"warnings,omitempty"`
	DocsURL          string            `json:"docs_url,omitempty"`
	HelpURL          string            `json:"help_url,omitempty"`
	UsesMetaAPI      bool              `json:"uses_metaapi"` // MT4/MT5 via MetaAPI cloud bridge
}

var allBrands = []Brand{
	{
		ID:          "mock",
		DisplayName: "Demo (Mock)",
		AdapterID:   "mock",
		Status:      "live",
		Description: "Simulated $10,000 account for testing. No real money.",
		Fields: []Field{
			{Key: "initial_balance", Label: "Starting balance (USD)", Type: "text", Required: false, Placeholder: "10000"},
		},
		Warnings: []string{"Market Mamba is not a broker. Demo only — no live funds."},
	},
	{
		ID:          "oanda",
		DisplayName: "OANDA",
		AdapterID:   "oanda",
		Status:      "live",
		Description: "OANDA v20 REST API (practice or live).",
		Fields: []Field{
			{Key: "api_token", Label: "API Token", Type: "password", Required: true},
			{Key: "account_id", Label: "Account ID", Type: "text", Required: true},
			{Key: "practice", Label: "Practice account", Type: "boolean", Required: false},
		},
		HelpURL:  "https://www.oanda.com/",
		Warnings: []string{"OANDA is not available in all countries. Use Demo or MetaAPI if signup is blocked."},
	},
	{
		ID:          "deriv",
		DisplayName: "Deriv",
		AdapterID:   "metaapi",
		UsesMetaAPI: true,
		Status:      "live",
		Description: "Connect your Deriv MT account via MetaAPI (MT4/MT5).",
		CredentialPreset: map[string]string{
			"platform": "mt5",
			"keywords": "Deriv.com Limited",
		},
		ServerExamples: []string{"Deriv-Demo", "Deriv-Server", "Deriv-Server-02"},
		Fields:         metaAPIBrandFields("Deriv-Demo"),
		HelpURL:        "https://app.metaapi.cloud/",
		DocsURL:        "/docs/BROKER_CONNECT.md#deriv",
		Warnings: []string{
			"Market Mamba is not a broker — you connect your own Deriv account.",
			"First connection may take 1–3 minutes while MetaAPI deploys your MT account.",
			"Synthetic indices on MT may differ from Deriv app API.",
		},
	},
	{
		ID:          "exness",
		DisplayName: "Exness",
		AdapterID:   "metaapi",
		UsesMetaAPI: true,
		Status:      "live",
		Description: "Connect your Exness MT account via MetaAPI (MT4/MT5).",
		CredentialPreset: map[string]string{
			"platform": "mt5",
		},
		ServerExamples: []string{"Exness-MT5Trial", "Exness-MT5Real", "Exness-Trial"},
		Fields:         metaAPIBrandFields("Exness-MT5Trial"),
		HelpURL:        "https://app.metaapi.cloud/",
		DocsURL:        "/docs/BROKER_CONNECT.md#exness",
		Warnings: []string{
			"Market Mamba is not a broker — you connect your own Exness account.",
			"Use the exact MT server name from Exness → My accounts.",
		},
	},
	{
		ID:          "tickmill",
		DisplayName: "Tickmill",
		AdapterID:   "metaapi",
		UsesMetaAPI: true,
		Status:      "live",
		Description: "Connect your Tickmill MT account via MetaAPI (MT4/MT5).",
		CredentialPreset: map[string]string{
			"platform": "mt5",
		},
		ServerExamples: []string{"Tickmill-Demo", "Tickmill-Live", "TickmillUK-Demo"},
		Fields:         metaAPIBrandFields("Tickmill-Demo"),
		HelpURL:        "https://app.metaapi.cloud/",
		DocsURL:        "/docs/BROKER_CONNECT.md#tickmill",
		Warnings: []string{
			"Market Mamba is not a broker — you connect your own Tickmill account.",
			"Copy the MT server name from Tickmill client area.",
		},
	},
	{
		ID:          "any_mt",
		DisplayName: "Any MT broker",
		AdapterID:   "metaapi",
		UsesMetaAPI: true,
		Status:      "live",
		Description: "Any MT4/MT5 broker supported by MetaAPI — enter your broker's server name.",
		CredentialPreset: map[string]string{
			"platform": "mt5",
		},
		ServerExamples: []string{"Deriv-Demo", "Exness-MT5Trial", "Tickmill-Demo", "XMGlobal-MT5", "YourBroker-Server"},
		Fields:         metaAPIBrandFields("YourBroker-Demo"),
		HelpURL:        "https://app.metaapi.cloud/",
		Warnings: []string{
			"Use the exact MT server name from your broker (copy from MT4/MT5 or broker website).",
			"Not sure? Pick Deriv, Exness, or Tickmill above if that is your broker.",
		},
	},
	{
		ID:          "icmarkets",
		DisplayName: "IC Markets (MT)",
		AdapterID:   "metaapi",
		UsesMetaAPI: true,
		Status:      "live",
		Description: "IC Markets via MetaAPI (or use “Any MT broker” for other servers).",
		CredentialPreset: map[string]string{
			"platform": "mt5",
		},
		ServerExamples: []string{"ICMarketsSC-Demo", "ICMarketsSC-MT5"},
		Fields:         metaAPIBrandFields("ICMarketsSC-Demo"),
		HelpURL:        "https://app.metaapi.cloud/",
		Warnings:       []string{"Use your broker's exact MT server name."},
	},
}

func metaAPIBrandFields(serverPlaceholder string) []Field {
	return metaAPIBrandFieldsOpts(serverPlaceholder, !UsesSharedMetaAPIToken())
}

func metaAPIBrandFieldsOpts(serverPlaceholder string, tokenRequired bool) []Field {
	return []Field{
		{Key: "metaapi_token", Label: "MetaAPI token", Type: "password", Required: tokenRequired, Placeholder: "From app.metaapi.cloud → API access"},
		{Key: "login", Label: "MT login (account number)", Type: "text", Required: true},
		{Key: "password", Label: "MT password", Type: "password", Required: true},
		{Key: "server", Label: "MT server name", Type: "text", Required: true, Placeholder: serverPlaceholder},
		{Key: "platform", Label: "Platform (mt5 or mt4)", Type: "text", Required: false, Placeholder: "mt5"},
		{Key: "metaapi_account_id", Label: "MetaAPI account id (optional)", Type: "text", Required: false},
		{Key: "region", Label: "MetaAPI region", Type: "text", Required: false, Placeholder: "new-york"},
		{Key: "keywords", Label: "Broker keywords (optional)", Type: "text", Required: false},
	}
}

// EnabledBrandIDs is set from config (ENABLED_BROKER_BRANDS). Empty = all brands.
var EnabledBrandIDs []string

// SetEnabledBrands configures which brand IDs are visible (from env).
func SetEnabledBrands(ids []string) {
	EnabledBrandIDs = ids
}

func brandEnabled(id string) bool {
	if len(EnabledBrandIDs) == 0 {
		return true
	}
	id = strings.ToLower(strings.TrimSpace(id))
	for _, e := range EnabledBrandIDs {
		if strings.EqualFold(strings.TrimSpace(e), id) {
			return true
		}
	}
	return false
}

// BrandByID returns a brand definition.
func BrandByID(id string) (*Brand, bool) {
	for i := range allBrands {
		if allBrands[i].ID == id {
			b := allBrands[i]
			return &b, true
		}
	}
	return nil, false
}

// SupportedBrands returns brands filtered by ENABLED_BROKER_BRANDS and adapter status.
func SupportedBrands() []Brand {
	out := make([]Brand, 0, len(allBrands))
	for _, b := range allBrands {
		if !brandEnabled(b.ID) {
			continue
		}
		if a, ok := getAdapter(b.AdapterID); ok && a.Status == "disabled" {
			continue
		}
		if b.Status == "live" {
			if a, ok := getAdapter(b.AdapterID); ok && a.Status != "live" && b.AdapterID != "mock" {
				b.Status = a.Status
			}
		}
		b.Fields = fieldsForBrand(b)
		out = append(out, b)
	}
	return out
}

func fieldsForBrand(b Brand) []Field {
	if !b.UsesMetaAPI || b.AdapterID != "metaapi" {
		return b.Fields
	}
	ph := "YourBroker-Demo"
	for _, f := range b.Fields {
		if f.Key == "server" && f.Placeholder != "" {
			ph = f.Placeholder
			break
		}
	}
	return metaAPIBrandFieldsOpts(ph, !UsesSharedMetaAPIToken())
}

// ResolveBrandConnection maps brand_id + user credentials to provider + merged credentials + label.
func ResolveBrandConnection(brandID, label string, creds Credentials) (provider string, merged Credentials, outLabel string, err error) {
	brand, ok := BrandByID(brandID)
	if !ok {
		return "", nil, "", fmt.Errorf("unknown broker brand: %s", brandID)
	}
	if !brandEnabled(brandID) {
		return "", nil, "", fmt.Errorf("broker %s is not enabled on this server", brand.DisplayName)
	}
	merged = Credentials{}
	for k, v := range brand.CredentialPreset {
		merged[k] = v
	}
	for k, v := range creds {
		if strings.TrimSpace(v) != "" {
			merged[k] = v
		}
	}
	merged = ApplySharedMetaAPIToken(merged)
	merged["brand_id"] = brandID
	provider = brand.AdapterID
	outLabel = label
	if outLabel == "" {
		outLabel = brand.DisplayName
		if s := credsHintServer(merged); s != "" {
			outLabel = brand.DisplayName + " · " + s
		}
	}
	return provider, merged, outLabel, nil
}

// MetaAPIBrands returns enabled brands that connect via the MetaAPI MT bridge.
func MetaAPIBrands() []Brand {
	var out []Brand
	for _, b := range SupportedBrands() {
		if b.UsesMetaAPI && b.AdapterID == "metaapi" {
			out = append(out, b)
		}
	}
	return out
}
