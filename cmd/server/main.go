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
	"forex-bot/internal/decision"
	"forex-bot/internal/feedback"
	"forex-bot/internal/marketdata"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/pairs"
	"forex-bot/internal/payments"
	"forex-bot/internal/risk"
	"forex-bot/internal/signals"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
	"forex-bot/internal/telegram"
	"forex-bot/internal/tier"
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

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := db.RunMigrations(migrationsDir); err != nil {
		logger.Error("Database migrations: %v", err)
	} else {
		logger.Info("Database migrations applied")
	}

	if len(os.Args) > 1 && os.Args[1] == "seed-admin" {
		if err := adminseed.Run(db); err != nil {
			log.Fatalf("seed-admin: %v", err)
		}
		log.Println("Web admin ready — log in with ADMIN_EMAIL on the dashboard")
		return
	}

	resolveBroker := func(userID int64) (broker.Broker, error) {
		return broker.ResolveBrokerAndSync(db, userID, cfg.App.BrokerEncryptionKey, cfg.Broker.Provider)
	}

	broker.SetEnabledBrands(cfg.Broker.EnabledBrokerBrands)
	broker.SetSharedMetaAPIToken(cfg.Broker.MetaAPISharedToken)

	subs := subscription.NewService(db, cfg)
	tierSvc := tier.NewService(db, cfg)
	subs.SetTier(tierSvc)
	paySvc := payments.NewService(db, subs, cfg)
	pairSvc := pairs.NewService(db, cfg)
	usersSvc := users.NewService(db, subs, cfg)

	validator := risk.NewRiskValidator(&models.RiskSettings{
		MaxRiskPerTrade: cfg.Risk.MaxRiskPerTrade,
		MaxDailyLoss:    cfg.Risk.MaxDailyLoss,
		MaxOpenTrades:   cfg.Risk.MaxOpenTrades,
		MaxTradesPerDay: cfg.Risk.MaxTradesPerDay,
		RiskRewardRatio: cfg.Risk.RiskRewardRatio,
	})

	tgBot, err := telegram.NewTelegramBot(cfg, resolveBroker, db, validator, usersSvc, subs, tierSvc)
	if err != nil {
		log.Fatalf("Telegram bot initialization failed: %v", err)
	}
	tgBot.ConfigureMiniApp(cfg.Payments.MiniAppURL)

	outcomeSvc := feedback.NewService(tgBot, db, subs, cfg.SignalSymbols())
	tgBot.SetOutcomeNotifier(outcomeSvc)

	var decisionEngine *decision.Engine
	if cfg.App.DecisionEnabled {
		marketSvc := marketdata.NewService(cfg.App.MarketDataAPIKey)
		cooldown := decision.NewCooldownTracker(cfg.SniperCooldown())
		decisionEngine = decision.NewEngine(
			marketSvc,
			validator,
			cooldown,
			cfg.App.SignalMinStrength,
			cfg.App.SniperMinConfidence,
			cfg.Risk.RiskRewardRatio,
		)
		tgBot.SetDecisionEngine(decisionEngine)
		logger.Info(
			"Real-time sniper decisions enabled (provider=%s, interval=%v, cooldown=%v, min conf=%.0f%%, mode advisory=%v auto=%v)",
			marketSvc.ProviderName(),
			cfg.DecisionInterval(),
			cfg.SniperCooldown(),
			cfg.App.SniperMinConfidence*100,
			cfg.App.DecisionAdvisory,
			cfg.App.DecisionAutoExecute,
		)
		go warmupMarketData(context.Background(), marketSvc, cfg.SignalSymbols())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sniperNotify := trading.SniperNotifier(nil)
	if cfg.App.DecisionAdvisory {
		sniperNotify = func(userID int64, d *decision.Decision) {
			_ = tgBot.NotifySniper(userID, d)
		}
	}

	coordinator := trading.NewCoordinator(db, cfg, subs, tierSvc, validator, resolveBroker, outcomeSvc, pairSvc, decisionEngine, sniperNotify)
	coordinator.Start(ctx)

	if cfg.App.SignalBroadcastEnabled {
		pub := signals.NewPublisher(
			db, subs, tierSvc, tgBot, validator, decisionEngine,
			cfg.SignalSymbols(), cfg.SignalBroadcastInterval(),
			cfg.App.SignalMinStrength,
			cfg.App.DecisionEnabled,
		)
		pub.Start(ctx)
	}

	if cfg.App.EnableWeb {
		apiServer := api.NewServer(cfg, db, subs, tierSvc, paySvc, usersSvc, resolveBroker, tgBot, validator, pairSvc)
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

func warmupMarketData(ctx context.Context, svc *marketdata.Service, symbols []string) {
	for _, sym := range symbols {
		if _, err := svc.Refresh(ctx, sym); err != nil {
			logger.Warn("Market warmup %s: %v", sym, err)
		}
	}
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
