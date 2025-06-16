export const ENDPOINTS = {
  // Auth
  LOGIN: '/auth/login',
  REGISTER: '/auth/register',
  REFRESH: '/auth/refresh',
  LOGOUT: '/auth/logout',
  PROFILE: '/auth/me',
  
  // Users
  USERS: '/users',
  USER_BY_ID: (id: string | number) => `/users/${id}`,
  
  // Products
  PRODUCTS: '/products',
  PRODUCT_BY_ID: (id: string) => `/products/${id}`,
  MY_PRODUCTS: '/products/me',
  
  // Chat (routes legacy du backend)
  CHAT_ROOMS: '/chat/rooms',
  CHAT_ROOM_MESSAGES: (roomId: string) => `/chat/rooms/${roomId}/messages`,
  CHAT_DM: (userId: string | number) => `/chat/dm/${userId}`,
  CHAT_CREATE_ROOM: '/chat/rooms',
  CHAT_USERS: '/users', // Utilise l'endpoint users standard
  
  // Tracks
  TRACKS: '/tracks',
  TRACK_BY_ID: (id: string | number) => `/tracks/${id}`,
  TRACK_UPLOAD: '/tracks/upload',
  
  // Resources
  RESOURCES: '/resources',
  RESOURCE_BY_ID: (id: string | number) => `/resources/${id}`,
  RESOURCE_UPLOAD: '/resources/upload',
  
  // Admin
  ADMIN_STATS: '/admin/stats',
  ADMIN_USERS: '/admin/users',
  ADMIN_USER_BY_ID: (id: string | number) => `/admin/users/${id}`,
  ADMIN_TRACKS: '/admin/tracks',
  ADMIN_TRACK_BY_ID: (id: string | number) => `/admin/tracks/${id}`,
  ADMIN_RESOURCES: '/admin/resources',
  ADMIN_RESOURCE_BY_ID: (id: string | number) => `/admin/resources/${id}`,
} as const; 