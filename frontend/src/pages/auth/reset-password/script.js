import { resetPassword } from '../../../utils/auth.js';
import notification from '../../../components/common/notification/script.js';

document.addEventListener('DOMContentLoaded', () => {
  const form = document.getElementById('reset-password-form');
  const passwordInput = document.getElementById('password');
  const confirmPasswordInput = document.getElementById('confirm-password');
  const submitButton = document.getElementById('submit-button');

  // Récupérer le token depuis l'URL
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');

  if (!token) {
    notification.error('Token de réinitialisation manquant');
    setTimeout(() => {
      window.location.href = '/src/pages/auth/login/index.html';
    }, 2000);
    return;
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const password = passwordInput.value;
    const confirmPassword = confirmPasswordInput.value;

    if (!password || !confirmPassword) {
      notification.error('Veuillez remplir tous les champs');
      return;
    }

    if (password !== confirmPassword) {
      notification.error('Les mots de passe ne correspondent pas');
      return;
    }

    if (password.length < 8) {
      notification.error('Le mot de passe doit contenir au moins 8 caractères');
      return;
    }

    try {
      submitButton.disabled = true;
      submitButton.textContent = 'Réinitialisation en cours...';

      await resetPassword(token, password);
      
      notification.success('Votre mot de passe a été réinitialisé avec succès');
      
      // Redirection vers la page de connexion après un court délai
      setTimeout(() => {
        window.location.href = '/src/pages/auth/login/index.html';
      }, 2000);
    } catch (error) {
      notification.error(error.message || 'Une erreur est survenue lors de la réinitialisation du mot de passe');
    } finally {
      submitButton.disabled = false;
      submitButton.textContent = 'Réinitialiser';
    }
  });
}); 