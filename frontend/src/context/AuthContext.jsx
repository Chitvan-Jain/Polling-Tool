import { createContext, useContext, useState, useCallback } from "react";
import { api } from "../api/client";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem("token"));

  const login = useCallback(async (email, password) => {
    const data = await api.login({ email, password });
    localStorage.setItem("token", data.token);
    setToken(data.token);
  }, []);

  const signup = useCallback(async (name, email, password) => {
    const data = await api.signup({ name, email, password });
    localStorage.setItem("token", data.token);
    setToken(data.token);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem("token");
    setToken(null);
  }, []);

  const value = { token, isAuthenticated: Boolean(token), login, signup, logout };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside an AuthProvider");
  return ctx;
}