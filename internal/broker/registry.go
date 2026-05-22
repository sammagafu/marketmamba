package broker

// BrokerType describes a supported broker integration for the web UI.
type BrokerType struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Status      string   `json:"status"` // live, coming_soon
	Fields      []Field  `json:"fields"`
}

type Field struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // text, password, url, boolean
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
}

func SupportedBrokerTypes() []BrokerType {
	return []BrokerType{
		{
			ID:          "mock",
			Name:        "Mock (Demo)",
			Description: "Simulated account for testing. No real money.",
			Status:      "live",
			Fields: []Field{
				{Key: "initial_balance", Label: "Starting balance (USD)", Type: "text", Required: false, Placeholder: "10000"},
			},
		},
		{
			ID:          "oanda",
			Name:        "OANDA",
			Description: "OANDA v20 REST API (practice or live). Not available in all countries — use Mock or MetaAPI if signup is blocked.",
			Status:      "live",
			Fields: []Field{
				{Key: "api_token", Label: "API Token", Type: "password", Required: true},
				{Key: "account_id", Label: "Account ID", Type: "text", Required: true},
				{Key: "practice", Label: "Practice account (fxTrade Practice)", Type: "boolean", Required: false},
			},
		},
		{
			ID:          "metaapi",
			Name:        "MetaAPI (MT4 / MT5)",
			Description: "Your MT broker via MetaAPI — Deriv, Exness, IC Markets, etc. Token from app.metaapi.cloud.",
			Status:      "live",
			Fields: []Field{
				{Key: "metaapi_token", Label: "MetaAPI token", Type: "password", Required: true, Placeholder: "From app.metaapi.cloud → API access"},
				{Key: "login", Label: "MT login (account number)", Type: "text", Required: true, Placeholder: "201620473"},
				{Key: "password", Label: "MT password", Type: "password", Required: true},
				{Key: "server", Label: "MT server name", Type: "text", Required: true, Placeholder: "Deriv-Demo"},
				{Key: "platform", Label: "Platform (mt5 or mt4)", Type: "text", Required: false, Placeholder: "mt5"},
				{Key: "metaapi_account_id", Label: "MetaAPI account id (optional)", Type: "text", Required: false, Placeholder: "UUID if already linked in MetaAPI"},
				{Key: "region", Label: "MetaAPI region", Type: "text", Required: false, Placeholder: "new-york"},
				{Key: "keywords", Label: "Broker keywords (optional)", Type: "text", Required: false, Placeholder: "Deriv.com Limited"},
			},
		},
		{
			ID:          "alpaca",
			Name:        "Alpaca",
			Description: "Alpaca Markets API (forex via supported pairs).",
			Status:      "coming_soon",
			Fields: []Field{
				{Key: "api_key", Label: "API Key", Type: "password", Required: true},
				{Key: "api_secret", Label: "API Secret", Type: "password", Required: true},
			},
		},
		{
			ID:          "custom",
			Name:        "Custom REST",
			Description: "Your own broker adapter HTTP endpoint (advanced).",
			Status:      "coming_soon",
			Fields: []Field{
				{Key: "base_url", Label: "Base URL", Type: "url", Required: true, Placeholder: "https://your-bridge.example.com"},
				{Key: "api_key", Label: "API Key", Type: "password", Required: true},
			},
		},
	}
}

func IsLiveProvider(provider string) bool {
	switch provider {
	case "mock", "oanda", "metaapi":
		return true
	default:
		return false
	}
}
