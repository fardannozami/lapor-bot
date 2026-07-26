import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from "react";
import type { ReactNode } from "react";
import type { EnrichedReport } from "../types";
import type { IAuthRepository, LoginResult } from "../domain/repositories";

const TOKEN_KEY = "lapor-bot-token";
const PROFILE_KEY = "lapor-bot-profile";

// Bridge so HttpClient (built before React mounts) can trigger logout.
let logoutHandler: (() => void) | null = null;

/** Call on API 401 — clears session and drops React auth state. */
export function handleUnauthorized(): void {
  if (logoutHandler) {
    logoutHandler();
    return;
  }
  clearSession();
}

function clearSession(): void {
  try {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(PROFILE_KEY);
  } catch {
    // ignore
  }
}

/** JWT exp (ms), or null if missing/malformed. */
function tokenExpiryMs(token: string): number | null {
  try {
    const part = token.split(".")[1];
    if (!part) return null;
    const json = atob(part.replace(/-/g, "+").replace(/_/g, "/"));
    const payload = JSON.parse(json) as { exp?: unknown };
    return typeof payload.exp === "number" ? payload.exp * 1000 : null;
  } catch {
    return null;
  }
}

export function isTokenValid(token: string | null | undefined): token is string {
  if (!token) return false;
  const exp = tokenExpiryMs(token);
  return exp !== null && exp > Date.now();
}

/** Read token from storage only if still unexpired; else wipe session. */
export function loadToken(): string | null {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    if (!isTokenValid(token)) {
      if (token) clearSession();
      return null;
    }
    return token;
  } catch {
    return null;
  }
}

function loadProfile(): EnrichedReport | null {
  try {
    // Drop cached profile when token is gone/expired.
    if (!loadToken()) return null;
    const raw = localStorage.getItem(PROFILE_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function saveSession(token: string, profile: EnrichedReport): void {
  try {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(PROFILE_KEY, JSON.stringify(profile));
  } catch {
    // storage full or unavailable — ignore
  }
}

export interface AuthState {
  token: string | null;
  user: EnrichedReport | null;
  isLoading: boolean;
  login: (phone: string) => Promise<EnrichedReport>;
  logout: () => void;
  getToken: () => string | null;
}

const AuthContext = createContext<AuthState | null>(null);

export interface AuthProviderProps {
  authRepo: IAuthRepository;
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({
  authRepo,
  children,
}) => {
  const [token, setToken] = useState<string | null>(loadToken);
  const [user, setUser] = useState<EnrichedReport | null>(loadProfile);
  const [isLoading, setIsLoading] = useState(false);

  const getToken = useCallback(() => {
    if (!isTokenValid(token)) {
      if (token) {
        setToken(null);
        setUser(null);
        clearSession();
      }
      return null;
    }
    return token;
  }, [token]);

  const logout = useCallback(() => {
    setToken(null);
    setUser(null);
    clearSession();
  }, []);

  const login = useCallback(
    async (phone: string): Promise<EnrichedReport> => {
      setIsLoading(true);
      try {
        const result: LoginResult = await authRepo.login(phone);
        setToken(result.token);
        setUser(result.profile);
        saveSession(result.token, result.profile);
        return result.profile;
      } finally {
        setIsLoading(false);
      }
    },
    [authRepo],
  );

  // HttpClient 401 → logout
  useEffect(() => {
    logoutHandler = logout;
    return () => {
      if (logoutHandler === logout) logoutHandler = null;
    };
  }, [logout]);

  // Drop session the moment JWT exp passes while tab is open.
  useEffect(() => {
    if (!token) return;
    const exp = tokenExpiryMs(token);
    if (exp === null) {
      logout();
      return;
    }
    const ms = exp - Date.now();
    if (ms <= 0) {
      logout();
      return;
    }
    const id = window.setTimeout(logout, ms);
    return () => window.clearTimeout(id);
  }, [token, logout]);

  // Sync token across tabs
  useEffect(() => {
    const handler = (e: StorageEvent) => {
      if (e.key !== TOKEN_KEY) return;
      if (!isTokenValid(e.newValue)) {
        setToken(null);
        setUser(null);
        if (e.newValue) clearSession();
        return;
      }
      setToken(e.newValue);
    };
    window.addEventListener("storage", handler);
    return () => window.removeEventListener("storage", handler);
  }, []);

  return (
    <AuthContext.Provider
      value={{ token, user, isLoading, login, logout, getToken }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuthContext = (): AuthState => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuthContext must be used within an AuthProvider");
  }
  return ctx;
};
