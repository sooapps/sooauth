# Changelog

All notable changes to Sooauth are documented here. The project is currently
in pre-launch development, so entries may describe unreleased work.

## Unreleased

### Fixed

- Dashboard accounts and project end users are fully isolated: disabling or
  deleting an app user with the same email no longer signs the owner out of the
  dashboard, `/oauth/authorize` never reuses the dashboard session, and the
  owner's sessions are no longer listed under a project.
- Disabled users can no longer refresh tokens or use existing sessions.
- "Password or code" mode now offers code sign-in (it behaved like password only).
- OTP endpoints respect the project's sign-in mode and identifiers; codes are
  random, rate limited, limited to 5 attempts and actually delivered by email.

### Added

- Dashboard sidebar navigation with per-page URLs, a setup wizard, a Getting
  started checklist, a live sign-in test and AI prompts filled in with project values.
- Two-step verification for project users (authenticator app or email code).
- Delete user from the dashboard, 2FA reset, `user.deleted` webhook.
- Hosted sign-in pages support code sign-in, 2FA and sign-up fields for projects.
- Docs: sign-in methods, one-time codes, two-step verification, sign-up fields,
  social login, users and sessions, webhooks.
- The Docker image embeds the React dashboard (the legacy static dashboard is removed).
- AGPL Community Edition licensing and Sooapps trademark notices.
- TOTP MFA setup, login challenges, and one-time backup codes for platform accounts.
- Public API, licensing, roadmap, and deployment documentation.
- Product marketing content for hosted Sooauth and self-hosted deployments.

### Planned

- End-user security center (manage authenticator from the widget).
- Auth activity analytics.

## M11 - Marketing site

- Added the Cloudflare-hosted marketing site and documentation hub.
- Added the embed widget, custom auth UI guide, and AI integration prompts.
