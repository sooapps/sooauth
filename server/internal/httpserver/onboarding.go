package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sooapps/sooauth/server/internal/store"
)

const defaultRedirectURI = "http://localhost:3000/api/auth/callback/oidc"

var onboardingSteps = []string{"project", "sign_in", "redirect", "integrate", "test"}

var appTypes = map[string]bool{"nextjs": true, "spa": true, "backend": true, "custom": true, "widget": true}

// handleDashboardOnboarding returns the setup checklist of the active project.
// Steps are ticked from real project state where possible (users, sessions,
// redirect URIs) and otherwise from marks the dashboard saved.
func (s *Server) handleDashboardOnboarding(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	state, err := s.tenants.Onboarding(ctx, dash.tenant.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "onboarding_failed"})
		return
	}

	if r.Method == http.MethodPut {
		var body struct {
			Dismissed  *bool   `json:"dismissed"`
			AppType    *string `json:"app_type"`
			Complete   string  `json:"complete"`
			Uncomplete string  `json:"uncomplete"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		if body.Dismissed != nil {
			state.Dismissed = *body.Dismissed
		}
		if body.AppType != nil {
			if !appTypes[*body.AppType] {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_app_type"})
				return
			}
			state.AppType = *body.AppType
		}
		if step := strings.TrimSpace(body.Complete); step != "" && isOnboardingStep(step) && !state.Has(step) {
			state.Completed = append(state.Completed, step)
		}
		if step := strings.TrimSpace(body.Uncomplete); step != "" {
			kept := state.Completed[:0]
			for _, c := range state.Completed {
				if c != step {
					kept = append(kept, c)
				}
			}
			state.Completed = kept
		}
		if err := s.tenants.SetOnboarding(ctx, dash.tenant.ID, state); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save_failed"})
			return
		}
	}

	signals, _ := s.tenants.OnboardingSignals(ctx, dash.tenant.ID)
	customRedirect := false
	if dash.client != nil {
		for _, uri := range dash.client.RedirectURIs {
			if uri != defaultRedirectURI {
				customRedirect = true
			}
		}
	}
	defaultAuth := store.DefaultAuthConfig()
	cfg := dash.tenant.AuthConfig
	authChanged := cfg.PrimaryAuthMode != defaultAuth.PrimaryAuthMode ||
		len(cfg.AllowedIdentifiers) != 1 || cfg.MFAPolicy != defaultAuth.MFAPolicy

	done := map[string]bool{
		"project":   state.Has("project") || state.AppType != "",
		"sign_in":   state.Has("sign_in") || authChanged || signals.Providers > 0,
		"redirect":  state.Has("redirect") || customRedirect,
		"integrate": state.Has("integrate") || signals.EndUsers > 0,
		"test":      state.Has("test") || signals.EndUserSignIn,
	}
	steps := make([]map[string]any, 0, len(onboardingSteps))
	count := 0
	for _, id := range onboardingSteps {
		if done[id] {
			count++
		}
		steps = append(steps, map[string]any{"id": id, "done": done[id]})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"dismissed":  state.Dismissed,
		"app_type":   state.AppType,
		"steps":      steps,
		"done_count": count,
		"total":      len(onboardingSteps),
		"complete":   count == len(onboardingSteps),
	})
}

func isOnboardingStep(step string) bool {
	for _, s := range onboardingSteps {
		if s == step {
			return true
		}
	}
	return false
}
