package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/sooapps/sooauth/server/internal/config"
	appjwt "github.com/sooapps/sooauth/server/internal/crypto/jwt"
	"github.com/sooapps/sooauth/server/internal/crypto/password"
	"github.com/sooapps/sooauth/server/internal/crypto/token"
	"github.com/sooapps/sooauth/server/internal/emailverify"
	"github.com/sooapps/sooauth/server/internal/mail"
	"github.com/sooapps/sooauth/server/internal/mfa"
	"github.com/sooapps/sooauth/server/internal/passwordpolicy"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/store"
	"github.com/sooapps/sooauth/server/internal/webhooks"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrRateLimited        = errors.New("rate limited")
	ErrInvalidToken       = errors.New("invalid token")
	ErrEmailDelivery      = errors.New("email delivery failed")
	ErrUserDisabled       = errors.New("user disabled")
)

type MFARequiredError struct{ Challenge string }

func (e *MFARequiredError) Error() string { return "mfa required" }

const (
	PersistentSessionTTL = 30 * 24 * time.Hour
	PersistentRefreshTTL = 30 * 24 * time.Hour
	TransientSessionTTL  = 24 * time.Hour
	TransientRefreshTTL  = 24 * time.Hour
)

type SignInOptions struct {
	RememberMe bool
}

type TokenBundle struct {
	AccessToken  string
	RefreshToken string
	SessionToken string
	CSRFToken    string
	ExpiresIn    int
	User         *store.User
	RememberMe   bool
}

type Service struct {
	cfg           config.Config
	users         *store.Users
	tokens        *store.VerificationTokens
	sess          *store.Sessions
	refresh       *store.RefreshTokens
	audit         *store.Audit
	mail          *mail.Mailer
	limit         *ratelimit.Limiter
	jwt           *appjwt.Issuer
	mfa           *mfa.Service
	hooks         *webhooks.Dispatcher
	emailSettings *store.TenantEmailSettingsStore
}

func NewService(
	cfg config.Config,
	users *store.Users,
	tokens *store.VerificationTokens,
	sess *store.Sessions,
	refresh *store.RefreshTokens,
	audit *store.Audit,
	mailer *mail.Mailer,
	limiter *ratelimit.Limiter,
	jwtIssuer *appjwt.Issuer,
	services ...any,
) *Service {
	var mfaSvc *mfa.Service
	var hooks *webhooks.Dispatcher
	var emailStore *store.TenantEmailSettingsStore
	for _, service := range services {
		switch value := service.(type) {
		case *mfa.Service:
			mfaSvc = value
		case *webhooks.Dispatcher:
			hooks = value
		case *store.TenantEmailSettingsStore:
			emailStore = value
		}
	}
	return &Service{
		cfg:           cfg,
		users:         users,
		tokens:        tokens,
		sess:          sess,
		refresh:       refresh,
		audit:         audit,
		mail:          mailer,
		limit:         limiter,
		jwt:           jwtIssuer,
		mfa:           mfaSvc,
		hooks:         hooks,
		emailSettings: emailStore,
	}
}

func (s *Service) SetEmailSettingsStore(emailStore *store.TenantEmailSettingsStore) {
	s.emailSettings = emailStore
}

func (s *Service) resolveMailer(ctx context.Context, tenantID *uuid.UUID) mail.Sender {
	if tenantID != nil && s.emailSettings != nil {
		settings, err := s.emailSettings.FindByTenant(ctx, *tenantID)
		if err == nil && settings != nil && settings.Enabled {
			cfg := s.emailSettings.ToProviderConfig(settings)
			if sender, err := mail.NewSender(cfg); err == nil && sender != nil {
				return sender
			}
		}
	}
	return s.mail
}

func (s *Service) SignUp(ctx context.Context, email, plainPassword, ip string) error {
	if ok, _ := s.limit.Allow(ctx, "signup_ip", ip, 5, time.Minute); !ok {
		return ErrRateLimited
	}

	user, err := s.users.CreatePlatformUser(ctx, email, plainPassword)
	if errors.Is(err, store.ErrEmailTaken) {
		existing, findErr := s.users.FindPlatformByEmailSimple(ctx, email)
		if findErr != nil || existing == nil {
			return err
		}
		if existing.EmailVerifiedAt != nil {
			platform, _ := s.users.IsPlatformOwner(ctx, existing.ID)
			if platform {
				return store.ErrEmailTaken
			}
			if err := s.users.PromoteToPlatformOwner(ctx, existing.ID); err != nil {
				return err
			}
			if err := s.users.UpdatePassword(ctx, existing.ID, plainPassword); err != nil {
				return err
			}
			uid := existing.ID
			s.audit.Log(ctx, &uid, "platform_upgrade", map[string]any{"email": existing.Email}, ip)
			return nil
		}
		if err := s.users.UpdatePassword(ctx, existing.ID, plainPassword); err != nil {
			return err
		}
		if err := s.sendVerificationEmail(ctx, existing); err != nil {
			return fmt.Errorf("%w: %v", ErrEmailDelivery, err)
		}
		uid := existing.ID
		s.audit.Log(ctx, &uid, "signup_resend", map[string]any{"email": existing.Email}, ip)
		return nil
	}
	if err != nil {
		return err
	}

	if err := s.sendVerificationEmail(ctx, user); err != nil {
		_ = s.users.DeleteUnverified(ctx, user.ID)
		return fmt.Errorf("%w: %v", ErrEmailDelivery, err)
	}

	uid := user.ID
	s.audit.Log(ctx, &uid, "signup", map[string]any{"email": user.Email}, ip)
	return nil
}

func (s *Service) ResendVerification(ctx context.Context, email, ip string) error {
	if ok, _ := s.limit.Allow(ctx, "verify_resend_ip", ip, 5, time.Minute); !ok {
		return ErrRateLimited
	}
	if ok, _ := s.limit.Allow(ctx, "verify_resend_email", email, 3, time.Minute); !ok {
		return ErrRateLimited
	}

	user, err := s.users.FindPlatformByEmailSimple(ctx, email)
	if err != nil {
		return err
	}
	if user == nil || user.EmailVerifiedAt != nil {
		return nil
	}

	if err := s.sendVerificationEmail(ctx, user); err != nil {
		return fmt.Errorf("%w: %v", ErrEmailDelivery, err)
	}
	uid := user.ID
	s.audit.Log(ctx, &uid, "verification_resent", map[string]any{"email": user.Email}, ip)
	return nil
}

func (s *Service) SignInApp(ctx context.Context, tenantID uuid.UUID, email, plainPassword, ip, userAgent string, requireVerified bool) (*TokenBundle, error) {
	return s.SignInAppWithOptions(ctx, tenantID, email, plainPassword, ip, userAgent, requireVerified, SignInOptions{RememberMe: true})
}

func (s *Service) SignInAppWithOptions(ctx context.Context, tenantID uuid.UUID, email, plainPassword, ip, userAgent string, requireVerified bool, opts SignInOptions) (*TokenBundle, error) {
	if ok, _ := s.limit.Allow(ctx, "signin_ip", ip, 20, time.Minute); !ok {
		return nil, ErrRateLimited
	}
	if ok, _ := s.limit.Allow(ctx, "signin_email", email, 8, time.Minute); !ok {
		return nil, ErrRateLimited
	}

	found, hash, err := s.users.FindByEmailInTenant(ctx, tenantID, email)
	if err != nil || found == nil {
		s.audit.Log(ctx, nil, "signin_failed", map[string]any{"email": email, "tenant_id": tenantID.String()}, ip)
		s.emit(ctx, &tenantID, "user.login_failed", nil, map[string]any{"email": email, "ip": ip})
		return nil, ErrInvalidCredentials
	}

	match, err := password.Verify(plainPassword, hash)
	if err != nil || !match {
		uid := found.ID
		s.audit.Log(ctx, &uid, "signin_failed", map[string]any{"email": found.Email, "tenant_id": tenantID.String()}, ip)
		s.emit(ctx, &tenantID, "user.login_failed", found, map[string]any{"ip": ip})
		return nil, ErrInvalidCredentials
	}

	if requireVerified && found.EmailVerifiedAt == nil {
		return nil, ErrEmailNotVerified
	}
	if found.DisabledAt != nil {
		return nil, ErrUserDisabled
	}

	return s.completeSignIn(ctx, found, ip, userAgent, "signin", opts.RememberMe)
}

func (s *Service) VerifyMFA(ctx context.Context, challenge, code, ip, userAgent string) (*TokenBundle, error) {
	if s.mfa == nil {
		return nil, ErrInvalidToken
	}
	userID, err := s.mfa.VerifyChallenge(ctx, challenge, code)
	if err != nil {
		return nil, err
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}
	return s.completeSignIn(ctx, user, ip, userAgent, "mfa_signin", true)
}

func (s *Service) SignInForTenant(ctx context.Context, email, plainPassword, ip, userAgent string, requireVerified bool) (*TokenBundle, error) {
	return s.SignIn(ctx, email, plainPassword, ip, userAgent)
}

func (s *Service) SignIn(ctx context.Context, email, plainPassword, ip, userAgent string) (*TokenBundle, error) {
	return s.SignInWithOptions(ctx, email, plainPassword, ip, userAgent, SignInOptions{RememberMe: true})
}

func (s *Service) SignInWithOptions(ctx context.Context, email, plainPassword, ip, userAgent string, opts SignInOptions) (*TokenBundle, error) {
	if ok, _ := s.limit.Allow(ctx, "signin_ip", ip, 20, time.Minute); !ok {
		return nil, ErrRateLimited
	}
	if ok, _ := s.limit.Allow(ctx, "signin_email", email, 8, time.Minute); !ok {
		return nil, ErrRateLimited
	}

	found, hash, err := s.users.FindPlatformByEmail(ctx, email)
	if err != nil || found == nil {
		s.audit.Log(ctx, nil, "signin_failed", map[string]any{"email": email}, ip)
		return nil, ErrInvalidCredentials
	}

	match, err := password.Verify(plainPassword, hash)
	if err != nil || !match {
		uid := found.ID
		s.audit.Log(ctx, &uid, "signin_failed", map[string]any{"email": found.Email}, ip)
		return nil, ErrInvalidCredentials
	}

	if found.EmailVerifiedAt == nil {
		return nil, ErrEmailNotVerified
	}

	if s.mfa != nil {
		enabled, err := s.mfa.Enabled(ctx, found.ID)
		if err != nil {
			return nil, err
		}
		if enabled {
			challenge, err := s.mfa.Challenge(ctx, found.ID)
			if err != nil {
				return nil, err
			}
			return nil, &MFARequiredError{Challenge: challenge}
		}
	}

	return s.completeSignIn(ctx, found, ip, userAgent, "signin", opts.RememberMe)
}

func (s *Service) SignInSocial(ctx context.Context, user *store.User, ip, userAgent string) (*TokenBundle, error) {
	return s.completeSignIn(ctx, user, ip, userAgent, "social_signin", true)
}

func (s *Service) SignInVerifiedUser(ctx context.Context, user *store.User, ip, userAgent string) (*TokenBundle, error) {
	return s.completeSignIn(ctx, user, ip, userAgent, "email_verified_signin", true)
}

func (s *Service) completeSignIn(ctx context.Context, user *store.User, ip, userAgent, auditAction string, rememberMe bool) (*TokenBundle, error) {
	if user.DisabledAt != nil {
		return nil, ErrUserDisabled
	}
	bundle, err := s.issueTokens(ctx, user, ip, userAgent, rememberMe)
	if err != nil {
		return nil, err
	}
	uid := user.ID
	s.audit.Log(ctx, &uid, auditAction, map[string]any{"email": user.Email, "remember_me": rememberMe}, ip)
	s.emit(ctx, user.TenantID, "user.login", user, map[string]any{"ip": ip, "method": auditAction, "remember_me": rememberMe})
	return bundle, nil
}

func (s *Service) issueTokens(ctx context.Context, user *store.User, ip, userAgent string, rememberMe bool) (*TokenBundle, error) {
	refreshTTL := PersistentRefreshTTL
	sessionTTL := PersistentSessionTTL
	if !rememberMe {
		refreshTTL = TransientRefreshTTL
		sessionTTL = TransientSessionTTL
	}

	bundle, err := s.issueAccessRefreshWithTTL(ctx, user, ip, refreshTTL)
	if err != nil {
		return nil, err
	}

	sessionPlain, sessionHash, err := token.Generate()
	if err != nil {
		return nil, err
	}
	if _, err := s.sess.CreateWithTTL(ctx, user.ID, sessionHash, ip, userAgent, sessionTTL); err != nil {
		return nil, err
	}

	csrf, err := randomToken(24)
	if err != nil {
		return nil, err
	}

	bundle.SessionToken = sessionPlain
	bundle.CSRFToken = csrf
	bundle.RememberMe = rememberMe
	return bundle, nil
}

func (s *Service) IssueOIDCTokens(ctx context.Context, user *store.User, ip string) (*TokenBundle, error) {
	return s.issueAccessRefresh(ctx, user, ip)
}

func (s *Service) issueAccessRefresh(ctx context.Context, user *store.User, ip string) (*TokenBundle, error) {
	return s.issueAccessRefreshWithTTL(ctx, user, ip, PersistentRefreshTTL)
}

func (s *Service) issueAccessRefreshWithTTL(ctx context.Context, user *store.User, ip string, refreshTTL time.Duration) (*TokenBundle, error) {
	access, expires, err := s.jwt.AccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshPlain, refreshHash, err := token.Generate()
	if err != nil {
		return nil, err
	}
	familyID := uuid.New()
	if _, err := s.refresh.IssueWithTTL(ctx, user.ID, familyID, refreshHash, refreshTTL); err != nil {
		return nil, err
	}

	return &TokenBundle{
		AccessToken:  access,
		RefreshToken: refreshPlain,
		ExpiresIn:    int(time.Until(expires).Seconds()),
		User:         user,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, plainRefresh, ip string) (*TokenBundle, error) {
	if ok, _ := s.limit.Allow(ctx, "refresh_ip", ip, 30, time.Minute); !ok {
		return nil, ErrRateLimited
	}
	if plainRefresh == "" {
		return nil, ErrInvalidToken
	}

	rt, err := s.refresh.FindByHash(ctx, token.Hash(plainRefresh))
	if err != nil || rt == nil {
		return nil, ErrInvalidToken
	}

	now := time.Now().UTC()
	if rt.ReplacedBy != nil {
		_ = s.refresh.RevokeFamily(ctx, rt.FamilyID)
		uid := rt.UserID
		s.audit.Log(ctx, &uid, "refresh_reuse_detected", map[string]any{"family_id": rt.FamilyID.String()}, ip)
		return nil, ErrInvalidToken
	}

	if rt.RevokedAt != nil || rt.ExpiresAt.Before(now) {
		return nil, ErrInvalidToken
	}

	user, err := s.users.FindByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	newPlain, newHash, err := token.Generate()
	if err != nil {
		return nil, err
	}
	newID, err := s.refresh.Issue(ctx, rt.UserID, rt.FamilyID, newHash)
	if err != nil {
		return nil, err
	}
	if err := s.refresh.MarkReplaced(ctx, rt.ID, newID); err != nil {
		return nil, err
	}

	access, expires, err := s.jwt.AccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	csrf, err := randomToken(24)
	if err != nil {
		return nil, err
	}

	uid := user.ID
	s.audit.Log(ctx, &uid, "token_refreshed", nil, ip)

	return &TokenBundle{
		AccessToken:  access,
		RefreshToken: newPlain,
		ExpiresIn:    int(time.Until(expires).Seconds()),
		User:         user,
		CSRFToken:    csrf,
	}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, plainToken, ip string) (alreadyVerified bool, err error) {
	if ok, _ := s.limit.Allow(ctx, "verify_ip", ip, 30, time.Minute); !ok {
		return false, ErrRateLimited
	}
	userID, err := s.tokens.Consume(ctx, "email_verify", token.Hash(plainToken))
	if err != nil {
		if errors.Is(err, store.ErrTokenInvalid) {
			if usedID, ok, lookupErr := s.tokens.UserFromUsed(ctx, "email_verify", token.Hash(plainToken)); lookupErr == nil && ok {
				user, findErr := s.users.FindByID(ctx, usedID)
				if findErr == nil && user != nil && user.EmailVerifiedAt != nil {
					return true, nil
				}
			}
		}
		return false, err
	}
	if err := s.users.MarkEmailVerified(ctx, userID); err != nil {
		return false, err
	}
	s.audit.Log(ctx, &userID, "email_verified", nil, ip)
	if user, findErr := s.users.FindByID(ctx, userID); findErr == nil {
		s.emit(ctx, user.TenantID, "user.email_verified", user, map[string]any{"ip": ip})
	}
	return false, nil
}

func (s *Service) VerifyEmailByCode(ctx context.Context, tenantID *uuid.UUID, email, rawCode, ip string) (user *store.User, alreadyVerified bool, err error) {
	if ok, _ := s.limit.Allow(ctx, "verify_ip", ip, 30, time.Minute); !ok {
		return nil, false, ErrRateLimited
	}
	if ok, _ := s.limit.Allow(ctx, "verify_code_email", strings.ToLower(strings.TrimSpace(email)), 10, time.Minute); !ok {
		return nil, false, ErrRateLimited
	}

	code := emailverify.NormalizeCode(rawCode)
	if !emailverify.ValidCode(code) {
		return nil, false, ErrInvalidToken
	}

	var found *store.User
	if tenantID != nil {
		found, err = s.users.FindByEmailSimpleInTenant(ctx, *tenantID, email)
	} else {
		found, err = s.users.FindPlatformByEmailSimple(ctx, email)
	}
	if err != nil {
		return nil, false, err
	}
	if found == nil {
		return nil, false, ErrInvalidToken
	}
	if found.EmailVerifiedAt != nil {
		return found, true, nil
	}

	userID, err := s.tokens.Consume(ctx, "email_verify_code", token.Hash(code))
	if err != nil {
		if errors.Is(err, store.ErrTokenInvalid) {
			if usedID, ok, lookupErr := s.tokens.UserFromUsed(ctx, "email_verify_code", token.Hash(code)); lookupErr == nil && ok && usedID == found.ID {
				user, findErr := s.users.FindByID(ctx, usedID)
				if findErr == nil && user != nil && user.EmailVerifiedAt != nil {
					return user, true, nil
				}
			}
		}
		return nil, false, err
	}
	if userID != found.ID {
		return nil, false, ErrInvalidToken
	}
	if err := s.users.MarkEmailVerified(ctx, userID); err != nil {
		return nil, false, err
	}
	s.audit.Log(ctx, &userID, "email_verified_code", map[string]any{"email": found.Email}, ip)
	s.emit(ctx, found.TenantID, "user.email_verified", found, map[string]any{"ip": ip})
	verified, err := s.users.FindByID(ctx, userID)
	return verified, false, err
}

type ForgotPasswordOpts struct {
	Email     string
	IP        string
	TenantID  *uuid.UUID
	ClientID  string
	ReturnTo  string
	Delivery  string
	BrandName string
	AppScoped bool
}

func (s *Service) ForgotPassword(ctx context.Context, email, ip string) error {
	return s.ForgotPasswordWithOpts(ctx, ForgotPasswordOpts{
		Email: email,
		IP:    ip,
	})
}

func (s *Service) ForgotPasswordWithOpts(ctx context.Context, opts ForgotPasswordOpts) error {
	if ok, _ := s.limit.Allow(ctx, "forgot_ip", opts.IP, 5, time.Minute); !ok {
		return ErrRateLimited
	}

	email := strings.ToLower(strings.TrimSpace(opts.Email))
	if email == "" {
		return nil
	}

	var user *store.User
	var err error
	if opts.TenantID != nil {
		user, err = s.users.FindByEmailSimpleInTenant(ctx, *opts.TenantID, email)
	} else {
		user, err = s.users.FindPlatformByEmailSimple(ctx, email)
	}
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	_ = s.tokens.DeletePending(ctx, user.ID, "password_reset", "password_reset_code")

	brand := strings.TrimSpace(opts.BrandName)
	if brand == "" {
		brand = s.cfg.BrandName
	}

	delivery := strings.ToLower(strings.TrimSpace(opts.Delivery))
	useCode := delivery == "code"

	d := mail.PasswordResetDelivery{}
	if useCode {
		code, hash, err := emailverify.GenerateCode()
		if err != nil {
			return err
		}
		expires := time.Now().UTC().Add(15 * time.Minute)
		if err := s.tokens.Create(ctx, user.ID, user.Email, "password_reset_code", hash, expires); err != nil {
			return err
		}
		d.Code = code
	} else {
		plain, hash, err := token.Generate()
		if err != nil {
			return err
		}
		expires := time.Now().UTC().Add(time.Hour)
		if err := s.tokens.Create(ctx, user.ID, user.Email, "password_reset", hash, expires); err != nil {
			return err
		}
		q := url.Values{"token": {plain}}
		if cid := strings.TrimSpace(opts.ClientID); cid != "" {
			q.Set("client_id", cid)
		}
		if rt := strings.TrimSpace(opts.ReturnTo); rt != "" {
			q.Set("return_to", rt)
		}
		d.LinkURL = fmt.Sprintf("%s/auth/reset-password?%s", trimSlash(s.cfg.AppURL), q.Encode())
	}

	tx := mail.Transactional{BrandName: brand, AppURL: s.cfg.AppURL}
	subject, plainBody, htmlBody := tx.PasswordResetDeliveryEmail(d, opts.AppScoped)
	mailer := s.resolveMailer(ctx, opts.TenantID)
	_ = mailer.SendOutbound(mail.Outbound{
		To:      user.Email,
		Subject: subject,
		Plain:   plainBody,
		HTML:    htmlBody,
	})

	uid := user.ID
	s.audit.Log(ctx, &uid, "password_reset_requested", nil, opts.IP)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, plainToken, newPassword, ip string) error {
	if ok, _ := s.limit.Allow(ctx, "reset_ip", ip, 10, time.Minute); !ok {
		return ErrRateLimited
	}
	if err := passwordpolicy.Default().Validate(newPassword); err != nil {
		return err
	}

	userID, err := s.tokens.Consume(ctx, "password_reset", token.Hash(plainToken))
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, newPassword); err != nil {
		return err
	}
	_ = s.sess.RevokeAllForUser(ctx, userID)
	_ = s.refresh.RevokeAllForUser(ctx, userID)
	s.audit.Log(ctx, &userID, "password_reset", nil, ip)
	if user, findErr := s.users.FindByID(ctx, userID); findErr == nil {
		s.emit(ctx, user.TenantID, "password.reset", user, map[string]any{"ip": ip})
	}
	return nil
}

func (s *Service) ResetPasswordWithCode(ctx context.Context, email, code, newPassword, ip string) error {
	if ok, _ := s.limit.Allow(ctx, "reset_ip", ip, 10, time.Minute); !ok {
		return ErrRateLimited
	}
	if err := passwordpolicy.Default().Validate(newPassword); err != nil {
		return err
	}

	code = emailverify.NormalizeCode(code)
	if !emailverify.ValidCode(code) {
		return ErrInvalidToken
	}

	userID, err := s.tokens.ConsumeCode(ctx, email, "password_reset_code", token.Hash(code))
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, newPassword); err != nil {
		return err
	}
	_ = s.sess.RevokeAllForUser(ctx, userID)
	_ = s.refresh.RevokeAllForUser(ctx, userID)
	s.audit.Log(ctx, &userID, "password_reset", nil, ip)
	if user, findErr := s.users.FindByID(ctx, userID); findErr == nil {
		s.emit(ctx, user.TenantID, "password.reset", user, map[string]any{"ip": ip})
	}
	return nil
}

func (s *Service) SignOut(ctx context.Context, sessionToken, refreshToken, ip string) error {
	var revokedUser *store.User
	if sessionToken != "" {
		if sess, _ := s.sess.FindByTokenHash(ctx, token.Hash(sessionToken)); sess != nil {
			revokedUser, _ = s.users.FindByID(ctx, sess.UserID)
		}
		_ = s.sess.RevokeByTokenHash(ctx, token.Hash(sessionToken))
	}
	if refreshToken != "" {
		rt, _ := s.refresh.FindByHash(ctx, token.Hash(refreshToken))
		if rt != nil {
			_ = s.refresh.RevokeFamily(ctx, rt.FamilyID)
		}
	}
	if revokedUser != nil {
		s.emit(ctx, revokedUser.TenantID, "session.revoked", revokedUser, map[string]any{"ip": ip})
	}
	return nil
}

func (s *Service) emit(ctx context.Context, tenantID *uuid.UUID, eventType string, user *store.User, data map[string]any) {
	if s.hooks != nil {
		s.hooks.Emit(ctx, tenantID, eventType, user, data)
	}
}

func (s *Service) SessionUser(ctx context.Context, sessionToken string) (*store.User, error) {
	if sessionToken == "" {
		return nil, nil
	}
	sess, err := s.sess.FindByTokenHash(ctx, token.Hash(sessionToken))
	if err != nil || sess == nil {
		return nil, err
	}
	return s.users.FindByID(ctx, sess.UserID)
}

func (s *Service) UserFromAccess(ctx context.Context, accessToken string) (*store.User, error) {
	if accessToken == "" {
		return nil, nil
	}
	claims, err := s.jwt.ParseAccess(accessToken)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.users.FindByID(ctx, userID)
}

func randomToken(n int) (string, error) {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func trimSlash(u string) string {
	for len(u) > 1 && u[len(u)-1] == '/' {
		u = u[:len(u)-1]
	}
	return u
}
