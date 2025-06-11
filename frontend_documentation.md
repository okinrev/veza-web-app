# 📖 Documentation Complète du Frontend Talas

## Table des matières
1. [Architecture Générale](#architecture-générale)
2. [Technologies Utilisées](#technologies-utilisées)
3. [Structure des Fichiers](#structure-des-fichiers)
4. [Pages Principales](#pages-principales)
5. [Composants JavaScript/Alpine.js](#composants-javascriptalpinejs)
6. [API Routes et Requêtes](#api-routes-et-requêtes)
7. [Gestion de l'État](#gestion-de-létat)
8. [WebSocket et Communication Temps Réel](#websocket-et-communication-temps-réel)

---

## 1. Architecture Générale

Le frontend de Talas suit une architecture **modulaire et légère** basée sur :
- HTML statique pour la structure
- Alpine.js pour la réactivité et gestion d'état local
- Tailwind CSS pour le styling
- JavaScript natif pour les fonctionnalités complexes
- WebSocket pour la communication temps réel

### Principes directeurs :
- **Pas de build process** : Tout fonctionne directement dans le navigateur
- **Progressive Enhancement** : Le site fonctionne sans JavaScript (forme basique)
- **État local** : Chaque page gère son propre état via Alpine.js
- **API REST** : Communication avec le backend Go via fetch API

---

## 2. Technologies Utilisées

| Technologie | Version | Usage |
|------------|---------|--------|
| Alpine.js | 3.x | Réactivité, gestion d'état local |
| Tailwind CSS | CDN | Styling utility-first |
| JavaScript | ES6+ | Logique métier, WebSocket |
| HTML5 | - | Structure des pages |
| LocalStorage | - | Persistance JWT, préférences |

---

## 3. Structure des Fichiers

```
frontend/
├── css/
│   └── style.css          # Styles personnalisés
├── js/
│   ├── app.js            # Logique principale (main.html)
│   ├── chat.js           # Gestion du chat unifié
│   ├── message.js        # Messages directs
│   ├── produits.js       # Gestion des produits
│   ├── register.js       # Inscription
│   ├── room.js           # Chat rooms
│   └── users.js          # Gestion utilisateurs
├── pages principales/
│   ├── login.html        # Page de connexion
│   ├── register.html     # Page d'inscription
│   ├── dashboard.html    # Tableau de bord dev
│   ├── main.html         # Application principale
│   ├── hub.html          # Hub intégré (ancien)
│   ├── hub_v2.html       # Hub v2 amélioré
│   ├── gg.html           # Version tout-en-un
│   ├── chat.html         # Chat unifié (rooms + DM)
│   ├── room.html         # Chat rooms dédié
│   ├── message.html      # Messages directs
│   ├── users.html        # Liste utilisateurs
│   ├── produits.html     # Gestion produits
│   ├── user_products.html # Produits utilisateur
│   ├── admin_products.html # Admin produits
│   ├── track.html        # Gestion pistes audio
│   ├── shared_ressources.html # Ressources partagées
│   ├── api.html          # Documentation API
│   └── test.html         # Console de test API
└── components/           # Composants réutilisables

```

---

## 4. Pages Principales

### 4.1 Page de Connexion (`login.html`)

**Fonctionnalités :**
- Formulaire de connexion avec email/mot de passe
- Validation côté client avec Alpine.js
- Gestion des erreurs et messages de succès
- Redirection après connexion réussie
- Liens vers inscription et récupération de mot de passe

**État Alpine.js :**
```javascript
{
    email: '',
    password: '',
    message: '',
    messageType: 'error',
    isLoading: false,
    isLoggedIn: false
}
```

**Routes API utilisées :**
- `POST /login` - Authentification utilisateur

**LocalStorage :**
- `access_token` - Token JWT d'accès
- `refresh_token` - Token de rafraîchissement
- `user_id` - ID de l'utilisateur connecté

---

### 4.2 Page d'Inscription (`register.html`)

**Fonctionnalités :**
- Formulaire complet avec validation en temps réel
- Indicateur de force du mot de passe
- Vérification de la correspondance des mots de passe
- Acceptation des conditions d'utilisation
- Auto-connexion après inscription réussie

**État Alpine.js :**
```javascript
{
    form: {
        email: '',
        username: '',
        password: '',
        confirmPassword: '',
        acceptTerms: false
    },
    errors: {},
    passwordStrength: {
        score: 0,
        label: '',
        color: '',
        percent: 0
    },
    isLoading: false,
    registrationSuccess: false,
    message: { text: '', type: 'error' }
}
```

**Méthodes principales :**
- `validateEmail()` - Validation format email
- `validateUsername()` - Validation nom d'utilisateur
- `checkPasswordStrength()` - Calcul force mot de passe
- `register()` - Soumission du formulaire

**Routes API :**
- `POST /signup` - Création de compte

---

### 4.3 Dashboard de Développement (`dashboard.html`)

**Description :**
Page de test permettant de visualiser plusieurs composants simultanément via des iframes.

**Composants affichés :**
- Login
- Register
- Produits
- Salons
- Messages directs
- Liste utilisateurs
- Console API de test

---

### 4.4 Application Principale (`main.html`)

**Architecture :**
- Layout avec sidebar de navigation
- Chargement dynamique des composants
- Gestion des onglets avec Alpine.js

**Navigation disponible :**
```javascript
{
    'dashboard': 'Dashboard',
    'dm': 'Messages Privés',
    'rooms': 'Salons',
    'tracks': 'Musiques',
    'products': 'Produits',
    'docs': 'Docs Produits',
    'shared': 'Ressources Partagées',
    'tags': 'Tags',
    'suggestions': 'Suggestions'
}
```

**Fonction de chargement dynamique :**
```javascript
loadComponent(id, url) {
    // Charge le HTML du composant dans le conteneur
}
```

---

### 4.5 Chat Unifié (`chat.html`)

**Fonctionnalités principales :**
- Interface à deux onglets : Salons et Messages Directs
- WebSocket pour communication temps réel
- Gestion des notifications navigateur
- Indicateurs de présence en ligne
- Historique des messages avec pagination
- Création de salons
- Recherche d'utilisateurs

**État complexe :**
```javascript
{
    // Authentification
    username: '',
    myUserId: null,
    isConnected: false,
    
    // Navigation
    activeTab: 'rooms',
    
    // Rooms
    rooms: [],
    currentRoom: '',
    roomMessages: [],
    newRoomName: '',
    
    // DM
    users: [],
    onlineUsers: [],
    otherUserId: null,
    otherUserInfo: {},
    messages: [],
    conversations: [],
    
    // UI States
    loadingMessages: false,
    sending: false,
    creating: false,
    notificationsEnabled: true,
    
    // WebSocket
    socket: null,
    reconnectAttempts: 0,
    maxReconnectAttempts: 5
}
```

**Protocole WebSocket :**
```javascript
// Envoi
{ type: "join", room: "general" }
{ type: "message", room: "general", content: "Hello" }
{ type: "dm", to: userId, content: "Hi!" }
{ type: "room_history", room: "general", limit: 50 }
{ type: "dm_history", with: userId, limit: 50 }

// Réception
{ type: "message", data: { room, username, content, timestamp } }
{ type: "dm", data: { fromUser, to, content, timestamp } }
{ type: "room_history", messages: [...] }
{ type: "user_joined", userId, username }
{ type: "user_left", userId }
{ type: "user_status", userId, isOnline }
```

---

### 4.6 Gestion des Produits (`produits.html`, `user_products.html`, `admin_products.html`)

**Fonctionnalités communes :**
- CRUD complet (Create, Read, Update, Delete)
- Upload de fichiers (images, PDF)
- Recherche et filtrage
- Pagination
- Tri par colonnes

**État produit :**
```javascript
{
    produits: [],
    filteredProduits: [],
    searchQuery: '',
    sortField: 'nom',
    sortDirection: 'asc',
    currentPage: 1,
    itemsPerPage: 10,
    
    form: {
        id: null,
        nom: '',
        description: '',
        prix: 0,
        stock: 0,
        fichiers: []
    },
    
    isEditing: false,
    isLoading: false,
    showDeleteModal: false,
    productToDelete: null
}
```

**Routes API produits :**
- `GET /products` - Liste des produits
- `POST /products` - Créer un produit
- `PUT /products/:id` - Modifier un produit
- `DELETE /products/:id` - Supprimer un produit
- `POST /products/:id/files` - Upload fichiers

**Différences entre versions :**
- `produits.html` : Version basique
- `user_products.html` : Vue utilisateur avec ses propres produits
- `admin_products.html` : Interface admin avec statistiques

---

### 4.7 Gestion des Pistes Audio (`track.html`)

**Fonctionnalités :**
- Liste des pistes avec métadonnées
- Lecteur audio intégré
- Upload de fichiers audio
- Gestion des tags et genres
- Partage communautaire
- Statistiques d'écoute

**État Alpine.js :**
```javascript
{
    tracks: [],
    currentTrack: null,
    isPlaying: false,
    audioPlayer: null,
    
    filters: {
        genre: '',
        artist: '',
        tags: []
    },
    
    uploadForm: {
        title: '',
        artist: '',
        album: '',
        genre: '',
        tags: '',
        audioFile: null,
        coverFile: null
    },
    
    stats: {
        totalTracks: 0,
        totalPlays: 0,
        totalDuration: 0
    }
}
```

**Méthodes audio :**
- `playTrack(track)` - Lecture d'une piste
- `pauseTrack()` - Pause
- `seekTo(position)` - Navigation dans la piste
- `updateVolume(level)` - Contrôle du volume

---

### 4.8 Ressources Partagées (`shared_ressources.html`)

**Types de ressources :**
- Samples audio
- Presets
- Templates
- Documents
- Tutoriels

**Fonctionnalités :**
- Upload avec drag & drop
- Catégorisation par tags
- Système de notation
- Commentaires
- Téléchargement avec compteur

**État :**
```javascript
{
    resources: [],
    categories: ['samples', 'presets', 'templates', 'docs', 'tutorials'],
    
    filters: {
        category: '',
        tags: [],
        rating: 0,
        uploader: ''
    },
    
    uploadForm: {
        title: '',
        description: '',
        category: '',
        tags: '',
        file: null
    },
    
    view: 'grid', // 'grid' ou 'list'
    sortBy: 'date', // 'date', 'rating', 'downloads'
}
```

---

### 4.9 Console de Test API (`test.html`)

**Fonctionnalités :**
- Interface pour tester tous les endpoints
- Historique des requêtes
- Sauvegarde des requêtes favorites
- Affichage formaté des réponses JSON
- Gestion des headers personnalisés

**Structure de requête :**
```javascript
{
    method: 'GET',
    endpoint: '/api/v1/users',
    headers: {
        'Authorization': 'Bearer [token]',
        'Content-Type': 'application/json'
    },
    body: '',
    response: null,
    status: null,
    duration: 0
}
```

---

## 5. Composants JavaScript/Alpine.js

### 5.1 Patterns Communs

**Initialisation :**
```javascript
x-data="componentName()" 
x-init="init()"
```

**Gestion des erreurs :**
```javascript
try {
    const response = await fetch(url, options);
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    // Traitement...
} catch (error) {
    this.showNotification('Erreur: ' + error.message, 'error');
}
```

**Notifications :**
```javascript
showNotification(message, type = 'info') {
    const notification = { id: Date.now(), message, type, show: true };
    this.notifications.push(notification);
    
    setTimeout(() => {
        notification.show = false;
        setTimeout(() => {
            this.notifications = this.notifications.filter(n => n.id !== notification.id);
        }, 300);
    }, 3000);
}
```

### 5.2 Gestion de l'authentification

**Vérification du token :**
```javascript
async checkAuth() {
    const token = localStorage.getItem('access_token');
    if (!token) {
        window.location.href = '/login.html';
        return false;
    }
    
    try {
        const response = await fetch('/api/v1/users/me', {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (!response.ok) {
            throw new Error('Token invalide');
        }
        
        const user = await response.json();
        this.username = user.data.username;
        this.myUserId = user.data.id;
        return true;
    } catch (error) {
        localStorage.removeItem('access_token');
        window.location.href = '/login.html';
        return false;
    }
}
```

**Déconnexion :**
```javascript
logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_id');
    
    if (this.socket) {
        this.socket.close();
    }
    
    window.location.href = '/login.html';
}
```

---

## 6. API Routes et Requêtes

### 6.1 Configuration des requêtes

**Headers par défaut :**
```javascript
const defaultHeaders = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${localStorage.getItem('access_token')}`
};
```

### 6.2 Endpoints principaux

| Domaine | Méthode | Endpoint | Description |
|---------|---------|----------|-------------|
| **Auth** | POST | `/login` | Connexion |
| | POST | `/signup` | Inscription |
| | POST | `/logout` | Déconnexion |
| | POST | `/refresh` | Rafraîchir token |
| **Users** | GET | `/users/me` | Profil actuel |
| | GET | `/users` | Liste utilisateurs |
| | GET | `/users/except-me` | Utilisateurs sauf moi |
| | GET | `/users/search?q=` | Recherche |
| | PUT | `/users/me` | Modifier profil |
| **Products** | GET | `/products` | Liste produits |
| | POST | `/products` | Créer produit |
| | PUT | `/products/:id` | Modifier |
| | DELETE | `/products/:id` | Supprimer |
| **Chat** | WS | `/ws` | WebSocket chat |
| | GET | `/chat/rooms` | Liste salons |
| | POST | `/chat/rooms` | Créer salon |
| **Messages** | GET | `/chat/dm/:userId` | Historique DM |
| | GET | `/chat/rooms/:room/messages` | Messages salon |
| **Tracks** | GET | `/tracks` | Liste pistes |
| | POST | `/tracks` | Upload piste |
| | GET | `/tracks/:id/stream` | Stream audio |
| **Resources** | GET | `/resources` | Liste ressources |
| | POST | `/resources` | Upload ressource |
| | GET | `/resources/:id/download` | Télécharger |

### 6.3 Gestion des erreurs HTTP

```javascript
async handleApiCall(url, options = {}) {
    try {
        const response = await fetch(url, {
            ...options,
            headers: {
                ...defaultHeaders,
                ...options.headers
            }
        });
        
        if (response.status === 401) {
            // Token expiré, tentative de refresh
            await this.refreshToken();
            return this.handleApiCall(url, options);
        }
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || `HTTP ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API Error:', error);
        throw error;
    }
}
```

---

## 7. Gestion de l'État

### 7.1 État Local (Alpine.js)

Chaque page gère son propre état via `x-data`. Les données ne persistent pas entre les pages sauf via :
- LocalStorage
- SessionStorage
- Paramètres URL
- Cookies

### 7.2 Patterns de réactivité

**Computed properties :**
```javascript
get filteredItems() {
    return this.items.filter(item => 
        item.name.toLowerCase().includes(this.searchQuery.toLowerCase())
    );
}
```

**Watchers :**
```javascript
x-init="
    $watch('searchQuery', value => {
        this.currentPage = 1;
        this.filterItems();
    })
"
```

### 7.3 Stores Alpine.js

Pour les données partagées (utilisées dans `gg.html`) :

```javascript
Alpine.store('app', {
    isLoggedIn: false,
    username: '',
    currentTab: 'auth',
    message: '',
    messageType: '',
    
    showMessage(msg, type = 'info') {
        this.message = msg;
        this.messageType = type;
        setTimeout(() => {
            this.message = '';
        }, 3000);
    }
});
```

---

## 8. WebSocket et Communication Temps Réel

### 8.1 Connexion WebSocket

```javascript
connectWebSocket() {
    const token = localStorage.getItem('access_token');
    const wsUrl = `ws://localhost:8080/ws?token=${token}`;
    
    this.socket = new WebSocket(wsUrl);
    
    this.socket.onopen = () => {
        console.log('WebSocket connecté');
        this.isConnected = true;
        this.reconnectAttempts = 0;
    };
    
    this.socket.onclose = () => {
        this.isConnected = false;
        this.handleReconnect();
    };
    
    this.socket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
    
    this.socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
        this.handleWebSocketMessage(data);
    };
}
```

### 8.2 Reconnexion automatique

```javascript
handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
        
        console.log(`Reconnexion dans ${delay/1000}s...`);
        
        setTimeout(() => {
            this.connectWebSocket();
        }, delay);
    } else {
        this.showNotification('Connexion perdue. Veuillez rafraîchir la page.', 'error');
    }
}
```

### 8.3 Gestion des messages

```javascript
handleWebSocketMessage(data) {
    switch(data.type) {
        case 'message':
            this.handleChatMessage(data);
            break;
        case 'dm':
            this.handleDirectMessage(data);
            break;
        case 'user_joined':
            this.handleUserJoined(data);
            break;
        case 'user_left':
            this.handleUserLeft(data);
            break;
        case 'typing':
            this.handleTypingIndicator(data);
            break;
        case 'error':
            this.showNotification(data.message, 'error');
            break;
        default:
            console.warn('Type de message inconnu:', data.type);
    }
}
```

---

## 9. Optimisations et Bonnes Pratiques

### 9.1 Performance

- **Lazy loading** : Chargement des composants à la demande
- **Debouncing** : Pour les recherches et validations
- **Virtualisation** : Pour les longues listes (non implémenté actuellement)
- **Pagination** : Limiter le nombre d'éléments affichés

### 9.2 Sécurité

- **Validation côté client** : Première barrière
- **Échappement HTML** : Protection XSS via Alpine.js
- **HTTPS** : Pour production
- **Token rotation** : Refresh tokens

### 9.3 Accessibilité

- **Labels ARIA** : Pour les lecteurs d'écran
- **Navigation clavier** : Support complet
- **Contraste** : Respect des standards WCAG
- **Messages d'erreur** : Clairs et informatifs

### 9.4 Responsive Design

- **Mobile-first** : Conception adaptative
- **Breakpoints Tailwind** : sm, md, lg, xl, 2xl
- **Touch-friendly** : Zones de clic suffisantes
- **Orientation** : Support portrait/paysage

---

## 10. Évolution Future

### Versions planifiées :

**V4 - Streaming Audio**
- Intégration lecteur audio avancé
- Playlists collaboratives
- Synchronisation multi-utilisateurs

**V5 - Partage Fichiers Avancé**
- Versioning des fichiers
- Collaboration temps réel
- Prévisualisation avancée

**V6 - Marketplace/Troc**
- Système d'enchères
- Réputation utilisateurs
- Paiements intégrés

**V7 - AudioGridder**
- Interface de contrôle VST
- Routing audio distant
- Sessions collaboratives

**V8-V12**
- Formation/tutoriels
- Radio communautaire
- Suivi produits/garanties
- Recherche globale
- Application Tauri

Cette documentation sera mise à jour au fur et à mesure de l'évolution du projet.
