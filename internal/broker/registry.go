package broker

// BrokerType describes a supported broker integration for the web UI (legacy technical list).
type BrokerType struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"` // live, coming_soon
	Fields      []Field `json:"fields"`
}

type Field struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // text, password, url, boolean
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
}

// SupportedBrokerTypes returns technical adapters for backward-compatible API clients.
func SupportedBrokerTypes() []BrokerType {
	out := make([]BrokerType, 0)
	for _, a := range ListAdapters() {
		bt := BrokerType{
			ID:     a.ID,
			Name:   a.Name,
			Status: a.Status,
		}
		if brand, ok := BrandByID(a.ID); ok {
			bt.Description = brand.Description
			bt.Fields = brand.Fields
		} else {
			switch a.ID {
			case "mock":
				bt.Description = "Simulated account for testing. No real money."
				bt.Fields = []Field{
					{Key: "initial_balance", Label: "Starting balance (USD)", Type: "text", Required: false, Placeholder: "10000"},
				}
			case "oanda":
				bt.Description = "OANDA v20 REST API (practice or live)."
				bt.Fields = []Field{
					{Key: "api_token", Label: "API Token", Type: "password", Required: true},
					{Key: "account_id", Label: "Account ID", Type: "text", Required: true},
					{Key: "practice", Label: "Practice account", Type: "boolean", Required: false},
				}
			case "metaapi":
				bt.Description = "MT4/MT5 via MetaAPI — Deriv, Exness, Tickmill, etc."
				bt.Fields = metaAPIBrandFields("Deriv-Demo")
			case "alpaca":
				bt.Description = "Alpaca Markets API."
				bt.Fields = []Field{
					{Key: "api_key", Label: "API Key", Type: "password", Required: true},
					{Key: "api_secret", Label: "API Secret", Type: "password", Required: true},
				}
			case "custom":
				bt.Description = "Your own broker adapter HTTP endpoint."
				bt.Fields = []Field{
					{Key: "base_url", Label: "Base URL", Type: "url", Required: true},
					{Key: "api_key", Label: "API Key", Type: "password", Required: true},
				}
			}
		}
		out = append(out, bt)
	}
	return out
}
