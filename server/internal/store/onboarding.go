package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// Onboarding is the dashboard setup state of a project.
type Onboarding struct {
	// Dismissed hides the setup wizard; the checklist stays available.
	Dismissed bool `json:"dismissed"`
	// AppType is what the owner is building (nextjs, spa, backend, custom, widget).
	AppType string `json:"app_type,omitempty"`
	// Completed holds checklist steps the owner marked done by hand.
	Completed []string `json:"completed,omitempty"`
}

func (o Onboarding) Has(step string) bool {
	for _, s := range o.Completed {
		if s == step {
			return true
		}
	}
	return false
}

func (t *Tenants) Onboarding(ctx context.Context, id uuid.UUID) (Onboarding, error) {
	var raw []byte
	var out Onboarding
	if err := t.db.QueryRow(ctx, `SELECT onboarding FROM tenants WHERE id = $1`, id).Scan(&raw); err != nil {
		return out, err
	}
	_ = json.Unmarshal(raw, &out)
	return out, nil
}

func (t *Tenants) SetOnboarding(ctx context.Context, id uuid.UUID, o Onboarding) error {
	raw, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = t.db.Exec(ctx, `UPDATE tenants SET onboarding = $2 WHERE id = $1`, id, raw)
	return err
}

// OnboardingSignals are facts about a project used to tick checklist steps.
type OnboardingSignals struct {
	EndUsers      int
	EndUserSignIn bool
	Providers     int
}

func (t *Tenants) OnboardingSignals(ctx context.Context, id uuid.UUID) (OnboardingSignals, error) {
	var out OnboardingSignals
	err := t.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users u WHERE u.tenant_id = $1 AND NOT u.platform_owner),
			EXISTS (
				SELECT 1 FROM sessions s JOIN users u ON u.id = s.user_id
				WHERE u.tenant_id = $1 AND NOT u.platform_owner
			),
			(SELECT COUNT(*) FROM oauth_providers p WHERE p.tenant_id = $1 AND p.enabled)
	`, id).Scan(&out.EndUsers, &out.EndUserSignIn, &out.Providers)
	return out, err
}
