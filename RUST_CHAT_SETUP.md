# Guide de connexion Chat React → Serveur Rust

## Architecture

Le chat React dans `talas-frontend` est maintenant configuré pour se connecter au serveur WebSocket Rust dans `backend/modules/chat_server/`.

## Démarrage du serveur Rust

### 1. Prérequis
```bash
# Installer Rust si pas encore fait
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Vérifier que PostgreSQL est démarré
sudo systemctl status postgresql
```

### 2. Configuration de la base de données
```bash
# Créer les tables nécessaires (si pas encore fait)
# Le serveur Rust attend ces tables :
# - users (id, username, email, password, etc.)
# - messages (id, from_user, to_user, room, content, timestamp)
# - rooms (id, name, description, created_at)
```

### 3. Variables d'environnement
Créer un fichier `.env` dans `backend/modules/chat_server/` :
```bash
# backend/modules/chat_server/.env
DATABASE_URL=postgresql://username:password@localhost/veza_dev
JWT_SECRET=votre_secret_jwt_du_backend_go
WS_BIND_ADDR=127.0.0.1:9001
RUST_LOG=chat_server=debug
```

### 4. Démarrage du serveur
```bash
cd backend/modules/chat_server
cargo run
```

Le serveur WebSocket Rust démarrera sur `ws://localhost:9001`

## Utilisation dans React

### 1. Démarrage du frontend
```bash
cd talas-frontend
npm run dev
```

### 2. Accès au chat
- Connectez-vous avec un utilisateur valide
- Allez sur la page Chat dans le dashboard
- Le chat tentera de se connecter au serveur Rust

### 3. Fonctionnalités supportées
- ✅ Connexion WebSocket avec authentification JWT
- ✅ Rejoindre des salons
- ✅ Envoyer messages dans les salons
- ✅ Messages directs entre utilisateurs
- ✅ Historique des messages
- ✅ Reconnexion automatique

## Protocole de communication

### Messages du client React vers Rust :
```typescript
// Rejoindre un salon
{ type: "join", room: "general" }

// Envoyer message salon
{ type: "message", room: "general", content: "Hello!" }

// Envoyer message direct
{ type: "dm", to: 123, content: "Hello!" }

// Demander historique salon
{ type: "room_history", room: "general", limit: 50 }

// Demander historique DM
{ type: "dm_history", with: 123, limit: 50 }
```

### Messages du serveur Rust vers React :
```typescript
// Nouveau message salon
{ type: "message", data: { id: 1, fromUser: 123, username: "user", content: "Hello!", timestamp: "...", room: "general" } }

// Nouveau message direct
{ type: "dm", data: { id: 1, fromUser: 123, username: "user", content: "Hello!", timestamp: "..." } }

// Historique salon
{ type: "room_history", data: [...messages] }

// Historique DM
{ type: "dm_history", data: [...messages] }

// Erreur
{ type: "error", message: "Description de l'erreur" }
```

## Résolution des problèmes

### Le serveur Rust ne démarre pas
1. Vérifier que PostgreSQL est démarré
2. Vérifier la DATABASE_URL dans le .env
3. Vérifier que les tables existent
4. Vérifier les permissions sur le port 9001

### React ne se connecte pas
1. Vérifier que le serveur Rust est démarré
2. Vérifier que le JWT est valide
3. Ouvrir les dev tools pour voir les erreurs WebSocket
4. Vérifier l'URL WebSocket dans les logs

### Messages ne s'affichent pas
1. Vérifier que l'utilisateur a rejoint le salon
2. Vérifier les logs du serveur Rust
3. Vérifier que les messages sont bien sauvés en base

## Logs utiles

### Serveur Rust
```bash
# Voir les logs détaillés
RUST_LOG=chat_server=debug cargo run

# Logs des connexions
2024-01-15 12:00:00 INFO  [chat_server] 🔌 Connexion TCP entrante
2024-01-15 12:00:00 INFO  [chat_server] 🔐 Authentification réussie user_id=123
2024-01-15 12:00:00 INFO  [chat_server] ✅ Connexion WS autorisée user_id=123
```

### React (Console du navigateur)
```javascript
// Connexion réussie
[Rust Chat WebSocket] Connecté au serveur Rust

// Message reçu
[Rust Chat WebSocket] Message reçu: {type: "message", data: {...}}

// Erreur
[Rust Chat WebSocket] Erreur: Failed to connect
```

## Tests

### Test de connexion
1. Démarrer le serveur Rust
2. Ouvrir la console navigateur
3. Aller sur la page chat
4. Vérifier le message "Connecté au serveur de chat Rust"

### Test d'envoi de message
1. Rejoindre un salon (ex: "general")
2. Envoyer un message
3. Vérifier qu'il apparaît dans l'interface
4. Vérifier dans les logs Rust qu'il est sauvé

### Test de messages directs
1. Sélectionner un utilisateur
2. Envoyer un message direct
3. Vérifier la réception côté destinataire 