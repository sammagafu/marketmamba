package broker

// BrokerCapabilities describes optional features and lot constraints per adapter.
type BrokerCapabilities struct {
	SupportsModifySL   bool
	SupportsModifyTP   bool
	MinLot           float64
	LotStep          float64
	RequiresMTBridge bool
}

func DefaultCapabilities() BrokerCapabilities {
	return BrokerCapabilities{
		SupportsModifySL: true,
		SupportsModifyTP: true,
		MinLot:           0.01,
		LotStep:          0.01,
	}
}
