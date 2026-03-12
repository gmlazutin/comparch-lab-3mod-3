import { createContext, useContext, useState, useEffect, type ReactNode } from "react";
import { setApiToken } from "../apiConfig";

type AuthCtx = {
  token: string | null;
  expires: number | null;
  setAuth: (token: string | null, expires?: number | null) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthCtx>({
  token: null,
  expires: null,
  setAuth: () => {},
  logout: () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState<string | null>(null);
  const [expires, setExpiresState] = useState<number | null>(null);

  useEffect(() => {
    const t = localStorage.getItem("token");
    const e = localStorage.getItem("token-expires");
    const eNum = e ? parseInt(e) : null;

    if (t && eNum) {
      const now = Date.now();
      if (eNum > now) {
        setTokenState(t);
        setExpiresState(eNum);
        setApiToken(t);
      } else {
        logout();
      }
    }
  }, []);

  const setAuth = (t: string | null, e?: number | null) => {
    if (t) {
      localStorage.setItem("token", t);
      if (e) localStorage.setItem("token-expires", e.toString());
      setTokenState(t);
      setExpiresState(e ?? null);
      setApiToken(t);
    } else {
      logout();
    }
  };

  const logout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("token-expires");
    setTokenState(null);
    setExpiresState(null);
    setApiToken(null);
  };

  useEffect(() => {
    if (!expires) return;

    const now = Date.now();
    const diff = expires - now;

    if (diff <= 0) {
      logout();
      return;
    }

    const timer = setTimeout(() => logout(), diff);
    return () => clearTimeout(timer);
  }, [expires]);

  return (
    <AuthContext.Provider value={{ token, expires, setAuth, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}