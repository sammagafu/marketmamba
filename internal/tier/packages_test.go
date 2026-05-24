package tier

import "testing"

func TestPublicPackages(t *testing.T) {
	pkgs := PublicPackages(10, 5)
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 public plans, got %d", len(pkgs))
	}
	if !pkgs[1].Recommended || pkgs[1].ID != "monthly" {
		t.Fatal("monthly should be recommended")
	}
	if !pkgs[2].ContactOnly {
		t.Fatal("pro should be contact only")
	}
}
