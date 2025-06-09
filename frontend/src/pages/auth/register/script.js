import { register } from '../../../utils/auth.js';
import notification from '../../../components/common/notification/script.js';

class RegisterPage {
    constructor() {
        this.form = document.getElementById('register-form');
        this.usernameInput = document.getElementById('username');
        this.emailInput = document.getElementById('email');
        this.passwordInput = document.getElementById('password');
        this.confirmPasswordInput = document.getElementById('confirm-password');

        this.init();
    }

    init() {
        this.setupEventListeners();
    }

    setupEventListeners() {
        this.form.addEventListener('submit', this.handleSubmit.bind(this));
    }

    validateForm() {
        if (this.passwordInput.value !== this.confirmPasswordInput.value) {
            notification.error('Les mots de passe ne correspondent pas');
            return false;
        }

        if (this.passwordInput.value.length < 8) {
            notification.error('Le mot de passe doit contenir au moins 8 caractères');
            return false;
        }

        return true;
    }

    async handleSubmit(event) {
        event.preventDefault();

        if (!this.validateForm()) {
            return;
        }

        const userData = {
            username: this.usernameInput.value,
            email: this.emailInput.value,
            password: this.passwordInput.value
        };

        try {
            const response = await register(userData);
            if (response.success) {
                notification.success('Inscription réussie !');
                window.location.href = '/login';
            } else {
                notification.error(response.message || 'Erreur lors de l\'inscription');
            }
        } catch (error) {
            console.error('Registration error:', error);
            notification.error('Une erreur est survenue lors de l\'inscription');
        }
    }
}

// Initialize the page
new RegisterPage(); 