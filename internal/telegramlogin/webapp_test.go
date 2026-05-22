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

func TestVerifyWebAppInitData(t *testing.T) {
	token := "123456:TEST"
	authDate := time.Now().Unix()
	user := `{"id":5311857635,"first_name":"Sam","username":"sam"}`
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
	initData := vals.Encode()

	wu, err := VerifyWebAppInitData(token, initData, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if wu.ID != 5311857635 || wu.FirstName != "Sam" {
		t.Fatalf("unexpected user: %+v", wu)
	}
}
