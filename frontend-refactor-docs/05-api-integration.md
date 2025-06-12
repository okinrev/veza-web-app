## 8. Communication avec le backend

### 8.1 Client API et WebSocket

```
// src/shared/api/client.ts
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { useAuthStore } from '@/features/auth/store/authStore';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

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
        const token = useAuthStore.getState().accessToken;
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
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

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            await useAuthStore.getState().refreshAuth();
            const newToken = useAuthStore.getState().accessToken;
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return this.client(originalRequest);
          } catch (refreshError) {
            // Refresh failed, logout user
            useAuthStore.getState().logout();
            window.location.href = '/login';
            return Promise.reject(refreshError);
          }
        }

        return Promise.reject(error);
      }
    );
  }

  // Generic request methods
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.patch<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
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
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  }
}

export const apiClient = new ApiClient();

// src/shared/api/websocket.ts
import { useAuthStore } from '@/features/auth/store/authStore';
import { useChatStore } from '@/features/chat/store/chatStore';

type MessageHandler = (data: any) => void;

class WebSocketManager {
  private ws: WebSocket | null = null;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map();
  private messageQueue: any[] = [];

  connect() {
    const token = useAuthStore.getState().accessToken;
    if (!token) {
      console.error('No auth token available');
      return;
    }

    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:9001'}/ws?token=${token}`;

    try {
      this.ws = new WebSocket(wsUrl);
      this.setupEventHandlers();
    } catch (error) {
      console.error('WebSocket connection error:', error);
      this.scheduleReconnect();
    }
  }

  private setupEventHandlers() {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      useChatStore.getState().setConnectionStatus(true);
      
      // Send queued messages
      while (this.messageQueue.length > 0) {
        const message = this.messageQueue.shift();
        this.send(message);
      }
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.handleMessage(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      useChatStore.getState().setConnectionStatus(false);
      this.ws = null;
      this.scheduleReconnect();
    };
  }

  private handleMessage(data: any) {
    const { type, ...payload } = data;
    
    // Notify all handlers for this message type
    const handlers = this.messageHandlers.get(type);
    if (handlers) {
      handlers.forEach(handler => handler(payload));
    }

    // Global message handling
    switch (type) {
      case 'message':
        useChatStore.getState().addMessage(payload.room, payload);
        break;
      case 'dm':
        useChatStore.getState().addDirectMessage(payload.from, payload);
        break;
      case 'user_joined':
      case 'user_left':
        // Update online users
        break;
      case 'typing':
        useChatStore.getState().setTypingUser(
          payload.room,
          payload.userId,
          payload.isTyping
        );
        break;
    }
  }

  send(data: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      // Queue message for later
      this.messageQueue.push(data);
    }
  }

  on(type: string, handler: MessageHandler) {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set());
    }
    this.messageHandlers.get(type)!.add(handler);
  }

  off(type: string, handler: MessageHandler) {
    const handlers = this.messageHandlers.get(type);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return;
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.messageQueue = [];
    this.reconnectAttempts = 0;
  }
}

export const wsManager = new WebSocketManager();
```

### 8.2 Services API par module

```
// src/features/auth/services/authService.ts
import { apiClient } from '@/shared/api/client';

interface LoginRequest {
  email: string;
  password: string;
}

interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

interface AuthResponse {
  user: {
    id: number;
    email: string;
    username: string;
    role: string;
  };
  accessToken: string;
  refreshToken: string;
}

export const authService = {
  async login(data: LoginRequest): Promise<AuthResponse> {
    return apiClient.post('/login', data);
  },

  async register(data: RegisterRequest): Promise<AuthResponse> {
    return apiClient.post('/signup', data);
  },

  async logout(): Promise<void> {
    return apiClient.post('/logout');
  },

  async refresh(refreshToken: string): Promise<AuthResponse> {
    return apiClient.post('/refresh', { refreshToken });
  },

  async forgotPassword(email: string): Promise<void> {
    return apiClient.post('/forgot-password', { email });
  },

  async resetPassword(token: string, password: string): Promise<void> {
    return apiClient.post('/reset-password', { token, password });
  },
};

// src/features/products/services/productService.ts
import { apiClient } from '@/shared/api/client';

export interface Product {
  id: number;
  nom: string;
  description: string;
  prix: number;
  stock: number;
  owner_id: number;
  owner_username: string;
  created_at: string;
  updated_at: string;
  files?: ProductFile[];
}

interface ProductFile {
  id: number;
  filename: string;
  file_type: string;
  file_size: number;
  uploaded_at: string;
}

interface CreateProductRequest {
  nom: string;
  description: string;
  prix: number;
  stock: number;
}

interface ProductsResponse {
  products: Product[];
  total: number;
  page: number;
  limit: number;
}

export const productService = {
  async getProducts(params?: {
    page?: number;
    limit?: number;
    search?: string;
    sort?: string;
  }): Promise<ProductsResponse> {
    return apiClient.get('/products', { params });
  },

  async getProduct(id: number): Promise<Product> {
    return apiClient.get(`/products/${id}`);
  },

  async createProduct(data: CreateProductRequest): Promise<Product> {
    return apiClient.post('/products', data);
  },

  async updateProduct(id: number, data: Partial<CreateProductRequest>): Promise<Product> {
    return apiClient.put(`/products/${id}`, data);
  },

  async deleteProduct(id: number): Promise<void> {
    return apiClient.delete(`/products/${id}`);
  },

  async uploadProductFile(productId: number, file: File, onProgress?: (progress: number) => void) {
    return apiClient.uploadFile(`/products/${productId}/files`, file, onProgress);
  },

  async deleteProductFile(productId: number, fileId: number): Promise<void> {
    return apiClient.delete(`/products/${productId}/files/${fileId}`);
  },
};

// src/features/tracks/services/audioService.ts
import { apiClient } from '@/shared/api/client';

export interface Track {
  id: number;
  title: string;
  artist: string;
  album?: string;
  filename: string;
  duration_seconds: number;
  tags: string[];
  is_public: boolean;
  uploader_id: number;
  uploader_username: string;
  play_count: number;
  created_at: string;
}

interface CreateTrackRequest {
  title: string;
  artist: string;
  album?: string;
  tags: string[];
  is_public: boolean;
}

interface TracksResponse {
  tracks: Track[];
  total: number;
  page: number;
  limit: number;
}

export const audioService = {
  async getTracks(params?: {
    page?: number;
    limit?: number;
    search?: string;
    tags?: string[];
    artist?: string;
  }): Promise<TracksResponse> {
    return apiClient.get('/tracks', { params });
  },

  async getTrack(id: number): Promise<Track> {
    return apiClient.get(`/tracks/${id}`);
  },

  async uploadTrack(file: File, metadata: CreateTrackRequest, onProgress?: (progress: number) => void) {
    const formData = new FormData();
    formData.append('audio', file);
    Object.entries(metadata).forEach(([key, value]) => {
      if (value !== undefined) {
        formData.append(key, typeof value === 'object' ? JSON.stringify(value) : String(value));
      }
    });

    return apiClient.post('/tracks/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  },

  async updateTrack(id: number, data: Partial<CreateTrackRequest>): Promise<Track> {
    return apiClient.put(`/tracks/${id}`, data);
  },

  async deleteTrack(id: number): Promise<void> {
    return apiClient.delete(`/tracks/${id}`);
  },

  async getStreamUrl(trackId: number): Promise<{ url: string }> {
    return apiClient.get(`/tracks/${trackId}/stream`);
  },

  async incrementPlayCount(trackId: number): Promise<void> {
    return apiClient.post(`/tracks/${trackId}/play`);
  },
};

// src/features/chat/services/chatService.ts
import { apiClient } from '@/shared/api/client';

export interface Room {
  id: string;
  name: string;
  description?: string;
  created_by: number;
  created_at: string;
  is_private: boolean;
  member_count: number;
}

export interface Message {
  id: number;
  content: string;
  user_id: number;
  username: string;
  room_id?: string;
  recipient_id?: number;
  created_at: string;
}

interface CreateRoomRequest {
  name: string;
  description?: string;
  is_private?: boolean;
}

export const chatService = {
  async getRooms(): Promise<Room[]> {
    return apiClient.get('/chat/rooms');
  },

  async getRoom(id: string): Promise<Room> {
    return apiClient.get(`/chat/rooms/${id}`);
  },

  async createRoom(data: CreateRoomRequest): Promise<Room> {
    return apiClient.post('/chat/rooms', data);
  },

  async deleteRoom(id: string): Promise<void> {
    return apiClient.delete(`/chat/rooms/${id}`);
  },

  async getRoomMessages(roomId: string, params?: {
    limit?: number;
    before?: string;
  }): Promise<Message[]> {
    return apiClient.get(`/chat/rooms/${roomId}/messages`, { params });
  },

  async getDirectMessages(userId: number, params?: {
    limit?: number;
    before?: string;
  }): Promise<Message[]> {
    return apiClient.get(`/chat/dm/${userId}`, { params });
  },

  async markMessagesAsRead(userId: number): Promise<void> {
    return apiClient.post(`/chat/dm/${userId}/read`);
  },
};

// src/features/resources/services/resourceService.ts
import { apiClient } from '@/shared/api/client';

export interface Resource {
  id: number;
  title: string;
  description: string;
  filename: string;
  file_type: string;
  file_size: number;
  category: 'sample' | 'preset' | 'template' | 'document' | 'other';
  tags: string[];
  downloads: number;
  rating: number;
  is_public: boolean;
  uploader_id: number;
  uploader_username: string;
  created_at: string;
}

interface CreateResourceRequest {
  title: string;
  description: string;
  category: Resource['category'];
  tags: string[];
  is_public: boolean;
}

interface ResourcesResponse {
  resources: Resource[];
  total: number;
  page: number;
  limit: number;
}

export const resourceService = {
  async getResources(params?: {
    page?: number;
    limit?: number;
    search?: string;
    category?: string;
    tags?: string[];
  }): Promise<ResourcesResponse> {
    return apiClient.get('/resources', { params });
  },

  async getResource(id: number): Promise<Resource> {
    return apiClient.get(`/resources/${id}`);
  },

  async uploadResource(file: File, metadata: CreateResourceRequest, onProgress?: (progress: number) => void) {
    const formData = new FormData();
    formData.append('file', file);
    Object.entries(metadata).forEach(([key, value]) => {
      if (value !== undefined) {
        formData.append(key, typeof value === 'object' ? JSON.stringify(value) : String(value));
      }
    });

    return apiClient.post('/resources/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  },

  async downloadResource(id: number): Promise<Blob> {
    const response = await apiClient.get(`/resources/${id}/download`, {
      responseType: 'blob',
    });
    return response as unknown as Blob;
  },

  async rateResource(id: number, rating: number): Promise<void> {
    return apiClient.post(`/resources/${id}/rate`, { rating });
  },

  async deleteResource(id: number): Promise<void> {
    return apiClient.delete(`/resources/${id}`);
  },
};
```

---
