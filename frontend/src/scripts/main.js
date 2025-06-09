// Configuration globale
const config = {
    apiUrl: '/api',
    wsUrl: 'ws://' + window.location.host + '/ws',
    debug: process.env.NODE_ENV === 'development'
};

// Gestionnaire d'état global
const state = {
    user: null,
    notifications: [],
    theme: localStorage.getItem('theme') || 'light',
    language: localStorage.getItem('language') || 'fr'
};

// Fonctions utilitaires
const utils = {
    // Gestion des dates
    formatDate: (date, format = 'DD/MM/YYYY') => {
        const d = new Date(date);
        const day = String(d.getDate()).padStart(2, '0');
        const month = String(d.getMonth() + 1).padStart(2, '0');
        const year = d.getFullYear();
        const hours = String(d.getHours()).padStart(2, '0');
        const minutes = String(d.getMinutes()).padStart(2, '0');
        
        return format
            .replace('DD', day)
            .replace('MM', month)
            .replace('YYYY', year)
            .replace('HH', hours)
            .replace('mm', minutes);
    },

    // Gestion des nombres
    formatNumber: (number, decimals = 2) => {
        return new Intl.NumberFormat(state.language, {
            minimumFractionDigits: decimals,
            maximumFractionDigits: decimals
        }).format(number);
    },

    // Gestion des devises
    formatCurrency: (amount, currency = 'EUR') => {
        return new Intl.NumberFormat(state.language, {
            style: 'currency',
            currency: currency
        }).format(amount);
    },

    // Validation des formulaires
    validateForm: (formData, rules) => {
        const errors = {};
        
        for (const [field, value] of formData.entries()) {
            if (rules[field]) {
                const fieldRules = rules[field];
                
                if (fieldRules.required && !value) {
                    errors[field] = 'Ce champ est requis';
                }
                
                if (fieldRules.minLength && value.length < fieldRules.minLength) {
                    errors[field] = `Minimum ${fieldRules.minLength} caractères`;
                }
                
                if (fieldRules.maxLength && value.length > fieldRules.maxLength) {
                    errors[field] = `Maximum ${fieldRules.maxLength} caractères`;
                }
                
                if (fieldRules.pattern && !fieldRules.pattern.test(value)) {
                    errors[field] = fieldRules.message || 'Format invalide';
                }
            }
        }
        
        return errors;
    },

    // Gestion des erreurs
    handleError: (error) => {
        if (config.debug) {
            console.error(error);
        }
        
        const message = error.response?.data?.message || error.message || 'Une erreur est survenue';
        showNotification(message, 'error');
    }
};

// Gestionnaire de notifications
const notifications = {
    show: (message, type = 'info', duration = 5000) => {
        const notification = {
            id: Date.now(),
            message,
            type,
            duration
        };
        
        state.notifications.push(notification);
        renderNotification(notification);
        
        if (duration > 0) {
            setTimeout(() => {
                notifications.remove(notification.id);
            }, duration);
        }
        
        return notification.id;
    },
    
    remove: (id) => {
        const index = state.notifications.findIndex(n => n.id === id);
        if (index !== -1) {
            state.notifications.splice(index, 1);
            const element = document.getElementById(`notification-${id}`);
            if (element) {
                element.remove();
            }
        }
    },
    
    success: (message, duration) => notifications.show(message, 'success', duration),
    error: (message, duration) => notifications.show(message, 'error', duration),
    warning: (message, duration) => notifications.show(message, 'warning', duration),
    info: (message, duration) => notifications.show(message, 'info', duration)
};

// Gestionnaire de thème
const theme = {
    toggle: () => {
        state.theme = state.theme === 'light' ? 'dark' : 'light';
        localStorage.setItem('theme', state.theme);
        document.documentElement.setAttribute('data-theme', state.theme);
    },
    
    init: () => {
        document.documentElement.setAttribute('data-theme', state.theme);
    }
};

// Gestionnaire de langue
const i18n = {
    setLanguage: (lang) => {
        state.language = lang;
        localStorage.setItem('language', lang);
        document.documentElement.setAttribute('lang', lang);
    },
    
    init: () => {
        document.documentElement.setAttribute('lang', state.language);
    }
};

// Gestionnaire d'authentification
const auth = {
    login: async (credentials) => {
        try {
            const response = await fetch(`${config.apiUrl}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(credentials)
            });
            
            if (!response.ok) {
                throw new Error('Identifiants invalides');
            }
            
            const data = await response.json();
            state.user = data.user;
            localStorage.setItem('token', data.token);
            
            return data;
        } catch (error) {
            utils.handleError(error);
            throw error;
        }
    },
    
    logout: () => {
        state.user = null;
        localStorage.removeItem('token');
        window.location.href = '/login';
    },
    
    checkAuth: () => {
        const token = localStorage.getItem('token');
        if (!token) {
            return false;
        }
        
        // Vérifier la validité du token
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            if (payload.exp * 1000 < Date.now()) {
                auth.logout();
                return false;
            }
            return true;
        } catch {
            auth.logout();
            return false;
        }
    }
};

// Gestionnaire de WebSocket
const ws = {
    connection: null,
    
    connect: () => {
        if (ws.connection) {
            return;
        }
        
        ws.connection = new WebSocket(config.wsUrl);
        
        ws.connection.onopen = () => {
            if (config.debug) {
                console.log('WebSocket connecté');
            }
        };
        
        ws.connection.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                handleWebSocketMessage(data);
            } catch (error) {
                utils.handleError(error);
            }
        };
        
        ws.connection.onclose = () => {
            if (config.debug) {
                console.log('WebSocket déconnecté');
            }
            setTimeout(ws.connect, 5000);
        };
        
        ws.connection.onerror = (error) => {
            utils.handleError(error);
        };
    },
    
    send: (message) => {
        if (ws.connection && ws.connection.readyState === WebSocket.OPEN) {
            ws.connection.send(JSON.stringify(message));
        }
    }
};

// Gestionnaire de messages WebSocket
function handleWebSocketMessage(data) {
    switch (data.type) {
        case 'notification':
            notifications.show(data.message, data.notificationType);
            break;
            
        case 'chat':
            // Gérer les messages de chat
            break;
            
        case 'status':
            // Gérer les mises à jour de statut
            break;
            
        default:
            if (config.debug) {
                console.log('Message WebSocket non géré:', data);
            }
    }
}

// Initialisation
document.addEventListener('DOMContentLoaded', () => {
    theme.init();
    i18n.init();
    
    if (auth.checkAuth()) {
        ws.connect();
    }
    
    // Initialiser les gestionnaires d'événements globaux
    initializeEventHandlers();
});

// Gestionnaires d'événements globaux
function initializeEventHandlers() {
    // Gestionnaire de formulaire global
    document.addEventListener('submit', async (event) => {
        const form = event.target;
        if (form.dataset.validate === 'true') {
            event.preventDefault();
            
            const formData = new FormData(form);
            const errors = utils.validateForm(formData, JSON.parse(form.dataset.rules || '{}'));
            
            if (Object.keys(errors).length > 0) {
                Object.entries(errors).forEach(([field, message]) => {
                    const input = form.querySelector(`[name="${field}"]`);
                    if (input) {
                        input.classList.add('is-invalid');
                        const feedback = input.nextElementSibling;
                        if (feedback && feedback.classList.contains('invalid-feedback')) {
                            feedback.textContent = message;
                        }
                    }
                });
                return;
            }
            
            try {
                const response = await fetch(form.action, {
                    method: form.method,
                    body: formData
                });
                
                if (!response.ok) {
                    throw new Error('Erreur lors de la soumission du formulaire');
                }
                
                const data = await response.json();
                notifications.success(data.message || 'Opération réussie');
                
                if (form.dataset.redirect) {
                    window.location.href = form.dataset.redirect;
                }
            } catch (error) {
                utils.handleError(error);
            }
        }
    });
    
    // Gestionnaire de validation en temps réel
    document.addEventListener('input', (event) => {
        const input = event.target;
        if (input.dataset.validate === 'true') {
            const rules = JSON.parse(input.dataset.rules || '{}');
            const errors = utils.validateForm(new FormData(input.form), { [input.name]: rules });
            
            if (errors[input.name]) {
                input.classList.add('is-invalid');
                const feedback = input.nextElementSibling;
                if (feedback && feedback.classList.contains('invalid-feedback')) {
                    feedback.textContent = errors[input.name];
                }
            } else {
                input.classList.remove('is-invalid');
                const feedback = input.nextElementSibling;
                if (feedback && feedback.classList.contains('invalid-feedback')) {
                    feedback.textContent = '';
                }
            }
        }
    });
    
    // Gestionnaire de thème
    const themeToggle = document.getElementById('theme-toggle');
    if (themeToggle) {
        themeToggle.addEventListener('click', theme.toggle);
    }
    
    // Gestionnaire de langue
    const languageSelect = document.getElementById('language-select');
    if (languageSelect) {
        languageSelect.addEventListener('change', (event) => {
            i18n.setLanguage(event.target.value);
        });
    }
}

// Exporter les modules
window.app = {
    config,
    state,
    utils,
    notifications,
    theme,
    i18n,
    auth,
    ws
};
