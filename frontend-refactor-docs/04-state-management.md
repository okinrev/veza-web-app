## 7. Gestion d'état et données

### 7.1 Store Zustand principal

```
// src/shared/store/useAppStore.ts
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

interface AppState {
  // UI State
  sidebarOpen: boolean;
  theme: 'light' | 'dark' | 'system';
  notifications: Notification[];
  
  // Actions
  toggleSidebar: () => void;
  setTheme: (theme: AppState['theme']) => void;
  addNotification: (notification: Omit<Notification, 'id'>) => void;
  removeNotification: (id: string) => void;
}

interface Notification {
  id: string;
  title: string;
  description?: string;
  type: 'info' | 'success' | 'warning' | 'error';
  timestamp: Date;
}

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      immer((set) => ({
        // Initial state
        sidebarOpen: true,
        theme: 'system',
        notifications: [],
        
        // Actions
        toggleSidebar: () =>
          set((state) => {
            state.sidebarOpen = !state.sidebarOpen;
          }),
          
        setTheme: (theme) =>
          set((state) => {
            state.theme = theme;
          }),
          
        addNotification: (notification) =>
          set((state) => {
            state.notifications.push({
              ...notification,
              id: Date.now().toString(),
              timestamp: new Date(),
            });
          }),
          
        removeNotification: (id) =>
          set((state) => {
            state.notifications = state.notifications.filter(
              (n) => n.id !== id
            );
          }),
      })),
      {
        name: 'app-store',
        partialize: (state) => ({ theme: state.theme }),
      }
    )
  )
);

// src/features/auth/store/authStore.ts
import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { authService } from '../services/authService';

interface User {
  id: number;
  email: string;
  username: string;
  role: 'user' | 'admin' | 'moderator';
  avatar?: string;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
  clearError: () => void;
}

interface RegisterData {
  email: string;
  username: string;
  password: string;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      immer((set, get) => ({
        // Initial state
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
        
        // Actions
        login: async (email, password) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });
          
          try {
            const response = await authService.login({ email, password });
            
            set((state) => {
              state.user = response.user;
              state.accessToken = response.accessToken;
              state.refreshToken = response.refreshToken;
              state.isAuthenticated = true;
              state.isLoading = false;
            });
          } catch (error) {
            set((state) => {
              state.error = error.message;
              state.isLoading = false;
            });
            throw error;
          }
        },
        
        register: async (data) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });
          
          try {
            const response = await authService.register(data);
            
            set((state) => {
              state.user = response.user;
              state.accessToken = response.accessToken;
              state.refreshToken = response.refreshToken;
              state.isAuthenticated = true;
              state.isLoading = false;
            });
          } catch (error) {
            set((state) => {
              state.error = error.message;
              state.isLoading = false;
            });
            throw error;
          }
        },
        
        logout: async () => {
          try {
            await authService.logout();
          } finally {
            set((state) => {
              state.user = null;
              state.accessToken = null;
              state.refreshToken = null;
              state.isAuthenticated = false;
              state.error = null;
            });
          }
        },
        
        refreshAuth: async () => {
          const refreshToken = get().refreshToken;
          if (!refreshToken) throw new Error('No refresh token');
          
          try {
            const response = await authService.refresh(refreshToken);
            
            set((state) => {
              state.accessToken = response.accessToken;
              state.refreshToken = response.refreshToken;
            });
          } catch (error) {
            // Si le refresh échoue, déconnecter l'utilisateur
            get().logout();
            throw error;
          }
        },
        
        clearError: () =>
          set((state) => {
            state.error = null;
          }),
      })),
      {
        name: 'auth-store',
        partialize: (state) => ({
          accessToken: state.accessToken,
          refreshToken: state.refreshToken,
          user: state.user,
        }),
      }
    )
  )
);

// src/features/chat/store/chatStore.ts
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

interface Message {
  id: string;
  content: string;
  userId: number;
  username: string;
  timestamp: Date;
  type: 'message' | 'system';
}

interface Room {
  id: string;
  name: string;
  description?: string;
  userCount: number;
  lastMessage?: Message;
}

interface ChatState {
  // State
  rooms: Room[];
  currentRoom: string | null;
  messages: Record<string, Message[]>; // roomId -> messages
  onlineUsers: number[];
  typingUsers: Record<string, number[]>; // roomId -> userIds
  isConnected: boolean;
  
  // Direct Messages
  conversations: Conversation[];
  currentConversation: number | null;
  directMessages: Record<number, Message[]>; // userId -> messages
  
  // Actions
  setRooms: (rooms: Room[]) => void;
  joinRoom: (roomId: string) => void;
  leaveRoom: () => void;
  addMessage: (roomId: string, message: Message) => void;
  addDirectMessage: (userId: number, message: Message) => void;
  setOnlineUsers: (users: number[]) => void;
  setTypingUser: (roomId: string, userId: number, isTyping: boolean) => void;
  setConnectionStatus: (status: boolean) => void;
}

interface Conversation {
  userId: number;
  username: string;
  avatar?: string;
  lastMessage?: Message;
  unreadCount: number;
}

export const useChatStore = create<ChatState>()(
  devtools(
    immer((set) => ({
      // Initial state
      rooms: [],
      currentRoom: null,
      messages: {},
      onlineUsers: [],
      typingUsers: {},
      isConnected: false,
      conversations: [],
      currentConversation: null,
      directMessages: {},
      
      // Actions
      setRooms: (rooms) =>
        set((state) => {
          state.rooms = rooms;
        }),
        
      joinRoom: (roomId) =>
        set((state) => {
          state.currentRoom = roomId;
          if (!state.messages[roomId]) {
            state.messages[roomId] = [];
          }
        }),
        
      leaveRoom: () =>
        set((state) => {
          state.currentRoom = null;
        }),
        
      addMessage: (roomId, message) =>
        set((state) => {
          if (!state.messages[roomId]) {
            state.messages[roomId] = [];
          }
          state.messages[roomId].push(message);
        }),
        
      addDirectMessage: (userId, message) =>
        set((state) => {
          if (!state.directMessages[userId]) {
            state.directMessages[userId] = [];
          }
          state.directMessages[userId].push(message);
        }),
        
      setOnlineUsers: (users) =>
        set((state) => {
          state.onlineUsers = users;
        }),
        
      setTypingUser: (roomId, userId, isTyping) =>
        set((state) => {
          if (!state.typingUsers[roomId]) {
            state.typingUsers[roomId] = [];
          }
          
          if (isTyping) {
            if (!state.typingUsers[roomId].includes(userId)) {
              state.typingUsers[roomId].push(userId);
            }
          } else {
            state.typingUsers[roomId] = state.typingUsers[roomId].filter(
              (id) => id !== userId
            );
          }
        }),
        
      setConnectionStatus: (status) =>
        set((state) => {
          state.isConnected = status;
        }),
    }))
  )
);
```

---
