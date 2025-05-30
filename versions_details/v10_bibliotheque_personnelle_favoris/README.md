# ❤️ Version 10 — Bibliothèque personnelle et favoris

🎯 Objectif : Permettre à chaque utilisateur de regrouper, organiser et retrouver facilement ses ressources préférées (fichiers, musiques, presets, tutoriels, produits) dans un espace personnel.

---

## 🧩 Fonctionnalités à Implémenter

### 📌 Ajout aux favoris
- Bouton “Favori” ou “Ajouter à ma bibliothèque” visible partout :
  - Fichiers partagés
  - Musiques de la radio
  - Tutoriels
  - Produits Talas achetés (pour mise en avant personnelle)
- Gestion des types de ressources favorisées

### 🗂️ Organisation personnalisée
- Classement par catégories : samples, presets, tutoriels, produits
- Possibilité de créer des "collections" (groupes personnalisés)
- Étiquetage personnel (tags privés)

### 📁 Espace “Ma Bibliothèque”
- Page utilisateur affichant tous les éléments favoris
- Tri, recherche, suppression
- Aperçu et lecture directe (audio, vidéo, PDF)

---

## 🔧 Stack Technique

| Composant         | Technologie                   |
|-------------------|-------------------------------|
| Backend API       | Go + PostgreSQL               |
| Base de données   | Table `favorites`, `collections` |
| Frontend          | React + Zustand               |
| Interface         | shadcn/ui                     |

---

## 🗃️ Structure Backend (extrait)

backend/
├── handlers/
│ └── favorites.go
├── models/
│ └── favorite.go
├── routes/
│ └── favorites.go
├── db/migrations/
│ └── add_favorites.sql

---

## 🧱 Tables PostgreSQL

- `favorites`: `id`, `user_id`, `resource_type`, `resource_id`, `tag`, `created_at`
- `collections`: `id`, `user_id`, `name`, `description`, `created_at`
- `collection_items`: `collection_id`, `favorite_id`

---

## 🔍 Endpoints REST

| Méthode | URL                          | Description                          |
|--------:|------------------------------|--------------------------------------|
| POST    | `/favorites`                 | Ajouter un favori                    |
| DELETE  | `/favorites/{id}`            | Supprimer un favori                  |
| GET     | `/favorites`                 | Lister tous les favoris de l’utilisateur |
| POST    | `/collections`               | Créer une collection                 |
| GET     | `/collections`               | Lister les collections               |

---

## 🖥️ Frontend

- Icône “❤️” dans chaque composant de ressource
- Page "Ma bibliothèque" :
  - Grille ou liste avec filtre par type
  - Accès rapide à la lecture ou au téléchargement
- Création de collections :
  - Interface simple de regroupement
  - Déplacement glisser-déposer (optionnel)

---

## ✅ Checklist de Validation

- [ ] Favoris ajoutables depuis tous les modules compatibles (radio, tutoriels, fichiers)
- [ ] Backend opérationnel avec permissions et sécurité
- [ ] Interface claire pour trier, supprimer, filtrer ses ressources
- [ ] Lecture directe des ressources favorites
- [ ] Collections créables, modifiables et supprimables

---

## 💡 Bonus possibles

- Export de bibliothèque (JSON, PDF)
- Suggestions de ressources similaires depuis les favoris
- Partage public (opt-in) de ses collections

---

## 🔁 **Plan de Développement — Version 10**

### ✅ **V10.1 – Modèle de données & migration**

| Objectifs                                                                         | Détails |
| --------------------------------------------------------------------------------- | ------- |
| 🧱 Tables : `favorites`, `collections`, `collection_items`                        |         |
| 📦 `favorites.resource_type` : `sample`, `preset`, `track`, `tutorial`, `product` |         |
| 🔐 Clé étrangère générique `resource_id` → vers ressource de type donnée          |         |
| 📂 Migration : `add_favorites.sql`                                                |         |

---

### ✅ **V10.2 – Ajout et suppression de favoris**

| Endpoint          | Méthode  | Auth | Description                                  |
| ----------------- | -------- | ---- | -------------------------------------------- |
| `/favorites`      | `POST`   | ✅    | Ajoute un favori (type + id de la ressource) |
| `/favorites/{id}` | `DELETE` | ✅    | Supprime un favori pour un utilisateur       |

**Backend** :

* Vérification que la ressource est valide et accessible
* Empêche les doublons

**Frontend** :

* Bouton "❤️ Ajouter à ma bibliothèque" réutilisable
* Changement d'état en temps réel (Zustand)

---

### ✅ **V10.3 – Listing & tri des favoris**

| Endpoint     | Méthode | Auth | Description                                               |
| ------------ | ------- | ---- | --------------------------------------------------------- |
| `/favorites` | `GET`   | ✅    | Liste tous les favoris de l’utilisateur, groupés par type |

**Filtrage par** :

* `?type=tutorial`, `?tag=ambient`
* Date d’ajout
* Recherche dans les titres (via jointure)

**UI :**

* Composants React `FavoriteGrid.tsx`, `FavoriteCard.tsx`, `FavoriteFilter.tsx`

---

### ✅ **V10.4 – Collections personnalisées**

| Endpoint                   | Méthode  | Auth | Description                                |
| -------------------------- | -------- | ---- | ------------------------------------------ |
| `/collections`             | `POST`   | ✅    | Crée une nouvelle collection               |
| `/collections`             | `GET`    | ✅    | Liste les collections de l’utilisateur     |
| `/collections/{id}/add`    | `POST`   | ✅    | Ajoute un favori existant à une collection |
| `/collections/{id}/remove` | `DELETE` | ✅    | Retire un favori d’une collection          |

**Frontend** :

* Composant `CollectionBuilder.tsx`
* Ajout par glisser-déposer (optionnel avec `dnd-kit`)
* Renommage et suppression possible

---

### ✅ **V10.5 – Accès et lecture directe**

| Objectif                                                               | Détails |
| ---------------------------------------------------------------------- | ------- |
| 📖 Rendu Markdown des tutoriels favoris                                |         |
| 🎧 Lecture audio directe pour tracks et samples                        |         |
| 📁 Téléchargement rapide pour fichiers et presets                      |         |
| 📦 Lien rapide vers la ressource source (page produit, tutoriel, etc.) |         |

---

## 🔍 **Résumé des Routes REST V10**

| Méthode | Endpoint                   | Description                           |
| ------: | -------------------------- | ------------------------------------- |
|    POST | `/favorites`               | Ajouter une ressource à ses favoris   |
|  DELETE | `/favorites/{id}`          | Supprimer un favori                   |
|     GET | `/favorites?type=tutorial` | Lister les favoris par type           |
|    POST | `/collections`             | Créer une collection                  |
|     GET | `/collections`             | Lister ses collections                |
|    POST | `/collections/{id}/add`    | Ajouter un favori dans une collection |
|  DELETE | `/collections/{id}/remove` | Retirer un favori d'une collection    |

---

## ✅ **Checklist Résumée**

| Composant                                | État |
| ---------------------------------------- | ---- |
| Ajout/suppression de favoris multi-type  | ✅    |
| Affichage unifié "Ma bibliothèque"       | ✅    |
| Lecture directe audio/vidéo/pdf          | ✅    |
| Collections personnalisées & tags privés | ✅    |
| Tri, filtre, recherche intégrés          | ✅    |

---

## 💡 Bonus possibles

| Fonction                 | Description                                  |
| ------------------------ | -------------------------------------------- |
| 📤 Export JSON/PDF       | Télécharger sa bibliothèque                  |
| 🧠 Suggestions auto      | "Vous aimerez aussi" (via tags partagés)     |
| 🌐 Collections publiques | Pages partagées optionnelles (liens publics) |

---
