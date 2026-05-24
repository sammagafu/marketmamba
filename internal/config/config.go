package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Telegram TelegramConfig
	Database DatabaseConfig
	Broker   BrokerConfig
	Risk     RiskConfig
	App      AppConfig
	Payments PaymentsConfig
	Phase    *CommunityPhaseRuntime
}

// PaymentsConfig — Binance USDT subscription (10 USDT / month after trial).
type PaymentsConfig struct {
	SubscriptionPriceUSDT float64
	SubscriptionDays      int
	BinancePayAPIKey      string
	BinancePaySecret      string
	BinancePayCertSN      string
	BinanceUID            string // manual USDT transfer to Binance account
	BinanceNetwork        string // TRC20, BEP20, etc.
	MiniAppURL            string
}

type TelegramConfig struct {
	BotToken       string
	BotClientID    string // OIDC client id (bot numeric id from token prefix)
	BotUsername    string // for Login Widget, e.g. market_mamba_bot
	LoginDomain    string // Allowed URL host in BotFather Web Login
	AllowedUserIDs []int64 // legacy; optional when public
	AdminUserIDs   []int64
}

type DatabaseConfig struct {
	URL string
}

type BrokerConfig struct {
	Provider            string
	EnabledBrokerBrands []string // empty = all brands from catalog
	MetaAPISharedToken  string   // optional — clients skip MetaAPI token field
}

type RiskConfig struct {
	MaxRiskPerTrade  float64
	MaxDailyLoss     float64
	MaxOpenTrades    int
	MaxTradesPerDay  int
	RiskRewardRatio  float64
}

type AppConfig struct {
	Environment                 string
	Port                        string
	HTTPPort                    string
	WebAPIKey                   string
	CORSOrigins                 []string
	BrokerEncryptionKey         string
	EnableWeb                   bool
	PublicMode                  bool
	SubscriptionRequired        bool
	FreeTrialDays               int
	SubscriptionContactMessage  string
	ValueProposition            string // marketing line: automation + risk + any broker
	ContactUsURL                string // Telegram, email, or support page
	ContactUsLabel              string
	WebSessionSecret            string
	WebSessionTTLDays           int
	PublicSiteURL               string // https://marketmamba.kkooapp.co.tz — Telegram Login callback origin
	SignalBroadcastEnabled      bool
	SignalBroadcastIntervalSec  int
	SignalBroadcastSymbol       string
	SignalSymbols               []string
	SignalMinStrength           float64
	// Real-time sniper decision support
	DecisionEnabled         bool
	DecisionIntervalSec     int
	SniperMinConfidence     float64
	SniperCooldownMin       int
	DecisionAdvisory        bool
	DecisionAutoExecute     bool
	MarketDataAPIKey           string // optional Twelve Data for richer live + seed bars
	AutoTradeRequiresApproval bool
	// Community launch phase (bitcoin-first; full catalog unlocks internally at threshold)
	AssetPhase                 string
	UnlockMinPaidSubscribers   int
	CommunityPhaseMessage      string
	CommunityLockedHint        string
	CommunityUnlockMessage     string
	AITrainingNote             string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken:       getEnv("TELEGRAM_BOT_TOKEN", ""),
			BotClientID:    botClientID(getEnv("TELEGRAM_BOT_CLIENT_ID", ""), getEnv("TELEGRAM_BOT_TOKEN", "")),
			BotUsername:    getEnv("TELEGRAM_BOT_USERNAME", "market_mamba_bot"),
			LoginDomain:    getEnv("TELEGRAM_LOGIN_DOMAIN", "marketmamba.kkooapp.co.tz"),
			AllowedUserIDs: parseUserIDs(getEnv("TELEGRAM_ALLOWED_USER_IDS", "")),
			AdminUserIDs:   parseAdminIDs(getEnv("TELEGRAM_ADMIN_USER_IDS", getEnv("TELEGRAM_ALLOWED_USER_IDS", ""))),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/forexbot"),
		},
		Broker: BrokerConfig{
			Provider:            getEnv("BROKER_PROVIDER", "mock"),
			EnabledBrokerBrands: parseCSV(getEnv("ENABLED_BROKER_BRANDS", "mock,deriv,exness,tickmill,any_mt")),
			MetaAPISharedToken:  getEnv("METAAPI_SHARED_TOKEN", ""),
		},
		Risk: RiskConfig{
			MaxRiskPerTrade:  parseFloat(getEnv("MAX_RISK_PER_TRADE", "0.005")),
			MaxDailyLoss:     parseFloat(getEnv("MAX_DAILY_LOSS", "0.02")),
			MaxOpenTrades:    parseInt(getEnv("MAX_OPEN_TRADES", "2")),
			MaxTradesPerDay:  parseInt(getEnv("MAX_TRADES_PER_DAY", "10")),
			RiskRewardRatio:  parseFloat(getEnv("RISK_REWARD_RATIO", "1.0")),
		},
		App: AppConfig{
			Environment:                getEnv("APP_ENV", "development"),
			Port:                       getEnv("PORT", "8080"),
			HTTPPort:                   getEnv("HTTP_PORT", "8090"),
			WebAPIKey:                  getEnv("WEB_API_KEY", ""),
			CORSOrigins:                parseCSV(getEnv("CORS_ORIGINS", "https://marketmamba.kkooapp.co.tz,http://localhost:8090,http://localhost:5173")),
			BrokerEncryptionKey:        getEnv("BROKER_ENCRYPTION_KEY", ""),
			EnableWeb:                  getEnv("ENABLE_WEB", "true") == "true",
			PublicMode:                 getEnv("PUBLIC_MODE", "true") == "true",
			SubscriptionRequired:       getEnv("SUBSCRIPTION_REQUIRED", "true") == "true",
			FreeTrialDays:              parseInt(getEnv("FREE_TRIAL_DAYS", "5")),
			SubscriptionContactMessage: getEnv("SUBSCRIPTION_CONTACT", "Pay in USDT via Binance only (no cards). Pro or team plans? Contact us on Telegram."),
			ValueProposition: getEnv("VALUE_PROPOSITION", "Automate with discipline: built-in risk limits, qualified signals, and execution on the MT broker you already use."),
			ContactUsURL:     getEnv("CONTACT_US_URL", ""),
			ContactUsLabel:   getEnv("CONTACT_US_LABEL", "Contact us"),
			WebSessionSecret:           getEnv("WEB_SESSION_SECRET", ""),
			WebSessionTTLDays:          parseInt(getEnv("WEB_SESSION_TTL_DAYS", "365")),
			PublicSiteURL:              strings.TrimRight(getEnv("PUBLIC_SITE_URL", "https://marketmamba.kkooapp.co.tz"), "/"),
			SignalBroadcastEnabled:     getEnv("SIGNAL_BROADCAST_ENABLED", "true") == "true",
			SignalBroadcastIntervalSec: parseInt(getEnv("SIGNAL_BROADCAST_INTERVAL_SEC", "300")),
			SignalBroadcastSymbol:      getEnv("SIGNAL_BROADCAST_SYMBOL", "BTCUSD"),
			SignalMinStrength:          parseFloat(getEnv("SIGNAL_MIN_STRENGTH", "0.7")),
			DecisionEnabled:            getEnv("DECISION_ENABLED", "true") == "true",
			DecisionIntervalSec:      parseInt(getEnv("DECISION_INTERVAL_SEC", "60")),
			SniperMinConfidence:      parseFloat(getEnv("SNIPER_MIN_CONFIDENCE", "0.75")),
			SniperCooldownMin:        parseInt(getEnv("SNIPER_COOLDOWN_MIN", "45")),
			DecisionAdvisory:         parseDecisionAdvisory(getEnv("DECISION_MODE", "both")),
			DecisionAutoExecute:      parseDecisionAuto(getEnv("DECISION_MODE", "both")),
			MarketDataAPIKey:           getEnv("MARKET_DATA_API_KEY", ""),
			AutoTradeRequiresApproval: getEnv("AUTO_TRADE_REQUIRES_APPROVAL", "false") == "true",
			AssetPhase: getEnv("ASSET_PHASE", "bitcoin"),
			UnlockMinPaidSubscribers: parseInt(getEnv("UNLOCK_MIN_PAID_SUBSCRIBERS", "100")),
			CommunityPhaseMessage: getEnv("COMMUNITY_PHASE_MESSAGE",
				"We're in community launch: Bitcoin and Ethereum while our AI learns precise entries. Forex, gold, and indexes open for everyone as membership grows."),
			CommunityLockedHint: getEnv("COMMUNITY_LOCKED_HINT",
				"Coming soon for the community — opens for all members as more traders join with a paid plan."),
			CommunityUnlockMessage: getEnv("COMMUNITY_UNLOCK_MESSAGE",
				"Community unlock — forex, indexes, and more pairs are now live for all members. Thank you for supporting Market Mamba."),
			AITrainingNote: getEnv("AI_TRAINING_NOTE",
				"We're training our bots with AI for more precise entries — signals improve as the community grows."),
		},
		Payments: PaymentsConfig{
			SubscriptionPriceUSDT: parseFloat(getEnv("SUBSCRIPTION_PRICE_USDT", "10")),
			SubscriptionDays:      parseInt(getEnv("SUBSCRIPTION_DAYS", "30")),
			BinancePayAPIKey:      getEnv("BINANCE_PAY_API_KEY", ""),
			BinancePaySecret:      getEnv("BINANCE_PAY_SECRET", ""),
			BinancePayCertSN:      getEnv("BINANCE_PAY_CERT_SN", ""),
			BinanceUID:            getEnv("BINANCE_PAY_UID", ""),
			BinanceNetwork:        getEnv("BINANCE_PAY_NETWORK", "TRC20"),
			MiniAppURL:            strings.TrimRight(getEnv("MINI_APP_URL", getEnv("PUBLIC_SITE_URL", "https://marketmamba.kkooapp.co.tz")), "/"),
		},
	}
	if cfg.Payments.MiniAppURL == "" {
		cfg.Payments.MiniAppURL = cfg.App.PublicSiteURL
	}
	if cfg.App.ContactUsURL == "" && cfg.Telegram.BotUsername != "" {
		cfg.App.ContactUsURL = "https://t.me/" + cfg.Telegram.BotUsername
	}
	broadcastDefault := "BTCUSD,ETHUSD"
	if cfg.App.AssetPhase == AssetPhaseFull {
		broadcastDefault = "EURUSD,BTCUSD"
	}
	cfg.App.SignalSymbols = ParseSignalSymbols(
		getEnv("SIGNAL_BROADCAST_SYMBOLS", broadcastDefault),
		cfg.App.SignalBroadcastSymbol,
	)
	cfg.Phase = NewCommunityPhaseRuntime()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if !c.App.PublicMode && len(c.Telegram.AllowedUserIDs) == 0 {
		return fmt.Errorf("TELEGRAM_ALLOWED_USER_IDS is required when PUBLIC_MODE=false")
	}
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.App.Environment == "production" {
		if len(c.App.WebAPIKey) < 16 {
			return fmt.Errorf("WEB_API_KEY must be at least 16 characters in production")
		}
		if len(c.App.WebSessionSecret) < 16 {
			return fmt.Errorf("WEB_SESSION_SECRET must be at least 16 characters in production")
		}
		if len(c.App.BrokerEncryptionKey) < 32 {
			return fmt.Errorf("BROKER_ENCRYPTION_KEY must be at least 32 characters in production")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseUserIDs(s string) []int64 {
	var ids []int64
	parts := strings.Split(strings.TrimSpace(s), ",")
	for _, part := range parts {
		if id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseAdminIDs(s string) []int64 {
	return parseUserIDs(s)
}

// SessionTTL returns how long web login sessions remain valid.
func (c *Config) SessionTTL() time.Duration {
	days := c.App.WebSessionTTLDays
	if days <= 0 {
		days = 365
	}
	return time.Duration(days) * 24 * time.Hour
}

func parseDecisionAdvisory(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "auto", "autostart", "execute":
		return false
	default:
		return true
	}
}

func parseDecisionAuto(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "advisory", "alert", "signals":
		return false
	default:
		return true
	}
}

func (c *Config) DecisionInterval() time.Duration {
	sec := c.App.DecisionIntervalSec
	if sec < 30 {
		sec = 30
	}
	return time.Duration(sec) * time.Second
}

func (c *Config) SniperCooldown() time.Duration {
	min := c.App.SniperCooldownMin
	if min < 5 {
		min = 45
	}
	return time.Duration(min) * time.Minute
}

func (c *Config) SignalBroadcastInterval() time.Duration {
	sec := c.App.SignalBroadcastIntervalSec
	if sec < 60 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

func (c *Config) IsAdmin(telegramID int64) bool {
	for _, id := range c.Telegram.AdminUserIDs {
		if id == telegramID {
			return true
		}
	}
	return false
}

// botClientID is only set when TELEGRAM_BOT_CLIENT_ID is explicit (OIDC / popup login).
// Do not derive from the bot token — that forces OIDC and hides the Login Widget when
// Web Login is not configured in @BotFather.
func botClientID(explicit, _ string) string {
	return strings.TrimSpace(explicit)
}

func parseCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
