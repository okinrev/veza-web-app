# 📂 Version 2 — Fichiers, Produits, Documentation

🎯 **Objectif** : Permettre aux utilisateurs de consulter leurs produits Talas achetés, d’accéder à la documentation technique associée, et d’envoyer/télécharger des fichiers via une interface sécurisée et intuitive.

---

## 📦 État de Développement : `🟢 Terminé (V2.x stable)`

---

## 🧩 Fonctionnalités Clés

### 🧾 Produits Talas liés à un compte
- Liste des produits associés à l’utilisateur connecté (`GET /products`)
- Détails complets d’un produit (version, achat, garantie)
- Affichage de l’état de garantie (active, expirée)

### 📁 Gestion de fichiers liés aux produits
- Upload de fichiers (guides, schémas, documents légaux)
- Organisation par type (`manuel`, `schéma`, `garantie`, etc.)
- Téléchargement sécurisé avec vérification de l’authentification

### 📑 Documentation embarquée
- Affichage ou téléchargement de :
  - Fiches techniques (PDF)
  - Schémas (SVG/JPG)
  - Guides Markdown (interactifs)
  - Vidéos (YouTube / fichiers locaux)

---

## 🔧 Stack Technique

| Composant         | Technologie                                 |
|-------------------|----------------------------------------------|
| Backend           | Go + SQLx + PostgreSQL                      |
| Stockage fichiers | ZFS local / Nextcloud (WebDAV ou API REST)  |
| Frontend          | React + Zustand + shadcn/ui + Tailwind      |
| Authentification  | JWT + Middleware d’accès sécurisé (hérité V1)|

---

## 📁 Structure Backend

backend/
├── handlers/
│ ├── product.go
│ └── file.go
├── models/
│ ├── product.go
│ └── file.go
├── routes/
│ ├── product.go
│ └── file.go
├── db/migrations/
│ └── add_products_and_files.sql


---

## 🧱 Modèle de Données (PostgreSQL)

- `products` :
  - `id`, `user_id`, `name`, `version`, `purchase_date`, `warranty_expires`
- `files` :
  - `id`, `product_id`, `filename`, `type`, `url`, `uploaded_at`

---

## 🔍 Endpoints REST

| Méthode | Endpoint               | Auth | Description                                   |
|--------:|------------------------|------|-----------------------------------------------|
| GET     | `/products`            | ✅    | Lister les produits d’un utilisateur          |
| GET     | `/products/{id}`       | ✅    | Détails d’un produit (version, garantie…)     |
| POST    | `/products/{id}/files` | ✅    | Upload d’un fichier lié au produit            |
| GET     | `/files/{id}`          | ✅    | Téléchargement sécurisé d’un fichier produit  |

---

## ✅ Checklist Technique

- [x] Authentification JWT active sur tous les endpoints
- [x] Liste des produits liés à l’utilisateur
- [x] Upload multi-format (PDF, images, Markdown)
- [x] Téléchargement protégé
- [x] Rendu frontend ergonomique (lecture/aperçu)
- [x] Stockage ZFS ou Nextcloud intégré (selon config)
- [x] Tests API upload/download/auth réalisés

---

## 💡 Améliorations UX recommandées

- Icônes dynamiques selon type MIME (`pdf`, `img`, `md`, etc.)
- Badges visuels pour garantie active/expirée
- Preview HTML pour schémas / Markdown
- Barre de progression pour upload

---

## 🪜 Étapes du Développement (Versionnées)

### ✅ **V2.1 — Modélisation**
- Tables PostgreSQL `products` & `files`
- Fichiers `models/`, `handlers/`, `routes/` créés

### ✅ **V2.2 — Liste des Produits**
- Endpoint `/products` protégé par JWT
- Retourne les produits liés à `user_id`

### ✅ **V2.3 — Détail Produit**
- Endpoint `/products/{id}`
- Retour complet avec garantie & version

### ✅ **V2.4 — Upload Fichier**
- Endpoint `/products/{id}/files` (POST)
- Upload avec validation MIME
- URL stockée en BDD (ZFS ou Nextcloud)

### ✅ **V2.5 — Téléchargement sécurisé**
- Endpoint `/files/{id}`
- Vérifie la propriété (`user_id`) du fichier

### ✅ **V2.6 — Intégration Frontend**
- Composants React : list, modal, upload
- Rendu Markdown, PDF viewer, vidéos embarquées

### ✅ **V2.7 — Intégration Nextcloud (optionnelle)**
- WebDAV ou API REST
- Arborescence : `/Users/{id}/Products/{product_id}/`

---

## 🧭 Synthèse des Routes (V2)

| Méthode | Endpoint               | Auth | Description                                   |
|--------:|------------------------|------|-----------------------------------------------|
| GET     | `/products`            | ✅    | Tous les produits utilisateur                 |
| GET     | `/products/{id}`       | ✅    | Détail produit                                |
| POST    | `/products/{id}/files` | ✅    | Upload de documentation                      |
| GET     | `/files/{id}`          | ✅    | Téléchargement conditionné à la propriété     |

---

## 📌 Étapes Suivantes (V3)

➡ Développement du module **chat & messagerie en temps réel** via WebSocket pour l’application communautaire Talas.

---
