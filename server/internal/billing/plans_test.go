package billing

import "testing"

func TestMaxProjects(t *testing.T) {
	if MaxProjects(PlanFree) != 1 {
		t.Fatal("free should allow 1 project")
	}
	if MaxProjects(PlanPro) != 3 {
		t.Fatal("pro should allow 3 projects")
	}
}

func TestCanCustomizeTheme(t *testing.T) {
	if CanCustomizeTheme(PlanFree) {
		t.Fatal("free cannot customize")
	}
	if !CanCustomizeTheme(PlanPro) {
		t.Fatal("pro can customize")
	}
}
