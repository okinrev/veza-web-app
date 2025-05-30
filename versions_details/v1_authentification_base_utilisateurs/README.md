# ✅ Version 1 — Authentification & Base Utilisateur

🎯 **Objectif** : Implémenter le socle fondamental de l’application **Talas** en mettant en place un système sécurisé de création de compte, connexion, et gestion de profil utilisateur, basé sur JWT.

---

## 📦 État de Développement : `🟢 Terminé`

---

## 🧩 Fonctionnalités à Implémenter

### 🔐 Authentification (JWT)
- [x] Création de compte (`POST /signup`)
- [x] Connexion (`POST /login`)
- [x] Déconnexion (client-side)
- [x] Middleware de vérification du token JWT
- [x] Hashage des mots de passe (`bcrypt`)

### 👤 Gestion du Profil
- [x] Récupération du profil (`GET /me`)
- [x] Modification des informations utilisateur (`PUT /users/{id}`)
- [ ] Suppression de compte (`DELETE /users/{id}` — optionnel)

---

## 🛡️ Sécurité

| Mesure                        | Détail technique                                 |
|------------------------------|--------------------------------------------------|
| Authentification             | JWT signé (`HS256` via `golang-jwt/jwt`)         |
| Hash de mot de passe         | `bcrypt` (`golang.org/x/crypto/bcrypt`)          |
| Middleware de sécurité       | Vérification du token dans `Authorization`       |
| Protection CORS              | Configuration stricte (`main.go`)                |
| Limitation brute-force login | Optionnel : Redis pour rate limiting             |

---

## 🔧 Stack Technique

| Composant        | Technologie                     |
|------------------|----------------------------------|
| Backend API      | Go (net/http + Gorilla Mux)      |
| Authentification | JWT (`github.com/golang-jwt/jwt`)|
| BDD              | PostgreSQL + `sqlx`              |
| Sécurité         | `bcrypt`, CORS, Rate Limit (Redis)|
| Frontend         | React + Vite + shadcn/ui + Tailwind |

---

## 🗃️ Structure des Fichiers (Backend)



backend/
├── main.go
├── routes/
│ └── auth.go
├── handlers/
│ └── user.go
├── middleware/
│ └── jwt.go
├── models/
│ └── user.go
├── db/
│ ├── database.go
│ └── migrations.sql
└── utils/
└── hash.go

---

## 🔍 Endpoints REST (V1)

| Méthode | Endpoint        | Auth | Description                                |
|--------:|------------------|------|--------------------------------------------|
| POST    | `/signup`        | ❌    | Crée un utilisateur                        |
| POST    | `/login`         | ❌    | Authentifie un utilisateur, retourne un JWT|
| GET     | `/me`            | ✅    | Donne les infos de l'utilisateur connecté  |
| PUT     | `/users/{id}`    | ✅    | Met à jour le profil utilisateur           |
| DELETE  | `/users/{id}`    | ✅    | Supprime le compte (optionnel)             |

---

## 🧪 Étapes d’Implémentation détaillées

### **V1.1 – Configuration de base**
- Mise en place de `main.go` avec gestion CORS et log.
- Connexion PostgreSQL avec `sqlx`.
- Migration SQL pour table `users`.

### **V1.2 – Création de compte**
- Endpoint `/signup`
- Hashage du mot de passe
- Validation & insertion BDD

### **V1.3 – Connexion et génération JWT**
- Endpoint `/login`
- Vérification mot de passe (`bcrypt`)
- Génération du JWT (payload : `user_id`, expiration)

### **V1.4 – Middleware JWT**
- Extraction du token
- Vérification de signature
- Injection `user_id` dans le contexte

### **V1.5 – Lecture du profil**
- Endpoint `/me`
- Récupération via JWT
- Renvoi des infos de base (email, id, etc.)

### **V1.6 – Mise à jour de profil**
- Endpoint `/users/{id}`
- Vérification identité (JWT vs ID)
- Mise à jour champs autorisés

### **V1.7 – (Optionnelle) Suppression du compte**
- Endpoint `/users/{id}` (DELETE)
- Authentification obligatoire

### **V1.8 – Sécurité complémentaire**
- Middleware CORS
- Rate limiting (optionnel, via Redis)
- Recommandation : HTTPS + refresh token (V1 simplifié)

---

## ✅ Checklist de Validation

- [x] JWT généré et vérifié via middleware
- [x] Création, connexion et récupération du profil fonctionnelles
- [x] Données stockées correctement en PostgreSQL
- [x] Mot de passe hashé
- [x] Frontend : formulaires signup/login fonctionnels
- [x] CORS protégé
- [ ] Rate limiting actif (si Redis configuré)
- [x] Tests manuels ou automatisés des endpoints

---

## 📌 Étapes Suivantes (V2)

➡ Intégrer la **gestion des produits et fichiers personnels**, posant les bases de l’espace utilisateur Talas.

---
