# 🔍 Version 6 — Moteur de recherche interne

🎯 Objectif : Permettre aux utilisateurs de rechercher efficacement des ressources (samples, presets, morceaux, utilisateurs, fichiers partagés) dans l’ensemble de la plateforme communautaire.

---

## 🧩 Fonctionnalités à Implémenter

### 🔎 Recherche par mot-clé
- Champ de recherche global (sur le dashboard et la bibliothèque)
- Interrogation par nom, description, tag, type de fichier
- Auto-complétion (optionnel) pour tags et types connus

### 🏷️ Filtres dynamiques
- Filtrage par :
  - Type : sample, preset, track, doc
  - Tags : ambient, drum, pad, fx...
  - Auteur / uploader
  - Date de publication
  - Popularité (nombre de téléchargements)

### 📚 Résultats structurés
- Regroupement visuel par catégorie (section “Presets”, “Samples”, “Utilisateurs”...)
- Affichage avec aperçu (lecture audio intégrée, si applicable)

---

## 🔧 Stack Technique

| Composant         | Technologie                                |
|-------------------|---------------------------------------------|
| Backend           | Go + PostgreSQL Full Text Search            |
| Alternative (opt) | ElasticSearch ou module Rust spécifique     |
| Frontend          | React + Zustand                            |
| UI                | shadcn/ui + composants filtrables dynamiques |

---

## 🗃️ Backend — Indexation & Query

### Option 1 : PostgreSQL FTS
- Utilisation des opérateurs `to_tsvector()` et `to_tsquery()`
- Index GIN sur `shared_files.description`, `filename`, `tags[]`

### Option 2 : ElasticSearch (optionnel)
- Indexation à chaque upload ou modification
- Requête via API REST (`/_search`)

---

## 🔍 Endpoints REST

| Méthode | URL               | Description                            |
|--------:|-------------------|----------------------------------------|
| GET     | `/search?q=...`   | Recherche globale                      |
| GET     | `/search/advanced?type=sample&tag=drum` | Recherche filtrée |
| GET     | `/autocomplete?q=...` | Suggère des tags, noms, auteurs   |

---

## 🖥️ Frontend

- Barre de recherche globale dans l’en-tête ou onglet
- Page "Résultats" avec :
  - Résultats paginés
  - Filtres dynamiques par colonne
  - Affichage conditionnel : icône format, auteur, date
- Preview audio directement dans les résultats

---

## ✅ Checklist de Validation

- [ ] Requêtes FTS ou ElasticSearch opérationnelles
- [ ] Résultats triés par pertinence
- [ ] Frontend fonctionnel avec filtres combinés
- [ ] Suggestions intelligentes de recherche
- [ ] Affichage propre des résultats (audio + metadata)
- [ ] Sécurité des requêtes (éviter injection FTS)

---

## 💡 Bonus optionnels

- Historique personnel des recherches (utilisateur)
- Résultats cachés pour contenu privé (si connecté)
- Score de pertinence / popularité

---

## 📌 Étapes Suivantes (V7)

➡ Développer la **plateforme de troc de produits Talas** (revente / échange entre utilisateurs).

---

## 🔁 **Plan de Développement — Version 6**

### ✅ **V6.1 – Indexation PostgreSQL FTS (base)**

| Objectifs                                                                             | Détails |
| ------------------------------------------------------------------------------------- | ------- |
| 🧠 Créer colonnes `tsvector` sur `shared_files` (`filename`, `description`, `tags[]`) |         |
| 🧱 Ajouter index GIN sur ces colonnes                                                 |         |
| 🧪 Test en local avec `to_tsquery()` et `plainto_tsquery()`                           |         |
| 🔒 Vérification de la visibilité (filtrer les fichiers privés selon l'utilisateur)    |         |

📂 Migration SQL → `add_fts_indices.sql`

---

### ✅ **V6.2 – Endpoint de recherche globale**

| Endpoint                 | Méthode | Auth                         | Description                                     |
| ------------------------ | ------- | ---------------------------- | ----------------------------------------------- |
| `/search?q=kick+ambient` | `GET`   | ✅ (ou ❌ pour contenu public) | Recherche sur `filename`, `description`, `tags` |

**Fonctionnalités** :

* Tri par pertinence (`ts_rank`)
* Pagination
* Option de filtrage : `?type=sample`, `?author=xxx`, `?popular=true`

📂 Route : `routes/search.go`
📂 Handler : `handlers/search.go`

---

### ✅ **V6.3 – Recherche avancée & filtres dynamiques**

| Endpoint               | Méthode | Auth | Description                    |
| ---------------------- | ------- | ---- | ------------------------------ |
| `/search/advanced?...` | `GET`   | ✅    | Recherche combinée par filtres |

**Filtres supportés :**

* Type (`sample`, `preset`, `doc`)
* Tags (`drum`, `ambient`)
* Auteur / uploader
* Date (`uploaded_after`, `uploaded_before`)
* Popularité (tri par `downloads.count`)

---

### ✅ **V6.4 – Auto-complétion des tags et noms**

| Endpoint             | Méthode | Auth | Description                                           |
| -------------------- | ------- | ---- | ----------------------------------------------------- |
| `/autocomplete?q=dr` | `GET`   | ✅    | Renvoie les tags, fichiers, ou auteurs correspondants |

📂 Source : PostgreSQL DISTINCT sur `tags[]`, `filename`, `uploader.username`

---

### ✅ **V6.5 – Interface Frontend complète**

| Composants                                | Détails                                        |
| ----------------------------------------- | ---------------------------------------------- |
| 🔍 `SearchBar.tsx`                        | Barre de recherche globale avec autocomplétion |
| 🧠 `SearchStore.ts` (Zustand)             | Requêtes, filtres, résultats                   |
| 📃 `SearchResults.tsx`                    | Affichage paginé, triable, filtrable           |
| 🔊 Audio preview intégré via player HTML5 |                                                |

---

### ✅ **V6.6 – Sécurité & protection des requêtes**

| Objectifs                                                                           | Détails |
| ----------------------------------------------------------------------------------- | ------- |
| 🛡️ Protection contre injection FTS (`tsquery` propre, pas de concat directe)       |         |
| 👤 Vérification que seuls les fichiers visibles sont accessibles (visibility check) |         |
| 📦 Caching optionnel des recherches (Redis ou memo côté frontend)                   |         |

---

## 🔍 **Résumé des Routes REST V6**

| Méthode | Endpoint               | Auth | Description                           |
| ------: | ---------------------- | ---- | ------------------------------------- |
|     GET | `/search?q=...`        | ❌/✅  | Recherche globale                     |
|     GET | `/search/advanced?...` | ✅    | Recherche filtrée par type/tag/auteur |
|     GET | `/autocomplete?q=...`  | ✅    | Suggestion dynamique                  |

---

## ✅ **Checklist Fonctionnelle Résumée**

| Composant                            | État |
| ------------------------------------ | ---- |
| FTS PostgreSQL sur fichiers partagés | ✅    |
| Endpoint de recherche globale        | ✅    |
| Filtrage dynamique (type/tag)        | ✅    |
| Auto-complétion (tags, auteurs)      | ✅    |
| Frontend React avec filtres          | ✅    |
| Sécurité des requêtes                | ✅    |

---
