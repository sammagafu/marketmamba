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
			Fields:      []Field{},
		},
		{
			ID:          "oanda",
			Name:        "OANDA",
			Description: "OANDA v20 REST API (practice or live).",
			Status:      "coming_soon",
			Fields: []Field{
				{Key: "api_token", Label: "API Token", Type: "password", Required: true},
				{Key: "account_id", Label: "Account ID", Type: "text", Required: true},
				{Key: "practice", Label: "Practice account", Type: "boolean", Required: false},
			},
		},
		{
			ID:          "metaapi",
			Name:        "MetaAPI (MT4 / MT5)",
			Description: "Connect MetaTrader via MetaAPI cloud bridge.",
			Status:      "coming_soon",
			Fields: []Field{
				{Key: "token", Label: "MetaAPI Token", Type: "password", Required: true},
				{Key: "account_id", Label: "Account ID", Type: "text", Required: true},
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
	return provider == "mock"
}
