# 💬 Version 3 — Chat & Messagerie Temps Réel

🎯 **Objectif** : Permettre aux utilisateurs de communiquer entre eux via des salons publics ou des messages directs (DM), en temps réel, avec authentification sécurisée et stockage persistant.

---

## 🚦 État : `🟡 En développement (V3.x)`

---

## 🧩 Fonctionnalités Principales

### 🔌 Serveur WebSocket
- Implémentation en **Go (`gorilla/websocket`)** ou **Rust (`tokio`, `axum`, `tungstenite`)**
- Authentification via **JWT** dans le handshake (token query ou header)
- Gestion des connexions simultanées via `hub.go`
- Prise en charge de :
  - Salons publics (`rooms`)
  - Messages privés (`DM`)
- Sauvegarde des messages dans PostgreSQL
- Historique accessible par événement `history`

### 👥 Interface utilisateur (Frontend)
- Liste des conversations (salons + utilisateurs disponibles)
- Chatbox dynamique avec envoi instantané
- Rechargement de l’historique au scroll
- Indicateur de statut en ligne (optionnel)

---

## 🔧 Stack Technique

| Composant         | Technologie                          |
|-------------------|---------------------------------------|
| WebSocket         | Go (`gorilla/websocket`) ou Rust (`tokio`) |
| Authentification  | JWT via header ou query (`Sec-WebSocket-Protocol`) |
| Stockage messages | PostgreSQL                           |
| Frontend          | React + Zustand (état chat)           |
| UI                | shadcn/ui + Tailwind                  |

---

## 🗃️ Backend : Structure des Fichiers

backend/
├── ws/
│ ├── hub.go # Gestion des connexions / broadcast
│ ├── client.go # Connexion WebSocket utilisateur
│ └── handlers.go # Traitement des événements (join, message, etc.)
├── routes/
│ └── ws.go # Endpoint WebSocket
├── models/
│ └── message.go
├── db/migrations/
│ └── add_messages_table.sql

---

## 🧱 Tables PostgreSQL

- `messages` :
  - `id`, `from_user`, `to_user` (nullable si message de salon), `room`, `content`, `timestamp`
- `rooms` :
  - `id`, `name`, `is_private`, `created_at`

---

## 📡 Événements WebSocket

| Événement | Payload                                     | Description                        |
|-----------|---------------------------------------------|------------------------------------|
| `join`    | `{ "room": "general" }`                    | Rejoindre un salon                 |
| `message` | `{ "room": "general", "content": "yo" }`   | Envoyer un message à un salon      |
| `dm`      | `{ "to": "user_id", "content": "yo" }`     | Message direct                     |
| `history` | `{ "room": "general", "limit": 50 }`       | Charger les messages précédents    |

---

## ✅ Checklist de Validation

- [ ] Authentification JWT à l’ouverture de WebSocket
- [ ] Mapping utilisateurs ↔ connexions actif
- [ ] Gestion des salons publics (`rooms`)
- [ ] Gestion des DMs entre utilisateurs
- [ ] Sauvegarde des messages en base PostgreSQL
- [ ] Rechargement de l’historique
- [ ] Intégration React : Chatbox, Sidebar, DMView
- [ ] Gestion frontend de l’état via Zustand
- [ ] Comportement fluide & responsive

---

## 💡 Bonus Fonctionnels

- 🔴 Statut "en ligne / hors ligne"
- 🔔 Notifications sur nouveaux messages (push/local)
- 💬 Formatage markdown simple (`marked.js`)
- 🙂 Support des emojis

---

## 🔁 Plan de Développement Étape par Étape

### ✅ V3.1 — Initialisation
- Démarrage serveur WS
- Auth JWT dans le handshake
- Structure `hub`, `client`, `handlers`

### ✅ V3.2 — Gestion des rooms
- Entrée dans room via `join`
- Broadcast des messages à la room
- Table `rooms` PostgreSQL
- Création auto de room si inexistante

### ✅ V3.3 — Échange de messages
- Envoi/Reception d’un message texte
- `dm` via socket map
- Enregistrement en base + horodatage

### ✅ V3.4 — Historique
- Requête `history` ou `dm_history`
- Rendu frontend + scrollback

### ✅ V3.5 — Intégration Frontend
- Composants : `ChatRoom`, `DMView`, `Sidebar`
- `chatStore.ts` avec Zustand

### ✅ V3.6 — UX/Performances
- Scroll automatique
- Ping régulier pour status en ligne
- Feedback utilisateur instantané

---

## 🔍 Résumé JSON des Protocoles WebSocket

```json
// Client → Serveur
{ "type": "join", "room": "general" }
{ "type": "message", "room": "general", "content": "yo" }
{ "type": "dm", "to": "user_id", "content": "salut" }
{ "type": "history", "room": "general", "limit": 50 }

// Serveur → Client
{ "type": "message", "room": "general", "from": "user_id", "content": "...", "timestamp": "..." }
{ "type": "dm", "from": "user_id", "content": "...", "timestamp": "..." }
{ "type": "history", "room": "general", "messages": [...] }
