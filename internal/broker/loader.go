package broker

import (
	"forex-bot/internal/secrets"
	"forex-bot/internal/storage"
)

// ResolveBroker returns DB-configured broker or falls back to provider from env.
func ResolveBroker(store *storage.PostgresStorage, userID int64, encryptionKey, envProvider string) (Broker, string, error) {
	b, provider, err := resolveBroker(store, userID, encryptionKey, envProvider)
	return b, provider, err
}

// ResolveBrokerAndSync also upserts the accounts table from broker balances.
func ResolveBrokerAndSync(store *storage.PostgresStorage, userID int64, encryptionKey, envProvider string) (Broker, error) {
	b, provider, err := resolveBroker(store, userID, encryptionKey, envProvider)
	if err != nil {
		return nil, err
	}
	if syncErr := syncTradingAccount(store, userID, provider, b); syncErr != nil {
		return b, syncErr
	}
	return b, nil
}

func resolveBroker(store *storage.PostgresStorage, userID int64, encryptionKey, envProvider string) (Broker, string, error) {
	conn, err := store.GetActiveBrokerConnection(userID)
	if err != nil {
		return nil, "", err
	}
	if conn != nil {
		var creds Credentials
		if err := secrets.DecryptJSON(encryptionKey, conn.CredentialsEnc, &creds); err != nil {
			return nil, "", err
		}
		b, err := NewFromProvider(conn.Provider, creds)
		if err != nil {
			return nil, "", err
		}
		return b, conn.Provider, nil
	}
	provider := envProvider
	if provider == "" {
		provider = "mock"
	}
	b, err := NewFromProvider(provider, nil)
	return b, provider, err
}
