package telegramlogin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signLoginWidget(botToken string, data LoginData) string {
	fields := map[string]string{
		"auth_date": strconv.FormatInt(data.AuthDate, 10),
		"id":        strconv.FormatInt(data.ID, 10),
	}
	if data.FirstName != "" {
		fields["first_name"] = data.FirstName
	}
	if data.LastName != "" {
		fields["last_name"] = data.LastName
	}
	if data.Username != "" {
		fields["username"] = data.Username
	}
	if data.PhotoURL != "" {
		fields["photo_url"] = data.PhotoURL
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fields[k])
	}
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(b.String()))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyLoginWidget(t *testing.T) {
	token := "123456:TEST-BOT-TOKEN"
	authDate := time.Now().Unix()
	data := LoginData{
		ID:        5311857635,
		FirstName: "Sam",
		Username:  "sam",
		AuthDate:  authDate,
	}
	data.Hash = signLoginWidget(token, data)
	if err := Verify(token, data, time.Hour); err != nil {
		t.Fatalf("expected valid login: %v", err)
	}
}

func TestVerifyLoginWidgetRejectsBadHash(t *testing.T) {
	data := LoginData{
		ID: 1, AuthDate: time.Now().Unix(), Hash: "deadbeef",
	}
	if err := Verify("123:TOKEN", data, time.Hour); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestVerifyLoginWidgetRejectsExpired(t *testing.T) {
	token := "123:TOKEN"
	data := LoginData{
		ID: 1, AuthDate: time.Now().Add(-48 * time.Hour).Unix(),
	}
	data.Hash = signLoginWidget(token, data)
	if err := Verify(token, data, time.Hour); err == nil {
		t.Fatal("expected expiry error")
	}
}
