import { useCallback, useEffect, useState } from "react";
import type { Session } from "./index.js";

type ClientLike = {
  getSession: () => Session | null;
  getUserInfo: (accessToken?: string) => Promise<unknown>;
  signInWithRedirect: (state?: string, nonce?: string) => Promise<void>;
  signOut: () => Promise<void>;
  getPasswordStatus?: (accessToken?: string) => Promise<{ has_password: boolean }>;
  changePassword?: (currentPassword: string, newPassword: string, accessToken?: string) => Promise<{ message: string }>;
  setPassword?: (newPassword: string, accessToken?: string) => Promise<{ message: string }>;
};

export function useSession(client: ClientLike) {
  const [session, setSession] = useState<Session | null>(() => client.getSession());
  const [user, setUser] = useState<unknown>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    const current = client.getSession();
    setSession(current);
    if (!current) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      setUser(await client.getUserInfo(current.accessToken));
    } catch {
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, [client]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return {
    session,
    user,
    loading,
    signIn: client.signInWithRedirect,
    signOut: async () => {
      await client.signOut();
      setSession(null);
      setUser(null);
    },
    getPasswordStatus: async () => {
      if (!client.getPasswordStatus) throw new Error("getPasswordStatus_not_supported");
      return client.getPasswordStatus(session?.accessToken);
    },
    changePassword: async (currentPassword: string, newPassword: string) => {
      if (!client.changePassword) throw new Error("changePassword_not_supported");
      return client.changePassword(currentPassword, newPassword, session?.accessToken);
    },
    setPassword: async (newPassword: string) => {
      if (!client.setPassword) throw new Error("setPassword_not_supported");
      return client.setPassword(newPassword, session?.accessToken);
    },
    refresh,
  };
}

export * from "./PasswordModal.js";

