// BFF-backed auth state (TanStack Query), mirroring CartContext's shape. GET /api/auth/me is the
// source of truth for "am I signed in"; signup/login/logout are mutations that update it directly
// (no need to invalidate — the BFF response IS the fresh user). accounts is the sole authority on
// identity; the SPA never sees the session token, only the cookie the BFF sets.
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createContext, useCallback, useContext, useMemo, type ReactNode } from "react";
import { ApiError, api } from "../lib/api";
import type { AuthUser } from "../lib/types";

export const AUTH_QUERY_KEY = ["auth", "me"] as const;

interface SignupInput {
  email: string;
  password: string;
  name: string;
}

interface LoginInput {
  email: string;
  password: string;
}

interface AuthValue {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  signup: (input: SignupInput) => Promise<AuthUser>;
  login: (input: LoginInput) => Promise<AuthUser>;
  logout: () => void;
  isSigningUp: boolean;
  isLoggingIn: boolean;
}

const AuthContext = createContext<AuthValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();

  // A 401 here just means "not signed in" — a normal, expected state, not a query error the UI
  // should show as broken. Resolve it to `null` rather than letting react-query mark it isError.
  const meQuery = useQuery({
    queryKey: AUTH_QUERY_KEY,
    queryFn: async ({ signal }): Promise<AuthUser | null> => {
      try {
        return await api.auth.me(signal);
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) return null;
        throw err;
      }
    },
    staleTime: 60_000,
    retry: false,
  });

  const signupMut = useMutation({
    mutationFn: api.auth.signup,
    onSuccess: (user) => queryClient.setQueryData(AUTH_QUERY_KEY, user),
  });
  const loginMut = useMutation({
    mutationFn: api.auth.login,
    onSuccess: (user) => queryClient.setQueryData(AUTH_QUERY_KEY, user),
  });
  const logoutMut = useMutation({
    mutationFn: api.auth.logout,
    onSuccess: () => queryClient.setQueryData(AUTH_QUERY_KEY, null),
  });

  const signup = useCallback((input: SignupInput) => signupMut.mutateAsync(input), [signupMut]);
  const login = useCallback((input: LoginInput) => loginMut.mutateAsync(input), [loginMut]);
  const logout = useCallback(() => logoutMut.mutate(), [logoutMut]);

  const user = meQuery.data ?? null;

  const value = useMemo<AuthValue>(
    () => ({
      user,
      isAuthenticated: !!user,
      isLoading: meQuery.isLoading,
      signup,
      login,
      logout,
      isSigningUp: signupMut.isPending,
      isLoggingIn: loginMut.isPending,
    }),
    [user, meQuery.isLoading, signup, login, logout, signupMut.isPending, loginMut.isPending],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
