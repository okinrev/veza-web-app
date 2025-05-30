# 📘 Talas — README des 12 Versions (Frontend Léger sans React)

Ce document décrit les 12 étapes d’implémentation de l’application Talas dans une stratégie **léger + modulaire**, orientée vers un **frontend statique HTML/CSS/JS** enrichi avec **HTMX** et **Alpine.js**, en conservant un backend Go (et Rust pour les modules performants). React ne sera utilisé **que si strictement nécessaire** à partir des versions avancées.

---

## ✅ Objectifs Généraux

* 🔧 Backend Go sécurisé et modulaire (API REST, JWT, PostgreSQL, Redis)
* 🧩 Modules Rust pour le streaming et le chat WebSocket
* 🖼️ Frontend HTML statique, avec interactions dynamiques grâce à HTMX et Alpine.js
* 📦 Déploiement ultra léger sur serveurs locaux (via Incus)
* 🚀 Aucune dépendance à Node/Vite sauf dans Tauri ou composants React isolés (V12)

---

## 🔁 Versions détaillées

### **V1 — Authentification (Login/Inscription avec JWT)**

* **Pages HTML :** `login.html`, `register.html`
* **Fonctions backend :** `POST /api/register`, `POST /api/login`, `GET /api/logout`
* **Sécurité :** hash avec `bcrypt`, token JWT avec expiration
* **UI :** HTMX pour les requêtes, Alpine.js pour l’état local (messages, validation, redirection)
* **Stockage :** PostgreSQL (table users)
* **Objectif :** Première authentification sécurisée, token localStorage

### **V2 — Fichiers et Produits**

* **Pages HTML :** `dashboard.html`, `produits.html`
* **Fonctions backend :** `GET/POST/PUT/DELETE /api/products`, `/api/files`
* **UI :** table dynamique, preview fichiers, upload fichiers (PDF, images)
* **Alpine.js :** interactions, état produit local
* **HTMX :** chargement et mise à jour sans JS manuel
* **Objectif :** tableau CRUD utilisateur + accès à ses fichiers/documents produits

### **V3 — Chat temps réel (rooms & DMs)**

* **Pages HTML :** `chat.html`
* **Backend :** Rust WebSocket server, authentifié via Go proxy JWT
* **UI :** messages temps réel avec template Alpine.js, websocket natif ou via extension HTMX
* **Objectif :** salon de discussion simple, messages persistés (optionnel)

### **V4 — Streaming audio communautaire**

* **Pages HTML :** `streaming.html`
* **Backend :** Rust module avec FFmpeg, index des pistes en Go
* **UI :** lecteur audio HTML5, lecture/pause, titre en cours
* **Objectif :** diffusion de pistes communautaires (écoute, navigation)

### **V5 — Partage de fichiers communautaires**

* **Pages HTML :** `partages.html`
* **Backend :** gestion des ressources (samples, presets), droits public/privé
* **UI :** drag & drop pour uploader, HTMX pour affichage/tri des fichiers
* **Objectif :** création d’une base communautaire ouverte de fichiers

### **V6 — Plateforme de troc (matériel Talas)**

* **Pages HTML :** `troc.html`
* **Backend :** API `POST /api/swap`, `/api/offers` (Go + PostgreSQL)
* **UI :** tableau d’offres, filtre par type/produit, match automatique
* **Objectif :** échanger du matériel entre utilisateurs de façon pair-à-pair

### **V7 — Hébergement de plugins VST (AudioGridder)**

* **Pages HTML :** `audiogridder.html`
* **Backend :** Rust serveur AudioGridder + Go UI controller (auth, assignation)
* **UI :** liste des plugins actifs, activation/désactivation distante
* **Objectif :** héberger des plugins VST sur serveur personnel pour DAW léger

### **V8 — Bibliothèque personnelle / favoris**

* **Pages HTML :** `favoris.html`
* **Fonctions :** marquer comme favori n’importe quelle ressource (fichier, preset, produit)
* **Stockage :** LocalStorage (offline) ou Go API (persisté)
* **UI :** affichage dynamique via Alpine.js `x-for`
* **Objectif :** créer un espace personnel rapide d’accès aux ressources préférées

### **V9 — Espace de formation (guides, vidéos)**

* **Pages HTML :** `formation.html`
* **Backend :** parsing Markdown, index des cours
* **UI :** HTMX pour navigation entre cours, Alpine.js pour toasts, transitions
* **Objectif :** héberger des guides, vidéos, fichiers de formation, accessibles librement

### **V10 — Suivi produit (garantie, upgrades)**

* **Pages HTML :** `produit.html`
* **Backend :** historique produit (Go + PostgreSQL), garantie, demandes de SAV
* **UI :** timeline d’événements, formulaire de contact interne
* **Objectif :** suivi détaillé d’un produit acheté : version, garantie, évolution, pannes

### **V11 — Moteur de recherche global**

* **Pages HTML :** `recherche.html`
* **Backend :** PostgreSQL full-text search ou trigrammes
* **UI :** champ de recherche avec HTMX (`hx-get` dynamique), filtres dynamiques (Alpine.js)
* **Objectif :** permettre la recherche de fichiers, utilisateurs, contenus, tags

### **V12 — Application standalone (Tauri)**

* **Compilation :** export des pages + API en local avec WebView
* **UI :** possibilité d’ajouter React pour les composants dynamiques critiques (chat, timeline audio)
* **Objectif :** fournir un client complet desktop Talas utilisable offline + synchronisation future

---

## 🛠️ Stack Technologique

* **Frontend :** HTML natif, Alpine.js (logique locale), HTMX (requêtes dynamiques)
* **Backend :** Go (REST API, JWT, PostgreSQL), Redis pour session/cache
* **Modules performants :** Rust (streaming audio, WebSocket, AudioGridder)
* **Déploiement :** Incus containers, Nginx ou Go http server, PostgreSQL, Nextcloud/ZFS
* **App Desktop :** Tauri (Rust + WebView2)

