"use client";

import { startTransition, useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import { usePathname, useRouter } from "next/navigation";
import { SessionTimeoutDialog } from "@/components/auth/SessionTimeoutDialog";
import { useToast } from "@/components/ui/Toast";
import { AuthContext } from "@/lib/auth";
import { auth as authApi, setAccessToken } from "@/lib/api-client";
import type { AuthResponse, User } from "@/types";

const IDLE_TIMEOUT_MS = 15 * 60 * 1000;
const WARNING_WINDOW_MS = 60 * 1000;
const REFRESH_BUFFER_MS = 60 * 1000;
const RECENT_ACTIVITY_WINDOW_MS = 60 * 1000;
const ACTIVITY_THROTTLE_MS = 5 * 1000;
const LAST_ACTIVITY_STORAGE_KEY = "gotube:last-activity-at";
const AUTH_EVENT_STORAGE_KEY = "gotube:auth-event";
const PROTECTED_ROUTE_PREFIXES = ["/profile", "/studio", "/upload"];

type SignOutReason = "idle" | "manual" | "refresh-failed" | "remote-logout";

function parseJwtExpiration(token: string): number | null {
  if (typeof window === "undefined") {
    return null;
  }

  const parts = token.split(".");
  if (parts.length < 2) {
    return null;
  }

  try {
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
    const payload = JSON.parse(window.atob(padded)) as { exp?: number };
    return typeof payload.exp === "number" ? payload.exp * 1000 : null;
  } catch {
    return null;
  }
}

function readStoredTimestamp(key: string): number | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(key);
  if (!raw) {
    return null;
  }

  const value = Number(raw);
  return Number.isFinite(value) ? value : null;
}

function isProtectedRoute(pathname: string) {
  return PROTECTED_ROUTE_PREFIXES.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { addToast } = useToast();
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [lastActivityAt, setLastActivityAt] = useState<number | null>(null);
  const [isSessionExpiring, setIsSessionExpiring] = useState(false);
  const [secondsUntilLogout, setSecondsUntilLogout] = useState<number | null>(null);
  const [isRefreshingSession, setIsRefreshingSession] = useState(false);
  const accessTokenExpiryRef = useRef<number | null>(null);
  const lastActivityRef = useRef<number | null>(null);
  const lastActivityWriteRef = useRef<number>(0);
  const userRef = useRef<User | null>(null);
  const expiringRef = useRef(false);
  const refreshPromiseRef = useRef<Promise<boolean> | null>(null);
  const signOutInProgressRef = useRef(false);

  const syncSessionWarningState = useCallback((nextExpiring: boolean, nextSeconds: number | null) => {
    expiringRef.current = nextExpiring;
    setIsSessionExpiring(nextExpiring);
    setSecondsUntilLogout(nextSeconds);
  }, []);

  const setCurrentUser = useCallback((nextUser: User | null) => {
    userRef.current = nextUser;
    setUser(nextUser);
  }, []);

  const setAccessTokenExpiry = useCallback((nextExpiry: number | null) => {
    accessTokenExpiryRef.current = nextExpiry;
  }, []);

  const setSharedLastActivity = useCallback((timestamp: number | null, persist: boolean) => {
    lastActivityRef.current = timestamp;
    setLastActivityAt(timestamp);

    if (typeof window === "undefined" || !persist) {
      return;
    }

    if (timestamp === null) {
      lastActivityWriteRef.current = 0;
      window.localStorage.removeItem(LAST_ACTIVITY_STORAGE_KEY);
      return;
    }

    lastActivityWriteRef.current = timestamp;
    window.localStorage.setItem(LAST_ACTIVITY_STORAGE_KEY, String(timestamp));
  }, []);

  const clearLocalSession = useCallback((clearSharedActivity: boolean) => {
    setAccessToken(null);
    setCurrentUser(null);
    setAccessTokenExpiry(null);
    setSharedLastActivity(null, clearSharedActivity);
    syncSessionWarningState(false, null);
    setIsRefreshingSession(false);
  }, [setAccessTokenExpiry, setCurrentUser, setSharedLastActivity, syncSessionWarningState]);

  const broadcastSignOut = useCallback((reason: SignOutReason) => {
    if (typeof window === "undefined") {
      return;
    }

    window.localStorage.setItem(
      AUTH_EVENT_STORAGE_KEY,
      JSON.stringify({
        type: "logout",
        reason,
        at: Date.now(),
      })
    );
  }, []);

  const applyAuthResponse = useCallback(
    (data: AuthResponse) => {
      setAccessToken(data.access_token);
      setCurrentUser(data.user);
      setAccessTokenExpiry(parseJwtExpiration(data.access_token));
    },
    [setAccessTokenExpiry, setCurrentUser]
  );

  const signOut = useCallback(
    async (
      reason: SignOutReason,
      options: {
        broadcast?: boolean;
        serverLogout?: boolean;
        toastMessage?: string;
      } = {}
    ) => {
      if (signOutInProgressRef.current) {
        return;
      }

      signOutInProgressRef.current = true;

      const shouldRedirect = isProtectedRoute(pathname);
      const logoutRequest =
        options.serverLogout && userRef.current
          ? authApi.logout().catch(() => undefined)
          : Promise.resolve();

      clearLocalSession(true);

      if (options.broadcast) {
        broadcastSignOut(reason);
      }

      if (options.toastMessage) {
        addToast(options.toastMessage, reason === "refresh-failed" ? "error" : "info");
      }

      if (shouldRedirect) {
        startTransition(() => {
          router.replace("/login");
        });
      }

      await logoutRequest;
      signOutInProgressRef.current = false;
    },
    [addToast, broadcastSignOut, clearLocalSession, pathname, router]
  );

  const refreshSession = useCallback(
    async (
      trigger: "activity" | "scheduled" | "stay-signed-in" | "visibility",
      options: {
        force?: boolean;
      } = {}
    ) => {
      if (typeof window === "undefined" || (!options.force && document.hidden)) {
        return false;
      }

      if (refreshPromiseRef.current) {
        return refreshPromiseRef.current;
      }

      const lastKnownActivity = lastActivityRef.current;
      const now = Date.now();
      if (!options.force) {
        if (!lastKnownActivity || now-lastKnownActivity >= IDLE_TIMEOUT_MS) {
          await signOut("idle", {
            broadcast: true,
            serverLogout: true,
            toastMessage: "Your session ended after 15 minutes of inactivity.",
          });
          return false;
        }

        if (trigger === "scheduled" && now-lastKnownActivity > RECENT_ACTIVITY_WINDOW_MS) {
          return false;
        }
      }

      setIsRefreshingSession(true);

      const promise = authApi
        .refresh()
        .then((data) => {
          applyAuthResponse(data);
          return true;
        })
        .catch(async () => {
          await signOut("refresh-failed", {
            broadcast: true,
            toastMessage: "Your session ended. Sign in again to continue.",
          });
          return false;
        })
        .finally(() => {
          setIsRefreshingSession(false);
          refreshPromiseRef.current = null;
        });

      refreshPromiseRef.current = promise;
      return promise;
    },
    [applyAuthResponse, signOut]
  );

  const loadUser = useCallback(async () => {
    try {
      const data = await authApi.refresh();
      applyAuthResponse(data);
      const now = Date.now();
      setSharedLastActivity(now, true);
      syncSessionWarningState(false, null);
    } catch {
      clearLocalSession(false);
    } finally {
      setIsLoading(false);
    }
  }, [applyAuthResponse, clearLocalSession, setSharedLastActivity, syncSessionWarningState]);

  useEffect(() => {
    void loadUser();
  }, [loadUser]);

  const recordActivity = useCallback(
    (persist = false) => {
      if (typeof window === "undefined" || document.hidden || !userRef.current) {
        return;
      }

      const now = Date.now();
      const shouldPersist = persist || expiringRef.current || now-lastActivityWriteRef.current >= ACTIVITY_THROTTLE_MS;
      setSharedLastActivity(now, shouldPersist);
      syncSessionWarningState(false, null);

      const accessTokenExpiresAt = accessTokenExpiryRef.current;
      if (accessTokenExpiresAt !== null && accessTokenExpiresAt-now <= REFRESH_BUFFER_MS) {
        void refreshSession("activity");
      }
    },
    [refreshSession, setSharedLastActivity, syncSessionWarningState]
  );

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const handleStorage = (event: StorageEvent) => {
      if (event.key === LAST_ACTIVITY_STORAGE_KEY) {
        const nextTimestamp = readStoredTimestamp(LAST_ACTIVITY_STORAGE_KEY);
        lastActivityRef.current = nextTimestamp;
        setLastActivityAt(nextTimestamp);
        return;
      }

      if (event.key !== AUTH_EVENT_STORAGE_KEY || !event.newValue) {
        return;
      }

      try {
        const payload = JSON.parse(event.newValue) as { type?: string };
        if (payload.type === "logout") {
          void signOut("remote-logout");
        }
      } catch {
        // Ignore malformed auth events from storage.
      }
    };

    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, [signOut]);

  useEffect(() => {
    if (!user) {
      return;
    }

    const events: Array<keyof WindowEventMap> = [
      "pointerdown",
      "pointermove",
      "mousemove",
      "keydown",
      "touchstart",
      "scroll",
      "focus",
    ];

    const handleVisibilityChange = () => {
      if (document.hidden || !userRef.current) {
        return;
      }

      const latestActivity = readStoredTimestamp(LAST_ACTIVITY_STORAGE_KEY) ?? lastActivityRef.current;
      if (latestActivity && Date.now()-latestActivity >= IDLE_TIMEOUT_MS) {
        void signOut("idle", {
          broadcast: true,
          serverLogout: true,
          toastMessage: "Your session ended after 15 minutes of inactivity.",
        });
        return;
      }

      recordActivity(true);
      const accessTokenExpiresAt = accessTokenExpiryRef.current;
      if (accessTokenExpiresAt !== null && accessTokenExpiresAt-Date.now() <= REFRESH_BUFFER_MS) {
        void refreshSession("visibility");
      }
    };

    const handleActivity = () => {
      recordActivity();
    };

    for (const eventName of events) {
      window.addEventListener(eventName, handleActivity, { passive: true });
    }
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      for (const eventName of events) {
        window.removeEventListener(eventName, handleActivity);
      }
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [recordActivity, refreshSession, signOut, user]);

  useEffect(() => {
    if (!user) {
      syncSessionWarningState(false, null);
      return;
    }

    const timer = window.setInterval(() => {
      const lastKnownActivity = lastActivityRef.current;
      if (!lastKnownActivity) {
        syncSessionWarningState(false, null);
        return;
      }

      const remainingMs = IDLE_TIMEOUT_MS - (Date.now() - lastKnownActivity);
      if (remainingMs <= 0) {
        void signOut("idle", {
          broadcast: true,
          serverLogout: true,
          toastMessage: "Your session ended after 15 minutes of inactivity.",
        });
        return;
      }

      if (remainingMs <= WARNING_WINDOW_MS) {
        syncSessionWarningState(true, Math.max(1, Math.ceil(remainingMs / 1000)));
        return;
      }

      syncSessionWarningState(false, null);
    }, 1000);

    return () => window.clearInterval(timer);
  }, [signOut, syncSessionWarningState, user]);

  useEffect(() => {
    if (!user) {
      return;
    }

    const expiresAt = accessTokenExpiryRef.current;
    if (expiresAt === null) {
      return;
    }

    const refreshAt = Math.max(expiresAt - Date.now() - REFRESH_BUFFER_MS, 0);
    const timer = window.setTimeout(() => {
      const lastKnownActivity = lastActivityRef.current;
      if (document.hidden || !lastKnownActivity) {
        return;
      }
      if (Date.now() - lastKnownActivity > RECENT_ACTIVITY_WINDOW_MS) {
        return;
      }
      void refreshSession("scheduled");
    }, refreshAt);

    return () => window.clearTimeout(timer);
  }, [refreshSession, user, lastActivityAt]);

  const login = async (email: string, password: string) => {
    const data = await authApi.login({ email, password });
    applyAuthResponse(data);
    recordActivity(true);
  };

  const register = async (username: string, email: string, password: string) => {
    const data = await authApi.register({ username, email, password });
    applyAuthResponse(data);
    recordActivity(true);
  };

  const logout = async () => {
    await signOut("manual", { broadcast: true, serverLogout: true });
  };

  const staySignedIn = async () => {
    recordActivity(true);
    await refreshSession("stay-signed-in", { force: true });
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isSessionExpiring,
        secondsUntilLogout,
        lastActivityAt,
        login,
        register,
        logout,
        staySignedIn,
      }}
    >
      {children}
      <SessionTimeoutDialog
        isOpen={isSessionExpiring}
        secondsRemaining={secondsUntilLogout}
        isRefreshing={isRefreshingSession}
        onStaySignedIn={staySignedIn}
      />
    </AuthContext.Provider>
  );
}
