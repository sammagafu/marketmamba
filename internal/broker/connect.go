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
// Use brandID non-empty to resolve a catalog brand (deriv, exness) to a technical adapter.
func SaveConnection(store ConnectionStore, encryptionKey string, userID int64, provider, label string, creds Credentials) error {
	return saveConnection(store, encryptionKey, userID, "", provider, label, creds)
}

// SaveBrandConnection saves using a user-facing brand id (deriv, exness, tickmill).
func SaveBrandConnection(store ConnectionStore, encryptionKey string, userID int64, brandID, label string, creds Credentials) error {
	return saveConnection(store, encryptionKey, userID, brandID, "", label, creds)
}

func saveConnection(store ConnectionStore, encryptionKey string, userID int64, brandID, provider, label string, creds Credentials) error {
	if brandID != "" {
		var err error
		provider, creds, label, err = ResolveBrandConnection(brandID, label, creds)
		if err != nil {
			return err
		}
	}
	if !IsLiveProvider(provider) {
		return fmt.Errorf("broker %q is not available yet — use mock (demo)", provider)
	}
	if label == "" {
		label = defaultLabel(provider, creds)
	}
	creds = ApplySharedMetaAPIToken(creds)
	if err := ValidateCredentials(provider, creds); err != nil {
		return err
	}
	b, err := NewFromProvider(provider, creds)
	if err != nil {
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
		IsPrimary:        true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := store.UpsertBrokerConnection(conn); err != nil {
		return err
	}
	// Best-effort: row syncs on /balance, dashboard test, or first trade if broker is offline here.
	_ = syncTradingAccount(store, userID, provider, b)
	return nil
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

// ValidateCredentials checks required fields for a technical provider or brand.
func ValidateCredentials(provider string, creds Credentials) error {
	return validateCredentials(provider, "", creds)
}

// ValidateBrandCredentials validates credentials for a catalog brand.
func ValidateBrandCredentials(brandID string, creds Credentials) error {
	brand, ok := BrandByID(brandID)
	if !ok {
		return fmt.Errorf("unknown broker brand: %s", brandID)
	}
	_, merged, _, err := ResolveBrandConnection(brandID, "", creds)
	if err != nil {
		return err
	}
	if a, ok := getAdapter(brand.AdapterID); ok && a.Validate != nil {
		if err := a.Validate(merged); err != nil {
			return err
		}
	}
	for _, f := range fieldsForBrand(*brand) {
		if !f.Required {
			continue
		}
		if merged[f.Key] == "" {
			return fmt.Errorf("%s is required", f.Label)
		}
	}
	return nil
}

func validateCredentials(provider, brandID string, creds Credentials) error {
	if creds == nil {
		creds = Credentials{}
	}
	if brandID != "" {
		return ValidateBrandCredentials(brandID, creds)
	}
	if a, ok := getAdapter(provider); ok && a.Validate != nil {
		return a.Validate(creds)
	}
	return fmt.Errorf("unknown broker provider: %s", provider)
}
