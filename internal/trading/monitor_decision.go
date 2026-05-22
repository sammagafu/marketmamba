package trading

import (
	"context"

	"forex-bot/internal/decision"
	"forex-bot/internal/logger"
)

// SniperNotifier sends advisory TAKE alerts to a Telegram user (chat_id = telegram user id).
type SniperNotifier func(userID int64, d *decision.Decision)

func (sm *SignalMonitor) evaluateDecision(ctx context.Context) error {
	if sm.engine == nil {
		return sm.generateAndExecuteSignalLegacy()
	}

	botState, err := sm.storage.GetBotState(sm.userID)
	if err != nil {
		return err
	}
	if botState.IsPaused || botState.DailyLossHit {
		return nil
	}

	sym := sm.symbols[sm.symbolIdx%len(sm.symbols)]
	sm.symbolIdx = (sm.symbolIdx + 1) % len(sm.symbols)

	d, err := sm.engine.Evaluate(ctx, sym)
	if err != nil {
		return err
	}

	switch d.Action {
	case decision.ActionTake:
		logger.Info(
			"Sniper TAKE user=%d %s %s confidence=%.0f%% | %s",
			sm.userID, d.Symbol, sideOrDash(d), d.Confidence*100, d.Reason,
		)
		marked := false
		if sm.advisoryNotify != nil && sm.advisoryEnabled {
			sm.advisoryNotify(sm.userID, d)
			marked = true
		}
		if botState.AutoTradingActive && sm.autoExecuteEnabled && sm.engine.AutoExecuteAllowed(d) && d.Signal != nil {
			pos, execErr := sm.executor.ExecuteSignal(d.Signal)
			if execErr != nil {
				logger.Warn("[%s] Failed to execute %s for user %d: %v", d.Symbol, d.Signal.Type, sm.userID, execErr)
			} else if pos != nil {
				marked = true
				logger.Info(
					"Assisted auto trade user=%d: %s %s qty=%.2f entry=%.5f id=%s",
					sm.userID, pos.Symbol, pos.Type, pos.Quantity, pos.EntryPrice, pos.ID,
				)
			}
		} else if d.Signal != nil {
			logger.Info(
				"Sniper TAKE user=%d — advisory only (autostart=%v auto_execute=%v conf=%.0f%% min=%.0f%%)",
				sm.userID, botState.AutoTradingActive, sm.autoExecuteEnabled, d.Confidence*100, sm.engine.SniperMinConfidence()*100,
			)
		}
		if marked {
			sm.engine.MarkTaken(d.Symbol)
		}
	case decision.ActionWait:
		logger.Debug("Sniper WAIT user=%d %s: %s", sm.userID, d.Symbol, d.Reason)
	default:
		logger.Debug("Sniper SKIP user=%d %s: %s", sm.userID, d.Symbol, d.Reason)
	}
	return nil
}

func sideOrDash(d *decision.Decision) string {
	if d.Signal != nil {
		return d.Signal.Type
	}
	return "—"
}

func (sm *SignalMonitor) generateAndExecuteSignalLegacy() error {
	signal := sm.generateSignal()
	if signal == nil {
		return nil
	}
	logger.Info(
		"Signal generated for user %d: %s %s | strength=%.2f SL=%.5f TP=%.5f | reason=%s",
		sm.userID, signal.Symbol, signal.Type, signal.Strength, signal.StopLoss, signal.TakeProfit, signal.Reason,
	)
	pos, err := sm.executor.ExecuteSignal(signal)
	if err != nil {
		logger.Warn("[%s] Failed to execute %s for user %d: %v", signal.Symbol, signal.Type, sm.userID, err)
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
