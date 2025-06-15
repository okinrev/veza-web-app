import { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';

const HomePage = lazy(() => import('@/app/HomePage'));
const NotFoundPage = lazy(() => import('@/app/NotFoundPage'));
const TracksPage = lazy(() => import('@/features/tracks/pages/TracksPage'));
const LoginPage = lazy(() => import('@/features/auth/pages/LoginPage'));

export const Router = () => (
  <Suspense fallback={<div>Chargement...</div>}>
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/tracks" element={<TracksPage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  </Suspense>
); 