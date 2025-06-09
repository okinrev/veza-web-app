import { login } from '../../../utils/auth.js';
import notification from '../../../components/common/notification/script.js';

document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('login-form');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');
  const submitButton = document.getElementById('submit-button');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const email = emailInput.value.trim();
    const password = passwordInput.value;

    if (!email || !password) {
      notification.error('Veuillez remplir tous les champs');
      return;
    }

    try {
      submitButton.disabled = true;
      submitButton.textContent = 'Connexion en cours...';

      const response = await login(email, password);
      
      notification.success('Connexion réussie !');
      
      // Redirection vers le tableau de bord après un court délai
      setTimeout(() => {
        window.location.href = '/src/pages/dashboard.html';
      }, 1000);
    } catch (error) {
      notification.error(error.message || 'Une erreur est survenue lors de la connexion');
    } finally {
      submitButton.disabled = false;
      submitButton.textContent = 'Se connecter';
    }
  });
}); 