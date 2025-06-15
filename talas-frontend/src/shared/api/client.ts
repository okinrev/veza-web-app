import axios, { AxiosError } from 'axios';
import type { AxiosResponse, AxiosInstance, InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

// Configuration de base
const API = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Intercepteur pour les requêtes (ajout token)
API.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Intercepteur pour les réponses (gestion erreurs)
API.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error: AxiosError) => {
    // Gestion des erreurs globales
    if (error.response) {
      const status = error.response.status;
      const message = (error.response.data as any)?.message || 'Erreur inconnue';
      
      switch (status) {
        case 401:
          // Token expiré ou invalide
          localStorage.removeItem('authToken');
          window.location.href = '/login';
          console.error('Session expirée - redirection vers login');
          break;
          
        case 403:
          console.error('Accès refusé');
          break;
          
        case 404:
          console.error('Ressource introuvable');
          break;
          
        case 500:
          console.error('Erreur serveur interne');
          break;
          
        default:
          console.error('Erreur API:', message);
      }
    } else if (error.request) {
      // Erreur réseau
      console.error('Erreur de connexion - impossible de joindre le serveur');
    }
    
    return Promise.reject(error);
  }
);

export default API;

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        const accessToken = localStorage.getItem('authToken');
        if (accessToken) {
          config.headers.Authorization = `Bearer ${accessToken}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        // Handle 401 errors
        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;
          
          // Pour l'instant, on déconnecte simplement l'utilisateur
          localStorage.removeItem('authToken');
          window.location.href = '/login';
          return Promise.reject(error);
        }

        // Handle other errors
        const errorMessage = error.response?.data?.message || error.message;
        console.error('API Error:', errorMessage);
        return Promise.reject(error);
      }
    );
  }

  // Generic request methods
  async get<T>(url: string, config?: object): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: object): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: object): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async patch<T>(url: string, data?: any, config?: object): Promise<T> {
    const response = await this.client.patch<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: object): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }

  // File upload
  async uploadFile(url: string, file: File, onProgress?: (progress: number) => void) {
    const formData = new FormData();
    formData.append('file', file);

    return this.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent: any) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  }
}

export const apiClient = new ApiClient();

// Fonction utilitaire pour créer un FormData pour les uploads de fichiers
export const createFormData = (file: File, additionalData?: Record<string, any>) => {
  const formData = new FormData();
  formData.append('file', file);
  
  if (additionalData) {
    Object.entries(additionalData).forEach(([key, value]) => {
      formData.append(key, value);
    });
  }
  
  return formData;
}; 