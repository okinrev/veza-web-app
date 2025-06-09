import { isAuthenticated, getUser, logout } from '../../../utils/auth.js';

class Header {
    constructor() {
        this.userMenu = document.getElementById('user-menu');
        this.authButtons = document.getElementById('auth-buttons');
        this.userMenuButton = document.getElementById('user-menu-button');
        this.userDropdown = document.getElementById('user-dropdown');
        this.logoutButton = document.getElementById('logout-button');
        this.username = document.getElementById('username');
        this.userAvatar = document.getElementById('user-avatar');

        this.init();
    }

    init() {
        this.updateAuthState();
        this.setupEventListeners();
    }

    updateAuthState() {
        if (isAuthenticated()) {
            const user = getUser();
            this.userMenu.style.display = 'block';
            this.authButtons.style.display = 'none';
            this.username.textContent = user.username;
            if (user.avatar) {
                this.userAvatar.src = user.avatar;
            }
        } else {
            this.userMenu.style.display = 'none';
            this.authButtons.style.display = 'flex';
        }
    }

    setupEventListeners() {
        // Toggle dropdown menu
        this.userMenuButton.addEventListener('click', () => {
            this.userDropdown.classList.toggle('show');
        });

        // Close dropdown when clicking outside
        document.addEventListener('click', (event) => {
            if (!this.userMenu.contains(event.target)) {
                this.userDropdown.classList.remove('show');
            }
        });

        // Handle logout
        this.logoutButton.addEventListener('click', async () => {
            try {
                await logout();
                window.location.href = '/login';
            } catch (error) {
                console.error('Logout failed:', error);
            }
        });
    }
}

export default Header; 