package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forex-bot/internal/config"
)

func TestStaticAssetsRoute(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("web", "dist", "assets"))
	if err != nil {
		t.Skip("web/dist/assets not built — run make web-build")
	}
	var cssName string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".css") {
			cssName = e.Name()
			break
		}
	}
	if cssName == "" {
		t.Skip("no CSS bundle in web/dist/assets")
	}

	cfg := &config.Config{App: config.AppConfig{EnableWeb: true}}
	s := NewServer(cfg, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/assets/"+cssName, nil)
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
