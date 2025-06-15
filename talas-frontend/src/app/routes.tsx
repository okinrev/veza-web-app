import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Layout } from './Layout';
import LoginPage from '@/features/auth/pages/LoginPage';
import RegisterPage from '@/features/auth/pages/RegisterPage';
import { DashboardPage } from '@/app/pages/DashboardPage';
// Temporary placeholders for pages not yet implemented
const ChatPage = () => <div className="p-8"><h1 className="text-2xl font-bold">Chat - À venir</h1></div>;
const ProductsPage = () => <div className="p-8"><h1 className="text-2xl font-bold">Produits - À venir</h1></div>;
const TracksPage = () => <div className="p-8"><h1 className="text-2xl font-bold">Pistes - À venir</h1></div>;
const ResourcesPage = () => <div className="p-8"><h1 className="text-2xl font-bold">Ressources - À venir</h1></div>;
const ProfilePage = () => <div className="p-8"><h1 className="text-2xl font-bold">Profil - À venir</h1></div>;
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