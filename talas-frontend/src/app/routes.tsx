import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Layout } from './Layout';
import { LoginPage } from '@/features/auth/LoginPage';
import { RegisterPage } from '@/features/auth/RegisterPage';
import { DashboardPage } from '@/features/dashboard/DashboardPage';
import { ChatPage } from '@/features/chat/ChatPage';
import { ProductsPage } from '@/features/products/ProductsPage';
import { TracksPage } from '@/features/tracks/TracksPage';
import { ResourcesPage } from '@/features/resources/ResourcesPage';
import { ProfilePage } from '@/features/profile/ProfilePage';
import { ProtectedRoute } from './ProtectedRoute';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />
      },
      {
        path: 'login',
        element: <LoginPage />,
      },
      {
        path: 'register',
        element: <RegisterPage />,
      },
      {
        path: 'dashboard',
        element: (
          <ProtectedRoute>
            <DashboardPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'chat',
        element: (
          <ProtectedRoute>
            <ChatPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'products',
        element: (
          <ProtectedRoute>
            <ProductsPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'tracks',
        element: (
          <ProtectedRoute>
            <TracksPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'resources',
        element: (
          <ProtectedRoute>
            <ResourcesPage />
          </ProtectedRoute>
        ),
      },
      {
        path: 'profile',
        element: (
          <ProtectedRoute>
            <ProfilePage />
          </ProtectedRoute>
        ),
      },
    ],
  },
]); 