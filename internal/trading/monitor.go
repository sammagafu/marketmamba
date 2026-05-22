package trading

import (
	"context"
	"fmt"
	"time"

	"forex-bot/internal/broker"
	"forex-bot/internal/decision"
	"forex-bot/internal/logger"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
)

// PositionMonitor continuously monitors open positions and checks for TP/SL
type PositionMonitor struct {
	broker   broker.Broker
	storage  storage.Storage
	userID   int64
	interval time.Duration
	stopChan chan struct{}
	done     chan struct{}
}

func NewPositionMonitor(b broker.Broker, s storage.Storage, userID int64, interval time.Duration) *PositionMonitor {
	return &PositionMonitor{
		broker:   b,
		storage:  s,
		userID:   userID,
		interval: interval,
		stopChan: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start begins monitoring positions
func (pm *PositionMonitor) Start(ctx context.Context, executor *TradeExecutor) {
	logger.Info("Starting position monitor for user %d (interval: %v)", pm.userID, pm.interval)

	go func() {
		defer close(pm.done)

		ticker := time.NewTicker(pm.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Position monitor stopped (context cancelled)")
				return
			case <-pm.stopChan:
				logger.Info("Position monitor stopped (manual)")
				return
			case <-ticker.C:
				if err := pm.checkPositions(executor); err != nil {
					logger.Error("Error checking positions: %v", err)
				}
			}
		}
	}()
}

// Stop halts the position monitor
func (pm *PositionMonitor) Stop() {
	close(pm.stopChan)
	<-pm.done
}

func (pm *PositionMonitor) checkPositions(executor *TradeExecutor) error {
	// Get open positions from broker
	positions, err := pm.broker.GetOpenPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	if len(positions) == 0 {
		return nil
	}

	// Check each position
	for _, pos := range positions {
		// Simulate price movement (in real system, get actual market price)
		if err := pm.checkPosition(pos, executor); err != nil {
			logger.Error("Error checking position %s: %v", pos.ID, err)
		}
	}

	return nil
}

func (pm *PositionMonitor) checkPosition(pos *models.Position, executor *TradeExecutor) error {
	// In production, you would get real price from market data
	// For now, simulate price movement
	simulatedPrice := pm.simulatePrice(pos)

	pos.CurrentPrice = simulatedPrice

	// Calculate unrealized profit
	if pos.Type == "BUY" {
		pos.Profit = (simulatedPrice - pos.EntryPrice) * pos.Quantity
		pos.ProfitPct = ((simulatedPrice - pos.EntryPrice) / pos.EntryPrice) * 100
	} else {
		pos.Profit = (pos.EntryPrice - simulatedPrice) * pos.Quantity
		pos.ProfitPct = ((pos.EntryPrice - simulatedPrice) / pos.EntryPrice) * 100
	}

	pos.UpdatedAt = time.Now()

	// Update position in storage
	if err := pm.storage.UpdatePosition(pos); err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	// Check if position should be closed
	if shouldCloseTakeProfit(pos) {
		logger.Info("Position %s hitting take profit: %.5f", pos.ID, pos.CurrentPrice)
		if mockBroker, ok := pm.broker.(*broker.MockBroker); ok {
			mockBroker.SimulatePrice(pos.ID, pos.CurrentPrice)
		}
		return executor.CheckAndClosePositions()
	}

	if shouldCloseStopLoss(pos) {
		logger.Info("Position %s hitting stop loss: %.5f", pos.ID, pos.CurrentPrice)
		if mockBroker, ok := pm.broker.(*broker.MockBroker); ok {
			mockBroker.SimulatePrice(pos.ID, pos.CurrentPrice)
		}
		return executor.CheckAndClosePositions()
	}

	return nil
}

// simulatePrice creates realistic price movement for testing
func (pm *PositionMonitor) simulatePrice(pos *models.Position) float64 {
	// Generate random walk with slight trend
	direction := 1.0
	if pos.Profit < 0 {
		direction = -0.5 // Slight downward bias if losing
	}

	// Small random movement (0-20 pips)
	movement := (time.Now().UnixNano() % 20) - 10 // -10 to +10
	pip := 0.0001
	if pos.Symbol == "USDJPY" {
		pip = 0.01
	}

	movement_price := float64(movement) * pip * direction

	if pos.CurrentPrice > 0 {
		return pos.CurrentPrice + movement_price
	}

	return pos.EntryPrice + movement_price
}

// SignalMonitor continuously generates and executes trading signals across symbols.
type SignalMonitor struct {
	symbols            []string
	rrRatio            float64
	symbolIdx          int
	executor           *TradeExecutor
	storage            storage.Storage
	userID             int64
	interval           time.Duration
	engine             *decision.Engine
	advisoryEnabled    bool
	autoExecuteEnabled bool
	advisoryNotify     SniperNotifier
	stopChan           chan struct{}
	done               chan struct{}
	lastTradeWarnMsg   string
	lastTradeWarnAt    time.Time
}

func NewSignalMonitor(
	symbols []string,
	rrRatio float64,
	exec *TradeExecutor,
	stor storage.Storage,
	userID int64,
	interval time.Duration,
	engine *decision.Engine,
	advisoryEnabled, autoExecuteEnabled bool,
	advisoryNotify SniperNotifier,
) *SignalMonitor {
	if len(symbols) == 0 {
		symbols = []string{"EURUSD", "BTCUSD"}
	}
	return &SignalMonitor{
		symbols:            symbols,
		rrRatio:            rrRatio,
		symbolIdx:          0,
		executor:           exec,
		storage:            stor,
		userID:             userID,
		interval:           interval,
		engine:             engine,
		advisoryEnabled:    advisoryEnabled,
		autoExecuteEnabled: autoExecuteEnabled,
		advisoryNotify:     advisoryNotify,
		stopChan:           make(chan struct{}),
		done:               make(chan struct{}),
	}
}

// Start begins signal generation and execution
func (sm *SignalMonitor) Start(ctx context.Context) {
	logger.Info("Starting signal monitor for user %d (interval: %v)", sm.userID, sm.interval)

	go func() {
		defer close(sm.done)

		ticker := time.NewTicker(sm.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Signal monitor stopped (context cancelled)")
				return
			case <-sm.stopChan:
				logger.Info("Signal monitor stopped (manual)")
				return
			case <-ticker.C:
				if err := sm.evaluateDecision(ctx); err != nil {
					logger.Error("Error in sniper decision: %v", err)
				}
			}
		}
	}()
}

// Stop halts the signal monitor
func (sm *SignalMonitor) Stop() {
	close(sm.stopChan)
	<-sm.done
}

func (sm *SignalMonitor) generateAndExecuteSignal() error {
	// Get bot state
	botState, err := sm.storage.GetBotState(sm.userID)
	if err != nil {
		return fmt.Errorf("failed to get bot state: %w", err)
	}

	// Don't generate signals if paused
	if botState.IsPaused {
		return nil
	}

	// Don't generate signals if daily loss limit hit
	if botState.DailyLossHit {
		return nil
	}

	// Simulate market data and generate signal
	signal := sm.generateSignal()
	if signal == nil {
		return nil // No signal this iteration
	}

	logger.Info(
		"Signal generated for user %d: %s %s | strength=%.2f SL=%.5f TP=%.5f | reason=%s",
		sm.userID, signal.Symbol, signal.Type, signal.Strength, signal.StopLoss, signal.TakeProfit, signal.Reason,
	)

	if err := sm.readyForAutoTrade(); err != nil {
		sm.logTradeBlocked(err)
		return nil
	}
	pos, err := sm.executor.ExecuteSignal(signal)
	if err != nil {
		sm.logTradeBlocked(err)
		return nil
	}
	if pos != nil {
		logger.Info(
			"Trade opened for user %d: %s %s qty=%.2f entry=%.5f SL=%.5f TP=%.5f id=%s",
			sm.userID, pos.Symbol, pos.Type, pos.Quantity, pos.EntryPrice, pos.StopLoss, pos.TakeProfit, pos.ID,
		)
	}

	return nil
}

func (sm *SignalMonitor) generateSignal() *models.TradeSignal {
	opportunityTypes := []string{"UPTREND_SCALP", "DOWNTREND_SCALP", "TREND_CONFIRMATION"}
	oppIdx := int(time.Now().UnixNano() % 3)

	for i := 0; i < len(sm.symbols); i++ {
		sym := sm.symbols[(sm.symbolIdx+i)%len(sm.symbols)]
		signal := SimulateScalpingOpportunity(sym, opportunityTypes[oppIdx])
		if signal != nil {
			sm.symbolIdx = (sm.symbolIdx + i + 1) % len(sm.symbols)
			return signal
		}
	}
	sm.symbolIdx = (sm.symbolIdx + 1) % len(sm.symbols)
	return nil
}
