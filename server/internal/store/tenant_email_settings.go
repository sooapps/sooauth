package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sooapps/sooauth/server/internal/crypto/encryption"
	"github.com/sooapps/sooauth/server/internal/mail"
)

type TenantEmailSettings struct {
	ID                    uuid.UUID
	TenantID              uuid.UUID
	Enabled               bool
	Provider              string
	FromName              string
	FromEmail             string
	ReplyTo               string
	SMTPHost              string
	SMTPPort              int
	SMTPUser              string
	SMTPPasswordEncrypted string
	SMTPTLSMode           string
	APIKeyEncrypted       string
	AWSAccessKeyID        string
	AWSSecretKeyEncrypted string
	AWSRegion             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TenantEmailSettingsInput struct {
	TenantID          uuid.UUID
	Enabled           bool
	Provider          string
	FromName          string
	FromEmail         string
	ReplyTo           string
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPasswordPlain *string // if non-nil, will encrypt; if nil, preserve existing
	SMTPTLSMode       string
	APIKeyPlain       *string // if non-nil, will encrypt; if nil, preserve existing
	AWSAccessKeyID    string
	AWSSecretKeyPlain *string // if non-nil, will encrypt; if nil, preserve existing
	AWSRegion         string
}

type TenantEmailSettingsStore struct {
	db     *pgxpool.Pool
	cipher *encryption.Cipher
}

func NewTenantEmailSettings(db *pgxpool.Pool, encryptionKey []byte) *TenantEmailSettingsStore {
	var cipher *encryption.Cipher
	if len(encryptionKey) == 32 {
		cipher, _ = encryption.New(encryptionKey)
	}
	return &TenantEmailSettingsStore{db: db, cipher: cipher}
}

func (s *TenantEmailSettingsStore) encrypt(plain string) string {
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return ""
	}
	if s.cipher != nil {
		if enc, err := s.cipher.Encrypt(plain); err == nil {
			return enc
		}
	}
	return plain
}

func (s *TenantEmailSettingsStore) decrypt(ciphertext string) string {
	ciphertext = strings.TrimSpace(ciphertext)
	if ciphertext == "" {
		return ""
	}
	if encryption.IsEncrypted(ciphertext) && s.cipher != nil {
		if dec, err := s.cipher.Decrypt(ciphertext); err == nil {
			return dec
		}
	}
	return ciphertext
}

func (s *TenantEmailSettingsStore) DecryptSMTPPassword(item *TenantEmailSettings) string {
	if item == nil {
		return ""
	}
	return s.decrypt(item.SMTPPasswordEncrypted)
}

func (s *TenantEmailSettingsStore) DecryptAPIKey(item *TenantEmailSettings) string {
	if item == nil {
		return ""
	}
	return s.decrypt(item.APIKeyEncrypted)
}

func (s *TenantEmailSettingsStore) DecryptAWSSecretKey(item *TenantEmailSettings) string {
	if item == nil {
		return ""
	}
	return s.decrypt(item.AWSSecretKeyEncrypted)
}

func (s *TenantEmailSettingsStore) ToProviderConfig(item *TenantEmailSettings) mail.ProviderConfig {
	if item == nil {
		return mail.ProviderConfig{}
	}
	port := item.SMTPPort
	if port <= 0 {
		port = 587
	}
	tlsMode := item.SMTPTLSMode
	if tlsMode == "" {
		tlsMode = "starttls"
	}
	region := item.AWSRegion
	if region == "" {
		region = "us-east-1"
	}

	return mail.ProviderConfig{
		Provider:     item.Provider,
		FromName:     item.FromName,
		FromEmail:    item.FromEmail,
		ReplyTo:      item.ReplyTo,
		BrandName:    item.FromName,
		SMTPHost:     item.SMTPHost,
		SMTPPort:     port,
		SMTPUser:     item.SMTPUser,
		SMTPPassword: s.DecryptSMTPPassword(item),
		SMTPTLSMode:  tlsMode,
		APIKey:       s.DecryptAPIKey(item),
		AWSAccessKey: item.AWSAccessKeyID,
		AWSSecretKey: s.DecryptAWSSecretKey(item),
		AWSRegion:    region,
	}
}

func (s *TenantEmailSettingsStore) scan(row pgx.Row) (*TenantEmailSettings, error) {
	var item TenantEmailSettings
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Enabled,
		&item.Provider,
		&item.FromName,
		&item.FromEmail,
		&item.ReplyTo,
		&item.SMTPHost,
		&item.SMTPPort,
		&item.SMTPUser,
		&item.SMTPPasswordEncrypted,
		&item.SMTPTLSMode,
		&item.APIKeyEncrypted,
		&item.AWSAccessKeyID,
		&item.AWSSecretKeyEncrypted,
		&item.AWSRegion,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *TenantEmailSettingsStore) FindByTenant(ctx context.Context, tenantID uuid.UUID) (*TenantEmailSettings, error) {
	return s.scan(s.db.QueryRow(ctx, `
		SELECT id, tenant_id, enabled, provider, from_name, from_email, reply_to,
		       smtp_host, smtp_port, smtp_user, smtp_password_encrypted, smtp_tls_mode,
		       api_key_encrypted, aws_access_key_id, aws_secret_key_encrypted, aws_region,
		       created_at, updated_at
		FROM tenant_email_settings
		WHERE tenant_id = $1
	`, tenantID))
}

func (s *TenantEmailSettingsStore) Upsert(ctx context.Context, in TenantEmailSettingsInput) error {
	existing, err := s.FindByTenant(ctx, in.TenantID)
	if err != nil {
		return err
	}

	smtpPasswordEnc := ""
	if existing != nil {
		smtpPasswordEnc = existing.SMTPPasswordEncrypted
	}
	if in.SMTPPasswordPlain != nil {
		smtpPasswordEnc = s.encrypt(*in.SMTPPasswordPlain)
	}

	apiKeyEnc := ""
	if existing != nil {
		apiKeyEnc = existing.APIKeyEncrypted
	}
	if in.APIKeyPlain != nil {
		apiKeyEnc = s.encrypt(*in.APIKeyPlain)
	}

	awsSecretKeyEnc := ""
	if existing != nil {
		awsSecretKeyEnc = existing.AWSSecretKeyEncrypted
	}
	if in.AWSSecretKeyPlain != nil {
		awsSecretKeyEnc = s.encrypt(*in.AWSSecretKeyPlain)
	}

	port := in.SMTPPort
	if port <= 0 {
		port = 587
	}
	tlsMode := in.SMTPTLSMode
	if tlsMode == "" {
		tlsMode = "starttls"
	}
	region := in.AWSRegion
	if region == "" {
		region = "us-east-1"
	}
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	if provider == "" {
		provider = "smtp"
	}

	now := time.Now().UTC()
	_, err = s.db.Exec(ctx, `
		INSERT INTO tenant_email_settings (
			id, tenant_id, enabled, provider, from_name, from_email, reply_to,
			smtp_host, smtp_port, smtp_user, smtp_password_encrypted, smtp_tls_mode,
			api_key_encrypted, aws_access_key_id, aws_secret_key_encrypted, aws_region,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $16
		)
		ON CONFLICT (tenant_id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			provider = EXCLUDED.provider,
			from_name = EXCLUDED.from_name,
			from_email = EXCLUDED.from_email,
			reply_to = EXCLUDED.reply_to,
			smtp_host = EXCLUDED.smtp_host,
			smtp_port = EXCLUDED.smtp_port,
			smtp_user = EXCLUDED.smtp_user,
			smtp_password_encrypted = EXCLUDED.smtp_password_encrypted,
			smtp_tls_mode = EXCLUDED.smtp_tls_mode,
			api_key_encrypted = EXCLUDED.api_key_encrypted,
			aws_access_key_id = EXCLUDED.aws_access_key_id,
			aws_secret_key_encrypted = EXCLUDED.aws_secret_key_encrypted,
			aws_region = EXCLUDED.aws_region,
			updated_at = EXCLUDED.updated_at
	`, in.TenantID, in.Enabled, provider, strings.TrimSpace(in.FromName), strings.TrimSpace(in.FromEmail), strings.TrimSpace(in.ReplyTo),
		strings.TrimSpace(in.SMTPHost), port, strings.TrimSpace(in.SMTPUser), smtpPasswordEnc, tlsMode,
		apiKeyEnc, strings.TrimSpace(in.AWSAccessKeyID), awsSecretKeyEnc, region, now)

	return err
}
