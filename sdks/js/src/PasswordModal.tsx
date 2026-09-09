import React, { useEffect, useState } from "react";

export type PasswordModalProps = {
  isOpen: boolean;
  onClose: () => void;
  client: {
    getPasswordStatus: (accessToken?: string) => Promise<{ has_password: boolean }>;
    changePassword: (currentPassword: string, newPassword: string, accessToken?: string) => Promise<{ message: string }>;
    setPassword: (newPassword: string, accessToken?: string) => Promise<{ message: string }>;
    getSession?: () => { accessToken: string } | null;
  };
  accessToken?: string;
  onSuccess?: (type: "set" | "change") => void;
  onError?: (err: Error) => void;
  className?: string;
};

export function PasswordModal({
  isOpen,
  onClose,
  client,
  accessToken,
  onSuccess,
  onError,
  className,
}: PasswordModalProps) {
  const [hasPassword, setHasPassword] = useState<boolean | null>(null);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [fetchingStatus, setFetchingStatus] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const token = accessToken ?? client.getSession?.()?.accessToken;

  const fetchStatus = React.useCallback(() => {
    setFetchingStatus(true);
    setError(null);

    client
      .getPasswordStatus(token)
      .then((res) => {
        setHasPassword(res.has_password);
      })
      .catch((err) => {
        setError(err.message || "Failed to load account security details");
        setHasPassword(null);
      })
      .finally(() => {
        setFetchingStatus(false);
      });
  }, [client, token]);

  useEffect(() => {
    if (!isOpen) {
      setHasPassword(null);
      setFetchingStatus(false);
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setError(null);
      setSuccessMsg(null);
      return;
    }

    fetchStatus();
  }, [isOpen, fetchStatus]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccessMsg(null);

    if (newPassword.length < 8) {
      setError("New password must be at least 8 characters long.");
      return;
    }
    if (newPassword !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }
    if (hasPassword && !currentPassword) {
      setError("Current password is required.");
      return;
    }

    setLoading(true);
    try {
      if (hasPassword) {
        await client.changePassword(currentPassword, newPassword, token);
        setSuccessMsg("Password changed successfully.");
        onSuccess?.("change");
      } else {
        await client.setPassword(newPassword, token);
        setSuccessMsg("Password set successfully.");
        setHasPassword(true);
        onSuccess?.("set");
      }
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
      setTimeout(() => {
        onClose();
      }, 1200);
    } catch (err: any) {
      const msg = err.message || "Operation failed";
      setError(msg);
      onError?.(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 9999,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "rgba(0, 0, 0, 0.5)",
        padding: "16px",
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className={className}
        style={{
          backgroundColor: "#ffffff",
          color: "#0f172a",
          borderRadius: "12px",
          width: "100%",
          maxWidth: "420px",
          padding: "24px",
          boxShadow: "0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)",
          fontFamily: "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
          boxSizing: "border-box",
        }}
      >
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: "16px" }}>
          <div>
            <h2 style={{ margin: 0, fontSize: "1.25rem", fontWeight: 600 }}>
              {fetchingStatus
                ? "Security"
                : hasPassword
                ? "Change password"
                : "Set password"}
            </h2>
            <p style={{ margin: "4px 0 0", fontSize: "0.875rem", color: "#64748b" }}>
              {hasPassword
                ? "Update your current account password."
                : "Set a password to enable email & password sign in."}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close modal"
            style={{
              background: "none",
              border: "none",
              fontSize: "1.25rem",
              cursor: "pointer",
              color: "#94a3b8",
              padding: "4px",
              lineHeight: 1,
            }}
          >
            &times;
          </button>
        </div>

        {error && (
          <div
            style={{
              backgroundColor: "#fef2f2",
              border: "1px solid #fee2e2",
              color: "#b91c1c",
              padding: "10px 12px",
              borderRadius: "6px",
              fontSize: "0.875rem",
              marginBottom: "16px",
            }}
          >
            {error}
          </div>
        )}

        {successMsg && (
          <div
            style={{
              backgroundColor: "#f0fdf4",
              border: "1px solid #dcfce7",
              color: "#15803d",
              padding: "10px 12px",
              borderRadius: "6px",
              fontSize: "0.875rem",
              marginBottom: "16px",
            }}
          >
            {successMsg}
          </div>
        )}

        {fetchingStatus ? (
          <div style={{ textAlign: "center", padding: "24px 0", color: "#64748b", fontSize: "0.875rem" }}>
            Loading security details...
          </div>
        ) : hasPassword === null ? (
          <div style={{ textAlign: "center", padding: "16px 0" }}>
            <button
              type="button"
              onClick={() => fetchStatus()}
              style={{
                padding: "8px 16px",
                borderRadius: "6px",
                border: "1px solid #cbd5e1",
                backgroundColor: "#ffffff",
                color: "#334155",
                fontSize: "0.875rem",
                fontWeight: 500,
                cursor: "pointer",
              }}
            >
              Retry
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "14px" }}>
            {hasPassword && (
              <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "0.875rem", fontWeight: 500 }}>
                Current password
                <input
                  type="password"
                  name="current_password"
                  autoComplete="current-password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  required
                  style={{
                    padding: "8px 12px",
                    border: "1px solid #cbd5e1",
                    borderRadius: "6px",
                    fontSize: "0.875rem",
                    outline: "none",
                  }}
                />
              </label>
            )}

            <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "0.875rem", fontWeight: 500 }}>
              New password
              <input
                type="password"
                name="new_password"
                autoComplete="new-password"
                minLength={8}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
                placeholder="At least 8 characters"
                style={{
                  padding: "8px 12px",
                  border: "1px solid #cbd5e1",
                  borderRadius: "6px",
                  fontSize: "0.875rem",
                  outline: "none",
                }}
              />
            </label>

            <label style={{ display: "flex", flexDirection: "column", gap: "6px", fontSize: "0.875rem", fontWeight: 500 }}>
              Confirm new password
              <input
                type="password"
                name="confirm_password"
                autoComplete="new-password"
                minLength={8}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                placeholder="Repeat new password"
                style={{
                  padding: "8px 12px",
                  border: "1px solid #cbd5e1",
                  borderRadius: "6px",
                  fontSize: "0.875rem",
                  outline: "none",
                }}
              />
            </label>

            <div style={{ display: "flex", justifyContent: "flex-end", gap: "10px", marginTop: "8px" }}>
              <button
                type="button"
                onClick={onClose}
                disabled={loading}
                style={{
                  padding: "8px 16px",
                  borderRadius: "6px",
                  border: "1px solid #cbd5e1",
                  backgroundColor: "#ffffff",
                  color: "#334155",
                  fontSize: "0.875rem",
                  fontWeight: 500,
                  cursor: "pointer",
                }}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                style={{
                  padding: "8px 16px",
                  borderRadius: "6px",
                  border: "none",
                  backgroundColor: "#0f172a",
                  color: "#ffffff",
                  fontSize: "0.875rem",
                  fontWeight: 500,
                  cursor: loading ? "not-allowed" : "pointer",
                  opacity: loading ? 0.7 : 1,
                }}
              >
                {loading
                  ? "Saving..."
                  : hasPassword
                  ? "Change password"
                  : "Set password"}
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
