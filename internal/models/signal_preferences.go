package models

import "time"

// SignalTypePreferences selects which asset classes receive signals and auto-trade scope.
type SignalTypePreferences struct {
	Forex   bool `json:"forex"`
	Indexes bool `json:"indexes"`
	Crypto  bool `json:"crypto"`
}

// DefaultSignalTypes enables all asset classes for new users.
func DefaultSignalTypes() SignalTypePreferences {
	return SignalTypePreferences{Forex: true, Indexes: true, Crypto: true}
}

// BitcoinPhaseSignalTypes is the community launch default (crypto only).
func BitcoinPhaseSignalTypes() SignalTypePreferences {
	return SignalTypePreferences{Forex: false, Indexes: false, Crypto: true}
}

// SignalAssetGroup describes one selectable signal type in the UI.
type SignalAssetGroup struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Symbols     []string `json:"symbols"`
	Enabled     bool     `json:"enabled"`
	Locked      bool     `json:"locked,omitempty"`
	ComingSoon  bool     `json:"coming_soon,omitempty"`
}

// UserSignalPreferencesRow is the DB shape.
type UserSignalPreferencesRow struct {
	UserID    int64     `db:"user_id"`
	Forex     bool      `db:"forex"`
	Indexes   bool      `db:"indexes"`
	Crypto    bool      `db:"crypto"`
	UpdatedAt time.Time `db:"updated_at"`
}
