package tier

import "fmt"

// PublicPlan is a user-facing subscription package for pricing UI.
type PublicPlan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceLabel  string   `json:"price_label"`
	PriceUSDT   float64  `json:"price_usdt,omitempty"`
	ContactOnly bool     `json:"contact_only"`
	Recommended bool     `json:"recommended"`
	Limits      Limits   `json:"limits"`
	Features    []string `json:"features"`
}

// PublicPackages returns display-ready plans (trial, monthly, pro). Manual/VIP omitted from public list.
func PublicPackages(monthlyPriceUSDT float64, trialDays int) []PublicPlan {
	if monthlyPriceUSDT <= 0 {
		monthlyPriceUSDT = 10
	}
	if trialDays <= 0 {
		trialDays = 5
	}
	trial := ForPlan("trial")
	monthly := ForPlan("monthly")
	pro := ForPlan("pro")

	return []PublicPlan{
		{
			ID:          "trial",
			Name:        "Free trial",
			Description: "Try controlled automation on demo or live broker.",
			PriceLabel:  fmt.Sprintf("%d days · $0", trialDays),
			ContactOnly: false,
			Recommended: false,
			Limits:      trial,
			Features:    planFeatures(trial),
		},
		{
			ID:          "monthly",
			Name:        "Monthly",
			Description: "Full automation for active traders. Billed in USDT via Binance.",
			PriceLabel:  fmt.Sprintf("%.0f USDT / month", monthlyPriceUSDT),
			PriceUSDT:   monthlyPriceUSDT,
			ContactOnly: false,
			Recommended: true,
			Limits:      monthly,
			Features:    planFeatures(monthly),
		},
		{
			ID:          "pro",
			Name:        "Pro",
			Description: "Higher limits, multiple brokers, priority support.",
			PriceLabel:  "Contact us",
			ContactOnly: true,
			Recommended: false,
			Limits:      pro,
			Features:    planFeatures(pro),
		},
	}
}

func planFeatures(l Limits) []string {
	return []string{
		fmt.Sprintf("%d broker account%s", l.MaxBrokerAccounts, plural(l.MaxBrokerAccounts)),
		fmt.Sprintf("%d signals per month", l.MaxSignalsPerPeriod),
		fmt.Sprintf("%d long + %d short trades / month", l.MaxLongTrades, l.MaxShortTrades),
		"Forex, indexes & crypto signal types",
		"Telegram Mini App dashboard",
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
