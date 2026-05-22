package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"forex-bot/internal/broker"
	"forex-bot/internal/models"
	"forex-bot/internal/utils"
)

// AccountStore persists trading account balances per user.
type AccountStore interface {
	CreateAccount(account *models.Account) error
	GetAccountByUser(userID int64) (*models.Account, error)
	UpdateAccount(account *models.Account) error
}

// SyncFromBroker creates or updates the persisted account row from live broker balances.
func SyncFromBroker(store AccountStore, userID int64, provider string, b broker.Broker) error {
	if store == nil || b == nil {
		return fmt.Errorf("account sync: missing store or broker")
	}
	if provider == "" {
		provider = "mock"
	}

	bal, err := b.GetBalance()
	if err != nil {
		return fmt.Errorf("broker balance: %w", err)
	}
	equity, err := b.GetEquity()
	if err != nil {
		equity = bal
	}

	now := time.Now()
	existing, err := store.GetAccountByUser(userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if errors.Is(err, sql.ErrNoRows) {
		return store.CreateAccount(&models.Account{
			ID:             utils.GenerateID("acc"),
			UserID:         userID,
			BrokerProvider: provider,
			Balance:        bal,
			Equity:         equity,
			UsedMargin:     0,
			FreeMargin:     equity,
			Leverage:       1,
			LastSyncedAt:   now,
			UpdatedAt:      now,
		})
	}

	existing.BrokerProvider = provider
	existing.Balance = bal
	existing.Equity = equity
	existing.FreeMargin = equity - existing.UsedMargin
	if existing.FreeMargin < 0 {
		existing.FreeMargin = equity
	}
	existing.LastSyncedAt = now
	existing.UpdatedAt = now
	return store.UpdateAccount(existing)
}
