# ♻️ Version 7 — Plateforme de troc de produits Talas

🎯 Objectif : Permettre aux utilisateurs d’échanger ou revendre leurs produits Talas via une plateforme interne de troc sécurisée, favorisant l’usage circulaire et la réutilisation.

---

## 🧩 Fonctionnalités à Implémenter

### 🪙 Création d’annonces
- Mise en ligne d’un produit Talas possédé par l’utilisateur :
  - Type, version, état, photos, description
  - Option : recherche d’un échange ou vente directe
  - Option : prix souhaité ou produit recherché (dans le cas d’un troc)

### 🔄 Échanges entre utilisateurs
- Proposition d’un échange produit ↔ produit
- Négociation ou contre-proposition (optionnelle)
- Historique des propositions et statut

### 💬 Contact sécurisé
- Discussion entre utilisateurs via système de messagerie (lié à la V3)
- Notification à l’auteur de l’annonce

### ✅ Finalisation de l’échange
- Marquer un échange comme réalisé (validation des deux parties)
- Retirer automatiquement l’annonce après validation

---

## 🔧 Stack Technique

| Composant        | Technologie                   |
|------------------|-------------------------------|
| Backend API      | Go (REST endpoints sécurisés) |
| Base de données  | PostgreSQL (`listings`, `offers`) |
| Frontend         | React + Tailwind + shadcn/ui  |
| Messagerie       | WebSocket (module existant)   |

---

## 🗃️ Structure Backend (extrait)

backend/
├── handlers/
│ └── listings.go
├── models/
│ ├── listing.go
│ └── offer.go
├── routes/
│ └── listings.go
├── db/migrations/
│ └── add_listings_and_offers.sql

---

## 🧱 Tables PostgreSQL

- `listings`: `id`, `user_id`, `product_id`, `description`, `state`, `price`, `exchange_for`, `images[]`, `status`
- `offers`: `id`, `listing_id`, `from_user_id`, `proposed_product_id`, `message`, `status`

---

## 🔍 Endpoints REST

| Méthode | URL                   | Description                                 |
|--------:|-----------------------|---------------------------------------------|
| POST    | `/listings`           | Créer une annonce                           |
| GET     | `/listings`           | Lister les annonces disponibles             |
| GET     | `/listings/{id}`      | Voir les détails d’une annonce              |
| POST    | `/listings/{id}/offer`| Proposer un échange                         |
| POST    | `/offers/{id}/accept` | Accepter une proposition                    |
| DELETE  | `/listings/{id}`      | Supprimer une annonce                       |

---

## 🖥️ Frontend

- Formulaire de mise en ligne avec preview image
- Galerie des annonces filtrable (type, produit, disponibilité)
- Page d’annonce détaillée
- Interface de proposition d’échange
- Zone de messages liée à chaque offre

---

## ✅ Checklist de Validation

- [ ] Création et affichage d’annonces fonctionnelles
- [ ] Propositions d’échange sécurisées
- [ ] Système de validation à double accord
- [ ] Suppression automatique des annonces conclues
- [ ] Intégration avec la messagerie existante
- [ ] Contrôle d’accès : seuls les propriétaires peuvent publier

---

## 💡 Bonus possibles

- Évaluation des utilisateurs (note après un échange)
- Historique des échanges (vue "mes transactions")
- Option de remise locale (ajout de ville ou région)

---

## 📌 Étapes Suivantes (V8)

➡ Développer un **module de formation** pour accompagner les utilisateurs avec des guides et tutoriels internes ou partenaires.

---

## 🔁 **Plan de Développement — Version 7**

### ✅ **V7.1 – Modélisation des Annonces et Offres**

| Objectifs                                                                         | Détails |
| --------------------------------------------------------------------------------- | ------- |
| 🧱 Création des tables PostgreSQL `listings` et `offers`                          |         |
| 🔐 Clés étrangères vers `users`, `products`                                       |         |
| 📦 Champs : `state`, `description`, `price`, `images[]`, `exchange_for`, `status` |         |
| 📂 Migration : `add_listings_and_offers.sql`                                      |         |

---

### ✅ **V7.2 – Création et gestion des annonces**

| Endpoint    | Méthode | Auth | Description                                     |
| ----------- | ------- | ---- | ----------------------------------------------- |
| `/listings` | `POST`  | ✅    | Créer une annonce à partir d’un produit possédé |

**Fonctionnalités** :

* Upload d’images
* Vérification de possession du produit
* Validation des champs (type, état, prix/troc)

📂 Handler : `handlers/listings.go`
📂 Model : `models/listing.go`

---

### ✅ **V7.3 – Consultation et recherche d’annonces**

| Endpoint         | Méthode | Auth | Description                          |
| ---------------- | ------- | ---- | ------------------------------------ |
| `/listings`      | `GET`   | ❌    | Liste toutes les annonces filtrables |
| `/listings/{id}` | `GET`   | ❌    | Voir les détails d’une annonce       |

**Filtres possibles** :

* Type produit
* Type d’offre : vente / troc
* Prix max / produit recherché
* Tri par date

---

### ✅ **V7.4 – Proposition d’échange / offre**

| Endpoint               | Méthode | Auth | Description                               |
| ---------------------- | ------- | ---- | ----------------------------------------- |
| `/listings/{id}/offer` | `POST`  | ✅    | Proposer un produit en échange d’un autre |

**Fonctionnalités** :

* Lier un `product_id` du proposeur
* Message personnalisé
* Enregistrement dans `offers`
* Notifier le créateur via WebSocket (V3)

---

### ✅ **V7.5 – Acceptation / finalisation de l’échange**

| Endpoint              | Méthode | Auth | Description                                               |
| --------------------- | ------- | ---- | --------------------------------------------------------- |
| `/offers/{id}/accept` | `POST`  | ✅    | Marquer l’offre comme acceptée et l’annonce comme conclue |

**Effets** :

* Statut `offer = accepted`, `listing = closed`
* Message système dans le salon de discussion
* Verrouillage des autres offres liées

---

### ✅ **V7.6 – Suppression et gestion des annonces**

| Endpoint         | Méthode  | Auth | Description                                     |
| ---------------- | -------- | ---- | ----------------------------------------------- |
| `/listings/{id}` | `DELETE` | ✅    | Supprimer une annonce (propriétaire uniquement) |

---

### ✅ **V7.7 – Intégration frontend complète**

| Composants                     | Détails                                                 |
| ------------------------------ | ------------------------------------------------------- |
| 🧾 `NewListingForm.jsx`        | Formulaire avec drag & drop image, description, type    |
| 🗃️ `ListingGallery.jsx`       | Galerie filtrable (état, type, troc/vente)              |
| 🧠 `ExchangeOfferModal.jsx`    | Sélection produit possédé + message                     |
| 💬 `OfferChat.jsx`             | Intégré avec WebSocket V3, messages liés à chaque offre |
| 🧠 `ListingStore.ts` (Zustand) | Gestion des annonces/offres localement                  |

---

## 🔍 **Résumé des Routes REST V7**

| Méthode | Endpoint               | Auth | Description                    |
| ------: | ---------------------- | ---- | ------------------------------ |
|    POST | `/listings`            | ✅    | Créer une annonce              |
|     GET | `/listings`            | ❌    | Lister les annonces            |
|     GET | `/listings/{id}`       | ❌    | Voir les détails d’une annonce |
|    POST | `/listings/{id}/offer` | ✅    | Proposer un échange            |
|    POST | `/offers/{id}/accept`  | ✅    | Valider une proposition        |
|  DELETE | `/listings/{id}`       | ✅    | Supprimer une annonce          |

---

## ✅ **Checklist Fonctionnelle Résumée**

| Composant                                   | État |
| ------------------------------------------- | ---- |
| Création d’annonces sécurisée               | ✅    |
| Upload images / métadonnées produit         | ✅    |
| Consultation publique filtrable             | ✅    |
| Propositions d’échange avec vérif ownership | ✅    |
| Système de double validation                | ✅    |
| Messagerie intégrée (via V3)                | ✅    |
| Suppression automatique post-échange        | ✅    |

---
