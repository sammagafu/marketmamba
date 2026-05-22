package broker

import (
	"fmt"

	"forex-bot/internal/accounts"
)

func syncTradingAccount(store interface{}, userID int64, provider string, b Broker) error {
	acct := accounts.AccountStoreFrom(store)
	if acct == nil {
		return fmt.Errorf("account storage unavailable")
	}
	if b == nil {
		return fmt.Errorf("broker unavailable")
	}
	if provider == "" {
		provider = "mock"
	}
	if err := accounts.SyncFromBroker(acct, userID, provider, b); err != nil {
		return fmt.Errorf("account sync: %w", err)
	}
	return nil
}
