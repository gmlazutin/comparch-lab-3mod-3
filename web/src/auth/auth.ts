import { OpenAPI } from "../api";
import { type Session } from "./types";

const TOKEN_KEY = "authToken";

export const saveSession = (user: Session) => {
    OpenAPI.TOKEN = user.token;
    localStorage.setItem(TOKEN_KEY, JSON.stringify(user));
};

export const loadSession = (): Session | null => {
    OpenAPI.TOKEN = undefined;
    const s = localStorage.getItem(TOKEN_KEY);
    if (!s) return null;
    const user: Session = JSON.parse(s);
    if (Date.now() > user.expires) {
        localStorage.removeItem(TOKEN_KEY);
        return null;
    }
    OpenAPI.TOKEN = user.token;
    return user;
};

export const clearSession = () => {
    OpenAPI.TOKEN = undefined;
    localStorage.removeItem(TOKEN_KEY);
};