package decision

import (
	"fmt"
	"strings"
	"time"

	"forex-bot/internal/models"
)

// Action is the real-time decision outcome.
type Action string

const (
	ActionTake Action = "TAKE"
	ActionSkip Action = "SKIP"
	ActionWait Action = "WAIT"
)

// Decision is sniper decision support for one symbol evaluation.
type Decision struct {
	Action     Action
	Confidence float64 // 0–1
	Reason     string
	Checks     []string
	Signal     *models.TradeSignal
	Symbol     string
	At         time.Time
}

// FormatTelegram builds a subscriber-facing sniper alert.
func FormatTelegram(d *Decision) string {
	if d == nil {
		return ""
	}
	checks := strings.Join(d.Checks, " | ")
	if checks == "" {
		checks = "—"
	}
	switch d.Action {
	case ActionTake:
		sig := d.Signal
		if sig == nil {
			return fmt.Sprintf("🎯 *SNIPER — %s*\n\n*TAKE* (%.0f%%)\n%s\n\nChecks: %s",
				d.Symbol, d.Confidence*100, d.Reason, checks)
		}
		return fmt.Sprintf(
			"🎯 *SNIPER — %s*\n\n*TAKE %s* (%.0f%%)\n%s\n\n"+
				"SL: %.5f | TP: %.5f | R:R %.2f\n\nChecks: %s\n\n"+
				"_Assisted auto runs only above SNIPER_MIN_CONFIDENCE with /autostart._",
			d.Symbol, sig.Type, d.Confidence*100, d.Reason,
			sig.StopLoss, sig.TakeProfit, sig.RiskRewardRatio, checks,
		)
	case ActionWait:
		return fmt.Sprintf(
			"⏳ *SNIPER — %s*\n\n*WAIT* (%.0f%%)\n%s\n\nChecks: %s",
			d.Symbol, d.Confidence*100, d.Reason, checks,
		)
	default:
		return fmt.Sprintf(
			"⏭️ *SNIPER — %s*\n\n*SKIP*\n%s\n\nChecks: %s",
			d.Symbol, d.Reason, checks,
		)
	}
}
