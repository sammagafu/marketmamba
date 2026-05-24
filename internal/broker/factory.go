package broker

import (
	"encoding/json"
	"fmt"
)

// Credentials is a flexible map of broker-specific secrets/settings.
type Credentials map[string]string

// NewFromProvider constructs a Broker for a technical adapter id (mock, oanda, metaapi).
func NewFromProvider(provider string, creds Credentials) (Broker, error) {
	a, ok := getAdapter(provider)
	if !ok {
		return nil, fmt.Errorf("unknown broker provider: %s", provider)
	}
	if a.Status != "live" {
		return nil, fmt.Errorf("broker %q is not available yet — use mock (demo)", provider)
	}
	return a.New(creds)
}

func ParseCredentialsJSON(raw string) (Credentials, error) {
	var creds Credentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, err
	}
	return creds, nil
}

// IsLiveProvider reports whether users may save this technical provider.
func IsLiveProvider(provider string) bool {
	a, ok := getAdapter(provider)
	return ok && a.Status == "live"
}
