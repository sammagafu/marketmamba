package broker

import (
	"fmt"
	"time"

	"forex-bot/internal/models"
	"forex-bot/internal/secrets"
	"forex-bot/internal/utils"
)

// ConnectionStore persists encrypted broker credentials per user.
type ConnectionStore interface {
	UpsertBrokerConnection(conn *models.BrokerConnection) error
	GetActiveBrokerConnection(userID int64) (*models.BrokerConnection, error)
}

// SaveConnection validates credentials, encrypts them, and activates the broker for a user.
func SaveConnection(store ConnectionStore, encryptionKey string, userID int64, provider, label string, creds Credentials) error {
	if !IsLiveProvider(provider) {
		return fmt.Errorf("broker %q is not available yet — use mock (demo)", provider)
	}
	if label == "" {
		label = defaultLabel(provider, creds)
	}
	if err := ValidateCredentials(provider, creds); err != nil {
		return err
	}
	if _, err := NewFromProvider(provider, creds); err != nil {
		return err
	}
	enc, err := secrets.EncryptJSON(encryptionKey, creds)
	if err != nil {
		return err
	}
	now := time.Now()
	conn := &models.BrokerConnection{
		ID:               utils.GenerateID("broker"),
		UserID:           userID,
		Provider:         provider,
		Label:            label,
		CredentialsEnc:   enc,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return store.UpsertBrokerConnection(conn)
}

func defaultLabel(provider string, creds Credentials) string {
	switch provider {
	case "mock":
		return "Demo account"
	case "metaapi":
		if s := credsHintServer(creds); s != "" {
			return "MT5 " + s
		}
		return "MetaAPI MT5"
	default:
		return provider
	}
}

// ValidateCredentials checks required fields from the broker registry.
func ValidateCredentials(provider string, creds Credentials) error {
	if creds == nil {
		creds = Credentials{}
	}
	if provider == "metaapi" {
		return ValidateMetaAPICredentials(creds)
	}
	for _, bt := range SupportedBrokerTypes() {
		if bt.ID != provider {
			continue
		}
		for _, f := range bt.Fields {
			if !f.Required {
				continue
			}
			if creds[f.Key] == "" {
				return fmt.Errorf("%s is required", f.Label)
			}
		}
		return nil
	}
	return fmt.Errorf("unknown broker provider: %s", provider)
}
