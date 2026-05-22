package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"forex-bot/internal/config"
)

func TestStaticAssetsRoute(t *testing.T) {
	cfg := &config.Config{App: config.AppConfig{EnableWeb: true}}
	s := NewServer(cfg, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/assets/index-BVvY9Vuq.css", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /assets/... status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/css; charset=utf-8" && ct != "text/css" {
		t.Fatalf("Content-Type = %q, want text/css", ct)
	}
}
