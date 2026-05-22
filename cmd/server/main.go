package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
	"forex-bot/internal/telegram"
	"forex-bot/internal/trading"
	"forex-bot/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Info("Starting Market Mamba")
	logger.Info("Environment: %s", cfg.App.Environment)

	// Initialize storage
	db, err := storage.NewPostgresStorage(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to connect to database: %v", err)
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Health(); err != nil {
		logger.Error("Database health check failed: %v", err)
		log.Fatalf("Database health check failed: %v", err)
	}

	logger.Info("Database connected successfully")

	// Initialize broker
	var b broker.Broker
	if cfg.Broker.Provider == "mock" {
		b = broker.NewMockBroker(10000) // Mock broker with $10k balance
		logger.Info("Using mock broker for development")
	} else {
		logger.Error("Unknown broker provider: %s", cfg.Broker.Provider)
		log.Fatalf("Unknown broker provider: %s", cfg.Broker.Provider)
	}

	// Initialize risk validator with default settings
	riskSettings := &models.RiskSettings{
		MaxRiskPerTrade:  cfg.Risk.MaxRiskPerTrade,
		MaxDailyLoss:     cfg.Risk.MaxDailyLoss,
		MaxOpenTrades:    cfg.Risk.MaxOpenTrades,
		MaxTradesPerDay:  cfg.Risk.MaxTradesPerDay,
		RiskRewardRatio:  cfg.Risk.RiskRewardRatio,
	}
	validator := risk.NewRiskValidator(riskSettings)

	// Initialize Telegram bot
	tgBot, err := telegram.NewTelegramBot(cfg.Telegram.BotToken, cfg.Telegram.AllowedUserIDs, b, db, validator)
	if err != nil {
		logger.Error("Failed to initialize Telegram bot: %v", err)
		log.Fatalf("Telegram bot initialization failed: %v", err)
	}

	// Initialize bot state for allowed users
	for _, userID := range cfg.Telegram.AllowedUserIDs {
		initializeBotState(db, userID)
	}

	logger.Info("Bot initialized successfully")
	logger.Info("Allowed users: %v", cfg.Telegram.AllowedUserIDs)

	// Demo: Log some test data
	logDemoInfo()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize trading services for the first allowed user (for testing)
	if len(cfg.Telegram.AllowedUserIDs) > 0 {
		primaryUserID := cfg.Telegram.AllowedUserIDs[0]

		// Create trading executor
		executor := trading.NewTradeExecutor(b, db, validator, primaryUserID)

		// Create position monitor
		posMonitor := trading.NewPositionMonitor(b, db, primaryUserID, 5*time.Second)
		posMonitor.Start(ctx, executor)
		logger.Info("Position monitor started for user %d", primaryUserID)

		// Create signal generator and monitor
		sigGen := trading.NewSignalGenerator("EURUSD", 0.7, cfg.Risk.RiskRewardRatio)
		sigMonitor := trading.NewSignalMonitor(sigGen, executor, db, primaryUserID, 10*time.Second)
		sigMonitor.Start(ctx)
		logger.Info("Signal monitor started for user %d", primaryUserID)

		// Run Telegram bot in goroutine
		go func() {
			if err := tgBot.Start(); err != nil {
				logger.Error("Bot error: %v", err)
			}
		}()

		// Wait for shutdown signal
		<-sigChan
		logger.Info("Shutdown signal received")

		// Graceful shutdown
		cancel()
		posMonitor.Stop()
		sigMonitor.Stop()

		logger.Info("Shutdown complete")
	} else {
		// Start bot if no users configured
		if err := tgBot.Start(); err != nil {
			logger.Error("Bot error: %v", err)
			log.Fatalf("Bot error: %v", err)
		}
	}
}

func initializeBotState(db storage.Storage, userID int64) {
	// Check if bot state exists
	_, err := db.GetBotState(userID)
	if err == nil {
		return // Already exists
	}

	// Create new bot state
	state := &models.BotState{
		ID:                utils.GenerateID("state"),
		UserID:            userID,
		IsPaused:          false,
		AutoTradingActive: false,
		DailyLossHit:      false,
		LastActiveAt:      time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := db.CreateBotState(state); err != nil {
		logger.Error("Failed to create bot state for user %d: %v", userID, err)
	}
}

func logDemoInfo() {
	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("Market Mamba - Development Mode")
	fmt.Println(separator)
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  /start        - Show help")
	fmt.Println("  /status       - Bot status")
	fmt.Println("  /balance      - Account balance")
	fmt.Println("  /positions    - Open positions")
	fmt.Println("  /open         - Open trade")
	fmt.Println("  /close        - Close position")
	fmt.Println("  /closeall     - Close all positions")
	fmt.Println("  /pause        - Pause trading")
	fmt.Println("  /resume       - Resume trading")
	fmt.Println("  /risk         - Risk settings")
	fmt.Println("  /dailyreport  - Daily report")
	fmt.Println("\nExample Trade:")
	fmt.Println("  /open EURUSD BUY 1.0 1.0900 1.1000")
	fmt.Println("\nWarning: This is a development bot. Use with caution in production.")
	fmt.Println(separator + "\n")
}
