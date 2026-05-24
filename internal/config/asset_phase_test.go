package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func testConfig(t *testing.T) *Config {
	t.Helper()
	t.Setenv("TELEGRAM_BOT_TOKEN", "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	return cfg
}

func TestPhasedSignalCatalog_bitcoinPhase(t *testing.T) {
	t.Setenv("ASSET_PHASE", "bitcoin")
	t.Setenv("FORCE_FULL_ASSETS", "false")
	cfg := testConfig(t)
	cfg.Phase.SetPaidCount(10)

	fx, idx, cry := cfg.PhasedSignalCatalog()
	if len(fx) != 0 || len(idx) != 0 {
		t.Fatalf("expected empty forex/indexes, got fx=%v idx=%v", fx, idx)
	}
	if len(cry) == 0 {
		t.Fatal("expected crypto symbols")
	}
}

func TestIsFullAssetCatalog_unlockAtThreshold(t *testing.T) {
	t.Setenv("ASSET_PHASE", "bitcoin")
	t.Setenv("UNLOCK_MIN_PAID_SUBSCRIBERS", "100")
	t.Setenv("FORCE_FULL_ASSETS", "false")
	cfg := testConfig(t)
	cfg.Phase.SetPaidCount(99)
	if cfg.IsFullAssetCatalog() {
		t.Fatal("expected locked at 99")
	}
	cfg.Phase.SetPaidCount(100)
	if !cfg.IsFullAssetCatalog() {
		t.Fatal("expected unlocked at 100")
	}
	fx, idx, cry := cfg.PhasedSignalCatalog()
	if len(fx) == 0 || len(cry) == 0 {
		t.Fatalf("expected full catalog: fx=%v idx=%v cry=%v", fx, idx, cry)
	}
}

func TestCommunityPhasePublic_noSubscriberCount(t *testing.T) {
	t.Setenv("ASSET_PHASE", "bitcoin")
	cfg := testConfig(t)
	copy := cfg.CommunityPhasePublic()
	b, _ := json.Marshal(copy)
	s := string(b)
	for _, forbidden := range []string{"paid_subscriber", "unlock_min", "/100", "42/"} {
		if strings.Contains(strings.ToLower(s), forbidden) {
			t.Fatalf("public copy must not contain %q: %s", forbidden, s)
		}
	}
	if copy.AITrainingNote == "" {
		t.Fatal("expected ai_training_note default")
	}
}
