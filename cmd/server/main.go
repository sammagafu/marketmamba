package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"forex-bot/internal/adminseed"
	"forex-bot/internal/api"
	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/risk"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/telegram"
	"forex-bot/internal/trading"
	"forex-bot/internal/users"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Info("Starting Market Mamba")
	logger.Info("Environment: %s | public=%v | subscription_required=%v",
		cfg.App.Environment, cfg.App.PublicMode, cfg.App.SubscriptionRequired)

	db, err := storage.NewPostgresStorage(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	logger.Info("Database connected successfully")

	if len(os.Args) > 1 && os.Args[1] == "seed-admin" {
		if err := adminseed.Run(db); err != nil {
			log.Fatalf("seed-admin: %v", err)
		}
		log.Println("Web admin ready — log in with ADMIN_EMAIL on the dashboard")
		return
	}

	resolveBroker := func(userID int64) (broker.Broker, error) {
		b, _, err := broker.ResolveBroker(db, userID, cfg.App.BrokerEncryptionKey, cfg.Broker.Provider)
		return b, err
	}

	subs := subscription.NewService(db, cfg)
	usersSvc := users.NewService(db, subs)

	validator := risk.NewRiskValidator(&models.RiskSettings{
		MaxRiskPerTrade: cfg.Risk.MaxRiskPerTrade,
		MaxDailyLoss:    cfg.Risk.MaxDailyLoss,
		MaxOpenTrades:   cfg.Risk.MaxOpenTrades,
		MaxTradesPerDay: cfg.Risk.MaxTradesPerDay,
		RiskRewardRatio: cfg.Risk.RiskRewardRatio,
	})

	tgBot, err := telegram.NewTelegramBot(cfg, resolveBroker, db, validator, usersSvc, subs)
	if err != nil {
		log.Fatalf("Telegram bot initialization failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	coordinator := trading.NewCoordinator(db, cfg, subs, validator, resolveBroker)
	coordinator.Start(ctx)

	if cfg.App.EnableWeb {
		apiServer := api.NewServer(cfg, db, subs, usersSvc, resolveBroker)
		go func() {
			addr := ":" + cfg.App.HTTPPort
			logger.Info("Web dashboard listening on %s", addr)
			if err := http.ListenAndServe(addr, apiServer.Handler()); err != nil {
				logger.Error("Web server error: %v", err)
			}
		}()
	}

	logStartupBanner(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := tgBot.Start(); err != nil {
			logger.Error("Bot error: %v", err)
		}
	}()

	<-sigChan
	logger.Info("Shutdown signal received")
	cancel()
	coordinator.StopAll()
	logger.Info("Shutdown complete")
}

func logStartupBanner(cfg *config.Config) {
	sep := strings.Repeat("=", 60)
	fmt.Println("\n" + sep)
	fmt.Println("Market Mamba")
	fmt.Println(sep)
	fmt.Printf("Public mode: %v\n", cfg.App.PublicMode)
	fmt.Printf("Admins: %v\n", cfg.Telegram.AdminUserIDs)
	fmt.Printf("Web: http://localhost:%s\n", cfg.App.HTTPPort)
	fmt.Println(sep + "\n")
}
