package filter

import "time"

// Category groups filters for reporting and UI layers.
type Category string

const (
	CategoryMarket    Category = "market"
	CategoryTechnical Category = "technical"
	CategorySetup     Category = "setup"
	CategoryRisk      Category = "risk"
	CategoryPlatform  Category = "platform"
)

// Status is the outcome of one filter step.
type Status string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"
	StatusWarn Status = "warn"
	StatusSkip Status = "skip"
)

// Step is one auditable gate in the pipeline.
type Step struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    Category               `json:"category"`
	Description string                 `json:"description,omitempty"`
	Status      Status                 `json:"status"`
	Message     string                 `json:"message"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
}

// Layer groups steps for stacked UI (market → technical → setup → risk).
type Layer struct {
	ID    Category `json:"id"`
	Title string   `json:"title"`
	Steps []Step   `json:"steps"`
}

// Report is a full filter audit for one symbol evaluation.
type Report struct {
	Symbol       string    `json:"symbol"`
	DataSource   string    `json:"data_source"`
	GeneratedAt  time.Time `json:"generated_at"`
	Verdict      Status    `json:"verdict"`
	Summary      string    `json:"summary"`
	Trend        string    `json:"trend,omitempty"`
	Layers       []Layer   `json:"layers"`
	Qualified    bool      `json:"qualified"`
	SignalSide   string    `json:"signal_side,omitempty"`
	Strength     float64   `json:"strength,omitempty"`
	RiskReward   float64   `json:"risk_reward,omitempty"`
	SetupReason  string    `json:"setup_reason,omitempty"`
	BarCount     int       `json:"bar_count,omitempty"`
	MinBars      int       `json:"min_bars,omitempty"`
	LiveReady    bool      `json:"live_ready"`
}

// CatalogEntry documents a filter for operators and API consumers.
type CatalogEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    Category `json:"category"`
	Description string   `json:"description"`
	Threshold   string   `json:"threshold,omitempty"`
}
