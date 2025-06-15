export const ENDPOINTS = {
  // Auth
  LOGIN: '/auth/login',
  REGISTER: '/auth/register',
  REFRESH: '/auth/refresh',
  LOGOUT: '/auth/logout',
  PROFILE: '/auth/profile',
  
  // Users
  USERS: '/users',
  USER_BY_ID: (id: string) => `/users/${id}`,
  
  // Products
  PRODUCTS: '/products',
  PRODUCT_BY_ID: (id: string) => `/products/${id}`,
  MY_PRODUCTS: '/products/me',
  
  // Chat
  ROOMS: '/chat/rooms',
  ROOM_BY_ID: (id: string) => `/chat/rooms/${id}`,
  MESSAGES: (roomId: string) => `/chat/rooms/${roomId}/messages`,
  DIRECT_MESSAGES: '/chat/direct',
  
  // Tracks
  TRACKS: '/tracks',
  TRACK_BY_ID: (id: string) => `/tracks/${id}`,
  TRACK_UPLOAD: '/tracks/upload',
  
  // Resources
  RESOURCES: '/resources',
  RESOURCE_BY_ID: (id: string) => `/resources/${id}`,
  RESOURCE_UPLOAD: '/resources/upload',
  
  // Admin
  ADMIN_STATS: '/admin/stats',
  ADMIN_USERS: '/admin/users',
} as const; 