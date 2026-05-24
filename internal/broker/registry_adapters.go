package broker

import (
	"fmt"
	"strconv"
	"sync"
)

// Adapter registers a technical broker integration (mock, oanda, metaapi).
type Adapter struct {
	ID           string
	Name         string
	Status       string // live, coming_soon, disabled
	New          func(Credentials) (Broker, error)
	Validate     func(Credentials) error
	Capabilities BrokerCapabilities
}

var (
	adaptersMu sync.RWMutex
	adapters   = map[string]*Adapter{}
)

// Register adds or replaces a broker adapter (typically from init()).
func Register(a *Adapter) {
	if a == nil || a.ID == "" {
		return
	}
	adaptersMu.Lock()
	adapters[a.ID] = a
	adaptersMu.Unlock()
}

func getAdapter(id string) (*Adapter, bool) {
	adaptersMu.RLock()
	a, ok := adapters[id]
	adaptersMu.RUnlock()
	return a, ok
}

// ListAdapters returns registered adapters sorted by ID.
func ListAdapters() []*Adapter {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	out := make([]*Adapter, 0, len(adapters))
	for _, a := range adapters {
		out = append(out, a)
	}
	// simple sort by id
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].ID < out[i].ID {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// AdapterCapabilities returns capabilities for a provider id.
func AdapterCapabilities(provider string) BrokerCapabilities {
	if a, ok := getAdapter(provider); ok {
		return a.Capabilities
	}
	return DefaultCapabilities()
}

func init() {
	Register(&Adapter{
		ID:     "mock",
		Name:   "Mock (Demo)",
		Status: "live",
		New: func(creds Credentials) (Broker, error) {
			bal := 10000.0
			if creds != nil {
				if s := creds["initial_balance"]; s != "" {
					if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
						bal = v
					}
				}
			}
			return NewMockBroker(bal), nil
		},
		Validate: func(creds Credentials) error { return nil },
		Capabilities: BrokerCapabilities{
			SupportsModifySL: true,
			SupportsModifyTP: true,
			MinLot:           0.01,
			LotStep:          0.01,
		},
	})

	Register(&Adapter{
		ID:     "oanda",
		Name:   "OANDA",
		Status: "live",
		New: func(creds Credentials) (Broker, error) {
			return NewOANDABroker(creds)
		},
		Validate: func(creds Credentials) error {
			if creds == nil {
				return fmt.Errorf("OANDA credentials required")
			}
			if creds["api_token"] == "" || creds["account_id"] == "" {
				return fmt.Errorf("OANDA api_token and account_id are required")
			}
			return nil
		},
		Capabilities: BrokerCapabilities{
			SupportsModifySL: false,
			SupportsModifyTP: false,
			MinLot:           0.01,
			LotStep:          0.01,
		},
	})

	Register(&Adapter{
		ID:     "metaapi",
		Name:   "MetaAPI (MT4/MT5)",
		Status: "live",
		New: func(creds Credentials) (Broker, error) {
			return NewMetaAPIBroker(creds)
		},
		Validate: ValidateMetaAPICredentials,
		Capabilities: BrokerCapabilities{
			SupportsModifySL:   true,
			SupportsModifyTP:   true,
			MinLot:             0.01,
			LotStep:            0.01,
			RequiresMTBridge:   true,
		},
	})

	Register(&Adapter{
		ID:     "alpaca",
		Name:   "Alpaca",
		Status: "coming_soon",
		New: func(creds Credentials) (Broker, error) {
			return nil, fmt.Errorf("Alpaca adapter is not implemented yet — use Mock for now")
		},
		Validate:     func(creds Credentials) error { return fmt.Errorf("Alpaca is not available yet") },
		Capabilities: DefaultCapabilities(),
	})

	Register(&Adapter{
		ID:     "custom",
		Name:   "Custom REST",
		Status: "coming_soon",
		New: func(creds Credentials) (Broker, error) {
			return nil, fmt.Errorf("custom REST adapter is not implemented yet — use Mock for now")
		},
		Validate:     func(creds Credentials) error { return fmt.Errorf("custom REST is not available yet") },
		Capabilities: DefaultCapabilities(),
	})
}
