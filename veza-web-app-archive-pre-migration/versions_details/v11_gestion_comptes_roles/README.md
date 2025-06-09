# 🛡️ Version 11 — Gestion avancée des comptes et rôles

🎯 Objectif : Mettre en place un système de rôles pour mieux structurer les permissions au sein de la plateforme Talas (utilisateur, artiste, modérateur, administrateur), et permettre une modération communautaire minimale.

---

## 🧩 Fonctionnalités à Implémenter

### 👤 Rôles utilisateur
- Création des rôles :
  - `user` (utilisateur simple)
  - `artist` (utilisateur créateur, peut publier plus de contenus)
  - `moderator` (peut modérer les contenus et échanges)
  - `admin` (accès complet à l’administration)
- Association d’un rôle à chaque utilisateur (champ `role` dans la table `users`)
- Middleware de contrôle d’accès (RBAC)

### 🛠️ Interface de gestion
- Tableau de bord admin :
  - Vue sur les utilisateurs, rôles, activités
  - Modification de rôle
  - Désactivation de compte (soft delete)
- Outils de modération :
  - Visualisation des contenus signalés
  - Suppression/modification par un modérateur

### 🚩 Système de signalement
- Bouton “Signaler” sur :
  - Fichiers
  - Messages
  - Annonces de troc
- Enregistrement en base avec motif, date et auteur du signalement

---

## 🔧 Stack Technique

| Composant        | Technologie                   |
|------------------|-------------------------------|
| Backend API      | Go (middleware + handlers RBAC) |
| Base de données  | PostgreSQL (`users`, `reports`) |
| Frontend         | React + shadcn/ui             |

---

## 🗃️ Structure Backend (extrait)

backend/
├── middleware/
│ └── rbac.go
├── handlers/
│ └── admin.go
│ └── reports.go
├── models/
│ └── user.go
│ └── report.go
├── db/migrations/
│ └── add_roles_and_reports.sql

---

## 🧱 Tables PostgreSQL

- `users`: ajout du champ `role`
- `reports`: `id`, `reporter_id`, `resource_type`, `resource_id`, `reason`, `status`, `created_at`

---

## 🔍 Endpoints REST

| Méthode | URL                       | Description                          |
|--------:|---------------------------|--------------------------------------|
| GET     | `/admin/users`            | Liste des utilisateurs (admin only)  |
| PUT     | `/admin/users/{id}/role`  | Modifier un rôle                     |
| GET     | `/reports`                | Liste des contenus signalés          |
| POST    | `/report`                 | Signaler un contenu                  |

---

## 🖥️ Frontend

- Interface d’administration :
  - Liste des utilisateurs et modification de rôle
  - Filtres par rôle, activité, date d’inscription
- Interface de modération :
  - Liste des signalements
  - Lien vers les contenus signalés
  - Actions disponibles (supprimer, archiver, marquer comme traité)

---

## ✅ Checklist de Validation

- [ ] Middleware RBAC opérationnel sur les routes protégées
- [ ] Ajout/modification des rôles via API sécurisée
- [ ] Dashboard admin fonctionnel avec gestion des utilisateurs
- [ ] Système de signalement intégré dans les modules existants
- [ ] Visualisation claire des contenus signalés
- [ ] Tests de permissions pour chaque type de rôle

---

## 💡 Bonus possibles

- Journal d’activité des modérateurs
- Restrictions spécifiques (upload max pour `user`, illimité pour `artist`)
- Statistiques sur les signalements et actions modératives

---

## 📌 Étapes Suivantes (V12)

➡ Compiler l’application Talas en une **version standalone multiplateforme** à l’aide de **Tauri**, avec un backend intégré.

---

## 🔁 **Plan de Développement — Version 11**

### ✅ **V11.1 – Migration & rôles dans la table `users`**

| Objectifs                                                             | Détails |
| --------------------------------------------------------------------- | ------- |
| 🧱 Ajouter le champ `role TEXT DEFAULT 'user'` dans `users`           |         |
| 🔐 Enum logique (`user`, `artist`, `moderator`, `admin`)              |         |
| 📂 Migration : `add_roles_and_reports.sql` avec contraintes possibles |         |

---

### ✅ **V11.2 – Middleware RBAC et vérification de rôle**

| Fichier                | Détail                                          |
| ---------------------- | ----------------------------------------------- |
| `middleware/rbac.go`   | Intercepteur basé sur les rôles extraits du JWT |
| `context.UserRole`     | Ajout via middleware d’auth                     |
| `RequireRole("admin")` | Fonction réutilisable par route                 |

**Fonctionnalités clés** :

* Refus avec 403 si rôle insuffisant
* Logging des accès bloqués (option)

---

### ✅ **V11.3 – Endpoints admin : gestion des rôles**

| Endpoint                    | Méthode | Rôle requis | Description                              |
| --------------------------- | ------- | ----------- | ---------------------------------------- |
| `/admin/users`              | `GET`   | `admin`     | Lister les comptes                       |
| `/admin/users/{id}/role`    | `PUT`   | `admin`     | Modifier le rôle d’un utilisateur        |
| `/admin/users/{id}/disable` | `PUT`   | `admin`     | Désactiver un compte (`enabled = false`) |

---

### ✅ **V11.4 – Système de signalement**

#### 📑 Table `reports`

| Colonne         | Type                                              |
| --------------- | ------------------------------------------------- |
| `id`            | UUID                                              |
| `reporter_id`   | UUID                                              |
| `resource_type` | `TEXT` (enum : `file`, `message`, `listing`)      |
| `resource_id`   | UUID                                              |
| `reason`        | `TEXT`                                            |
| `status`        | `TEXT` (enum : `pending`, `resolved`, `archived`) |
| `created_at`    | `TIMESTAMP`                                       |

#### 🔍 Endpoints REST

| Méthode | URL             | Rôle requis | Description                      |
| ------: | --------------- | ----------- | -------------------------------- |
|  `POST` | `/report`       | `auth`      | Créer un signalement             |
|   `GET` | `/reports`      | `moderator` | Voir les signalements en attente |
|   `PUT` | `/reports/{id}` | `moderator` | Marquer comme résolu/archivé     |

---

### ✅ **V11.5 – Dashboard frontend**

**Admin UI**

* Composants :

  * `UserTable.tsx`
  * `RoleSelector.tsx`
  * `DisableUserButton.tsx`

**Modération UI**

* Composants :

  * `ReportList.tsx`
  * `ReportDetails.tsx`
  * `ModerationAction.tsx`

**Bouton de signalement**

* Intégrable à :

  * `SharedFileCard.tsx`
  * `ChatMessage.tsx`
  * `ListingCard.tsx`

---

## 🔍 **Résumé des Routes REST V11**

| Méthode | Endpoint                    | Rôle requis     | Description                 |
| ------: | --------------------------- | --------------- | --------------------------- |
|     GET | `/admin/users`              | admin           | Liste tous les utilisateurs |
|     PUT | `/admin/users/{id}/role`    | admin           | Modifier rôle               |
|     PUT | `/admin/users/{id}/disable` | admin           | Désactiver le compte        |
|    POST | `/report`                   | user            | Signaler un contenu         |
|     GET | `/reports`                  | moderator/admin | Voir tous les signalements  |
|     PUT | `/reports/{id}`             | moderator/admin | Traiter un signalement      |

---

## ✅ **Checklist de Validation**

| Élément                                      | État |
| -------------------------------------------- | ---- |
| Champ `role` et middleware RBAC fonctionnels | ✅    |
| Endpoints admin sécurisés                    | ✅    |
| Signalements stockés avec typage ressource   | ✅    |
| Modération visible dans dashboard            | ✅    |
| Permissions validées via tests unitaires     | ✅    |

---

## 💡 Bonus potentiels

| Idée                      | Description                                          |
| ------------------------- | ---------------------------------------------------- |
| 📘 Journal d’activité     | Log automatique : changements de rôle, désactivation |
| 🔢 Statistiques           | Nb de signalements, actions modératives par mois     |
| 🧾 Historique utilisateur | Liste des signalements faits / reçus                 |

---
