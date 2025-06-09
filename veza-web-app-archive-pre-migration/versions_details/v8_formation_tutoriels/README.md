# 📚 Version 8 — Module de formation et tutoriels

🎯 Objectif : Fournir un espace pédagogique aux utilisateurs pour apprendre à utiliser leurs produits Talas, découvrir des techniques de production audio, et accéder à des tutoriels internes ou partenaires.

---

## 🧩 Fonctionnalités à Implémenter

### 🧠 Accès aux contenus de formation
- Pages listant :
  - Tutoriels internes (vidéos, textes, guides interactifs)
  - Ressources externes (via partenaires ou liens)
- Classement par :
  - Produit concerné (microphone, carte son, etc.)
  - Thématique (réglages, enregistrement, mixage, réparation...)

### 📝 Formats pris en charge
- Markdown enrichi (pour guides internes)
- Vidéo embarquée (YouTube, Peertube ou fichiers .mp4)
- PDF téléchargeables
- Liens externes vers plateformes partenaires (BandLab, Skillshare...)

### 🎓 Progression utilisateur (optionnel)
- Marquage "vu / non vu"
- Historique des ressources consultées
- Suggestions de parcours (débutant / avancé)

---

## 🔧 Stack Technique

| Composant        | Technologie                       |
|------------------|-----------------------------------|
| Backend API      | Go (modules `tutorials`, `tags`)  |
| Base de données  | PostgreSQL                        |
| Frontend         | React + shadcn/ui                 |
| Formats supportés| Markdown, MP4, PDF, liens externes|

---

## 🗃️ Structure Backend (extrait)

backend/
├── handlers/
│ └── tutorials.go
├── models/
│ └── tutorial.go
├── routes/
│ └── tutorials.go
├── db/migrations/
│ └── add_tutorials.sql

---

## 🧱 Tables PostgreSQL

- `tutorials`: `id`, `title`, `slug`, `description`, `type`, `url`, `markdown`, `tags[]`, `product`, `author`, `created_at`
- `user_tutorials`: `user_id`, `tutorial_id`, `viewed`, `progress`

---

## 🔍 Endpoints REST

| Méthode | URL                   | Description                           |
|--------:|-----------------------|---------------------------------------|
| GET     | `/tutorials`          | Lister tous les tutoriels             |
| GET     | `/tutorials/{id}`     | Contenu d’un tutoriel                 |
| POST    | `/tutorials/view`     | Marquer un tutoriel comme consulté    |
| GET     | `/tutorials/by-tag`   | Recherche par thématique              |

---

## 🖥️ Frontend

- Page "Formation" avec liste filtrable :
  - Tags, produits, niveaux (débutant, avancé)
- Carte de tutoriel : aperçu, durée, format
- Page de lecture :
  - Vidéo intégrée / markdown affiché / bouton téléchargement
  - Bouton "J’ai terminé ce tutoriel"
- Profil utilisateur → onglet "Mes tutoriels"

---

## ✅ Checklist de Validation

- [ ] Backend fonctionnel pour lister, lire et enregistrer les vues de tutoriels
- [ ] Upload et lecture de vidéos intégrés ou externes
- [ ] Lecture propre des fichiers Markdown
- [ ] Frontend ergonomique et responsive
- [ ] Historique utilisateur stocké et mis à jour
- [ ] Tri par thématique et type de support opérationnel

---

## 💡 Bonus possibles

- Badge de progression ou certification utilisateur
- Quiz après chaque module (avec correction automatique)
- Tutoriels communautaires validés par l’équipe Talas

---

## 📌 Étapes Suivantes (V9)

➡ Mettre en place une **radio communautaire** pour découvrir les musiques partagées par les utilisateurs et encourager l’écoute collaborative.

---

## 🔁 **Plan de Développement — Version 8**

### ✅ **V8.1 — Base de données et modèle de tutoriels**

| Objectifs                                                        | Détails |
| ---------------------------------------------------------------- | ------- |
| 🧱 Création des tables `tutorials` et `user_tutorials`           |         |
| 🎯 Champs : `type`, `product`, `tags[]`, `markdown`, `url`, etc. |         |
| 📂 Migration SQL `add_tutorials.sql`                             |         |
| 📦 Modèle `models/tutorial.go` avec gestion multi-format         |         |

---

### ✅ **V8.2 — Listing & lecture de tutoriels**

| Endpoints         | Méthodes | Auth | Description                              |
| ----------------- | -------- | ---- | ---------------------------------------- |
| `/tutorials`      | `GET`    | ❌    | Liste tous les tutoriels avec filtres    |
| `/tutorials/{id}` | `GET`    | ❌/✅  | Affiche le contenu complet d’un tutoriel |

**Backend :**

* Support des formats : `markdown`, `mp4`, `pdf`, `external_link`
* Tri par date, produit, tag, type

**Frontend :**

* `TutorialCard.tsx` : titre, type, durée estimée
* `TutorialList.tsx` : filtres (tags, produits, niveaux)
* `TutorialView.tsx` : rendu markdown ou player vidéo intégré

---

### ✅ **V8.3 — Suivi utilisateur (progression)**

| Endpoint            | Méthode | Auth | Description                                   |
| ------------------- | ------- | ---- | --------------------------------------------- |
| `/tutorials/view`   | `POST`  | ✅    | Marque un tutoriel comme vu par l’utilisateur |
| `/tutorials/by-tag` | `GET`   | ❌    | Liste des tutoriels filtrés par thème         |

**Backend :**

* Écriture dans `user_tutorials`
* Option : champs `progress` si tutoriels longs/structurés

**Frontend :**

* Bouton “J’ai terminé ce tutoriel”
* Affichage dans `Profile → Mes tutoriels`

---

### ✅ **V8.4 — Rendu Markdown & fichiers embarqués**

**Fonctionnalités** :

* Lecture directe des guides au format `.md`
* Liens vers PDF téléchargeables
* Players vidéo (YouTube, MP4, Peertube)

📂 Utilisation de `goldmark` côté Go pour `markdown → HTML` sécurisé

---

### ✅ **V8.5 — Intégration frontend complète**

| Composants                      | Détails                                |
| ------------------------------- | -------------------------------------- |
| 📚 `FormationPage.tsx`          | Vue globale avec tutoriels filtrables  |
| 📘 `TutorialView.tsx`           | Page de tutoriel (vidéo ou markdown)   |
| 🧠 `TutorialStore.ts` (Zustand) | Suivi de vue, progression, liste       |
| 🧑‍🎓 `UserProfile.tsx`         | Onglet “Mes tutoriels” avec historique |

---

## 🔍 **Résumé des Routes REST V8**

| Méthode | Endpoint            | Auth | Description                  |
| ------: | ------------------- | ---- | ---------------------------- |
|     GET | `/tutorials`        | ❌    | Liste globale                |
|     GET | `/tutorials/{id}`   | ❌/✅  | Tutoriel complet             |
|    POST | `/tutorials/view`   | ✅    | Marquer comme vu             |
|     GET | `/tutorials/by-tag` | ❌    | Recherche par tag/thématique |

---

## ✅ **Checklist Résumée**

| Composant                                | État |
| ---------------------------------------- | ---- |
| Listing multi-format                     | ✅    |
| Rendu markdown + player vidéo            | ✅    |
| Enregistrement “vu” + profil utilisateur | ✅    |
| Filtres frontend (produit, type, tag)    | ✅    |
| UI responsive + ergonomique              | ✅    |
| Historique personnel / suggestions       | ✅    |

---
