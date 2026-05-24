package telegramlogin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func mustTestWebAppInitData(t *testing.T, token string, id int64, firstName, username string) string {
	t.Helper()
	authDate := time.Now().Unix()
	user := `{"id":` + strconv.FormatInt(id, 10) + `,"first_name":"` + firstName + `","username":"` + username + `"}`
	vals := url.Values{}
	vals.Set("auth_date", strconv.FormatInt(authDate, 10))
	vals.Set("user", user)
	var pairs []string
	for k, v := range vals {
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)
	dataCheck := strings.Join(pairs, "\n")
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))
	key := secret.Sum(nil)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(dataCheck))
	vals.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return vals.Encode()
}

func TestVerifyWebAppInitData(t *testing.T) {
	initData := mustTestWebAppInitData(t, "123456:TEST", 5311857635, "Sam", "sam")
	wu, err := VerifyWebAppInitData("123456:TEST", initData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if wu.ID != 5311857635 || wu.FirstName != "Sam" {
		t.Fatalf("unexpected user: %+v", wu)
	}
}

func TestVerifyWebAppInitDataRejectsBadHash(t *testing.T) {
	initData := mustTestWebAppInitData(t, "123:TOK", 1, "A", "a")
	vals, err := url.ParseQuery(initData)
	if err != nil {
		t.Fatal(err)
	}
	vals.Set("hash", "0000000000000000000000000000000000000000000000000000000000000000")
	if _, err := VerifyWebAppInitData("123:TOK", vals.Encode(), time.Hour); err == nil {
		t.Fatal("expected invalid signature")
	}
}

func TestVerifyWebAppInitDataRejectsExpired(t *testing.T) {
	token := "123:TOK"
	authDate := time.Now().Add(-48 * time.Hour).Unix()
	user := `{"id":1,"first_name":"A"}`
	vals := url.Values{}
	vals.Set("auth_date", strconv.FormatInt(authDate, 10))
	vals.Set("user", user)
	var pairs []string
	for k, v := range vals {
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))
	key := secret.Sum(nil)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(strings.Join(pairs, "\n")))
	vals.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	if _, err := VerifyWebAppInitData(token, vals.Encode(), time.Hour); err == nil {
		t.Fatal("expected expiry error")
	}
}
