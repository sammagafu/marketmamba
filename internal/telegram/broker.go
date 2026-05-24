package telegram

import (
	"fmt"
	"strings"

	"forex-bot/internal/broker"
	"forex-bot/internal/storage"
)

func (tb *TelegramBot) handleBroker(chatID, userID int64, args []string) {
	if len(args) == 0 {
		tb.showBrokerStatus(chatID, userID)
		return
	}
	switch strings.ToLower(args[0]) {
	case "connect":
		provider := "mock"
		if len(args) > 1 {
			provider = strings.ToLower(args[1])
		}
		if provider == "web" || provider == "wizard" {
			tb.sendBrokerConnectLink(chatID)
			return
		}
		tb.connectBroker(chatID, userID, provider)
	case "mock", "demo":
		tb.connectBroker(chatID, userID, "mock")
	default:
		tb.sendMessage(chatID, `*Broker connection*

/broker — show current connection
/broker connect — Mock demo ($10,000)
/broker connect web — open connection wizard

*Deriv, Exness, Tickmill, OANDA* — use the web wizard:
`+tb.brokerConnectURL())
	}
}

func (tb *TelegramBot) brokerConnectURL() string {
	base := tb.cfg.App.PublicSiteURL
	if u := tb.cfg.Payments.MiniAppURL; u != "" {
		base = u
	}
	if base == "" {
		base = "https://marketmamba.kkooapp.co.tz"
	}
	return strings.TrimRight(base, "/") + "/#/connect"
}

func (tb *TelegramBot) sendBrokerConnectLink(chatID int64) {
	tb.sendMessage(chatID, fmt.Sprintf(`*Connect your broker*

Open the dashboard and choose Deriv, Exness, Tickmill, or OANDA:

%s

Or use /broker connect for a free Mock demo.`, tb.brokerConnectURL()))
}

func (tb *TelegramBot) showBrokerStatus(chatID, userID int64) {
	pg, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ Broker status unavailable")
		return
	}
	conn, err := pg.GetActiveBrokerConnection(userID)
	if err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	if conn == nil {
		tb.sendMessage(chatID, fmt.Sprintf(`*No broker connected*

Demo: /broker connect
Live brokers: %s`, tb.brokerConnectURL()))
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf(`*Broker connected*
Provider: *%s*
Label: %s
Updated: %s

/balance — view account
/broker connect web — change broker`, conn.Provider, conn.Label, conn.UpdatedAt.Format("2006-01-02 15:04")))
}

func (tb *TelegramBot) connectBroker(chatID, userID int64, provider string) {
	pg, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ Connect via web dashboard")
		return
	}
	if provider != "mock" {
		tb.sendBrokerConnectLink(chatID)
		return
	}
	if tb.tier != nil {
		if err := tb.tier.CanAddBroker(userID); err != nil {
			tb.sendMessage(chatID, "❌ "+err.Error())
			return
		}
	}
	creds := broker.Credentials{"initial_balance": "10000"}
	if err := broker.SaveBrandConnection(pg, tb.cfg.App.BrokerEncryptionKey, userID, "mock", "Demo account", creds); err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "✅ Saved, but account sync failed: "+err.Error())
		return
	}
	bal, _ := b.GetBalance()
	tb.sendMessage(chatID, fmt.Sprintf("✅ Connected *%s* (demo)\nBalance: $%.2f\n\nTry /balance or /positions", provider, bal))
}
