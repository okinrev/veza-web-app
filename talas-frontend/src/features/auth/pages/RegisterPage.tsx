
import { AuthGuard } from '../components/AuthGuard';
import { RegisterForm } from '../components/RegisterForm';

const RegisterPage = () => {
  return (
    <AuthGuard requireAuth={false}>
      <RegisterForm />
    </AuthGuard>
  );
};

export default RegisterPage; 