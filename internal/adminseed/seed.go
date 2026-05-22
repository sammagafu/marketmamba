package adminseed

import (
	"fmt"
	"os"
	"time"

	"forex-bot/internal/auth"
	"forex-bot/internal/models"
	"forex-bot/internal/storage"
	"forex-bot/internal/utils"
)

// Run creates or updates web admin from ADMIN_EMAIL, ADMIN_PASSWORD, ADMIN_TELEGRAM_ID.
func Run(store *storage.PostgresStorage) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("set ADMIN_EMAIL and ADMIN_PASSWORD in .env")
	}
	var telegramID int64 = 5311857635
	if s := os.Getenv("ADMIN_TELEGRAM_ID"); s != "" {
		if _, err := fmt.Sscan(s, &telegramID); err != nil {
			return fmt.Errorf("invalid ADMIN_TELEGRAM_ID: %w", err)
		}
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now()
	return store.UpsertWebAdmin(&models.WebAdmin{
		ID:           utils.GenerateID("adm"),
		Email:        email,
		PasswordHash: hash,
		TelegramID:   telegramID,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}
