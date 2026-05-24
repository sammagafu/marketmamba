package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forex-bot/internal/config"
)

func TestTelegramWebAppAuthRejectsInvalidInitData(t *testing.T) {
	cfg := &config.Config{
		App:      config.AppConfig{WebSessionSecret: "test-web-session-secret"},
		Telegram: config.TelegramConfig{BotToken: "123456:TEST"},
	}
	s := NewServer(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	body, _ := json.Marshal(map[string]string{"init_data": "auth_date=1&hash=bad&user=%7B%22id%22%3A1%7D"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram/webapp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.handleTelegramWebAppAuth(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d want 401 body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "signature") && !strings.Contains(rec.Body.String(), "invalid") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestTelegramWebAppAuthRejectsEmptyBody(t *testing.T) {
	cfg := &config.Config{
		Telegram: config.TelegramConfig{BotToken: "123:TEST"},
	}
	s := NewServer(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/telegram/webapp", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.handleTelegramWebAppAuth(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d want 401", rec.Code)
	}
}
