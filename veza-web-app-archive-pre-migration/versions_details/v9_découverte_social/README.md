# 🌍 Version 9 — Découverte communautaire / streaming social

🎯 Objectif : Offrir une expérience d’écoute collaborative des musiques et morceaux partagés sur la plateforme, via une radio communautaire avec suggestions, playlists et interactions sociales.

---

## 🧩 Fonctionnalités à Implémenter

### 📻 Radio communautaire
- Flux musical continu basé sur les fichiers partagés en public
- Lecture automatique avec transition douce entre morceaux
- Moteur de suggestion simple basé sur :
  - Popularité (likes, écoutes)
  - Tags communs
  - Nouveautés

### 💞 Interaction utilisateur
- Bouton “J’aime” (like) pour chaque morceau
- Compteur d’écoutes et de likes
- Commentaires optionnels sous chaque morceau (via messagerie ou fil de discussion)

### 🎼 Playlists & découverte
- Génération de playlists thématiques ou aléatoires (tag, style, humeur)
- Interface "Découverte du jour"
- Navigation possible sans interaction (mode passif)

---

## 🔧 Stack Technique

| Composant             | Technologie                        |
|------------------------|------------------------------------|
| Backend API            | Go + PostgreSQL                   |
| Module de streaming    | Rust (utilisé depuis la V4)       |
| Frontend               | React + Audio player + Zustand    |
| Moteur de recommandation | Logique simple interne (likes, tags) |

---

## 🗃️ Structure Backend (extrait)

backend/
├── handlers/
│ ├── stream_feed.go
│ └── likes.go
├── models/
│ ├── stream_item.go
│ └── like.go
├── routes/
│ └── stream.go
├── db/migrations/
│ └── add_streaming_feed.sql

---

## 🧱 Tables PostgreSQL

- `stream_feed`: `id`, `file_id`, `played_at`, `like_count`, `play_count`
- `likes`: `user_id`, `file_id`, `timestamp`
- `comments` (optionnel) : `id`, `file_id`, `user_id`, `content`, `created_at`

---

## 🔍 Endpoints REST / WebSocket

| Méthode | URL                        | Description                        |
|--------:|----------------------------|------------------------------------|
| GET     | `/stream/radio`           | Récupère le prochain morceau       |
| POST    | `/stream/like/{file_id}`  | Aimer un morceau                   |
| GET     | `/stream/playlist?tag=fx` | Playlist filtrée par tag          |
| GET     | `/stream/discover`        | Flux de découverte du jour         |

---

## 🖥️ Frontend

- Lecteur audio persisté (affiché même en navigation)
- Panneau latéral “Découverte” avec liste des morceaux à venir
- Bouton Like, compteur d’écoutes
- Tag + nom d’uploader affiché
- Affichage optionnel de commentaires

---

## ✅ Checklist de Validation

- [ ] Génération automatique du flux de lecture audio
- [ ] Intégration du player frontend fonctionnelle
- [ ] Gestion des likes par utilisateur
- [ ] Playlist dynamique avec filtres (tag, date, popularité)
- [ ] UI fluide et accessible (mobile-friendly)
- [ ] Intégration fluide avec le module de streaming Rust

---

## 💡 Bonus possibles

- Crossfade audio entre morceaux (Rust)
- Bouton “ajouter à ma bibliothèque” ou "favoris"
- Mode radio silencieuse (lecture mais pas visible dans profil public)

---

## 📌 Étapes Suivantes (V10)

➡ Ajouter une **bibliothèque personnelle** permettant à chaque utilisateur de regrouper ses ressources et fichiers favoris.

---

## 🔁 **Plan de Développement — Version 9**

### ✅ **V9.1 — Backend de flux communautaire**

| Objectifs                                                                                | Détails |
| ---------------------------------------------------------------------------------------- | ------- |
| 🧱 Créer la table `stream_feed` pour suivre les morceaux diffusés                        |         |
| 🔁 Implémenter une logique simple : fichiers publics triés par `like_count`, `played_at` |         |
| 🔍 Endpoint :                                                                            |         |
| `GET /stream/radio` → retourne le *prochain morceau* basé sur critères                   |         |
| 📂 Migration : `add_streaming_feed.sql`                                                  |         |

---

### ✅ **V9.2 — Likes & statistiques d’écoute**

| Endpoint                   | Méthode | Auth | Description                                         |
| -------------------------- | ------- | ---- | --------------------------------------------------- |
| `/stream/like/{file_id}`   | `POST`  | ✅    | Like d’un fichier par un utilisateur                |
| `/stream/playlist?tag=xxx` | `GET`   | ❌    | Retourne une playlist filtrée par tag ou popularité |
| `/stream/discover`         | `GET`   | ❌    | Playlist quotidienne générée dynamiquement          |

📦 Table `likes`: clé `(user_id, file_id)` + timestamp
📦 `stream_feed.play_count` auto-incrémenté à chaque lecture

---

### ✅ **V9.3 — Intégration au module de streaming Rust (V4)**

| Tâches                                                                    | Détails |
| ------------------------------------------------------------------------- | ------- |
| 🧩 Appels au serveur Rust `/stream/{id}?token=...` depuis `/stream/radio` |         |
| 🎧 Pré-sélection Go → réponse Rust : un fichier = un flux audio sécurisé  |         |
| 🧠 Option future : `crossfade` en Rust avec FFmpeg `concat + afade`       |         |

---

### ✅ **V9.4 — Frontend : Radio & Découverte**

| Composants                    | Détails                                          |
| ----------------------------- | ------------------------------------------------ |
| 🎛️ `RadioPlayer.tsx`         | Player audio persistent (lecture automatique)    |
| 🎶 `DiscoverPanel.tsx`        | Liste latérale avec les prochains titres à venir |
| ❤️ `LikeButton.tsx`           | Intégré au player ou dans les cartes morceau     |
| 🧠 `StreamStore.ts` (Zustand) | Gestion de la file, écoute en cours, likes       |

---

### ✅ **V9.5 — Playlists personnalisées**

| Endpoint                        | Méthode | Description                                       |
| ------------------------------- | ------- | ------------------------------------------------- |
| `/stream/playlist?tag=ambient`  | `GET`   | Retourne une liste de morceaux par tag            |
| `/stream/playlist?popular=true` | `GET`   | Trié par likes ou écoutes                         |
| `/stream/discover`              | `GET`   | Sélection auto (nouveautés ou profil utilisateur) |

**Frontend :**

* Liste type "Netflix" avec lecteur inline ou bouton "ajouter à la file"
* Possibilité de marquer comme favori

---

### ✅ **V9.6 — (Optionnel) Commentaires**

| Table                                                        | `comments` |
| ------------------------------------------------------------ | ---------- |
| Champs : `id`, `file_id`, `user_id`, `content`, `created_at` |            |

**Frontend** :
Zone de commentaires sous chaque carte morceau (si activée)

---

## 🔍 **Résumé des Routes REST V9**

| Méthode | Endpoint                 | Auth | Description                       |
| ------: | ------------------------ | ---- | --------------------------------- |
|     GET | `/stream/radio`          | ❌    | Retourne le prochain morceau      |
|    POST | `/stream/like/{file_id}` | ✅    | Aime un fichier audio             |
|     GET | `/stream/playlist`       | ❌    | Playlist par tag/popularité       |
|     GET | `/stream/discover`       | ❌    | Playlist de découverte du jour    |
|    POST | `/comments`              | ✅    | Ajoute un commentaire (optionnel) |

---

## ✅ **Checklist Résumée**

| Composant                             | État |
| ------------------------------------- | ---- |
| Génération dynamique du flux audio    | ✅    |
| Intégration complète au module Rust   | ✅    |
| Lecteur frontend persistant           | ✅    |
| Likes et compteurs fonctionnels       | ✅    |
| Playlists thématiques opérationnelles | ✅    |
| Commentaires intégrables (option)     | ✅    |

---
