package trading

import (
	"context"
	"sync"
	"time"

	"forex-bot/internal/broker"
	"forex-bot/internal/config"
	"forex-bot/internal/logger"
	"forex-bot/internal/risk"
	"forex-bot/internal/feedback"
	"forex-bot/internal/storage"
	"forex-bot/internal/subscription"
)

type BrokerResolver func(userID int64) (broker.Broker, error)

type userRunner struct {
	cancel     context.CancelFunc
	posMonitor *PositionMonitor
	sigMonitor *SignalMonitor
}

type Coordinator struct {
	store           *storage.PostgresStorage
	cfg             *config.Config
	subs            *subscription.Service
	validator       *risk.RiskValidator
	outcomeNotifier feedback.OutcomeNotifier
	resolve         BrokerResolver
	mu              sync.Mutex
	runners         map[int64]*userRunner
	interval        time.Duration
}

func NewCoordinator(
	store *storage.PostgresStorage,
	cfg *config.Config,
	subs *subscription.Service,
	v *risk.RiskValidator,
	resolve BrokerResolver,
	outcomeNotifier feedback.OutcomeNotifier,
) *Coordinator {
	return &Coordinator{
		store:           store,
		cfg:             cfg,
		subs:            subs,
		validator:       v,
		outcomeNotifier: outcomeNotifier,
		resolve:         resolve,
		runners:         make(map[int64]*userRunner),
		interval:        15 * time.Second,
	}
}

func (c *Coordinator) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		c.sync(ctx)
		for {
			select {
			case <-ctx.Done():
				c.StopAll()
				return
			case <-ticker.C:
				c.sync(ctx)
			}
		}
	}()
	logger.Info("Multi-user trading coordinator started")
}

func (c *Coordinator) sync(ctx context.Context) {
	ids, err := c.store.ListAutoTradingUserIDs()
	if err != nil {
		logger.Error("Coordinator list users: %v", err)
		return
	}
	want := make(map[int64]bool)
	for _, id := range ids {
		ok, _ := c.subs.CanTrade(id)
		if !ok {
			continue
		}
		want[id] = true
		c.ensureRunner(ctx, id)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, r := range c.runners {
		if !want[id] {
			r.cancel()
			r.posMonitor.Stop()
			r.sigMonitor.Stop()
			delete(c.runners, id)
			logger.Info("Stopped monitors for user %d", id)
		}
	}
}

func (c *Coordinator) ensureRunner(ctx context.Context, userID int64) {
	c.mu.Lock()
	if _, ok := c.runners[userID]; ok {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	b, err := c.resolve(userID)
	if err != nil {
		logger.Error("Coordinator broker user %d: %v", userID, err)
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	executor := NewTradeExecutor(b, c.store, c.validator, userID, c.outcomeNotifier)
	posMonitor := NewPositionMonitor(b, c.store, userID, 5*time.Second)
	sigMonitor := NewSignalMonitor(c.cfg.SignalSymbols(), c.cfg.Risk.RiskRewardRatio, executor, c.store, userID, 10*time.Second)
	posMonitor.Start(runCtx, executor)
	sigMonitor.Start(runCtx)

	c.mu.Lock()
	c.runners[userID] = &userRunner{cancel: cancel, posMonitor: posMonitor, sigMonitor: sigMonitor}
	c.mu.Unlock()
	logger.Info("Started monitors for user %d", userID)
}

func (c *Coordinator) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, r := range c.runners {
		r.cancel()
		r.posMonitor.Stop()
		r.sigMonitor.Stop()
		delete(c.runners, id)
	}
}
