# 🎵 Version 4 — Streaming Audio Talas

🎯 Objectif : Permettre aux utilisateurs d’écouter des morceaux audio (productions, samples, maquettes) en streaming depuis le navigateur, en intégrant un module performant basé sur Rust.

---

## 🧩 Fonctionnalités à Implémenter

### 📡 Serveur de Streaming (Rust)
- Serveur HTTP ou WebSocket capable de streamer des fichiers audio
- Lecture en continu avec **bufferisation** dynamique
- Transcodage à la volée via FFmpeg (MP3, WAV, FLAC)
- Gestion des headers MIME pour lecture dans un navigateur
- Support de la **lecture progressive** (HTTP Range)

### 🎧 Intégration Go ↔ Rust
- Appels gRPC pour déléguer le streaming depuis l’API Go
- Transfert sécurisé du token d’accès
- Déploiement en tant que service à part (Rust standalone ou sidecar)

### 🖥️ Frontend React
- Intégration d’un lecteur audio minimal :
  - Bouton play/pause
  - Barre de progression
  - Affichage du nom, durée, artiste
- Récupération des morceaux disponibles depuis l’API

---

## 🔧 Stack Technique

| Composant        | Technologie                                |
|------------------|---------------------------------------------|
| Module streaming | Rust + Tokio + FFmpeg + hyper / axum       |
| Backend bridge   | Go → Rust via gRPC                         |
| Codec support    | MP3, WAV, FLAC (transcodage à la volée)    |
| Frontend         | React + HTML `<audio>` + Tailwind          |
| Auth             | Token JWT envoyé via headers ou query param|

---

## 🗃️ Structure des Fichiers (Rust)

rust_streaming/
├── src/
│ ├── main.rs
│ ├── server.rs
│ ├── stream.rs
│ └── ffmpeg.rs
├── proto/
│ └── stream.proto
└── Cargo.toml

---

## 🔍 API & Routes

### 🎶 Backend Go → Rust (gRPC)
| Méthode gRPC | Description                                 |
|--------------|---------------------------------------------|
| `GetStreamURL(audio_id)` | Retourne une URL temporaire de streaming |

### 🌐 Endpoints HTTP (Rust)
| Méthode | Route                  | Description                     |
|--------:|------------------------|---------------------------------|
| GET     | `/stream/{id}`         | Stream audio avec auth token    |
| GET     | `/preview/{id}`        | Preview court (30s)             |

---

## ✅ Checklist de Validation

- [ ] Serveur Rust opérationnel, capable de streamer en local un fichier
- [ ] Transcodage FFmpeg à la volée testé et sans latence excessive
- [ ] Endpoint Go → Rust fonctionnel (gRPC)
- [ ] Sécurisation via token temporaire ou signed URL
- [ ] Frontend avec lecteur React lisant le flux
- [ ] Test de lecture sur formats MP3, FLAC, WAV
- [ ] Lecture fluide sur navigateur mobile et desktop

---

## 💡 Bonus optionnels

- Adaptation bitrate en fonction de la bande passante (HLS ou qualité fixe ?)
- Mode preview ou "écoute collaborative" à venir
- Playlist continue pour les morceaux d’un artiste ou d’un salon

---

## 📌 Étapes Suivantes (V5)

➡ Mettre en place le **partage de fichiers et ressources audio** dans la communauté (samples, presets, tracks).

---

## 🔁 **Plan de Développement — Version 4**

### ✅ **V4.1 – Serveur Rust local (MVP)**

| Objectif                                                           | Détails |
| ------------------------------------------------------------------ | ------- |
| 🚀 Serveur HTTP `axum` ou `hyper` basique (`/stream/{id}`)         |         |
| 📦 Fonction de lecture d’un fichier `.mp3` local                   |         |
| 🎧 Envoi des headers MIME (`Content-Type`, `Accept-Ranges`, etc.)  |         |
| 🧪 Test lecture dans navigateur avec `<audio src="/stream/xxx" />` |         |

📂 Fichier clé : `stream.rs`, `server.rs`

---

### ✅ **V4.2 – Support de Lecture Progressive (HTTP Range)**

| Objectif                                               | Description |
| ------------------------------------------------------ | ----------- |
| 📍 Lecture à la position demandée (`Range: bytes=...`) |             |
| ⚡ Bufferisation dynamique pour navigation dans le flux |             |
| 🎯 Lecture fluide sur navigateur desktop/mobile        |             |

📂 Ajout dans `stream.rs`

---

### ✅ **V4.3 – Transcodage FFmpeg à la Volée**

| Objectif                                                  | Description |
| --------------------------------------------------------- | ----------- |
| 🔁 Utilisation de FFmpeg pour convertir FLAC/WAV en MP3   |             |
| ⚙️ Pipe standard (`Command + stdout`) pour streaming live |             |
| 🧪 Benchmark de la latence transcode + stream             |             |

📂 Module `ffmpeg.rs`

---

### ✅ **V4.4 – API gRPC entre Go ↔ Rust**

| Méthode gRPC             | Description                                                   |
| ------------------------ | ------------------------------------------------------------- |
| `GetStreamURL(audio_id)` | Retourne une URL temporaire signée par Go pour le Rust server |

**Objectif :**

* Le backend Go vérifie l’accès et génère une URL avec token JWT.
* Le Rust ne vérifie que le token et stream si valide.

📂 `proto/stream.proto` + `grpc_server.rs`

---

### ✅ **V4.5 – Sécurisation de l’Accès**

| Méthode                                                        | Détails |
| -------------------------------------------------------------- | ------- |
| 🔐 Token JWT court envoyé en query param ou header             |         |
| 🧱 Middleware Rust : vérifie le token avant d’ouvrir le stream |         |
| 🎯 Protection contre hotlinking ou accès non autorisé          |         |

📂 Fichier `auth.rs` ou dans `main.rs`

---

### ✅ **V4.6 – Frontend Lecteur Audio**

| Composants                                               | Description                                    |
| -------------------------------------------------------- | ---------------------------------------------- |
| 🎛️ `<AudioPlayer />`                                    | Bouton play/pause, durée, barre de progression |
| 📦 Chargement des morceaux (`GET /tracks`) depuis API Go |                                                |
| 🔗 Lecture via URL `/stream/{id}?token=...`              |                                                |

📂 `components/AudioPlayer.tsx`

---

### ✅ **V4.7 – Pré-écoute & Qualité**

| Bonus                                      | Description                                          |
| ------------------------------------------ | ---------------------------------------------------- |
| `/preview/{id}`                            | Stream 30s de la piste (découpée via FFmpeg `-t 30`) |
| Option HLS ou version 128kbit/s uniquement |                                                      |

---

## 🔍 **Résumé des Routes V4**

### 📡 Rust — HTTP

| Méthode | Route           | Description                 |
| ------: | --------------- | --------------------------- |
|     GET | `/stream/{id}`  | Stream principal (JWT req.) |
|     GET | `/preview/{id}` | Extrait 30s (optionnel)     |

### 🔌 Go — gRPC vers Rust

| RPC            | Payload        | Réponse                      |
| -------------- | -------------- | ---------------------------- |
| `GetStreamURL` | `{ audio_id }` | `{ url_signed, expires_at }` |

---

## ✅ Checklist Résumé

| Étape                              | État |
| ---------------------------------- | ---- |
| Serveur Rust avec stream basique   | ✅    |
| Support HTTP Range + bufferisation | ✅    |
| Transcodage à la volée FFmpeg      | ✅    |
| Intégration Go ↔ Rust via gRPC     | ✅    |
| Sécurisation avec token JWT        | ✅    |
| Lecteur React intégré              | ✅    |

---