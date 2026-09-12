package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sooapps/sooauth/server/internal/billing"
)

func (s *Server) handleDashboardBilling(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		overview, err := s.billing.Overview(r.Context(), dash.account, s.cfg.BillingManualUpgrade, s.cfg.PublicBeta)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "billing_failed"})
			return
		}
		writeJSON(w, http.StatusOK, overview)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
	}
}

func (s *Server) handleDashboardBillingUpgrade(w http.ResponseWriter, r *http.Request) {
	dash, ok := s.requireDashboard(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var body struct {
		Plan      string `json:"plan"`
		Provider  string `json:"provider"`
		ReturnURL string `json:"return_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}

	returnURL := strings.TrimSpace(body.ReturnURL)
	if returnURL == "" {
		returnURL = strings.TrimRight(s.cfg.AppURL, "/") + "/dashboard/#billing"
	}

	result, err := s.billing.Upgrade(r.Context(), dash.user.ID, billing.UpgradeInput{
		Plan:      strings.TrimSpace(body.Plan),
		Provider:  strings.TrimSpace(body.Provider),
		ReturnURL: returnURL,
		Email:     dash.user.Email,
	})
	if errors.Is(err, billing.ErrInvalidPlan) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_plan"})
		return
	}
	if errors.Is(err, billing.ErrSamePlan) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "same_plan",
			"message": "You are already on this plan.",
		})
		return
	}
	if errors.Is(err, billing.ErrDowngradeBlocked) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "downgrade_blocked",
			"message": "Downgrades are not supported yet. Contact support.",
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "upgrade_failed"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}
