package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forex-bot/internal/config"
)

func TestHandlePublicConfig_noSubscriberCounts(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	t.Setenv("ASSET_PHASE", "bitcoin")
	t.Setenv("FORCE_FULL_ASSETS", "false")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.Phase = config.NewCommunityPhaseRuntime()
	cfg.Phase.SetPaidCount(42)

	s := &Server{cfg: cfg, mux: http.NewServeMux()}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	rec := httptest.NewRecorder()
	s.handlePublicConfig(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"paid_subscriber_count", "unlock_min_paid_subscribers"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("public config must not expose %q", key)
		}
	}
	if payload["ai_training_note"] == nil || payload["ai_training_note"] == "" {
		t.Fatal("expected ai_training_note")
	}
	if payload["asset_phase"] != config.PublicPhaseCommunityLaunch {
		t.Fatalf("asset_phase=%v", payload["asset_phase"])
	}
	body := strings.ToLower(rec.Body.String())
	if strings.Contains(body, "42/") || strings.Contains(body, "/100") {
		t.Fatalf("public config must not contain quota fractions: %s", rec.Body.String())
	}
}
