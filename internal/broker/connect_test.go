package broker

import (
	"database/sql"
	"testing"

	"forex-bot/internal/models"
)

type memConnStore struct {
	conn     *models.BrokerConnection
	accounts map[int64]*models.Account
}

func (m *memConnStore) UpsertBrokerConnection(c *models.BrokerConnection) error {
	m.conn = c
	return nil
}

func (m *memConnStore) GetActiveBrokerConnection(userID int64) (*models.BrokerConnection, error) {
	if m.conn != nil && m.conn.UserID == userID {
		return m.conn, nil
	}
	return nil, nil
}

func (m *memConnStore) CreateAccount(a *models.Account) error {
	if m.accounts == nil {
		m.accounts = map[int64]*models.Account{}
	}
	m.accounts[a.UserID] = a
	return nil
}

func (m *memConnStore) GetAccountByUser(userID int64) (*models.Account, error) {
	if m.accounts == nil {
		return nil, sql.ErrNoRows
	}
	a, ok := m.accounts[userID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return a, nil
}

func (m *memConnStore) UpdateAccount(a *models.Account) error {
	if m.accounts == nil {
		m.accounts = map[int64]*models.Account{}
	}
	m.accounts[a.UserID] = a
	return nil
}

func TestSaveConnectionMock(t *testing.T) {
	store := &memConnStore{}
	if err := SaveConnection(store, "test-encryption-key-32bytes!!", 123, "mock", "", Credentials{"initial_balance": "5000"}); err != nil {
		t.Fatal(err)
	}
	if store.conn == nil || store.conn.Provider != "mock" {
		t.Fatalf("expected mock connection, got %+v", store.conn)
	}
}

func TestSaveConnectionOANDA(t *testing.T) {
	store := &memConnStore{}
	err := SaveConnection(store, "test-encryption-key-32bytes!!", 123, "oanda", "", Credentials{
		"api_token":  "test-token",
		"account_id": "101-001-1234567-001",
		"practice":   "true",
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.conn.Provider != "oanda" {
		t.Fatalf("expected oanda, got %s", store.conn.Provider)
	}
}

func TestSaveConnectionRejectsAlpaca(t *testing.T) {
	store := &memConnStore{}
	err := SaveConnection(store, "test-encryption-key-32bytes!!", 123, "alpaca", "", Credentials{"api_key": "x", "api_secret": "y"})
	if err == nil {
		t.Fatal("expected error for alpaca")
	}
}
