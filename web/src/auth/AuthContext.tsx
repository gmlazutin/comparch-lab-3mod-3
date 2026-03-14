import React, { createContext, useState, useEffect, useContext, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { loadSession, saveSession, clearSession as clearSessionStorage } from "./auth";
import { type Session } from "./types";

interface AuthContextType {
    session: Session | null;
    setSession: (u: Session | null) => void;
}

const AuthContext = createContext<AuthContextType>({
    session: null,
    setSession: () => {}
});

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [session, setSessionState] = useState<Session | null>(loadSession());
    const navigate = useNavigate();

    const logout = () => {
        clearSessionStorage();
        setSessionState(null);
        navigate("/login");
    };

    useEffect(() => {
        let timer: NodeJS.Timeout | null = null;

        if (session) {
            const delay = session.expires - Date.now();

            if (delay <= 0) {
                logout();
            } else {
                timer = setTimeout(logout, delay);
            }
        }

        return () => {
            if (timer) clearTimeout(timer);
        };
    }, [session]);

    const setSession = (u: Session | null) => {
        if (u) {
            saveSession(u);
            setSessionState(u);
            navigate("/");
        } else {
            logout();
        }
    };

    return (
        <AuthContext.Provider value={{ session, setSession }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);