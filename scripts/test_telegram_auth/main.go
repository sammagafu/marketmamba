// One-off smoke test: validates TELEGRAM_BOT_TOKEN from .env and Mini App initData signing.
// Usage: go run ./scripts/test_telegram_auth/
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"forex-bot/internal/config"
	"forex-bot/internal/telegramlogin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fatal("config: %v", err)
	}
	token := strings.TrimSpace(cfg.Telegram.BotToken)
	if token == "" || strings.HasPrefix(token, "your_") {
		fatal("set TELEGRAM_BOT_TOKEN in .env")
	}

	// 1) Telegram getMe — proves token is accepted by Telegram API
	meURL := "https://api.telegram.org/bot" + token + "/getMe"
	resp, err := http.Get(meURL)
	if err != nil {
		fatal("getMe request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var me struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
			ID       int64  `json:"id"`
		} `json:"result"`
		Description string `json:"description"`
	}
	_ = json.Unmarshal(body, &me)
	if !me.OK {
		fatal("getMe failed: %s", me.Description)
	}
	fmt.Printf("OK  Telegram bot @%s (id %d)\n", me.Result.Username, me.Result.ID)

	// 2) Sign + verify initData like the Mini App sends
	uid := cfg.Telegram.AllowedUserIDs[0]
	if uid == 0 && len(cfg.Telegram.AdminUserIDs) > 0 {
		uid = cfg.Telegram.AdminUserIDs[0]
	}
	if uid == 0 {
		fatal("set TELEGRAM_ALLOWED_USER_IDS or TELEGRAM_ADMIN_USER_IDS in .env for test user id")
	}
	initData := signWebAppInitData(token, uid, "Test", "user")
	wu, err := telegramlogin.VerifyWebAppInitData(token, initData, 24*time.Hour)
	if err != nil {
		fatal("VerifyWebAppInitData: %v", err)
	}
	fmt.Printf("OK  initData signature valid for user %d (%s)\n", wu.ID, wu.FirstName)

	// 3) Optional: hit local API if running
	apiBase := os.Getenv("API_BASE")
	if apiBase == "" {
		apiBase = "http://127.0.0.1:" + cfg.App.HTTPPort
	}
	payload, _ := json.Marshal(map[string]string{"init_data": initData})
	req, _ := http.NewRequest(http.MethodPost, strings.TrimRight(apiBase, "/")+"/api/v1/auth/telegram/webapp", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	aresp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("SKIP API auth (server not reachable at %s): %v\n", apiBase, err)
		os.Exit(0)
	}
	defer aresp.Body.Close()
	abody, _ := io.ReadAll(aresp.Body)
	if aresp.StatusCode != http.StatusOK {
		fmt.Printf("WARN API auth HTTP %d: %s\n", aresp.StatusCode, truncate(string(abody), 200))
		fmt.Println("     Start API + Postgres (docker compose up) for full registration test.")
		os.Exit(1)
	}
	var login map[string]interface{}
	_ = json.Unmarshal(abody, &login)
	if login["session_token"] == nil {
		fatal("API response missing session_token: %s", truncate(string(abody), 200))
	}
	fmt.Printf("OK  API auth returned session for telegram_id %v\n", login["telegram_id"])
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FAIL "+format+"\n", args...)
	os.Exit(1)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// signWebAppInitData builds Telegram Mini App initData (same algorithm as webapp_test.go).
func signWebAppInitData(botToken string, userID int64, firstName, username string) string {
	authDate := time.Now().Unix()
	userJSON := fmt.Sprintf(`{"id":%d,"first_name":"%s","username":"%s"}`, userID, firstName, username)
	vals := url.Values{}
	vals.Set("auth_date", strconv.FormatInt(authDate, 10))
	vals.Set("user", userJSON)
	var pairs []string
	for k, v := range vals {
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)
	dataCheck := strings.Join(pairs, "\n")
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	key := secret.Sum(nil)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(dataCheck))
	vals.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return vals.Encode()
}
