# 📚 Documentation Complète du Backend Talas

## 📑 Table des matières

1. [Vue d'ensemble](#vue-densemble)
2. [Architecture](#architecture)
3. [Stack Technique](#stack-technique)
4. [Configuration](#configuration)
5. [Base de données](#base-de-données)
6. [API REST - Endpoints détaillés](#api-rest---endpoints-détaillés)
7. [WebSocket - Chat temps réel](#websocket---chat-temps-réel)
8. [Modules Rust](#modules-rust)
9. [Authentification & Sécurité](#authentification--sécurité)
10. [Tests & Validation](#tests--validation)

---

## 🎯 Vue d'ensemble

Talas est une plateforme audio collaborative permettant le partage de fichiers audio, la communication en temps réel, et la gestion de produits. Le backend est construit avec une architecture modulaire utilisant Go pour l'API REST et Rust pour les fonctionnalités hautes performances.

### Versions du projet

- **V1** : Authentification & gestion utilisateurs ✅
- **V2** : Fichiers, produits & documentation ✅  
- **V3** : Chat & messagerie temps réel 🟡
- **V4** : Streaming audio (Rust) 🔴
- **V5** : Partage de ressources 🔴
- **V6** : Tags & recherche avancée 🔴
- **V7** : Plateforme de troc 🔴
- **V8** : Messagerie directe avancée 🔴
- **V9** : Radio communautaire 🔴

---

## 🏗️ Architecture

### Structure des dossiers

```
backend/
├── cmd/server/          # Point d'entrée principal
│   └── main.go
├── internal/            # Code interne (non exporté)
│   ├── api/            # Modules API
│   │   ├── auth/       # Authentification
│   │   ├── user/       # Gestion utilisateurs
│   │   ├── admin/      # Administration
│   │   ├── product/    # Catalogue produits
│   │   ├── file/       # Gestion fichiers
│   │   ├── track/      # Gestion audio
│   │   ├── listing/    # Marketplace
│   │   ├── offer/      # Offres d'échange
│   │   ├── message/    # Messages
│   │   ├── room/       # Salons de chat
│   │   ├── search/     # Recherche
│   │   ├── tag/        # Tags
│   │   └── shared_resources/ # Ressources partagées
│   ├── common/         # Utilitaires partagés
│   ├── config/         # Configuration
│   ├── database/       # Connexion DB
│   ├── middleware/     # Middlewares
│   ├── models/         # Modèles de données
│   └── utils/          # Utilitaires
├── modules/            # Modules Rust
│   └── chat_server/    # Serveur WebSocket
├── scripts/            # Scripts utilitaires
└── docs/              # Documentation
```

### Architecture modulaire

Chaque module dans `internal/api/` suit cette structure :
- `handler.go` : Gestionnaires HTTP
- `service.go` : Logique métier
- `routes.go` : Configuration des routes

---

## 🛠️ Stack Technique

| Composant | Technologie | Version |
|-----------|-------------|---------|
| **Backend API** | Go | 1.21+ |
| **Framework Web** | Gin | v1.9.1 |
| **Base de données** | PostgreSQL | 15+ |
| **ORM/Query Builder** | SQLx | v1.3.5 |
| **Authentification** | JWT (dgrijalva/jwt-go) | v3.2.0 |
| **WebSocket Chat** | Rust + Tokio + Tungstenite | - |
| **Cache** | Redis (prévu) | - |
| **Déploiement** | Docker/Incus | - |

---

## ⚙️ Configuration

### Variables d'environnement

```bash
# Base de données
DATABASE_URL=postgres://user:password@localhost:5432/talas_db
DATABASE_MAX_CONNECTIONS=25
DATABASE_MAX_IDLE_CONNECTIONS=5

# Serveur
PORT=8080
ENVIRONMENT=development

# JWT
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h

# WebSocket
WS_BIND_ADDR=127.0.0.1:9001

# Fichiers
UPLOAD_PATH=./uploads
INTERNAL_DOCS_PATH=./internal_docs
MAX_FILE_SIZE=10485760  # 10MB
MAX_INTERNAL_DOC_SIZE=52428800  # 50MB

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8000
```

---

## 🗄️ Base de données

### Tables principales

#### `users`
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    bio TEXT,
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `products`
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    category_id INT REFERENCES categories(id),
    brand VARCHAR(100),
    model VARCHAR(100),
    description TEXT,
    price INT,
    warranty_months INT,
    warranty_conditions TEXT,
    manufacturer_website VARCHAR(255),
    specifications JSONB,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `user_products`
```sql
CREATE TABLE user_products (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    product_id INT REFERENCES products(id),
    version VARCHAR(50),
    serial_number VARCHAR(100),
    purchase_date DATE,
    purchase_price INT,
    warranty_expires DATE,
    notes TEXT,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);
```

#### `files`
```sql
CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES user_products(id),
    filename VARCHAR(255) NOT NULL,
    url VARCHAR(500),
    type VARCHAR(50),
    mime_type VARCHAR(100),
    size BIGINT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `messages`
```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    from_user INT REFERENCES users(id),
    to_user INT REFERENCES users(id),
    room VARCHAR(100),
    content TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### `rooms`
```sql
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_private BOOLEAN DEFAULT false,
    created_by INT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🔌 API REST - Endpoints détaillés

### 🔐 Authentication (`/api/v1/auth`)

#### `POST /auth/register`
Création d'un nouveau compte utilisateur.

**Request:**
```json
{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePass123!"
}
```

**Response (201):**
```json
{
    "success": true,
    "message": "User registered successfully",
    "data": {
        "user_id": 1,
        "username": "john_doe",
        "email": "john@example.com"
    }
}
```

**Errors:**
- `400` : Données invalides
- `409` : Email/username déjà existant

#### `POST /auth/login`
Authentification d'un utilisateur.

**Request:**
```json
{
    "email": "john@example.com",
    "password": "SecurePass123!"
}
```

**Response (200):**
```json
{
    "success": true,
    "message": "Login successful",
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 86400,
        "user": {
            "id": 1,
            "username": "john_doe",
            "email": "john@example.com",
            "role": "user"
        }
    }
}
```

#### `POST /auth/refresh`
Renouvellement du token d'accès.

**Request:**
```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200):**
```json
{
    "success": true,
    "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "expires_in": 86400
    }
}
```

#### `POST /auth/logout`
Déconnexion (révoque le refresh token).

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
    "success": true,
    "message": "Logout successful"
}
```

#### `GET /auth/me`
Récupère le profil de l'utilisateur connecté.

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
    "success": true,
    "data": {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com",
        "role": "user",
        "first_name": "John",
        "last_name": "Doe",
        "bio": "Audio enthusiast",
        "avatar": "/uploads/avatars/1.jpg",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

---

### 👤 Users (`/api/v1/users`)

#### `GET /users`
Liste tous les utilisateurs (avec pagination).

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `page` (int) : Numéro de page (défaut: 1)
- `limit` (int) : Nombre d'éléments par page (défaut: 20)
- `search` (string) : Recherche par nom/email

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "username": "john_doe",
            "email": "john@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "avatar": "/uploads/avatars/1.jpg"
        }
    ],
    "pagination": {
        "page": 1,
        "limit": 20,
        "total": 150,
        "pages": 8
    }
}
```

#### `GET /users/:id`
Récupère un utilisateur spécifique.

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
    "success": true,
    "data": {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "bio": "Audio enthusiast",
        "avatar": "/uploads/avatars/1.jpg",
        "created_at": "2024-01-15T10:30:00Z"
    }
}
```

#### `GET /users/me`
Alias pour `/auth/me`.

#### `PUT /users/me`
Met à jour le profil de l'utilisateur connecté.

**Headers:** `Authorization: Bearer {access_token}`

**Request:**
```json
{
    "first_name": "John",
    "last_name": "Smith",
    "bio": "Updated bio"
}
```

**Response (200):**
```json
{
    "success": true,
    "message": "Profile updated successfully",
    "data": {
        "id": 1,
        "username": "john_doe",
        "first_name": "John",
        "last_name": "Smith",
        "bio": "Updated bio"
    }
}
```

#### `PUT /users/password`
Change le mot de passe.

**Headers:** `Authorization: Bearer {access_token}`

**Request:**
```json
{
    "current_password": "OldPass123!",
    "new_password": "NewPass456!"
}
```

#### `GET /users/except-me`
Liste tous les utilisateurs sauf l'utilisateur connecté.

**Headers:** `Authorization: Bearer {access_token}`

---

### 📦 Products (`/api/v1/products`)

#### `GET /products/search`
Recherche publique de produits (sans auth).

**Query Parameters:**
- `q` (string) : Terme de recherche
- `limit` (int) : Limite de résultats (défaut: 10)

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "name": "Talas MK-1",
            "brand": "Talas Audio",
            "model": "MK-1",
            "description": "Professional audio interface"
        }
    ]
}
```

---

### 🗂️ User Products (`/api/v1/user-products`)

#### `GET /user-products`
Liste les produits possédés par l'utilisateur.

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `page` (int) : Page
- `limit` (int) : Limite
- `status` (string) : Filtrer par statut (active, sold, broken, archived)

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "user_id": 1,
            "product_id": 5,
            "product_name": "Talas MK-1",
            "category_name": "Audio Interfaces",
            "brand": "Talas Audio",
            "model": "MK-1",
            "version": "2.0",
            "serial_number": "TAL-001234",
            "purchase_date": "2023-06-15",
            "purchase_price": 299,
            "warranty_expires": "2025-06-15",
            "is_under_warranty": true,
            "status": "active",
            "notes": "Bought from official store",
            "files_count": 3,
            "docs_count": 2
        }
    ]
}
```

#### `GET /user-products/:id`
Détails d'un produit possédé.

**Headers:** `Authorization: Bearer {access_token}`

#### `POST /user-products`
Ajoute un produit à la collection.

**Headers:** `Authorization: Bearer {access_token}`

**Request:**
```json
{
    "product_id": 5,
    "version": "2.0",
    "serial_number": "TAL-001234",
    "purchase_date": "2023-06-15",
    "purchase_price": 299,
    "warranty_expires": "2025-06-15",
    "notes": "Bought from official store"
}
```

#### `PUT /user-products/:id`
Met à jour un produit possédé.

**Headers:** `Authorization: Bearer {access_token}`

**Request:**
```json
{
    "version": "2.1",
    "status": "active",
    "notes": "Updated firmware"
}
```

#### `DELETE /user-products/:id`
Supprime un produit de la collection.

**Headers:** `Authorization: Bearer {access_token}`

**Errors:**
- `409` : Impossible de supprimer (fichiers associés)

#### `GET /user-products/warranty`
Récupère le statut de garantie.

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `filter` (string) : expiring, expired, active

---

### 📄 Files (`/api/v1/files`)

#### `POST /products/:id/files`
Upload un fichier lié à un produit.

**Headers:** 
- `Authorization: Bearer {access_token}`
- `Content-Type: multipart/form-data`

**Form Data:**
- `file` : Le fichier à uploader
- `type` : Type de fichier (manual, warranty, invoice, image, document)

**Response (201):**
```json
{
    "success": true,
    "message": "File uploaded successfully",
    "data": {
        "id": 1,
        "filename": "manual_mk1.pdf",
        "url": "/uploads/files/1_manual_mk1.pdf",
        "type": "manual",
        "mime_type": "application/pdf",
        "size": 2048576
    }
}
```

**Errors:**
- `413` : Fichier trop large
- `415` : Type de fichier non supporté

#### `GET /products/:id/files`
Liste les fichiers d'un produit.

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "product_id": 1,
            "filename": "manual_mk1.pdf",
            "url": "/uploads/files/1_manual_mk1.pdf",
            "type": "manual",
            "mime_type": "application/pdf",
            "size": 2048576,
            "uploaded_at": "2024-01-15T10:30:00Z"
        }
    ]
}
```

#### `GET /files/:id/download`
Télécharge un fichier.

**Headers:** `Authorization: Bearer {access_token}`

**Response:** Fichier binaire avec headers appropriés

#### `DELETE /files/:id`
Supprime un fichier.

**Headers:** `Authorization: Bearer {access_token}`

---

### 💬 Chat & Messages (`/api/v1/chat`)

#### `GET /chat/rooms`
Liste les salons publics.

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "name": "general",
            "description": "General discussion",
            "is_private": false,
            "member_count": 42,
            "created_at": "2024-01-01T00:00:00Z"
        }
    ]
}
```

#### `POST /chat/rooms`
Crée un nouveau salon.

**Headers:** `Authorization: Bearer {access_token}`

**Request:**
```json
{
    "name": "audio-production",
    "description": "Discussion about audio production",
    "is_private": false
}
```

#### `GET /chat/rooms/:room/messages`
Récupère l'historique d'un salon.

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `limit` (int) : Nombre de messages (défaut: 50)
- `before` (string) : Messages avant cette date

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 123,
            "from_user": 1,
            "username": "john_doe",
            "content": "Hello everyone!",
            "timestamp": "2024-01-15T14:30:00Z"
        }
    ]
}
```

#### `GET /chat/dm/:user_id`
Récupère l'historique des messages directs.

**Headers:** `Authorization: Bearer {access_token}`

---

### 🔍 Search (`/api/v1/search`)

#### `GET /search`
Recherche globale.

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `q` (string) : Terme de recherche
- `type` (string) : Type de résultat (users, products, files, tracks)
- `limit` (int) : Limite par type

**Response (200):**
```json
{
    "success": true,
    "data": {
        "users": [...],
        "products": [...],
        "files": [...],
        "tracks": [...]
    }
}
```

---

### 🏷️ Tags (`/api/v1/tags`)

#### `GET /tags`
Liste tous les tags.

**Headers:** `Authorization: Bearer {access_token}`

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "name": "ambient",
            "usage_count": 45
        }
    ]
}
```

#### `GET /tags/search`
Recherche de tags.

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `q` (string) : Terme de recherche

---

### 🎵 Tracks (`/api/v1/tracks`)

#### `GET /tracks`
Liste les pistes audio.

**Headers:** `Authorization: Bearer {access_token}`

**Query Parameters:**
- `page` (int) : Page
- `limit` (int) : Limite
- `tag` (string) : Filtrer par tag
- `user_id` (int) : Filtrer par utilisateur

**Response (200):**
```json
{
    "success": true,
    "data": [
        {
            "id": 1,
            "title": "Ambient Loop #1",
            "artist": "john_doe",
            "duration": 180,
            "file_url": "/uploads/tracks/1.mp3",
            "waveform_url": "/uploads/waveforms/1.png",
            "tags": ["ambient", "loop"],
            "plays": 234,
            "likes": 45,
            "created_at": "2024-01-15T10:00:00Z"
        }
    ]
}
```

#### `POST /tracks`
Upload une nouvelle piste.

**Headers:** 
- `Authorization: Bearer {access_token}`
- `Content-Type: multipart/form-data`

**Form Data:**
- `file` : Fichier audio
- `title` : Titre de la piste
- `description` : Description
- `tags` : Tags (séparés par virgules)

---

### 🛡️ Admin (`/api/v1/admin`)

#### `GET /admin/dashboard`
Tableau de bord administrateur.

**Headers:** `Authorization: Bearer {access_token}` (role: admin)

**Response (200):**
```json
{
    "success": true,
    "data": {
        "users_count": 1234,
        "products_count": 56,
        "files_count": 789,
        "tracks_count": 345,
        "new_users_today": 12,
        "active_sessions": 89
    }
}
```

#### `GET /admin/users`
Liste des utilisateurs (admin).

**Headers:** `Authorization: Bearer {access_token}` (role: admin)

**Query Parameters:**
- `page` (int) : Page
- `limit` (int) : Limite
- `role` (string) : Filtrer par rôle
- `status` (string) : Filtrer par statut

#### `PUT /admin/users/:id`
Met à jour un utilisateur.

**Headers:** `Authorization: Bearer {access_token}` (role: admin)

**Request:**
```json
{
    "role": "moderator",
    "status": "active"
}
```

#### `GET /admin/products`
Gestion du catalogue produits.

**Headers:** `Authorization: Bearer {access_token}` (role: admin)

#### `POST /admin/products`
Ajoute un produit au catalogue.

**Headers:** `Authorization: Bearer {access_token}` (role: admin)

**Request:**
```json
{
    "name": "Talas MK-2",
    "category_id": 1,
    "brand": "Talas Audio",
    "model": "MK-2",
    "description": "Next generation audio interface",
    "price": 399,
    "warranty_months": 24,
    "specifications": {
        "inputs": 4,
        "outputs": 4,
        "sample_rate": "192kHz"
    }
}
```

---

## 🔌 WebSocket - Chat temps réel

### Connexion

```javascript
const token = localStorage.getItem('access_token');
const socket = new WebSocket(`ws://localhost:9001/?token=${token}`);
```

### Événements Client → Serveur

#### Rejoindre un salon
```json
{
    "type": "join",
    "room": "general"
}
```

#### Envoyer un message dans un salon
```json
{
    "type": "message",
    "room": "general",
    "content": "Hello everyone!"
}
```

#### Envoyer un message direct
```json
{
    "type": "dm",
    "to": 123,
    "content": "Hi there!"
}
```

#### Récupérer l'historique d'un salon
```json
{
    "type": "room_history",
    "room": "general",
    "limit": 50
}
```

#### Récupérer l'historique DM
```json
{
    "type": "dm_history",
    "with": 123,
    "limit": 50
}
```

### Événements Serveur → Client

#### Message de salon
```json
{
    "type": "message",
    "data": {
        "room": "general",
        "from": 1,
        "username": "john_doe",
        "content": "Hello everyone!",
        "timestamp": "2024-01-15T14:30:00Z"
    }
}
```

#### Message direct
```json
{
    "type": "dm",
    "data": {
        "from": 123,
        "username": "jane_smith",
        "content": "Hi there!",
        "timestamp": "2024-01-15T14:31:00Z"
    }
}
```

#### Confirmation d'action
```json
{
    "type": "joined",
    "data": {
        "room": "general",
        "status": "ok"
    }
}
```

#### Erreur
```json
{
    "type": "error",
    "data": {
        "message": "Room not found"
    }
}
```

---

## 🦀 Modules Rust

### Chat Server (WebSocket)

**Emplacement:** `backend/modules/chat_server/`

**Architecture:**
- `main.rs` : Point d'entrée, serveur Tokio
- `auth.rs` : Validation JWT
- `client.rs` : Gestion des connexions client
- `messages.rs` : Types de messages
- `hub/` : Logique de distribution
  - `common.rs` : Hub principal
  - `room.rs` : Gestion des salons
  - `dm.rs` : Messages directs

**Fonctionnalités:**
- Authentification JWT lors du handshake
- Gestion des salons publics/privés
- Messages directs entre utilisateurs
- Persistance dans PostgreSQL
- Broadcast efficace avec Tokio

**Configuration:**
```bash
WS_BIND_ADDR=127.0.0.1:9001
DATABASE_URL=postgres://user:pass@localhost/talas_db
```

### Streaming Audio (Prévu V4)

**Architecture prévue:**
- Serveur HTTP/gRPC en Rust
- Transcodage FFmpeg à la volée
- Support HTTP Range pour streaming progressif
- Cache intelligent des segments audio

---

## 🔐 Authentification & Sécurité

### JWT (JSON Web Tokens)

**Structure du token:**
```json
{
    "user_id": 1,
    "username": "john_doe",
    "role": "user",
    "exp": 1705325400,
    "iat": 1705239000
}
```

**Middleware d'authentification:**
```go
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Missing authorization header"})
            c.Abort()
            return
        }
        
        // Validate token
        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateToken(token, jwtSecret)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

### Rôles & Permissions

| Rôle | Permissions |
|------|------------|
| `user` | CRUD propres ressources, chat, recherche |
| `moderator` | + Modération chat, gestion tags |
| `admin` | + Gestion utilisateurs, catalogue |
| `super_admin` | Accès total |

### Sécurité des fichiers

- Validation MIME type
- Limite de taille par type
- Noms de fichiers sanitizés
- Stockage hors webroot
- Accès via endpoint sécurisé

---

## 🧪 Tests & Validation

### Script de test des endpoints

**Localisation:** `backend/scripts/test_endpoints.sh`

**Utilisation:**
```bash
./scripts/test_endpoints.sh
```

**Tests automatisés:**
1. Inscription/Connexion
2. Récupération du profil
3. CRUD produits
4. Upload/Download fichiers
5. Chat WebSocket
6. Recherche
7. Administration

### Validation des migrations

```bash
# Appliquer les migrations
migrate -path ./db/migrations -database $DATABASE_URL up

# Rollback si nécessaire
migrate -path ./db/migrations -database $DATABASE_URL down
```

### Tests unitaires

```bash
# Lancer tous les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests d'un module spécifique
go test ./internal/api/auth/...
```

---

## 📊 Monitoring & Logs

### Structure des logs

```go
log.Printf("[%s] %s %s - Status: %d, Duration: %v",
    time.Now().Format("2006-01-02 15:04:05"),
    method,
    path,
    statusCode,
    duration,
)
```

### Métriques importantes

- Temps de réponse API
- Nombre de connexions WebSocket actives
- Taille de la file d'attente de messages
- Utilisation mémoire/CPU
- Taux d'erreur par endpoint

---

## 🚀 Déploiement

### Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: talas_db
      POSTGRES_USER: talas
      POSTGRES_PASSWORD: secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build: ./backend
    environment:
      DATABASE_URL: postgres://talas:secure_password@postgres:5432/talas_db
      JWT_SECRET: ${JWT_SECRET}
      PORT: 8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    volumes:
      - ./uploads:/app/uploads
      - ./internal_docs:/app/internal_docs

  chat_server:
    build: ./backend/modules/chat_server
    environment:
      DATABASE_URL: postgres://talas:secure_password@postgres:5432/talas_db
      WS_BIND_ADDR: 0.0.0.0:9001
    ports:
      - "9001:9001"
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
      - chat_server

volumes:
  postgres_data:
```

### Build & Run

```bash
# Build
docker-compose build

# Run
docker-compose up -d

# Logs
docker-compose logs -f backend

# Stop
docker-compose down
```

---

## 📝 Notes importantes

1. **Sécurité** : Toujours valider les entrées utilisateur et utiliser des requêtes préparées
2. **Performance** : Utiliser la pagination pour les listes longues
3. **Fichiers** : Nettoyer régulièrement les fichiers orphelins
4. **WebSocket** : Implémenter un système de heartbeat/ping
5. **Monitoring** : Surveiller les métriques critiques en production

---

## 🔗 Ressources

- [Documentation Go](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [SQLx](https://github.com/jmoiron/sqlx)
- [JWT Go](https://github.com/dgrijalva/jwt-go)
- [Tokio (Rust)](https://tokio.rs/)
- [Tungstenite (Rust WebSocket)](https://github.com/snapview/tungstenite-rs)
