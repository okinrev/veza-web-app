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
          const response = await API.post<ApiResponse<{ 
            user: User; 
            access_token: string; 
            refresh_token: string;
            expires_in: number;
          }>>(
            ENDPOINTS.LOGIN,
            credentials
          );
          
          console.log('Réponse login complète:', response);
          console.log('Données de réponse:', response.data);
          
          const { user, access_token, refresh_token } = response.data.data;
          
          console.log('Access token reçu:', access_token ? 'Présent' : 'Manquant');
          console.log('Refresh token reçu:', refresh_token ? 'Présent' : 'Manquant');
          console.log('User reçu:', user);
          
          // Stocker l'access token comme token principal
          localStorage.setItem('authToken', access_token);
          localStorage.setItem('refreshToken', refresh_token);
          set({
            user,
            token: access_token,
            isAuthenticated: true,
            isLoading: false,
          });
          
          console.log('État authStore après login:', { user, token: access_token, isAuthenticated: true });
        } catch (error) {
          console.error('Erreur login authStore:', error);
          set({ isLoading: false });
          throw error;
        }
      },
      
      register: async (data) => {
        set({ isLoading: true });
        try {
          const response = await API.post<ApiResponse<{ 
            user: User; 
            access_token: string; 
            refresh_token: string;
            expires_in: number;
          }>>(
            ENDPOINTS.REGISTER,
            data
          );
          
          const { user, access_token, refresh_token } = response.data.data;
          
          localStorage.setItem('authToken', access_token);
          localStorage.setItem('refreshToken', refresh_token);
          set({
            user,
            token: access_token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },
      
      logout: () => {
        console.log('[AuthStore] Déconnexion - nettoyage des tokens');
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        });
      },
      
      refreshToken: async () => {
        try {
          console.log('[AuthStore] Tentative de refresh du token...');
          const refreshToken = localStorage.getItem('refreshToken');
          if (!refreshToken) {
            throw new Error('No refresh token found');
          }

          const response = await API.post<ApiResponse<{ 
            access_token: string; 
            refresh_token?: string;
            expires_in: number 
          }>>(ENDPOINTS.REFRESH, {
            refresh_token: refreshToken,
          });
          
          const { access_token, refresh_token: newRefreshToken, expires_in } = response.data.data;
          
          // Mettre à jour les tokens dans localStorage
          localStorage.setItem('authToken', access_token);
          if (newRefreshToken) {
            localStorage.setItem('refreshToken', newRefreshToken);
          }
          
          console.log('[AuthStore] Token refreshé avec succès');
          set({ token: access_token });
        } catch (error) {
          console.error('[AuthStore] Erreur lors du refresh du token:', error);
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