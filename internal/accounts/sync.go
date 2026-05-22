package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

// IsNoRows reports whether err is a missing-row result from the database.
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, sql.ErrNoRows) || strings.Contains(strings.ToLower(err.Error()), "no rows")
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
	if err != nil && !IsNoRows(err) {
		return err
	}
	if IsNoRows(err) {
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
