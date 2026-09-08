package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/emailverify"
	"github.com/sooapps/sooauth/server/internal/mail"
	"github.com/sooapps/sooauth/server/internal/store"
)

type AppVerificationEmailOpts struct {
	BrandName string
	User      *store.User
	Delivery  emailverify.Delivery
	ClientID  string
	ReturnTo  string
}

type verificationMailOpts struct {
	BrandName string
	User        *store.User
	Delivery    emailverify.Delivery
	ClientID    string
	ReturnTo    string
	AppScoped   bool
}

func (s *Service) sendVerificationMail(ctx context.Context, opts verificationMailOpts) error {
	if s.cfg.Env == "production" && !s.cfg.SMTPConfigured() {
		return fmt.Errorf("SMTP not configured")
	}
	if opts.User == nil {
		return fmt.Errorf("user required")
	}

	_ = s.tokens.DeletePending(ctx, opts.User.ID, "email_verify", "email_verify_code")

	expires := time.Now().UTC().Add(24 * time.Hour)
	content := mail.VerificationDelivery{}

	if opts.Delivery.IncludesLink() {
		plain, hash, err := token.Generate()
		if err != nil {
			return err
		}
		if err := s.tokens.Create(ctx, opts.User.ID, opts.User.Email, "email_verify", hash, expires); err != nil {
			return err
		}
		q := url.Values{"token": {plain}}
		if cid := strings.TrimSpace(opts.ClientID); cid != "" {
			q.Set("client_id", cid)
		}
		if rt := strings.TrimSpace(opts.ReturnTo); rt != "" {
			q.Set("return_to", rt)
		}
		content.LinkURL = fmt.Sprintf("%s/auth/verify?%s", trimSlash(s.cfg.AppURL), q.Encode())
	}

	if opts.Delivery.IncludesCode() {
		code, hash, err := emailverify.GenerateCode()
		if err != nil {
			return err
		}
		if err := s.tokens.Create(ctx, opts.User.ID, opts.User.Email, "email_verify_code", hash, expires); err != nil {
			return err
		}
		content.Code = code
	}

	tx := mail.Transactional{BrandName: opts.BrandName, AppURL: s.cfg.AppURL}
	subject, plainBody, htmlBody := tx.VerificationDeliveryEmail(content, opts.AppScoped)
	return s.mail.SendOutbound(mail.Outbound{
		To:      opts.User.Email,
		Subject: subject,
		Plain:   plainBody,
		HTML:    htmlBody,
	})
}

func (s *Service) sendVerificationEmail(ctx context.Context, user *store.User) error {
	return s.sendVerificationMail(ctx, verificationMailOpts{
		BrandName: s.cfg.BrandName,
		User:      user,
		Delivery:  emailverify.DeliveryLink,
		AppScoped: false,
	})
}

func (s *Service) ResendAppVerification(ctx context.Context, opts AppVerificationEmailOpts, tenantID uuid.UUID, email, ip string) error {
	if ok, _ := s.limit.Allow(ctx, "verify_resend_ip", ip, 5, time.Minute); !ok {
		return ErrRateLimited
	}
	if ok, _ := s.limit.Allow(ctx, "verify_resend_email", email, 3, time.Minute); !ok {
		return ErrRateLimited
	}

	user, err := s.users.FindByEmailSimpleInTenant(ctx, tenantID, email)
	if err != nil {
		return err
	}
	if user == nil || user.EmailVerifiedAt != nil {
		return nil
	}
	opts.User = user
	if err := s.sendVerificationMail(ctx, verificationMailOpts{
		BrandName: opts.BrandName,
		User:      opts.User,
		Delivery:  opts.Delivery,
		ClientID:  opts.ClientID,
		ReturnTo:  opts.ReturnTo,
		AppScoped: true,
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrEmailDelivery, err)
	}
	uid := user.ID
	s.audit.Log(ctx, &uid, "verification_resent", map[string]any{"email": user.Email}, ip)
	return nil
}

func (s *Service) SendAppVerificationEmail(ctx context.Context, opts AppVerificationEmailOpts) error {
	return s.sendVerificationMail(ctx, verificationMailOpts{
		BrandName: opts.BrandName,
		User:      opts.User,
		Delivery:  opts.Delivery,
		ClientID:    opts.ClientID,
		ReturnTo:    opts.ReturnTo,
		AppScoped:   true,
	})
}
