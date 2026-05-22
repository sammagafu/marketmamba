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
	Provider string
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
	WebSessionSecret            string
	WebSessionTTLDays           int
	PublicSiteURL               string // https://marketmamba.kkooapp.co.tz — Telegram Login callback origin
	SignalBroadcastEnabled      bool
	SignalBroadcastIntervalSec  int
	SignalBroadcastSymbol       string
	SignalMinStrength           float64
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
			Provider: getEnv("BROKER_PROVIDER", "mock"),
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
			SubscriptionRequired:       getEnv("SUBSCRIPTION_REQUIRED", "false") == "true",
			FreeTrialDays:              parseInt(getEnv("FREE_TRIAL_DAYS", "30")),
			SubscriptionContactMessage: getEnv("SUBSCRIPTION_CONTACT", "Free testing period. Contact @codexxl on Telegram to extend after launch."),
			WebSessionSecret:           getEnv("WEB_SESSION_SECRET", ""),
			WebSessionTTLDays:          parseInt(getEnv("WEB_SESSION_TTL_DAYS", "365")),
			PublicSiteURL:              strings.TrimRight(getEnv("PUBLIC_SITE_URL", "https://marketmamba.kkooapp.co.tz"), "/"),
			SignalBroadcastEnabled:     getEnv("SIGNAL_BROADCAST_ENABLED", "true") == "true",
			SignalBroadcastIntervalSec: parseInt(getEnv("SIGNAL_BROADCAST_INTERVAL_SEC", "300")),
			SignalBroadcastSymbol:      getEnv("SIGNAL_BROADCAST_SYMBOL", "EURUSD"),
			SignalMinStrength:          parseFloat(getEnv("SIGNAL_MIN_STRENGTH", "0.7")),
		},
	}

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

func botClientID(explicit, botToken string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	if i := strings.Index(botToken, ":"); i > 0 {
		return botToken[:i]
	}
	return ""
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
