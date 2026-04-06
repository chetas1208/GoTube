"use client";

import { createContext, useContext } from "react";
import type { User } from "@/types";

export interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isSessionExpiring: boolean;
  secondsUntilLogout: number | null;
  lastActivityAt: number | null;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  staySignedIn: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType>({
  user: null,
  isLoading: true,
  isSessionExpiring: false,
  secondsUntilLogout: null,
  lastActivityAt: null,
  login: async () => {},
  register: async () => {},
  logout: async () => {},
  staySignedIn: async () => {},
});

export function useAuth() {
  return useContext(AuthContext);
}
