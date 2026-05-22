package accounts

import (
	"database/sql"
	"testing"

	"forex-bot/internal/broker"
	"forex-bot/internal/models"
)

type memAccountStore struct {
	accounts map[int64]*models.Account
}

func (m *memAccountStore) CreateAccount(a *models.Account) error {
	m.accounts[a.UserID] = a
	return nil
}

func (m *memAccountStore) GetAccountByUser(userID int64) (*models.Account, error) {
	a, ok := m.accounts[userID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return a, nil
}

func (m *memAccountStore) UpdateAccount(a *models.Account) error {
	m.accounts[a.UserID] = a
	return nil
}

func TestSyncFromBrokerCreatesAccount(t *testing.T) {
	store := &memAccountStore{accounts: map[int64]*models.Account{}}
	b := broker.NewMockBroker(7500)

	if err := SyncFromBroker(store, 99, "mock", b); err != nil {
		t.Fatal(err)
	}
	acc, err := store.GetAccountByUser(99)
	if err != nil {
		t.Fatal(err)
	}
	if acc.Balance != 7500 || acc.BrokerProvider != "mock" {
		t.Fatalf("unexpected account: %+v", acc)
	}
}

func TestSyncFromBrokerUpdatesBalance(t *testing.T) {
	store := &memAccountStore{accounts: map[int64]*models.Account{}}
	b := broker.NewMockBroker(5000)
	if err := SyncFromBroker(store, 1, "mock", b); err != nil {
		t.Fatal(err)
	}

	b2 := broker.NewMockBroker(12000)
	if err := SyncFromBroker(store, 1, "mock", b2); err != nil {
		t.Fatal(err)
	}
	acc, _ := store.GetAccountByUser(1)
	if acc.Balance != 12000 {
		t.Fatalf("expected 12000 balance, got %.2f", acc.Balance)
	}
}
