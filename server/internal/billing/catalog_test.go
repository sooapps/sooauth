package billing

import "testing"

func TestValidPlan(t *testing.T) {
	if !ValidPlan(PlanPro) || ValidPlan("enterprise") {
		t.Fatal("plan validation mismatch")
	}
}

func TestPlanRank(t *testing.T) {
	if PlanRank(PlanBusiness) <= PlanRank(PlanPro) {
		t.Fatal("business should rank above pro")
	}
}

func TestAllPlansIncludesFree(t *testing.T) {
	plans := AllPlans()
	if len(plans) != 3 || plans[0].ID != PlanFree {
		t.Fatal("expected three plans starting with free")
	}
}
