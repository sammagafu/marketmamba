package config

import (
	"os"
	"strings"
	"sync"
)

const (
	AssetPhaseBitcoin = "bitcoin"
	AssetPhaseFull    = "full"

	PublicPhaseCommunityLaunch = "community_launch"
	PublicPhaseFull            = "full"
)

// CommunityPhaseRuntime holds the cached paid-subscriber count (internal unlock only).
type CommunityPhaseRuntime struct {
	mu        sync.RWMutex
	paidCount int
}

func NewCommunityPhaseRuntime() *CommunityPhaseRuntime {
	return &CommunityPhaseRuntime{}
}

func (r *CommunityPhaseRuntime) SetPaidCount(n int) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.paidCount = n
	r.mu.Unlock()
}

func (r *CommunityPhaseRuntime) PaidCount() int {
	if r == nil {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.paidCount
}

// CommunityPhaseCopy is user-facing messaging (no subscriber counts).
type CommunityPhaseCopy struct {
	AssetPhase            string
	AssetPhaseUnlocked    bool
	CommunityPhaseMessage string
	CommunityLockedHint   string
	CommunityUnlockMessage string
	AITrainingNote        string
}

// IsFullAssetCatalog reports whether forex/indexes are live for everyone.
func IsFullAssetCatalog(c *Config, paidCount int) bool {
	if c == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("FORCE_FULL_ASSETS")), "true") {
		return true
	}
	phase := strings.ToLower(strings.TrimSpace(c.App.AssetPhase))
	if phase == "" {
		phase = AssetPhaseBitcoin
	}
	if phase == AssetPhaseFull {
		return true
	}
	min := c.App.UnlockMinPaidSubscribers
	if min <= 0 {
		min = 100
	}
	return paidCount >= min
}

func (c *Config) communityCopy(unlocked bool) CommunityPhaseCopy {
	phase := PublicPhaseCommunityLaunch
	if unlocked {
		phase = PublicPhaseFull
	}
	msg := c.App.CommunityPhaseMessage
	if unlocked && strings.TrimSpace(c.App.CommunityUnlockMessage) != "" {
		msg = c.App.CommunityUnlockMessage
	}
	return CommunityPhaseCopy{
		AssetPhase:             phase,
		AssetPhaseUnlocked:     unlocked,
		CommunityPhaseMessage:  msg,
		CommunityLockedHint:    c.App.CommunityLockedHint,
		CommunityUnlockMessage: c.App.CommunityUnlockMessage,
		AITrainingNote:         c.App.AITrainingNote,
	}
}

// CommunityPhasePublic returns API-safe community fields (no counts).
func (c *Config) CommunityPhasePublic() CommunityPhaseCopy {
	return c.communityCopy(c.IsFullAssetCatalog())
}

func (c *Config) IsFullAssetCatalog() bool {
	return IsFullAssetCatalog(c, c.paidCountForPhase())
}

func (c *Config) paidCountForPhase() int {
	if c == nil || c.Phase == nil {
		return 0
	}
	return c.Phase.PaidCount()
}

// PhasedSignalCatalog returns forex, indexes, crypto for the current community phase.
func (c *Config) PhasedSignalCatalog() (forex, indexes, crypto []string) {
	if c == nil {
		return nil, nil, []string{"BTCUSD", "ETHUSD"}
	}
	fx, idx, cry := c.signalCatalogFull()
	if c.IsFullAssetCatalog() {
		return fx, idx, cry
	}
	if len(cry) == 0 {
		cry = []string{"BTCUSD", "ETHUSD"}
	}
	return nil, nil, cry
}

// SignalCatalogFull returns the unrestricted symbol catalog (for locked UI previews).
func (c *Config) SignalCatalogFull() (forex, indexes, crypto []string) {
	return c.signalCatalogFull()
}

func (c *Config) signalCatalogFull() (forex, indexes, crypto []string) {
	if c == nil {
		return []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "EURJPY"},
			[]string{"US500", "USTEC", "GER40", "UK100", "VOL75"},
			[]string{"BTCUSD", "ETHUSD"}
	}
	forexDef := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "EURJPY"}
	indexDef := []string{"US500", "USTEC", "GER40", "UK100", "VOL75"}
	cryptoDef := []string{"BTCUSD", "ETHUSD"}
	forex = parseSymbolCSV(getEnv("SIGNAL_FOREX_SYMBOLS", ""), forexDef)
	indexes = parseSymbolCSV(getEnv("SIGNAL_INDEX_SYMBOLS", ""), indexDef)
	crypto = parseSymbolCSV(getEnv("SIGNAL_CRYPTO_SYMBOLS", ""), cryptoDef)
	if getEnv("SIGNAL_FOREX_SYMBOLS", "") == "" && getEnv("SIGNAL_INDEX_SYMBOLS", "") == "" &&
		getEnv("SIGNAL_CRYPTO_SYMBOLS", "") == "" && len(c.App.SignalSymbols) > 0 {
		for _, sym := range c.App.SignalSymbols {
			sym = strings.ToUpper(strings.TrimSpace(sym))
			switch sym {
			case "BTCUSD", "ETHUSD":
				if !containsSym(crypto, sym) {
					crypto = append(crypto, sym)
				}
			case "US500", "USTEC", "GER40", "UK100", "VOL75":
				if !containsSym(indexes, sym) {
					indexes = append(indexes, sym)
				}
			default:
				if !containsSym(forex, sym) {
					forex = append(forex, sym)
				}
			}
		}
	}
	return forex, indexes, crypto
}
