export interface ApiResponse<T = any> {
  success: boolean;
  data: T;
  message?: string;
  errors?: string[];
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface User {
  id: number;
  username?: string;
  first_name: any; // sql.NullString from Go backend
  last_name: any;  // sql.NullString from Go backend
  email: string;
  role?: string;
  avatar?: any;    // sql.NullString from Go backend
  bio?: any;       // sql.NullString from Go backend
  is_active?: boolean;
  is_verified?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  category: string;
  tags: string[];
  images: string[];
  ownerId: string;
  owner: User;
  createdAt: string;
  updatedAt: string;
}

export interface ChatRoom {
  id: string;
  name: string;
  description?: string;
  isPrivate: boolean;
  members: User[];
  createdAt: string;
  updatedAt: string;
}

export interface ChatMessage {
  id: string;
  content: string;
  type: 'text' | 'image' | 'file';
  senderId: string;
  sender: User;
  roomId?: string;
  receiverId?: string;
  createdAt: string;
}

export interface Track {
  id: string;
  title: string;
  artist: string;
  genre?: string;
  duration: number;
  fileUrl: string;
  coverUrl?: string;
  tags: string[];
  ownerId: string;
  owner: User;
  createdAt: string;
  updatedAt: string;
}

export interface Resource {
  id: string;
  name: string;
  description?: string;
  fileUrl: string;
  fileType: string;
  fileSize: number;
  tags: string[];
  ownerId: string;
  owner: User;
  createdAt: string;
  updatedAt: string;
} 