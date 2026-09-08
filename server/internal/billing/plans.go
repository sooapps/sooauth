package billing

const (
	PlanFree     = "free"
	PlanPro      = "pro"
	PlanBusiness = "business"
)

func MaxProjects(plan string) int {
	switch plan {
	case PlanPro:
		return 3
	case PlanBusiness:
		return 10
	default:
		return 1
	}
}

func CanCustomizeTheme(plan string) bool {
	return plan != PlanFree
}

func CanHideBranding(plan string) bool {
	return plan != PlanFree
}

func CanCustomSocialCallback(plan string) bool {
	return plan != PlanFree
}

func PlanLabel(plan string) string {
	switch plan {
	case PlanPro:
		return "Pro"
	case PlanBusiness:
		return "Business"
	default:
		return "Free"
	}
}
