package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sooapps/sooauth/server/internal/adminui"
	"github.com/sooapps/sooauth/server/internal/auth"
	"github.com/sooapps/sooauth/server/internal/billing"
	billingproviders "github.com/sooapps/sooauth/server/internal/billing/providers"
	"github.com/sooapps/sooauth/server/internal/config"
	appjwt "github.com/sooapps/sooauth/server/internal/crypto/jwt"
	"github.com/sooapps/sooauth/server/internal/crypto/signing"
	"github.com/sooapps/sooauth/server/internal/ephemeral"
	"github.com/sooapps/sooauth/server/internal/health"
	"github.com/sooapps/sooauth/server/internal/mail"
	"github.com/sooapps/sooauth/server/internal/mfa"
	"github.com/sooapps/sooauth/server/internal/oidc"
	"github.com/sooapps/sooauth/server/internal/pages"
	"github.com/sooapps/sooauth/server/internal/passkeys"
	"github.com/sooapps/sooauth/server/internal/ratelimit"
	"github.com/sooapps/sooauth/server/internal/social"
	"github.com/sooapps/sooauth/server/internal/store"
	"github.com/sooapps/sooauth/server/internal/webhooks"
)

type Server struct {
	cfg            config.Config
	db             *pgxpool.Pool
	rdb            *redis.Client
	auth           *auth.Service
	mfa            *mfa.Service
	oidc           *oidc.Service
	social         *social.Service
	passkeys       *passkeys.Service
	credentials    *store.WebAuthnCredentials
	jwt            *appjwt.Issuer
	pages          *pages.Renderer
	theme          *store.ThemeStore
	users          *store.Users
	sessions       *store.Sessions
	refresh        *store.RefreshTokens
	audit          *store.Audit
	oauthProviders *store.OAuthProviders
	tenants        *store.Tenants
	oauthClients   *store.OAuthClients
	tenantUsers    *store.TenantUsers
	accounts       *store.Accounts
	ephemeral      *ephemeral.Store
	billing        *billing.Service
	webhooks       *webhooks.Dispatcher
	webhookStore   *store.Webhooks
	identities     *store.SocialIdentities
}

func New(cfg config.Config, db *pgxpool.Pool, signingKey *signing.Key) (*Server, error) {
	var rdb *redis.Client
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			slog.Warn("redis parse skipped", "err", err)
		} else {
			rdb = redis.NewClient(opt)
		}
	}

	jwtIssuer, err := appjwt.NewIssuer(signingKey, cfg.AppURL)
	if err != nil {
		return nil, fmt.Errorf("jwt issuer: %w", err)
	}

	users := store.NewUsers(db)
	webhookStore := store.NewWebhooks(db)
	webhookDispatcher := webhooks.NewDispatcher(webhookStore)
	mailer := mail.NewWithSettings(mail.Settings{
		URL:       cfg.SMTPURL,
		Host:      cfg.SMTPHost,
		Port:      cfg.SMTPPort,
		User:      cfg.SMTPUser,
		Password:  cfg.SMTPPassword,
		From:      cfg.SMTPFrom,
		BrandName: cfg.BrandName,
	})
	limiter := ratelimit.New(rdb)
	ephem := ephemeral.New(rdb)
	mfaSvc := mfa.NewService(store.NewMFA(db, cfg.MFAEncryptionKey))
	authSvc := auth.NewService(
		cfg,
		users,
		store.NewVerificationTokens(db),
		store.NewSessions(db),
		store.NewRefreshTokens(db),
		store.NewAudit(db),
		mailer,
		limiter,
		jwtIssuer,
		mfaSvc,
		webhookDispatcher,
	)

	pageRenderer, err := pages.NewRenderer()
	if err != nil {
		return nil, fmt.Errorf("pages: %w", err)
	}

	credentials := store.NewWebAuthnCredentials(db)
	passkeySvc, err := passkeys.NewService(
		cfg,
		credentials,
		users,
		authSvc,
		store.NewAudit(db),
		ephem,
	)
	if err != nil {
		return nil, fmt.Errorf("passkeys: %w", err)
	}

	s := &Server{
		cfg:  cfg,
		db:   db,
		rdb:  rdb,
		jwt:  jwtIssuer,
		auth: authSvc,
		mfa:  mfaSvc,
		oidc: oidc.NewService(
			cfg.AppURL,
			store.NewOAuthClients(db),
			store.NewAuthorizationCodes(db),
			users,
			authSvc,
			jwtIssuer,
			limiter,
		),
		social: social.NewService(
			cfg,
			users,
			store.NewAudit(db),
			authSvc,
			store.NewOAuthProviders(db),
			store.NewTenants(db),
			store.NewAccounts(db),
			ephem,
			store.NewSocialIdentities(db),
		),
		passkeys:       passkeySvc,
		credentials:    credentials,
		pages:          pageRenderer,
		theme:          store.NewThemeStore(db),
		users:          users,
		sessions:       store.NewSessions(db),
		refresh:        store.NewRefreshTokens(db),
		audit:          store.NewAudit(db),
		oauthProviders: store.NewOAuthProviders(db),
		tenants:        store.NewTenants(db),
		oauthClients:   store.NewOAuthClients(db),
		tenantUsers:    store.NewTenantUsers(db),
		accounts:       store.NewAccounts(db),
		ephemeral:      ephem,
		billing: billing.NewService(
			store.NewSubscriptions(db),
			store.NewBillingCheckouts(db),
			store.NewBillingEvents(db),
			store.NewAccounts(db),
			[]billing.Provider{
				billingproviders.NewManual(cfg.BillingManualUpgrade),
				billingproviders.NewPayTR(),
				billingproviders.NewStripe(),
			},
		),
		webhooks:     webhookDispatcher,
		webhookStore: webhookStore,
		identities:   store.NewSocialIdentities(db),
	}

	return s, nil
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", health.Handler(s.db, s.rdb))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/auth/sign-in", http.StatusFound)
	})

	mux.HandleFunc("GET /v1/widget/config", s.handleWidgetConfig)
	mux.HandleFunc("OPTIONS /v1/widget/config", s.handlePublicCORS)
	mux.HandleFunc("POST /v1/widget/exchange", s.handleWidgetExchange)
	mux.HandleFunc("OPTIONS /v1/widget/exchange", s.handlePublicCORS)
	mux.HandleFunc("GET /v1/widget/embed.js", s.handleWidgetEmbedJS)
	mux.HandleFunc("POST /auth/sign-up", s.handleSignUp)
	mux.HandleFunc("OPTIONS /auth/sign-up", s.handlePublicCORS)
	mux.HandleFunc("POST /auth/sign-in", s.handleSignIn)
	mux.HandleFunc("OPTIONS /auth/sign-in", s.handlePublicCORS)
	mux.HandleFunc("POST /auth/mfa/start", s.handleMFAStart)
	mux.HandleFunc("POST /auth/mfa/confirm", s.handleMFAConfirm)
	mux.HandleFunc("POST /auth/mfa/disable", s.handleMFADisable)
	mux.HandleFunc("POST /auth/mfa/verify", s.handleMFAVerify)
	mux.HandleFunc("POST /auth/resend-verification", s.handleResendVerification)
	mux.HandleFunc("POST /auth/verify-code", s.handleVerifyCode)
	mux.HandleFunc("GET /auth/verify-email", s.handleVerifyEmail)
	mux.HandleFunc("POST /auth/forgot-password", s.handleForgotPassword)
	mux.HandleFunc("POST /auth/reset-password", s.handleResetPassword)
	mux.HandleFunc("POST /auth/sign-out", s.handleSignOut)
	mux.HandleFunc("GET /auth/me", s.handleMe)
	mux.HandleFunc("GET /auth/account", s.handlePageAccount)
	mux.HandleFunc("GET /auth/account/sessions", s.handleAccountSessions)
	mux.HandleFunc("DELETE /auth/account/sessions", s.handleAccountRevokeAllSessions)
	mux.HandleFunc("DELETE /auth/account/sessions/{id}", s.handleAccountRevokeSession)
	mux.HandleFunc("GET /auth/account/passkeys", s.handleAccountPasskeys)
	mux.HandleFunc("DELETE /auth/account/passkeys/{id}", s.handleAccountDeletePasskey)
	mux.HandleFunc("GET /auth/account/identities", s.handleAccountIdentities)
	mux.HandleFunc("DELETE /auth/account/identities/{id}", s.handleAccountUnlinkIdentity)
	mux.HandleFunc("GET /auth/account/identities/{provider}/link", s.handleAccountLinkIdentity)
	mux.HandleFunc("GET /auth/account/mfa", s.handleAccountMFA)
	mux.HandleFunc("GET /auth/account/mfa/qr", s.handleAccountMFAQRCode)
	mux.HandleFunc("PUT /auth/account/password", s.handleAccountChangePassword)

	mux.HandleFunc("GET /auth/sign-in", s.handlePageSignIn)
	mux.HandleFunc("GET /auth/sign-up", s.handlePageSignUp)
	mux.HandleFunc("GET /auth/forgot-password", s.handlePageForgotPassword)
	mux.HandleFunc("GET /auth/reset-password", s.handlePageResetPassword)
	mux.HandleFunc("GET /auth/verify", s.handlePageVerifyEmail)
	mux.Handle("GET /auth/static/", http.StripPrefix("/auth/static/", pages.StaticHandler()))

	mux.HandleFunc("GET /auth/social/google/start", s.handleSocialStart)
	mux.HandleFunc("GET /auth/social/google/callback", s.handleSocialCallback)
	mux.HandleFunc("GET /auth/social/github/start", s.handleSocialStart)
	mux.HandleFunc("GET /auth/social/github/callback", s.handleSocialCallback)

	mux.HandleFunc("POST /auth/passkey/register/begin", s.handlePasskeyRegisterBegin)
	mux.HandleFunc("POST /auth/passkey/register/finish", s.handlePasskeyRegisterFinish)
	mux.HandleFunc("POST /auth/passkey/sign-in/begin", s.handlePasskeySignInBegin)
	mux.HandleFunc("POST /auth/passkey/sign-in/finish", s.handlePasskeySignInFinish)

	mux.HandleFunc("POST /oauth/token", s.handleOAuthToken)
	mux.HandleFunc("GET /oauth/authorize", s.handleAuthorize)
	mux.HandleFunc("GET /oauth/userinfo", s.handleUserInfo)
	mux.HandleFunc("GET /.well-known/jwks.json", s.handleJWKS)
	mux.HandleFunc("GET /.well-known/openid-configuration", s.handleOpenIDConfiguration)

	mux.HandleFunc("GET /dashboard/api/me", s.handleDashboardMe)
	mux.HandleFunc("GET /dashboard/api/projects", s.handleDashboardProjects)
	mux.HandleFunc("POST /dashboard/api/projects", s.handleDashboardProjects)
	mux.HandleFunc("POST /dashboard/api/projects/{id}/select", s.handleDashboardSelectProject)
	mux.HandleFunc("GET /dashboard/api/project", s.handleDashboardProject)
	mux.HandleFunc("PUT /dashboard/api/project", s.handleDashboardProject)
	mux.HandleFunc("GET /dashboard/api/integration", s.handleDashboardIntegration)
	mux.HandleFunc("PUT /dashboard/api/integration", s.handleDashboardIntegration)
	mux.HandleFunc("GET /dashboard/api/users", s.handleDashboardUsers)
	mux.HandleFunc("DELETE /dashboard/api/users/{id}", s.handleDashboardDeleteUser)
	mux.HandleFunc("POST /dashboard/api/users/{id}/disable", s.handleDashboardDisableUser)
	mux.HandleFunc("POST /dashboard/api/users/{id}/enable", s.handleDashboardEnableUser)
	mux.HandleFunc("POST /dashboard/api/users/{id}/resend-verification", s.handleDashboardResendUserVerification)
	mux.HandleFunc("GET /dashboard/api/sessions", s.handleDashboardSessions)
	mux.HandleFunc("POST /dashboard/api/sessions/{id}/revoke", s.handleDashboardRevokeSession)
	mux.HandleFunc("GET /dashboard/api/audit", s.handleDashboardAudit)
	mux.HandleFunc("GET /dashboard/api/providers", s.handleDashboardProviders)
	mux.HandleFunc("PUT /dashboard/api/providers/{provider}", s.handleDashboardProviderUpdate)
	mux.HandleFunc("GET /dashboard/api/theme", s.handleDashboardThemeGet)
	mux.HandleFunc("PUT /dashboard/api/theme", s.handleDashboardThemePut)
	mux.HandleFunc("GET /dashboard/api/billing", s.handleDashboardBilling)
	mux.HandleFunc("POST /dashboard/api/billing/upgrade", s.handleDashboardBillingUpgrade)
	mux.HandleFunc("GET /dashboard/api/webhooks", s.handleDashboardWebhooks)
	mux.HandleFunc("POST /dashboard/api/webhooks", s.handleDashboardWebhooks)
	mux.HandleFunc("DELETE /dashboard/api/webhooks/{id}", s.handleDashboardWebhookDelete)
	mux.HandleFunc("POST /dashboard/api/webhooks/{id}/test", s.handleDashboardWebhookTest)
	mux.HandleFunc("GET /dashboard/api/webhooks/{id}/deliveries", s.handleDashboardWebhookDeliveries)
	mux.Handle("GET /dashboard/", http.StripPrefix("/dashboard/", adminui.Handler()))

	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"} {
		mux.HandleFunc(method+" /admin/", redirectLegacyAdmin)
	}

	return withSecurityHeaders(withDashboardCORS(s.cfg.DashboardOrigin, mux))
}

func withDashboardCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/dashboard/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if origin != "" && r.Header.Get("Origin") == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func redirectLegacyAdmin(w http.ResponseWriter, r *http.Request) {
	target := strings.Replace(r.URL.Path, "/admin", "/dashboard", 1)
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}

func (s *Server) Close() {
	if s.db != nil {
		s.db.Close()
	}
	if s.rdb != nil {
		_ = s.rdb.Close()
	}
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) PingRedis(ctx context.Context) error {
	if s.rdb == nil {
		return nil
	}
	return s.rdb.Ping(ctx).Err()
}
