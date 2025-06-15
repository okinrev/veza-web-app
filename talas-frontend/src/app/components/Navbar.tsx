import { Link } from 'react-router-dom';
import { useAuth } from '../providers/AuthProvider';
import { useTheme } from '../providers/ThemeProvider';

export default function Navbar() {
  const { isAuthenticated, user, logout } = useAuth();
  const { theme, setTheme } = useTheme();

  return (
    <nav className="bg-card border-b">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <Link to="/" className="text-xl font-bold">
              Talas
            </Link>
          </div>

          <div className="flex items-center space-x-4">
            {isAuthenticated ? (
              <>
                <Link
                  to="/chat"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Chat
                </Link>
                <Link
                  to="/products"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Produits
                </Link>
                <Link
                  to="/tracks"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Pistes
                </Link>
                <Link
                  to="/resources"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Ressources
                </Link>
                <Link
                  to="/marketplace"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Marketplace
                </Link>
                <Link
                  to="/profile"
                  className="text-foreground/60 hover:text-foreground"
                >
                  {user?.email}
                </Link>
                <button
                  onClick={() => logout()}
                  className="text-foreground/60 hover:text-foreground"
                >
                  Déconnexion
                </button>
              </>
            ) : (
              <>
                <Link
                  to="/auth/login"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Connexion
                </Link>
                <Link
                  to="/auth/register"
                  className="text-foreground/60 hover:text-foreground"
                >
                  Inscription
                </Link>
              </>
            )}

            <button
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="p-2 rounded-md hover:bg-accent"
            >
              {theme === 'dark' ? '☀️' : '🌙'}
            </button>
          </div>
        </div>
      </div>
    </nav>
  );
} 