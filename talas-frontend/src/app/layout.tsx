import { Outlet } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ToastContainer } from '@/shared/components/ui/ToastContainer';
import { useSessionTimeout } from '@/features/auth/hooks/useSessionTimeout';

const queryClient = new QueryClient();

export const Layout = () => {
  useSessionTimeout();

  return (
    <html lang="fr">
      <body>
        <QueryClientProvider client={queryClient}>
          <div className="min-h-screen bg-background">
            <Outlet />
            <ToastContainer />
          </div>
        </QueryClientProvider>
      </body>
    </html>
  );
}; 