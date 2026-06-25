// Lightweight auth state backed by GET /me. Keeps the logged-in user/profile
// in module state so the sidebar and pages share one source of truth.
import { useEffect, useState, useCallback } from "react";
import { api, getToken, clearToken, type User, type Profile } from "../api/client";

export type AuthState = {
  loading: boolean;
  user: User | null;
  profile: Profile | null;
};

export function useAuth() {
  const [state, setState] = useState<AuthState>({
    loading: true,
    user: null,
    profile: null,
  });

  const refresh = useCallback(async () => {
    if (!getToken()) {
      setState({ loading: false, user: null, profile: null });
      return;
    }
    try {
      const { user, profile } = await api.me();
      setState({ loading: false, user, profile });
    } catch {
      clearToken();
      setState({ loading: false, user: null, profile: null });
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const logout = useCallback(() => {
    clearToken();
    setState({ loading: false, user: null, profile: null });
  }, []);

  return { ...state, refresh, logout };
}
