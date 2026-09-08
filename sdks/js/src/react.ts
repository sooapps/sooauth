import { useCallback, useEffect, useState } from "react";
import type { Session } from "./index.js";

type ClientLike = {
  getSession: () => Session | null;
  getUserInfo: (accessToken?: string) => Promise<unknown>;
  signInWithRedirect: (state?: string, nonce?: string) => Promise<void>;
  signOut: () => Promise<void>;
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
    refresh,
  };
}
