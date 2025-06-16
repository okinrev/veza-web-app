import { apiClient } from './client';
import type { ChatRoom, ChatMessage, User } from './types';

export interface CreateRoomData {
  name: string;
  description?: string;
  is_private?: boolean;
}

export interface SendMessageData {
  content: string;
  room_id?: string;
  receiver_id?: string;
}

export interface MessageHistory {
  messages: ChatMessage[];
  total: number;
  page: number;
  limit: number;
}

export interface ChatUser {
  id: number;
  username: string;
  email: string;
  first_name?: string;
  last_name?: string;
}

export const chatApi = {
  // Rooms
  async getRooms(): Promise<ChatRoom[]> {
    return apiClient.get('/chat/rooms');
  },

  async createRoom(data: CreateRoomData): Promise<ChatRoom> {
    return apiClient.post('/chat/rooms', data);
  },

  async joinRoom(roomId: string): Promise<void> {
    return apiClient.post(`/chat/rooms/${roomId}/join`);
  },

  async leaveRoom(roomId: string): Promise<void> {
    return apiClient.post(`/chat/rooms/${roomId}/leave`);
  },

  async getRoomMembers(roomId: string): Promise<User[]> {
    return apiClient.get(`/chat/rooms/${roomId}/members`);
  },

  // Messages
  async getMessageHistory(roomId: string, page = 1, limit = 50): Promise<MessageHistory> {
    return apiClient.get(`/chat/rooms/${roomId}/messages?page=${page}&limit=${limit}`);
  },

  async sendMessage(data: SendMessageData): Promise<ChatMessage> {
    return apiClient.post('/chat/messages', data);
  },

  async getDirectMessages(userId: string, page = 1, limit = 50): Promise<MessageHistory> {
    return apiClient.get(`/chat/direct/${userId}?page=${page}&limit=${limit}`);
  },

  // Users
  async getUsers(): Promise<ChatUser[]> {
    try {
      const response = await apiClient.get('/users');
      return response.data;
    } catch (error) {
      console.error('Erreur lors du chargement des utilisateurs:', error);
      return [];
    }
  },

  async getOnlineUsers(): Promise<User[]> {
    return apiClient.get('/chat/users/online');
  },

  // Statistics
  async getChatStats(): Promise<{
    totalRooms: number;
    activeUsers: number;
    todayMessages: number;
  }> {
    return apiClient.get('/chat/stats');
  },

  // Récupérer un utilisateur spécifique
  async getUser(userId: number): Promise<ChatUser | null> {
    try {
      const response = await apiClient.get(`/users/${userId}`);
      return response.data;
    } catch (error) {
      console.error('Erreur lors du chargement de l\'utilisateur:', error);
      return null;
    }
  },

  // Créer un nouveau salon
  async createRoom(name: string, description?: string): Promise<ChatRoom | null> {
    try {
      // Pour l'instant, on simule la création
      const newRoom: ChatRoom = {
        id: Date.now(),
        name,
        description,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      return newRoom;
    } catch (error) {
      console.error('Erreur lors de la création du salon:', error);
      return null;
    }
  }
}; 