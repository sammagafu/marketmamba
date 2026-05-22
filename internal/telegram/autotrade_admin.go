package telegram

import (
	"fmt"
	"strconv"
	"strings"
)

func (tb *TelegramBot) handleApproveAuto(chatID, adminID int64, args []string) {
	if !tb.cfg.IsAdmin(adminID) {
		tb.sendMessage(chatID, "❌ Admin only")
		return
	}
	target := adminID
	if len(args) > 0 {
		v, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
		if err != nil {
			tb.sendMessage(chatID, "Usage: /approveauto [telegram_user_id]")
			return
		}
		target = v
	}
	if err := tb.storage.SetAutoTradeApproved(target, true); err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf("✅ Auto-trade approved for user %d", target))
	tb.sendMessage(target, "✅ Admin approved your account for *assisted auto-trading*. Use /autostart if not already on.")
}

func (tb *TelegramBot) handleRevokeAuto(chatID, adminID int64, args []string) {
	if !tb.cfg.IsAdmin(adminID) {
		tb.sendMessage(chatID, "❌ Admin only")
		return
	}
	target := adminID
	if len(args) > 0 {
		v, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64)
		if err != nil {
			tb.sendMessage(chatID, "Usage: /revokeauto [telegram_user_id]")
			return
		}
		target = v
	}
	_ = tb.storage.SetAutoTradeApproved(target, false)
	_ = tb.storage.UpdateBotState(target, false, false, false)
	tb.sendMessage(chatID, fmt.Sprintf("⏹️ Auto-trade approval revoked for user %d", target))
}
