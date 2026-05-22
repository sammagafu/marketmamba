package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const defaultSessionTTL = 365 * 24 * time.Hour

// Issue creates a signed session token for a Telegram user ID.
func Issue(secret string, telegramID int64, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("session secret not configured")
	}
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}
	exp := time.Now().Add(ttl).Unix()
	payload := fmt.Sprintf("%d:%d", telegramID, exp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
	return token, nil
}

// Verify parses and validates a session token.
func Verify(secret, token string) (int64, error) {
	if secret == "" || token == "" {
		return 0, fmt.Errorf("invalid session")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid session format")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid session payload")
	}
	payload := string(payloadBytes)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBytes)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return 0, fmt.Errorf("invalid session signature")
	}

	seg := strings.Split(payload, ":")
	if len(seg) != 2 {
		return 0, fmt.Errorf("invalid session payload")
	}
	uid, err := strconv.ParseInt(seg[0], 10, 64)
	if err != nil {
		return 0, err
	}
	exp, err := strconv.ParseInt(seg[1], 10, 64)
	if err != nil {
		return 0, err
	}
	if time.Now().Unix() > exp {
		return 0, fmt.Errorf("session expired")
	}
	return uid, nil
}
