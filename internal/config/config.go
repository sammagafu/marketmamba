package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	BotToken         string
	AllowedUserIDs   []int64
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
	Environment string
	Port        string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Telegram: TelegramConfig{
			BotToken:       getEnv("TELEGRAM_BOT_TOKEN", ""),
			AllowedUserIDs: parseUserIDs(getEnv("TELEGRAM_ALLOWED_USER_IDS", "")),
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
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("PORT", "8080"),
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
	if len(c.Telegram.AllowedUserIDs) == 0 {
		return fmt.Errorf("TELEGRAM_ALLOWED_USER_IDS is required")
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
