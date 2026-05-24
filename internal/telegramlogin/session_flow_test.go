package telegramlogin

import (
	"testing"
	"time"

	"forex-bot/internal/auth"
)

// TestWebAppToSession mirrors Mini App auth: validate initData → issue session → verify Bearer token.
func TestWebAppToSession(t *testing.T) {
	initData := mustTestWebAppInitData(t, "123456:TEST", 5311857635, "Sam", "sam")

	wu, err := VerifyWebAppInitData("123456:TEST", initData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	const secret = "test-session-secret-32bytes!!"
	token, err := auth.Issue(secret, wu.ID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	uid, err := auth.Verify(secret, token)
	if err != nil {
		t.Fatalf("session verify: %v", err)
	}
	if uid != wu.ID {
		t.Fatalf("uid %d want %d", uid, wu.ID)
	}
}
