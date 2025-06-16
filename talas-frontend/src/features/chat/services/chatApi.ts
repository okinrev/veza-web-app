import { apiClient } from '../../../shared/api/client';
import { ENDPOINTS } from '../../../shared/api/endpoints';

// Types pour les réponses API du backend
interface BackendRoom {
  id: number;
  name: string;
  description?: string;
  is_private: boolean;
  member_count?: number;
  created_at: string;
}

interface BackendMessage {
  id: number;
  content: string;
  from_user: number;
  user_id?: number;
  username?: string;
  from_username?: string;
  timestamp?: string;
  created_at?: string;
  room?: string;
  to_user?: number;
}

// Type pour les utilisateurs du backend Go (avec champs nullable)
interface BackendUser {
  id: number;
  username: string;
  email: string;
  first_name: {
    String: string;
    Valid: boolean;
  };
  last_name: {
    String: string;
    Valid: boolean;
  };
  bio: {
    String: string;
    Valid: boolean;
  };
  avatar: {
    String: string;
    Valid: boolean;
  };
  role: string;
  is_active: boolean;
  is_verified: boolean;
  last_login_at: {
    Time: string;
    Valid: boolean;
  };
  created_at: string;
  updated_at: string;
}

// Type pour les réponses API
interface ApiResponse<T> {
  data: T;
  message?: string;
  success?: boolean;
}

// Types pour les salons
export interface ChatRoom {
  id: number;
  name: string;
  description?: string;
  is_private: boolean;
  creator_id?: number;
  member_count?: number;
  created_at: string;
  updated_at?: string;
}

// Types pour les messages
export interface ChatMessage {
  id: number;
  room_id?: string;
  user_id: number;
  username: string;
  content: string;
  timestamp: string;
  message_type?: 'text' | 'image' | 'file';
}

// Types pour les messages privés
export interface DirectMessage {
  id: number;
  from_user_id: number;
  to_user_id: number;
  from_username: string;
  to_username: string;
  content: string;
  timestamp: string;
  is_read: boolean;
}

// Types pour les utilisateurs (pour les DM)
export interface ChatUser {
  id: number;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
  avatar_url?: string;
  isOnline: boolean;
  lastSeen: string;
  unreadCount: number;
}

export const chatApi = {
  // Récupérer la liste des salons
  async getRooms(): Promise<ChatRoom[]> {
    try {
      const response = await apiClient.get(ENDPOINTS.CHAT_ROOMS) as { data: { rooms: BackendRoom[] } };
      console.log('[Chat API] Réponse salons:', response.data);
      
      // Le backend retourne { rooms: [...] }
      const rooms = response.data.rooms || [];
      return rooms.map((room: BackendRoom) => ({
        id: room.id,
        name: room.name,
        description: room.description || '',
        is_private: room.is_private || false,
        member_count: room.member_count || 0,
        created_at: room.created_at || new Date().toISOString(),
      }));
    } catch (error) {
      console.error('Erreur lors de la récupération des salons:', error);
      // Fallback avec des salons de démonstration
      return [
        {
          id: 1,
          name: 'general',
          description: 'Salon général pour tous',
          is_private: false,
          member_count: 5,
          created_at: new Date().toISOString(),
        },
        {
          id: 2,
          name: 'random',
          description: 'Discussions libres',
          is_private: false,
          member_count: 3,
          created_at: new Date().toISOString(),
        }
      ];
    }
  },

  // Créer un nouveau salon
  async createRoom(name: string, description?: string, isPrivate: boolean = false): Promise<ChatRoom> {
    try {
      const response = await apiClient.post(ENDPOINTS.CHAT_CREATE_ROOM, {
        name,
        description,
        is_private: isPrivate
      }) as { data: BackendRoom };
      
      const room = response.data;
      return {
        id: room.id,
        name: room.name,
        description: room.description || '',
        is_private: room.is_private,
        member_count: room.member_count || 1,
        created_at: room.created_at,
      };
    } catch (error) {
      console.error('Erreur lors de la création du salon:', error);
      throw error;
    }
  },

  // Récupérer les messages d'un salon
  async getRoomMessages(roomId: string, limit: number = 50): Promise<ChatMessage[]> {
    try {
      const response = await apiClient.get(ENDPOINTS.CHAT_ROOM_MESSAGES(roomId), {
        params: { limit }
      }) as { data: { messages: BackendMessage[] } };
      
      const messages = response.data.messages || [];
      return messages.map((msg: BackendMessage) => ({
        id: msg.id,
        room_id: roomId,
        user_id: msg.from_user || msg.user_id || 0,
        username: msg.username || msg.from_username || 'Utilisateur',
        content: msg.content,
        timestamp: msg.timestamp || msg.created_at || new Date().toISOString(),
        message_type: 'text' as const,
      }));
    } catch (error) {
      console.error('Erreur lors de la récupération des messages du salon:', error);
      return [];
    }
  },

  // Récupérer les messages privés avec un utilisateur
  async getDirectMessages(userId: number, limit: number = 50): Promise<DirectMessage[]> {
    try {
      const response = await apiClient.get(ENDPOINTS.CHAT_DM(userId), {
        params: { limit }
      }) as { data: { messages: BackendMessage[] } };
      
      const messages = response.data.messages || [];
      return messages.map((msg: BackendMessage) => ({
        id: msg.id,
        from_user_id: msg.from_user,
        to_user_id: msg.to_user || userId,
        from_username: msg.username || msg.from_username || 'Utilisateur',
        to_username: `User ${msg.to_user || userId}`,
        content: msg.content,
        timestamp: msg.timestamp || msg.created_at || new Date().toISOString(),
        is_read: false,
      }));
    } catch (error) {
      console.error('Erreur lors de la récupération des messages privés:', error);
      // Fallback avec des messages de démonstration
      return [
        {
          id: Date.now(),
          from_user_id: userId === 1 ? 2 : 1,
          to_user_id: userId,
          from_username: userId === 1 ? 'alice' : 'vous',
          to_username: userId === 1 ? 'vous' : 'alice',
          content: `Salut ! C'est le début de votre conversation privée.`,
          timestamp: new Date(Date.now() - 3600000).toISOString(), // Il y a 1h
          is_read: false,
        },
        {
          id: Date.now() + 1,
          from_user_id: userId === 1 ? 2 : 1,
          to_user_id: userId,
          from_username: userId === 1 ? 'alice' : 'vous',
          to_username: userId === 1 ? 'vous' : 'alice',
          content: `Comment ça va ?`,
          timestamp: new Date(Date.now() - 1800000).toISOString(), // Il y a 30min
          is_read: false,
        }
      ];
    }
  },

  // Récupérer la liste des utilisateurs pour les DM
  async getUsers(): Promise<ChatUser[]> {
    try {
      const response = await apiClient.get(ENDPOINTS.CHAT_USERS) as { data: { data: BackendUser[] } };
      
      // Le backend retourne { success: true, data: [...], message: "...", meta: {...} }
      const users = response.data.data || [];
      return users.map((user: BackendUser) => ({
        id: user.id,
        username: user.username,
        email: user.email || '',
        // Gérer les champs nullable du backend Go
        first_name: user.first_name?.String || '',
        last_name: user.last_name?.String || '',
        avatar_url: user.avatar?.String || '',
        isOnline: user.is_active || false,
        lastSeen: user.last_login_at?.Time || new Date().toISOString(),
        unreadCount: 0,
      }));
    } catch (error) {
      console.error('Erreur lors de la récupération des utilisateurs:', error);
      // Fallback avec des utilisateurs de démonstration
      return [
        {
          id: 1,
          username: 'alice',
          email: 'alice@example.com',
          first_name: 'Alice',
          last_name: 'Martin',
          avatar_url: '',
          isOnline: true,
          lastSeen: new Date().toISOString(),
          unreadCount: 2,
        },
        {
          id: 2,
          username: 'bob',
          email: 'bob@example.com',
          first_name: 'Bob',
          last_name: 'Dupont',
          avatar_url: '',
          isOnline: false,
          lastSeen: new Date(Date.now() - 3600000).toISOString(),
          unreadCount: 0,
        },
        {
          id: 3,
          username: 'charlie',
          email: 'charlie@example.com',
          first_name: 'Charlie',
          last_name: 'Durand',
          avatar_url: '',
          isOnline: true,
          lastSeen: new Date().toISOString(),
          unreadCount: 1,
        }
      ];
    }
  },

  // Marquer les messages comme lus
  async markMessagesAsRead(userId: number): Promise<void> {
    try {
      console.log('Marquer messages comme lus pour utilisateur:', userId);
      // Endpoint pour marquer comme lu (à implémenter côté backend)
      // await apiClient.post(`/chat/dm/${userId}/read`);
    } catch (error) {
      console.error('Erreur lors du marquage des messages comme lus:', error);
    }
  },
}; 