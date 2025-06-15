export interface Message {
  id: string;
  content: string;
  userId: number;
  username: string;
  timestamp: Date;
  type: 'message' | 'system';
}

export interface Room {
  id: string;
  name: string;
  description?: string;
  userCount: number;
  lastMessage?: Message;
}

export interface Conversation {
  userId: number;
  username: string;
  avatar?: string;
  lastMessage?: Message;
  unreadCount: number;
} 