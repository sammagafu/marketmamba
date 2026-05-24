package tier

import "testing"

func TestForPlan(t *testing.T) {
	trial := ForPlan("trial")
	if trial.MaxBrokerAccounts != 1 || trial.MaxSignalsPerPeriod != 30 {
		t.Fatalf("trial limits: %+v", trial)
	}
	pro := ForPlan("pro")
	if pro.MaxBrokerAccounts != 5 {
		t.Fatalf("pro: %+v", pro)
	}
	unknown := ForPlan("something_else")
	if unknown.Plan != "trial" {
		t.Fatalf("unknown should map to trial defaults, got %+v", unknown)
	}
}

func TestAllPlans(t *testing.T) {
	if len(AllPlans()) < 4 {
		t.Fatal("expected at least 4 plans")
	}
}
