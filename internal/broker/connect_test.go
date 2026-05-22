package broker

import (
	"testing"

	"forex-bot/internal/models"
)

type memConnStore struct {
	conn *models.BrokerConnection
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

func TestSaveConnectionMock(t *testing.T) {
	store := &memConnStore{}
	if err := SaveConnection(store, "test-encryption-key-32bytes!!", 123, "mock", "", Credentials{"initial_balance": "5000"}); err != nil {
		t.Fatal(err)
	}
	if store.conn == nil || store.conn.Provider != "mock" {
		t.Fatalf("expected mock connection, got %+v", store.conn)
	}
}

func TestSaveConnectionRejectsComingSoon(t *testing.T) {
	store := &memConnStore{}
	err := SaveConnection(store, "test-encryption-key-32bytes!!", 123, "oanda", "", Credentials{"api_token": "x", "account_id": "y"})
	if err == nil {
		t.Fatal("expected error for oanda")
	}
}
