package telegramlogin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WebAppUser is parsed from validated Mini App initData.
type WebAppUser struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

// VerifyWebAppInitData validates Telegram Mini App initData (see core.telegram.org/bots/webapps).
func VerifyWebAppInitData(botToken, initData string, maxAge time.Duration) (*WebAppUser, error) {
	if botToken == "" {
		return nil, fmt.Errorf("bot token not configured")
	}
	initData = strings.TrimSpace(initData)
	if initData == "" {
		return nil, fmt.Errorf("missing init data")
	}
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}

	vals, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("invalid init data")
	}
	receivedHash := vals.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("missing hash")
	}

	var pairs []string
	for k, v := range vals {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)
	dataCheck := strings.Join(pairs, "\n")

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	key := secret.Sum(nil)

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(dataCheck))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(receivedHash)) {
		return nil, fmt.Errorf("invalid init data signature")
	}

	authDate, _ := strconv.ParseInt(vals.Get("auth_date"), 10, 64)
	if authDate > 0 && time.Now().Unix()-authDate > int64(maxAge.Seconds()) {
		return nil, fmt.Errorf("init data expired")
	}

	userJSON := vals.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("missing user in init data")
	}
	var raw struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.Unmarshal([]byte(userJSON), &raw); err != nil {
		return nil, fmt.Errorf("invalid user payload")
	}
	if raw.ID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	return &WebAppUser{
		ID:        raw.ID,
		Username:  raw.Username,
		FirstName: raw.FirstName,
		LastName:  raw.LastName,
	}, nil
}
