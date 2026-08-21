import { createContext, useContext, useState, type ReactNode } from "react";
import { apiFetch, getToken, setToken } from "../api/client";

interface AuthContextValue {
  isAuthenticated: boolean;
  login: (password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

interface AuthenticateByNameResponse {
  AccessToken: string;
  ServerId: string;
  User: { Id: string; Name: string };
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [isAuthenticated, setIsAuthenticated] = useState(() => Boolean(getToken()));

  async function login(password: string) {
    const res = await apiFetch<AuthenticateByNameResponse>("/Users/AuthenticateByName", {
      method: "POST",
      body: JSON.stringify({ Username: "magicboxie-web", Pw: password }),
    });
    setToken(res.AccessToken);
    setIsAuthenticated(true);
  }

  function logout() {
    setToken(null);
    setIsAuthenticated(false);
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
