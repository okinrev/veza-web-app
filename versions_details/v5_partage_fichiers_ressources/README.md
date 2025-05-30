# 📁 Version 5 — Partage de fichiers et ressources

🎯 Objectif : Permettre aux utilisateurs de partager des fichiers audio (samples, presets, maquettes), avec contrôle des droits d’accès, tagging, et visualisation communautaire.

---

## 🧩 Fonctionnalités à Implémenter

### ⬆️ Upload / ⬇️ Téléchargement de Ressources
- Upload de fichiers audio ou presets (MP3, WAV, FLAC, .fxp, .zip, etc.)
- Téléchargement avec permissions selon visibilités : privé / public / entre amis
- Association d’un fichier à des métadonnées (type, tags, description)

### 🔖 Tagging & Classement
- Ajout de tags libres ou via liste prédéfinie (e.g. "drum", "synth", "ambient")
- Type de ressource : sample, preset, mix, maquette
- Date d’upload, nombre de téléchargements, nom d’auteur

### 🌐 Partage Communautaire
- Page publique de partage (type "library")
- Filtres par type, date, tags, popularité
- Affichage des fichiers en ligne avec aperçu (lecture directe si audio)

---

## 🔧 Stack Technique

| Composant        | Technologie                              |
|------------------|-------------------------------------------|
| Backend API      | Go (auth, permissions, metadata)          |
| Stockage fichiers| Nextcloud via WebDAV ou ZFS               |
| Base de données  | PostgreSQL (tables `files`, `tags`)       |
| Frontend         | React + shadcn/ui + Axios                 |

---

## 🗃️ Structure Backend (extrait)

backend/
├── handlers/
│ └── share.go
├── routes/
│ └── share.go
├── models/
│ └── file.go
├── db/migrations/
│ └── add_shared_files.sql

---

## 🧱 Tables PostgreSQL

- `shared_files` : `id`, `user_id`, `filename`, `type`, `tags[]`, `description`, `visibility`, `uploaded_at`
- `downloads` : `file_id`, `user_id`, `timestamp`

---

## 🔍 Endpoints REST

| Méthode | URL                        | Description                         |
|--------:|----------------------------|-------------------------------------|
| POST    | `/share/upload`            | Upload d’un fichier partagé         |
| GET     | `/share/library`           | Liste des fichiers publics          |
| GET     | `/share/file/{id}`         | Télécharger ou afficher un fichier  |
| GET     | `/share/search?tag=xxx`    | Recherche par tag ou type           |

---

## 🖥️ Frontend

- Page "Partage" avec :
  - Filtres (type, tag, popularité)
  - Carte de fichier (nom, extension, durée si audio)
  - Bouton "Écouter" / "Télécharger"
  - Formulaire d’upload avec choix de visibilité et tags

---

## ✅ Checklist de Validation

- [ ] Upload fonctionnel et sécurisé de fichiers audio ou presets
- [ ] Stockage avec nom unique, chemin utilisateur
- [ ] Contrôle d’accès respecté : public / privé / amis
- [ ] Recherche de fichiers via tags ou type
- [ ] Affichage frontend propre et rapide
- [ ] Lecteur audio intégré pour les fichiers partagés
- [ ] Statistiques de téléchargement par fichier

---

## 💡 Bonus possibles

- Preview audio (limite 30s en mode public)
- Like / bookmark communautaire
- Système de validation (modération)

---

## 📌 Étapes Suivantes (V6)

➡ Créer un **moteur de recherche interne** global pour tous les contenus partagés.

---

## 🔁 **Plan de Développement — Version 5**

### ✅ **V5.1 — Structure BDD & Authentification**

| Objectifs                                                    | Détails |
| ------------------------------------------------------------ | ------- |
| 🧱 Création des tables `shared_files`, `downloads`           |         |
| 🔐 Vérification JWT obligatoire pour upload                  |         |
| 🧑‍🤝‍🧑 Champs `visibility`: `private`, `public`, `friends` |         |
| 🔁 Intégration dans modèle `file.go` et migrations SQL       |         |

---

### ✅ **V5.2 — Upload sécurisé et métadonnées**

| Endpoint        | Méthode | Auth | Description                                                          |
| --------------- | ------- | ---- | -------------------------------------------------------------------- |
| `/share/upload` | `POST`  | ✅    | Upload d’un fichier partagé avec tags, type, description, visibilité |

**Fonctionnalités :**

* Format accepté : MP3, FLAC, WAV, ZIP, FXP…
* Attribution automatique à l’utilisateur
* Génération de nom de fichier unique
* Upload vers `ZFS` ou `Nextcloud` via WebDAV

---

### ✅ **V5.3 — Affichage communautaire (library publique)**

| Endpoint         | Méthode | Auth | Description                                   |
| ---------------- | ------- | ---- | --------------------------------------------- |
| `/share/library` | `GET`   | ❌    | Liste des fichiers `public` visibles par tous |

**Backend :**

* Filtrage par visibilité
* Pagination (par défaut 20)

**Frontend :**

* Affichage avec cartes : titre, auteur, date, type, tags

---

### ✅ **V5.4 — Visualisation & Téléchargement**

| Endpoint           | Méthode | Auth | Description                                            |
| ------------------ | ------- | ---- | ------------------------------------------------------ |
| `/share/file/{id}` | `GET`   | ✅    | Récupère et sert le fichier si l’utilisateur y a droit |

**Logique d'accès :**

* ✅ Public → accessible à tous
* 🔒 Privé → seulement propriétaire
* 👥 Friends → via table `friends` (optionnel à venir)

**Fonction :**

* Ajout d’un enregistrement `downloads` pour statistiques

---

### ✅ **V5.5 — Recherche avancée & filtrage**

| Endpoint                            | Méthode | Auth | Description                               |
| ----------------------------------- | ------- | ---- | ----------------------------------------- |
| `/share/search?tag=xxx&type=sample` | `GET`   | ❌    | Filtres combinés sur type, tag, nom, date |

**Backend :**

* Recherche simple (ILIKE) ou trigram PostgreSQL (optionnel)
* Tri par popularité (via `downloads`)

---

### ✅ **V5.6 — Intégration frontend complète**

| Composants                   | Détails                                            |
| ---------------------------- | -------------------------------------------------- |
| 🔍 `LibraryView.jsx`         | Liste de fichiers partagés, filtres interactifs    |
| 🎧 `AudioCard.jsx`           | Lecture inline avec lecteur HTML5 ou ReactPlayer   |
| 📤 `UploadForm.jsx`          | Formulaire complet : drag'n'drop, tags, visibilité |
| 🧠 `shareStore.ts` (Zustand) | État : fichiers visibles, tags actifs, recherche   |

---

## 🔍 **Résumé des Routes REST V5**

| Méthode | Endpoint                | Auth | Description                   |
| ------: | ----------------------- | ---- | ----------------------------- |
|    POST | `/share/upload`         | ✅    | Upload d’un fichier partagé   |
|     GET | `/share/library`        | ❌    | Liste publique de fichiers    |
|     GET | `/share/file/{id}`      | ✅    | Récupération / téléchargement |
|     GET | `/share/search?tag=xxx` | ❌    | Recherche par tag/type        |

---

## ✅ **Checklist Fonctionnelle Résumée**

| Composant                             | État |
| ------------------------------------- | ---- |
| Upload sécurisé avec visibilité       | ✅    |
| Stockage structuré (ZFS/Nextcloud)    | ✅    |
| Affichage de la bibliothèque publique | ✅    |
| Recherche par tag/type                | ✅    |
| Lecture audio inline (HTML5)          | ✅    |
| Contrôle d'accès (visibility)         | ✅    |
| Statistiques téléchargements          | ✅    |

---
