import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { Message, Room, Conversation } from '../types';

interface ChatState {
  // State
  currentUserId: number | null;
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
  setCurrentUserId: (userId: number) => void;
  setRooms: (rooms: Room[]) => void;
  joinRoom: (roomId: string) => void;
  leaveRoom: () => void;
  addMessage: (roomId: string, message: Message) => void;
  addDirectMessage: (userId: number, message: Message) => void;
  setOnlineUsers: (users: number[]) => void;
  setTypingUser: (roomId: string, userId: number, isTyping: boolean) => void;
  setConnectionStatus: (status: boolean) => void;
}

export const useChatStore = create<ChatState>()(
  devtools(
    immer((set) => ({
      // Initial state
      currentUserId: null,
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
      setCurrentUserId: (userId) =>
        set((state) => {
          state.currentUserId = userId;
        }),
        
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