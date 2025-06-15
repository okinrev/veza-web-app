import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import API from '@/shared/api/client';
import { ENDPOINTS } from '@/shared/api/endpoints';
import type { User, LoginCredentials, RegisterData, ApiResponse } from '@/shared/api/types';

interface AuthState {
  // État
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  
  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  updateProfile: (data: Partial<User>) => Promise<void>;
  checkAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      
      login: async (credentials) => {
        set({ isLoading: true });
        try {
          const response = await API.post<ApiResponse<{ user: User; token: string }>>(
            ENDPOINTS.LOGIN,
            credentials
          );
          
          const { user, token } = response.data.data;
          
          localStorage.setItem('authToken', token);
          set({
            user,
            token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },
      
      register: async (data) => {
        set({ isLoading: true });
        try {
          const response = await API.post<ApiResponse<{ user: User; token: string }>>(
            ENDPOINTS.REGISTER,
            data
          );
          
          const { user, token } = response.data.data;
          
          localStorage.setItem('authToken', token);
          set({
            user,
            token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },
      
      logout: () => {
        localStorage.removeItem('authToken');
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        });
      },
      
      refreshToken: async () => {
        try {
          const response = await API.post<ApiResponse<{ token: string }>>(ENDPOINTS.REFRESH);
          const { token } = response.data.data;
          
          localStorage.setItem('authToken', token);
          set({ token });
        } catch (error) {
          get().logout();
          throw error;
        }
      },
      
      updateProfile: async (data) => {
        try {
          const response = await API.put<ApiResponse<User>>(ENDPOINTS.PROFILE, data);
          const user = response.data.data;
          
          set({ user });
        } catch (error) {
          throw error;
        }
      },
      
      checkAuth: async () => {
        const token = localStorage.getItem('authToken');
        if (!token) {
          set({ isAuthenticated: false });
          return;
        }
        
        try {
          const response = await API.get<ApiResponse<User>>(ENDPOINTS.PROFILE);
          const user = response.data.data;
          
          set({
            user,
            token,
            isAuthenticated: true,
          });
        } catch (error) {
          get().logout();
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
); 