package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/emailverify"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
)

type AuthConfig struct {
	AllowedIdentifiers []string `json:"allowed_identifiers"` // "email", "username", "phone"
	// PrimaryAuthMode: "password" (password only), "otp" (one-time code only) or
	// "both" (the user picks password OR code on the sign-in screen).
	PrimaryAuthMode string `json:"primary_auth_mode"`
	// MFAPolicy controls a second step after a password sign-in:
	// "off", "optional" (users opt in) or "required" (everyone).
	MFAPolicy string `json:"mfa_policy"`
	// MFAMethods lists the second-step methods users may use: "email_otp", "totp".
	MFAMethods []string `json:"mfa_methods"`
	// PhoneOTPViaWebhook delivers phone codes inside the otp.sent webhook so the
	// project can send the SMS with its own provider.
	PhoneOTPViaWebhook bool `json:"phone_otp_via_webhook"`
}

const (
	AuthModePassword = "password"
	AuthModeOTP      = "otp"
	AuthModeBoth     = "both"

	MFAPolicyOff      = "off"
	MFAPolicyOptional = "optional"
	MFAPolicyRequired = "required"

	MFAMethodEmailOTP = "email_otp"
	MFAMethodTOTP     = "totp"
)

func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		AllowedIdentifiers: []string{"email"},
		PrimaryAuthMode:    AuthModePassword,
		MFAPolicy:          MFAPolicyOff,
		MFAMethods:         []string{MFAMethodEmailOTP, MFAMethodTOTP},
	}
}

func (c AuthConfig) Normalize() AuthConfig {
	c.AllowedIdentifiers = filterKnown(c.AllowedIdentifiers, "email", "username", "phone")
	if len(c.AllowedIdentifiers) == 0 {
		c.AllowedIdentifiers = []string{"email"}
	}
	switch c.PrimaryAuthMode {
	case AuthModePassword, AuthModeOTP, AuthModeBoth:
	default:
		c.PrimaryAuthMode = AuthModePassword
	}
	switch c.MFAPolicy {
	case MFAPolicyOff, MFAPolicyOptional, MFAPolicyRequired:
	default:
		c.MFAPolicy = MFAPolicyOff
	}
	c.MFAMethods = filterKnown(c.MFAMethods, MFAMethodEmailOTP, MFAMethodTOTP)
	if len(c.MFAMethods) == 0 {
		c.MFAMethods = []string{MFAMethodEmailOTP, MFAMethodTOTP}
	}
	return c
}

func (c AuthConfig) AllowsPassword() bool {
	return c.PrimaryAuthMode != AuthModeOTP
}

func (c AuthConfig) AllowsOTP() bool {
	return c.PrimaryAuthMode == AuthModeOTP || c.PrimaryAuthMode == AuthModeBoth
}

func (c AuthConfig) AllowsIdentifier(kind string) bool {
	for _, id := range c.AllowedIdentifiers {
		if id == kind {
			return true
		}
	}
	return false
}

func (c AuthConfig) AllowsMFAMethod(method string) bool {
	for _, m := range c.MFAMethods {
		if m == method {
			return true
		}
	}
	return false
}

// MFAEnabled reports whether a second step can run for this project at all.
func (c AuthConfig) MFAEnabled() bool {
	return c.MFAPolicy == MFAPolicyOptional || c.MFAPolicy == MFAPolicyRequired
}

// IdentifierKind classifies a sign-in identifier as "email", "phone" or "username".
func IdentifierKind(identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		return "email"
	}
	if strings.HasPrefix(identifier, "+") {
		return "phone"
	}
	digits := 0
	for _, r := range identifier {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			return "username"
		}
	}
	if digits >= 7 {
		return "phone"
	}
	return "username"
}

func filterKnown(values []string, known ...string) []string {
	allowed := make(map[string]bool, len(known))
	for _, k := range known {
		allowed[k] = true
	}
	seen := make(map[string]bool)
	var out []string
	for _, v := range values {
		v = strings.ToLower(strings.TrimSpace(v))
		if allowed[v] && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

type SelectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type OptionsSource struct {
	Type            string            `json:"type"` // "static" | "dynamic_api"
	Options         []SelectOption    `json:"options,omitempty"`
	URL             string            `json:"url,omitempty"`
	Method          string            `json:"method,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	ItemsPath       string            `json:"items_path,omitempty"`
	LabelKey        string            `json:"label_key,omitempty"`
	ValueKey        string            `json:"value_key,omitempty"`
	CacheTTLSeconds int               `json:"cache_ttl_seconds,omitempty"`
}

type RegistrationField struct {
	ID            string         `json:"id"`
	Label         string         `json:"label"`
	Type          string         `json:"type"` // "text" | "textarea" | "number" | "select" | "checkbox"
	Required      bool           `json:"required"`
	Placeholder   string         `json:"placeholder,omitempty"`
	Description   string         `json:"description,omitempty"`
	OptionsSource *OptionsSource `json:"options_source,omitempty"`
}

type Tenant struct {
	ID                     uuid.UUID
	AccountID              *uuid.UUID
	Name                   string
	OwnerUserID            uuid.UUID
	EmailVerifyRequired    bool
	EmailVerifyDelivery    string
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	SocialCallbackOrigin   string
	DefaultLocale          string
	AuthConfig             AuthConfig
	RegistrationSchema     []RegistrationField
	CreatedAt              time.Time
}

func (t Tenant) VerifyDelivery() emailverify.Delivery {
	return emailverify.ParseDelivery(t.EmailVerifyDelivery)
}

func (t Tenant) PasswordPolicy() passwordpolicy.Policy {
	return passwordpolicy.Policy{
		MinLength:        t.PasswordMinLength,
		RequireUppercase: t.PasswordRequireUpper,
		RequireNumber:    t.PasswordRequireNumber,
		RequireSpecial:   t.PasswordRequireSpecial,
	}.Normalize()
}

type Tenants struct {
	db *pgxpool.Pool
}

func NewTenants(db *pgxpool.Pool) *Tenants {
	return &Tenants{db: db}
}

const tenantSelectCols = `id, account_id, name, owner_user_id, email_verify_required, email_verify_delivery,
		       password_min_length, password_require_uppercase, password_require_number, password_require_special,
		       social_callback_origin, default_locale, created_at,
		       COALESCE(auth_config, '{"allowed_identifiers":["email"],"primary_auth_mode":"password"}'::jsonb),
		       COALESCE(registration_schema, '[]'::jsonb)`

func (t *Tenants) scanTenant(row pgx.Row) (*Tenant, error) {
	var tenant Tenant
	var authConfigJSON []byte
	var regSchemaJSON []byte
	if err := row.Scan(
		&tenant.ID,
		&tenant.AccountID,
		&tenant.Name,
		&tenant.OwnerUserID,
		&tenant.EmailVerifyRequired,
		&tenant.EmailVerifyDelivery,
		&tenant.PasswordMinLength,
		&tenant.PasswordRequireUpper,
		&tenant.PasswordRequireNumber,
		&tenant.PasswordRequireSpecial,
		&tenant.SocialCallbackOrigin,
		&tenant.DefaultLocale,
		&tenant.CreatedAt,
		&authConfigJSON,
		&regSchemaJSON,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if tenant.DefaultLocale == "" {
		tenant.DefaultLocale = "en"
	}
	tenant.AuthConfig = DefaultAuthConfig()
	if len(authConfigJSON) > 0 {
		_ = json.Unmarshal(authConfigJSON, &tenant.AuthConfig)
	}
	tenant.AuthConfig = tenant.AuthConfig.Normalize()
	if len(regSchemaJSON) > 0 {
		_ = json.Unmarshal(regSchemaJSON, &tenant.RegistrationSchema)
	}
	if tenant.RegistrationSchema == nil {
		tenant.RegistrationSchema = []RegistrationField{}
	}
	return &tenant, nil
}

func (t *Tenants) FindByOwner(ctx context.Context, ownerID uuid.UUID) (*Tenant, error) {
	return t.scanTenant(t.db.QueryRow(ctx, `
		SELECT `+tenantSelectCols+`
		FROM tenants WHERE owner_user_id = $1
		ORDER BY created_at ASC LIMIT 1
	`, ownerID))
}

func (t *Tenants) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]Tenant, error) {
	rows, err := t.db.Query(ctx, `
		SELECT `+tenantSelectCols+`
		FROM tenants WHERE account_id = $1
		ORDER BY created_at ASC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tenant
	for rows.Next() {
		var tenant Tenant
		var authConfigJSON []byte
		var regSchemaJSON []byte
		if err := rows.Scan(
			&tenant.ID,
			&tenant.AccountID,
			&tenant.Name,
			&tenant.OwnerUserID,
			&tenant.EmailVerifyRequired,
			&tenant.EmailVerifyDelivery,
			&tenant.PasswordMinLength,
			&tenant.PasswordRequireUpper,
			&tenant.PasswordRequireNumber,
			&tenant.PasswordRequireSpecial,
			&tenant.SocialCallbackOrigin,
			&tenant.DefaultLocale,
			&tenant.CreatedAt,
			&authConfigJSON,
			&regSchemaJSON,
		); err != nil {
			return nil, err
		}
		if tenant.DefaultLocale == "" {
			tenant.DefaultLocale = "en"
		}
		tenant.AuthConfig = DefaultAuthConfig()
		if len(authConfigJSON) > 0 {
			_ = json.Unmarshal(authConfigJSON, &tenant.AuthConfig)
		}
		tenant.AuthConfig = tenant.AuthConfig.Normalize()
		if len(regSchemaJSON) > 0 {
			_ = json.Unmarshal(regSchemaJSON, &tenant.RegistrationSchema)
		}
		if tenant.RegistrationSchema == nil {
			tenant.RegistrationSchema = []RegistrationField{}
		}
		out = append(out, tenant)
	}
	return out, rows.Err()
}

func (t *Tenants) CountByAccount(ctx context.Context, accountID uuid.UUID) (int, error) {
	var n int
	err := t.db.QueryRow(ctx, `SELECT COUNT(*) FROM tenants WHERE account_id = $1`, accountID).Scan(&n)
	return n, err
}

func (t *Tenants) FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return t.scanTenant(t.db.QueryRow(ctx, `
		SELECT `+tenantSelectCols+`
		FROM tenants WHERE id = $1
	`, id))
}

func (t *Tenants) FindByIDForAccount(ctx context.Context, id, accountID uuid.UUID) (*Tenant, error) {
	return t.scanTenant(t.db.QueryRow(ctx, `
		SELECT `+tenantSelectCols+`
		FROM tenants 
		WHERE id = $1 AND (account_id = $2 OR owner_user_id = (SELECT owner_user_id FROM accounts WHERE id = $2))
	`, id, accountID))
}

func (t *Tenants) Create(ctx context.Context, accountID, ownerID uuid.UUID, name string) (*Tenant, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := t.db.Exec(ctx, `
		INSERT INTO tenants (id, account_id, name, owner_user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, accountID, name, ownerID, now)
	if err != nil {
		return nil, err
	}
	return &Tenant{
		ID:                     id,
		AccountID:              &accountID,
		Name:                   name,
		OwnerUserID:            ownerID,
		EmailVerifyRequired:    true,
		EmailVerifyDelivery:    string(emailverify.DeliveryLink),
		PasswordMinLength:      8,
		PasswordRequireUpper:   false,
		PasswordRequireNumber:  false,
		PasswordRequireSpecial: false,
		DefaultLocale:          "en",
		AuthConfig:             DefaultAuthConfig(),
		RegistrationSchema:     []RegistrationField{},
		CreatedAt:              now,
	}, nil
}

func NormalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "tr" {
		return "tr"
	}
	return "en"
}

func (t *Tenants) UpdateSettings(ctx context.Context, id uuid.UUID, name string, emailVerifyRequired bool, delivery emailverify.Delivery, policy passwordpolicy.Policy, defaultLocale string) error {
	if !delivery.Valid() {
		delivery = emailverify.DeliveryLink
	}
	policy = policy.Normalize()
	defaultLocale = NormalizeLocale(defaultLocale)
	_, err := t.db.Exec(ctx, `
		UPDATE tenants SET
			name = $2,
			email_verify_required = $3,
			email_verify_delivery = $4,
			password_min_length = $5,
			password_require_uppercase = $6,
			password_require_number = $7,
			password_require_special = $8,
			default_locale = $9
		WHERE id = $1
	`, id, name, emailVerifyRequired, delivery.String(), policy.MinLength, policy.RequireUppercase, policy.RequireNumber, policy.RequireSpecial, defaultLocale)
	return err
}

func (t *Tenants) UpdateAuthConfigAndRegistrationSchema(ctx context.Context, id uuid.UUID, authConfig AuthConfig, schema []RegistrationField) error {
	authConfig = authConfig.Normalize()
	authConfigJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}
	if schema == nil {
		schema = []RegistrationField{}
	}
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	_, err = t.db.Exec(ctx, `
		UPDATE tenants SET
			auth_config = $2,
			registration_schema = $3
		WHERE id = $1
	`, id, authConfigJSON, schemaJSON)
	return err
}

func (t *Tenants) UpdateSocialCallbackOrigin(ctx context.Context, id uuid.UUID, origin string) error {
	_, err := t.db.Exec(ctx, `
		UPDATE tenants SET social_callback_origin = $2 WHERE id = $1
	`, id, strings.TrimSpace(origin))
	return err
}
