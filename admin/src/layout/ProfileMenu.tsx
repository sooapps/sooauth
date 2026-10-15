import { useEffect, useRef, useState } from "react";
import { API, csrf, type Data } from "../api";
import { useI18n } from "../context";

export type AppTheme = "light" | "dark" | "system";

export function ProfileMenu({ me, theme, onThemeChange }: { me: Data | null; theme: AppTheme; onThemeChange: (theme: AppTheme) => void }) {
  const { tr } = useI18n();
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    if (open) {
      document.addEventListener("mousedown", handleOutsideClick);
      document.addEventListener("keydown", handleKeyDown);
    }
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  const handleSignOut = () => {
    fetch(`${API}/auth/sign-out`, {
      method: "POST",
      credentials: "include",
      headers: { "X-CSRF-Token": csrf() }
    }).then(() => {
      location.href = "/auth/sign-in";
    });
  };

  return (
    <div className="user-profile-menu-wrap" ref={menuRef}>
      <button
        type="button"
        className={`profile-card-trigger ${open ? "open" : ""}`}
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-haspopup="true"
        aria-label={tr("profile.ariaLabel")}
      >
        <div className="profile-identity">
          <span className="avatar">{(me?.email || "?").slice(0, 1).toUpperCase()}</span>
          <div className="profile-info">
            <strong title={me?.email}>{me?.email || tr("common.loadingAccount")}</strong>
            <span className="profile-badge">{tr("profile.platformAdmin")}</span>
          </div>
        </div>
        <svg className="profile-chevron" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>

      {open && (
        <div className="profile-popover" role="menu">
          <div className="popover-header">
            <strong>{me?.email}</strong>
            <small>{tr("profile.accountSettings")}</small>
          </div>

          <div className="popover-divider" />

          <div className="popover-section">
            <span className="popover-label">{tr("profile.appearance")}</span>
            <div className="theme-toggle-group">
              {(["light", "dark", "system"] as AppTheme[]).map((item) => (
                <button
                  key={item}
                  type="button"
                  className={`theme-pill ${theme === item ? "active" : ""}`}
                  onClick={() => { onThemeChange(item); setOpen(false); }}
                >
                  {item === "light" && tr("profile.themeLight")}
                  {item === "dark" && tr("profile.themeDark")}
                  {item === "system" && tr("profile.themeSystem")}
                </button>
              ))}
            </div>
          </div>

          <div className="popover-divider" />

          <div className="popover-links">
            <a href="/auth/account" className="popover-item">
              <span>{tr("common.securityCenter")}</span>
              <span>↗</span>
            </a>
            <a href="https://sooauth.com/docs" target="_blank" rel="noopener noreferrer" className="popover-item">
              <span>{tr("common.documentation")}</span>
              <span>↗</span>
            </a>
          </div>

          <div className="popover-divider" />

          <button type="button" className="popover-item danger" onClick={handleSignOut}>
            <span>{tr("profile.signOut")}</span>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
              <polyline points="16 17 21 12 16 7"></polyline>
              <line x1="21" y1="12" x2="9" y2="12"></line>
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}

