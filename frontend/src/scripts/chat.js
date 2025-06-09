function unifiedChatApp() {
    return {
        // État de l'application
        username: '',
        isConnected: false,
        activeTab: 'rooms',
        currentRoom: null,
        otherUserId: null,
        otherUserInfo: {},
        rooms: [],
        connectedUsers: [],
        filteredUsers: [],
        recentConversations: [],
        userSearch: '',
        newRoomName: '',
        creating: false,
        loadingUsers: false,
        roomStats: {
            totalRooms: 0,
            activeUsers: 0,
            todayMessages: 0
        },
        dmStats: {
            totalMessages: 0,
            todayMessages: 0
        },

        // Initialisation
        init() {
            this.checkAuth();
            this.initWebSocket();
            this.loadInitialData();
        },

        // Vérification de l'authentification
        checkAuth() {
            const token = localStorage.getItem('token');
            if (!token) {
                window.location.href = '/login.html';
                return;
            }
            this.username = localStorage.getItem('username');
        },

        // Initialisation WebSocket
        initWebSocket() {
            this.ws = new WebSocket('ws://localhost:8000/ws');
            
            this.ws.onopen = () => {
                this.isConnected = true;
                this.ws.send(JSON.stringify({
                    type: 'auth',
                    token: localStorage.getItem('token')
                }));
            };

            this.ws.onclose = () => {
                this.isConnected = false;
                setTimeout(() => this.initWebSocket(), 5000);
            };

            this.ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                this.handleWebSocketMessage(data);
            };
        },

        // Gestion des messages WebSocket
        handleWebSocketMessage(data) {
            switch (data.type) {
                case 'room_list':
                    this.rooms = data.rooms;
                    this.roomStats.totalRooms = data.rooms.length;
                    break;
                case 'user_list':
                    this.connectedUsers = data.users;
                    this.roomStats.activeUsers = data.users.length;
                    break;
                case 'message':
                    this.handleNewMessage(data);
                    break;
                case 'user_status':
                    this.updateUserStatus(data);
                    break;
            }
        },

        // Gestion des nouveaux messages
        handleNewMessage(data) {
            if (data.room) {
                // Message dans un salon
                this.roomStats.todayMessages++;
            } else {
                // Message privé
                this.dmStats.todayMessages++;
                this.dmStats.totalMessages++;
                
                // Mise à jour des conversations récentes
                const convIndex = this.recentConversations.findIndex(c => c.userId === data.senderId);
                if (convIndex > -1) {
                    this.recentConversations.splice(convIndex, 1);
                }
                this.recentConversations.unshift({
                    userId: data.senderId,
                    username: data.sender,
                    lastMessage: data.content
                });
            }
        },

        // Mise à jour du statut utilisateur
        updateUserStatus(data) {
            if (data.userId === this.otherUserId) {
                this.otherUserInfo.isOnline = data.isOnline;
            }
            const userIndex = this.filteredUsers.findIndex(u => u.id === data.userId);
            if (userIndex > -1) {
                this.filteredUsers[userIndex].isOnline = data.isOnline;
            }
        },

        // Chargement des données initiales
        loadInitialData() {
            if (this.activeTab === 'rooms') {
                this.refreshRooms();
            } else {
                this.refreshUsers();
            }
        },

        // Rafraîchissement des salons
        refreshRooms() {
            this.ws.send(JSON.stringify({ type: 'get_rooms' }));
        },

        // Rafraîchissement des utilisateurs
        refreshUsers() {
            this.loadingUsers = true;
            this.ws.send(JSON.stringify({ type: 'get_users' }));
            setTimeout(() => this.loadingUsers = false, 1000);
        },

        // Création d'un salon
        createRoom() {
            if (!this.newRoomName.trim()) return;
            
            this.creating = true;
            this.ws.send(JSON.stringify({
                type: 'create_room',
                name: this.newRoomName
            }));
            
            setTimeout(() => {
                this.creating = false;
                this.newRoomName = '';
            }, 1000);
        },

        // Rejoindre un salon
        joinRoom(roomName) {
            if (this.currentRoom === roomName) return;
            
            this.currentRoom = roomName;
            this.ws.send(JSON.stringify({
                type: 'join_room',
                room: roomName
            }));
        },

        // Sélection d'un utilisateur pour les messages privés
        selectUser(userId) {
            this.otherUserId = userId;
            this.ws.send(JSON.stringify({
                type: 'get_user_info',
                userId: userId
            }));
        },

        // Recherche d'utilisateurs
        searchUsers() {
            if (!this.userSearch.trim()) {
                this.filteredUsers = [];
                return;
            }
            
            this.ws.send(JSON.stringify({
                type: 'search_users',
                query: this.userSearch
            }));
        },

        // Déconnexion
        logout() {
            localStorage.removeItem('token');
            localStorage.removeItem('username');
            window.location.href = '/login.html';
        },

        // Retour à la page précédente
        goBack() {
            window.history.back();
        }
    };
} 