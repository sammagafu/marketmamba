package tier

import "strings"

// Limits defines per-billing-period quotas for a subscription plan.
type Limits struct {
	Plan                string `json:"plan"`
	MaxBrokerAccounts   int    `json:"max_broker_accounts"`
	MaxSignalsPerPeriod int    `json:"max_signals_per_period"`
	MaxLongTrades       int    `json:"max_long_trades"`
	MaxShortTrades      int    `json:"max_short_trades"`
}

// Usage tracks consumption for the current period.
type Usage struct {
	PeriodStart      string `json:"period_start"`
	SignalsReceived  int    `json:"signals_received"`
	LongTrades       int    `json:"long_trades"`
	ShortTrades      int    `json:"short_trades"`
	BrokerAccounts   int    `json:"broker_accounts"`
}

// Snapshot combines limits and usage for API responses.
type Snapshot struct {
	Limits Limits `json:"limits"`
	Usage  Usage  `json:"usage"`
}

var planLimits = map[string]Limits{
	"trial": {
		Plan: "trial", MaxBrokerAccounts: 1, MaxSignalsPerPeriod: 30,
		MaxLongTrades: 5, MaxShortTrades: 5,
	},
	"monthly": {
		Plan: "monthly", MaxBrokerAccounts: 2, MaxSignalsPerPeriod: 200,
		MaxLongTrades: 30, MaxShortTrades: 30,
	},
	"pro": {
		Plan: "pro", MaxBrokerAccounts: 5, MaxSignalsPerPeriod: 1000,
		MaxLongTrades: 100, MaxShortTrades: 100,
	},
	"manual": {
		Plan: "manual", MaxBrokerAccounts: 10, MaxSignalsPerPeriod: 10000,
		MaxLongTrades: 1000, MaxShortTrades: 1000,
	},
}

// ForPlan returns limits for a subscription plan name (unknown → trial).
func ForPlan(plan string) Limits {
	plan = strings.ToLower(strings.TrimSpace(plan))
	if l, ok := planLimits[plan]; ok {
		return l
	}
	return planLimits["trial"]
}

// AllPlans returns tier definitions for docs/admin UI.
func AllPlans() []Limits {
	order := []string{"trial", "monthly", "pro", "manual"}
	out := make([]Limits, 0, len(order))
	for _, p := range order {
		out = append(out, planLimits[p])
	}
	return out
}
