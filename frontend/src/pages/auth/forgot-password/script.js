import { forgotPassword } from '../../../utils/auth.js';
import notification from '../../../components/common/notification/script.js';

document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('forgot-password-form');
  const emailInput = document.getElementById('email');
  const submitButton = document.getElementById('submit-button');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const email = emailInput.value.trim();

    if (!email) {
      notification.error('Veuillez entrer votre adresse email');
      return;
    }

    try {
      submitButton.disabled = true;
      submitButton.textContent = 'Envoi en cours...';

      await forgotPassword(email);
      
      notification.success('Un email de réinitialisation a été envoyé à votre adresse');
      
      // Redirection vers la page de connexion après un court délai
      setTimeout(() => {
        window.location.href = '/src/pages/auth/login/index.html';
      }, 2000);
    } catch (error) {
      notification.error(error.message || 'Une erreur est survenue lors de l\'envoi de l\'email');
    } finally {
      submitButton.disabled = false;
      submitButton.textContent = 'Envoyer';
    }
  });
}); 