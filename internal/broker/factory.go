package broker

import (
	"encoding/json"
	"fmt"
)

// Credentials is a flexible map of broker-specific secrets/settings.
type Credentials map[string]string

func NewFromProvider(provider string, creds Credentials) (Broker, error) {
	switch provider {
	case "mock":
		return NewMockBroker(10000), nil
	case "oanda":
		return nil, fmt.Errorf("OANDA adapter is not implemented yet — use Mock for now")
	case "metaapi":
		return nil, fmt.Errorf("MetaAPI adapter is not implemented yet — use Mock for now")
	case "alpaca":
		return nil, fmt.Errorf("Alpaca adapter is not implemented yet — use Mock for now")
	case "custom":
		return nil, fmt.Errorf("custom REST adapter is not implemented yet — use Mock for now")
	default:
		return nil, fmt.Errorf("unknown broker provider: %s", provider)
	}
}

func ParseCredentialsJSON(raw string) (Credentials, error) {
	var creds Credentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, err
	}
	return creds, nil
}
