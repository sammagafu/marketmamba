package telegram

import (
	"fmt"
	"strings"

	"forex-bot/internal/config"
	"forex-bot/internal/models"
	"forex-bot/internal/pairs"
	"forex-bot/internal/storage"
)

func (tb *TelegramBot) pairService() *pairs.Service {
	ps, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		return nil
	}
	return pairs.NewService(ps, tb.cfg)
}

func (tb *TelegramBot) handlePairs(chatID, userID int64, args []string) {
	svc := tb.pairService()
	if svc == nil {
		tb.sendMessage(chatID, "❌ Pair preferences unavailable")
		return
	}
	if len(args) == 0 {
		resp, err := svc.GetResponse(userID)
		if err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		var b strings.Builder
		b.WriteString("*Your signal setup*\n\n")
		if note := communityLaunchNote(tb.cfg); note != "" {
			b.WriteString(note)
			b.WriteString("\n\n")
		}
		b.WriteString(formatSignalTypes(resp.SignalTypes))
		b.WriteString("\n")
		for _, g := range resp.AssetGroups {
			if g.ComingSoon {
				b.WriteString(fmt.Sprintf("*%s:* _coming soon for the community_\n", g.Label))
				continue
			}
			if !g.Enabled {
				continue
			}
			b.WriteString(fmt.Sprintf("*%s:* %s\n", g.Label, strings.Join(g.Symbols, ", ")))
		}
		b.WriteString("\n*Per-pair flags*\n")
		for _, p := range resp.Pairs {
			sig := "—"
			if p.ReceiveSignals {
				sig = "📡"
			}
			auto := "—"
			if p.AutoTrade {
				auto = "🤖"
			}
			b.WriteString(fmt.Sprintf("• *%s* signals %s · auto %s\n", p.Symbol, sig, auto))
		}
		b.WriteString("\n*Commands*\n")
		b.WriteString("`/signaltypes forex crypto` — asset classes\n")
		b.WriteString("`/pairs EURUSD BTCUSD` — enable specific pairs\n")
		b.WriteString("_📡 = Telegram signals · 🤖 = auto-trade with /autostart_")
		tb.sendMessage(chatID, b.String())
		return
	}

	sub := args[0]
	if sub == "all" {
		if err := svc.SeedDefaults(userID); err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		tb.sendMessage(chatID, "✅ All signal types and platform pairs enabled")
		return
	}

	if err := svc.SetSymbolsQuick(userID, args); err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	resp, _ := svc.GetResponse(userID)
	tb.sendMessage(chatID, fmt.Sprintf(
		"✅ Pairs updated\n\nSignals: %s\nAuto-trade: %s\n\nUse /autostart to run automation on your auto pairs.",
		strings.Join(resp.SignalSymbols, ", "),
		strings.Join(resp.AutoTradeSymbols, ", "),
	))
}

func (tb *TelegramBot) handleSignalTypes(chatID, userID int64, args []string) {
	svc := tb.pairService()
	if svc == nil {
		tb.sendMessage(chatID, "❌ Signal preferences unavailable")
		return
	}
	if len(args) == 0 {
		resp, err := svc.GetResponse(userID)
		if err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		var b strings.Builder
		b.WriteString("*Signal types*\n\n")
		b.WriteString(formatSignalTypes(resp.SignalTypes))
		b.WriteString("\n")
		if note := communityLaunchNote(tb.cfg); note != "" {
			b.WriteString(note)
			b.WriteString("\n\n")
		}
		for _, g := range resp.AssetGroups {
			state := "off"
			if g.Enabled {
				state = "on"
			}
			if g.ComingSoon {
				state = "coming soon"
			}
			b.WriteString(fmt.Sprintf("• *%s* (%s): %s\n", g.Label, state, strings.Join(g.Symbols, ", ")))
		}
		b.WriteString("\nEnable types:\n`/signaltypes forex indexes crypto`\n")
		if tb.cfg != nil && !tb.cfg.IsFullAssetCatalog() {
			b.WriteString("`/signaltypes crypto` — Bitcoin & Ethereum (launch phase)\n")
		} else {
			b.WriteString("`/signaltypes all` — enable everything\n")
		}
		tb.sendMessage(chatID, b.String())
		return
	}

	if strings.EqualFold(args[0], "all") {
		prefs := models.DefaultSignalTypes()
		if tb.cfg != nil && !tb.cfg.IsFullAssetCatalog() {
			prefs = models.BitcoinPhaseSignalTypes()
		}
		if err := svc.SetSignalTypes(userID, prefs); err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
		if tb.cfg != nil && !tb.cfg.IsFullAssetCatalog() {
			tb.sendMessage(chatID, "✅ Community launch: Bitcoin & Ethereum signals enabled")
			return
		}
		tb.sendMessage(chatID, "✅ Forex, indexes, and crypto signals enabled")
		return
	}

	prefs, ok := pairs.ParseSignalTypesFromArgs(args)
	if !ok {
		tb.sendMessage(chatID, "❌ Unknown option. Use: `forex`, `indexes`, `crypto`, or `all`")
		return
	}
	if err := svc.SetSignalTypes(userID, prefs); err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	resp, _ := svc.GetResponse(userID)
	tb.sendMessage(chatID, fmt.Sprintf(
		"✅ Signal types updated\n\n%s\n\nActive pairs: %s",
		formatSignalTypes(resp.SignalTypes),
		strings.Join(resp.AvailableSymbols, ", "),
	))
}

func communityLaunchNote(cfg *config.Config) string {
	if cfg == nil || cfg.IsFullAssetCatalog() {
		return ""
	}
	note := strings.TrimSpace(cfg.App.AITrainingNote)
	if note == "" {
		note = "We're training our bots with AI for more precise entries."
	}
	phase := strings.TrimSpace(cfg.App.CommunityPhaseMessage)
	if phase != "" {
		return phase + "\n_" + note + "_"
	}
	return "_Community launch: Bitcoin & Ethereum. " + note + "_"
}

func formatSignalTypes(p models.SignalTypePreferences) string {
	var on []string
	if p.Forex {
		on = append(on, "Forex")
	}
	if p.Indexes {
		on = append(on, "Indexes")
	}
	if p.Crypto {
		on = append(on, "Bitcoin & crypto")
	}
	if len(on) == 0 {
		return "Enabled: _none_"
	}
	return "Enabled: " + strings.Join(on, " · ")
}
