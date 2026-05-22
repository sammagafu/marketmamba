package telegramlogin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LoginData is the payload from Telegram Login Widget.
type LoginData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

func Verify(botToken string, data LoginData, maxAge time.Duration) error {
	if botToken == "" {
		return fmt.Errorf("bot token not configured")
	}
	if data.Hash == "" {
		return fmt.Errorf("missing hash")
	}
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}
	if time.Now().Unix()-data.AuthDate > int64(maxAge.Seconds()) {
		return fmt.Errorf("login data expired")
	}

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
	dataCheck := b.String()

	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheck))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(data.Hash)) {
		return fmt.Errorf("invalid telegram login signature")
	}
	return nil
}
