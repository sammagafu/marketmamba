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
		tb.connectBroker(chatID, userID, provider)
	case "mock", "demo":
		tb.connectBroker(chatID, userID, "mock")
	default:
		tb.sendMessage(chatID, `*Broker connection*

/broker — show current connection
/broker connect — connect Mock (Demo) $10,000
/broker connect mock — same as above

Or use the web dashboard:
https://marketmamba.kkooapp.co.tz`)
	}
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
		tb.sendMessage(chatID, `*No broker connected*

Connect a demo account:
/broker connect

Or open the web dashboard.`)
		return
	}
	tb.sendMessage(chatID, fmt.Sprintf(`*Broker connected*
Provider: *%s*
Label: %s
Updated: %s

/balance — view account`, conn.Provider, conn.Label, conn.UpdatedAt.Format("2006-01-02 15:04")))
}

func (tb *TelegramBot) connectBroker(chatID, userID int64, provider string) {
	pg, ok := tb.storage.(*storage.PostgresStorage)
	if !ok {
		tb.sendMessage(chatID, "❌ Connect via web dashboard")
		return
	}
	creds := broker.Credentials{}
	if provider == "mock" {
		creds["initial_balance"] = "10000"
	}
	if err := broker.SaveConnection(pg, tb.cfg.App.BrokerEncryptionKey, userID, provider, "", creds); err != nil {
		tb.sendMessage(chatID, "❌ "+err.Error())
		return
	}
	b, err := tb.brokerFor(userID)
	if err != nil {
		tb.sendMessage(chatID, "✅ Saved, but could not load broker: "+err.Error())
		return
	}
	bal, _ := b.GetBalance()
	tb.sendMessage(chatID, fmt.Sprintf("✅ Connected *%s* (demo)\nBalance: $%.2f\n\nTry /balance or /positions", provider, bal))
}
