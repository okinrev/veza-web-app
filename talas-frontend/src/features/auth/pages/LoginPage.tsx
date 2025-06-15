
import { AuthGuard } from '../components/AuthGuard';
import { LoginForm } from '../components/LoginForm';

const LoginPage = () => {
  return (
    <AuthGuard requireAuth={false}>
      <LoginForm />
    </AuthGuard>
  );
};

export default LoginPage; 