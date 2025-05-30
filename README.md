# 🎙️ Talas — Écosystème Audio Éthique, Modulaire et Accessible

**Talas** est une plateforme audio complète dédiée aux artistes indépendants. Elle combine matériel audio réparable (microphones, accessoires) et une application communautaire web pour le partage, la formation, le streaming et la collaboration musicale.

---

## 🚀 Vision

> Réconcilier **qualité professionnelle**, **évolutivité** et **éthique environnementale** dans l’audio.

- 🎧 Du matériel **modulaire**, **open-source** et **durable**
- 🌐 Une plateforme web pour **partager**, **apprendre**, **échanger** et **créer**
- 💚 Un modèle **anti-obsolescence** avec réparation, mise à niveau et troc

---

## 🧱 Stack Technique

| Côté          | Technologie principale                         |
|---------------|-------------------------------------------------|
| Backend       | Go (API REST) + PostgreSQL + JWT + Redis        |
| Modules       | Rust (chat WebSocket, streaming audio performant) |
| Frontend      | HTML/JS (Vite/React à venir)                    |
| Base de données | PostgreSQL                                    |
| Authentification | JWT + Bcrypt                                |
| Déploiement   | Conteneurs Incus / Docker                      |

---

## 🗂️ Architecture

```

talas/
├── backend/                  # API REST en Go + modules Rust
│   ├── routes/              # Routes API (auth, user, fichiers, etc.)
│   ├── handlers/            # Logique métier Go
│   ├── models/              # Schémas Go + SQL
│   ├── db/                  # Migrations SQL
│   ├── utils/               # JWT, hash, signed URLs
│   ├── modules/
│   │   ├── chat\_server/     # Rust WebSocket Chat (rooms & DM)
│   │   └── stream\_server/   # Rust Streaming Audio
├── frontend/                # Pages HTML et scripts JS
├── versions\_details/        # README des 12 versions progressives
├── .env                     # Variables d’environnement (non commité)
└── README.md                # Ce fichier

````

---

## ⚙️ Installation locale

### 1. Cloner le projet

```bash
git clone https://github.com/ton-compte/talas.git
cd talas
````

### 2. Configuration

Créer un fichier `.env` dans `backend/` :

```env
JWT_SECRET="votre_clé_secrète"
DATABASE_URL="postgres://user:password@localhost:5432/talas_db"
```

Assurez-vous d’avoir PostgreSQL avec une base `talas_db` configurée.

### 3. Lancer le backend

```bash
cd backend
go run main.go
```

Les routes seront disponibles sur `http://localhost:8080`.

### 4. Lancer les modules Rust

#### Streaming Audio

```bash
cd backend/modules/stream_server
cargo run
```

#### Chat WebSocket

```bash
cd backend/modules/chat_server
cargo run
```

---

## 📌 Fonctionnalités par Version

| Version | Fonctionnalité                               | Statut    |
| ------- | -------------------------------------------- | --------- |
| V1      | Authentification / Base utilisateurs         | ✅ Fait    |
| V2      | Fichiers / Produits / Documentation          | ✅ Fait    |
| V3      | Chat & messagerie temps réel                 | ✅ Fait    |
| V4      | Streaming audio intégré                      | ✅ Fait    |
| V5      | Partage de fichiers & ressources             | ✅ Fait    |
| V6      | Moteur de recherche                          | ⬜ À faire |
| V7      | Plateforme de troc (matériel, presets, etc.) | ⬜ À faire |
| V8      | Formation (guides, tutos, parcours)          | ⬜ À faire |
| V9      | Découverte sociale (likes, écoutes, feed)    | ⬜ À faire |
| V10     | Bibliothèque personnelle (favoris, presets)  | ⬜ À faire |
| V11     | Gestion des rôles & comptes avancés          | ⬜ À faire |
| V12     | Application standalone (Electron ou PWA)     | ⬜ À faire |

---

## 🧪 Tests & Sécurité

* ✅ JWT sécurisé avec Bcrypt
* 🔒 Uploads protégés par signed URLs
* 🚧 Tests automatisés à venir
* 🚧 CI/CD à mettre en place

---

## 📄 Documentation

* `api_doc.md` : endpoints et routes
* `versions_details/` : détails par version
* `main_doc.md`, `Liste Complète des Aspects.md` : vision et stratégie

---

## 🤝 Contribution

Tu veux participer ? Clone, code, propose une PR !
Suggestions, critiques, améliorations sont les bienvenues ❤️

---

## 🧠 À venir

* Migration vers une **interface React + Zustand**
* **Déploiement CI/CD** automatisé avec tests
* Intégration complète de **Nextcloud + AudioGridder**
* Application mobile ou **PWA standalone (V12)**

---

## 📫 Contact

Projet fondé par **Mark Milo**
📧 [contact@talas.fr](mailto:contact@talas.fr)
🌐 [https://talas.fr](https://talas.fr) *(à venir)*

---

**Talas** — L’audio open, évolutif et durable.

---