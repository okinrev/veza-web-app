import { create } from "zustand";
import { api } from "@/shared/api/api";

export interface User {
  id: string;
  username: string;
  email: string;
  token: string;
  avatar?: string;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,

  login: async (email: string, password: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await api.post('/auth/login', { email, password });
      
      if (response.data.success) {
        const userData = response.data.data;
        set({
          user: userData,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        });
        localStorage.setItem('user', JSON.stringify(userData));
      } else {
        set({
          isAuthenticated: false,
          isLoading: false,
          error: response.data.error || 'Erreur lors de la connexion',
        });
      }
    } catch (error: any) {
      set({
        isAuthenticated: false,
        isLoading: false,
        error: error.response?.data?.error || 'Erreur lors de la connexion',
      });
    }
  },

  register: async (username: string, email: string, password: string) => {
    try {
      set({ isLoading: true, error: null });
      const response = await api.post('/auth/register', {
        username,
        email,
        password,
      });

      if (response.data.success) {
        const userData = response.data.data;
        set({
          user: userData,
          isAuthenticated: true,
          isLoading: false,
          error: null,
        });
        localStorage.setItem('user', JSON.stringify(userData));
      } else {
        set({
          isAuthenticated: false,
          isLoading: false,
          error: response.data.error || 'Erreur lors de l\'inscription',
        });
      }
    } catch (error: any) {
      set({
        isAuthenticated: false,
        isLoading: false,
        error: error.response?.data?.error || 'Erreur lors de l\'inscription',
      });
    }
  },

  logout: () => {
    localStorage.removeItem('user');
    set({
      user: null,
      isAuthenticated: false,
      error: null,
    });
  },
})); 