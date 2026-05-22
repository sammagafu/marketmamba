package trading

import (
	"strings"
	"time"

	"forex-bot/internal/logger"
)

func (sm *SignalMonitor) readyForAutoTrade() error {
	if sm.executor == nil {
		return nil
	}
	return sm.executor.ReadyForTrade()
}

func (sm *SignalMonitor) logTradeBlocked(err error) {
	if err == nil {
		return
	}
	msg := err.Error()
	now := time.Now()
	if msg == sm.lastTradeWarnMsg && now.Sub(sm.lastTradeWarnAt) < 5*time.Minute {
		return
	}
	sm.lastTradeWarnMsg = msg
	sm.lastTradeWarnAt = now
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "no rows"),
		strings.Contains(lower, "sync account"),
		strings.Contains(lower, "connect a broker"),
		strings.Contains(lower, "broker"):
		logger.Warn(
			"Auto-trade skipped for user %d (connect broker: /broker connect or dashboard, then /balance): %v",
			sm.userID, err,
		)
	default:
		logger.Warn("Auto-trade blocked for user %d: %v", sm.userID, err)
	}
}
