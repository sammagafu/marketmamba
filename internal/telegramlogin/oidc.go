package telegramlogin

import (
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	telegramIssuer = "https://oauth.telegram.org"
	telegramJWKS   = "https://oauth.telegram.org/.well-known/jwks.json"
)

// OIDCUser is decoded Telegram Login id_token claims.
type OIDCUser struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	PhotoURL   string
	Phone      string
}

type oidcClaims struct {
	jwt.RegisteredClaims
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
	PhoneNumber       string `json:"phone_number"`
}

var jwks keyfunc.Keyfunc

func initJWKS() (keyfunc.Keyfunc, error) {
	if jwks != nil {
		return jwks, nil
	}
	k, err := keyfunc.NewDefault([]string{telegramJWKS})
	if err != nil {
		return nil, err
	}
	jwks = k
	return jwks, nil
}

// VerifyIDToken validates Telegram OIDC id_token (oauth.telegram.org widget).
func VerifyIDToken(clientID, idToken string) (*OIDCUser, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, fmt.Errorf("telegram client id not configured")
	}
	if idToken == "" {
		return nil, fmt.Errorf("missing id_token")
	}
	keys, err := initJWKS()
	if err != nil {
		return nil, fmt.Errorf("jwks: %w", err)
	}
	claims := &oidcClaims{}
	token, err := jwt.ParseWithClaims(idToken, claims, keys.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("invalid id_token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid id_token")
	}
	if claims.Issuer != telegramIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	audOK := false
	for _, aud := range claims.Audience {
		if aud == clientID {
			audOK = true
			break
		}
	}
	if !audOK {
		return nil, fmt.Errorf("audience mismatch — check TELEGRAM_BOT_CLIENT_ID matches BotFather")
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("id_token expired")
	}

	uid := claims.ID
	if uid == 0 && claims.Subject != "" {
		fmt.Sscan(claims.Subject, &uid)
	}
	if uid == 0 {
		return nil, fmt.Errorf("missing user id in token")
	}

	first, last := splitName(claims.Name)
	return &OIDCUser{
		TelegramID: uid,
		Username:   claims.PreferredUsername,
		FirstName:  first,
		LastName:   last,
		PhotoURL:   claims.Picture,
		Phone:      claims.PhoneNumber,
	}, nil
}

func splitName(name string) (first, last string) {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	if len(parts) > 0 {
		first = parts[0]
	}
	if len(parts) > 1 {
		last = parts[1]
	}
	return first, last
}
