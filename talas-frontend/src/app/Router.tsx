import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Loader2 } from 'lucide-react';

// Lazy loading des pages
const HomePage = lazy(() => import('@/app/HomePage'));
const DashboardPage = lazy(() => import('@/app/pages/DashboardPage'));
const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage'));
const RegisterPage = lazy(() => import('@/features/auth/pages/RegisterPage'));
const TracksPage = lazy(() => import('@/features/tracks/pages/TracksPage'));
const NotFoundPage = lazy(() => import('@/app/NotFoundPage'));

// Composant de fallback pour le loading
const LoadingFallback = () => (
  <div className="min-h-screen flex items-center justify-center">
    <div className="flex flex-col items-center space-y-4">
      <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      <p className="text-gray-600">Chargement...</p>
    </div>
  </div>
);

export const Router = () => (
  <Suspense fallback={<LoadingFallback />}>
    <Routes>
      {/* Redirect root to dashboard */}
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      
      {/* Pages publiques */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      
      {/* Pages protégées */}
      <Route path="/dashboard" element={<DashboardPage />} />
      <Route path="/tracks" element={<TracksPage />} />
      <Route path="/home" element={<HomePage />} />
      
      {/* 404 */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  </Suspense>
); 