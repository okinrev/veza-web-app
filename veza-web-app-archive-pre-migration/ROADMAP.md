# 🛣️ **Roadmap Talas – V1 à V12**

---

## ✅ V1 – Authentification & Base Utilisateurs *(Fait)*

### Objectifs :

* Création de compte, login, logout
* Hashage sécurisé avec Bcrypt
* JWT (avec expiration, refresh à venir)

### Statut : **Implémenté et fonctionnel**

---

## ✅ V2 – Gestion Fichiers, Produits & Documentation *(Fait)*

### Objectifs :

* Upload de fichiers (produits, ressources, docs)
* Modèle produit + tagging
* Stockage PostgreSQL + signed URLs

### Statut : **Implémenté**

---

## ✅ V3 – Chat & Messagerie temps réel *(Fait)*

### Objectifs :

* Serveur WebSocket Rust (rooms et DM)
* Authentification via JWT
* Stockage historique PostgreSQL
* UI HTML fonctionnelle

### Statut : **Implémenté**

🔜 À faire : meilleure UI frontend (React ?)

---

## ✅ V4 – Streaming Audio *(Fait)*

### Objectifs :

* Serveur Rust de streaming (`stream_server`)
* Lecture de `.mp3` à partir de la DB ou d’un dossier
* Endpoints protégés via token

### Statut : **Implémenté**

🔜 À faire : contrôle du flux (pause, seek), UI audio

---

## ✅ V5 – Partage de Fichiers & Ressources *(Fait)*

### Objectifs :

* Ressources publiques et partagées
* Ajout de métadonnées
* Possibilité de marquer comme "collaboratif" ou "lecture seule"

### Statut : **Implémenté**

🔜 À faire : interface plus claire pour partager/inviter

---

## ⏳ V6 – Moteur de Recherche *(À faire)*

### Objectifs :

* Recherche sur :

  * Tags, titres, artistes (tracks)
  * Utilisateurs
  * Ressources partagées
* Filtres : public / privé, format, date
* Pagination + indexation

### Tâches :

* [ ] Routes `search.go` (Go)
* [ ] Endpoint `/search?q=...`
* [ ] Intégration UI (champ + résultats dynamiques)

---

## ⏳ V7 – Plateforme de Troc Produits & Modules *(À faire)*

### Objectifs :

* Uploads d’articles à échanger (type produit)
* Système de demande / acceptation
* Historique des échanges

### Tâches :

* [ ] Modèle `ExchangeOffer`
* [ ] Routes pour créer, répondre, valider
* [ ] Notification / message en cas d’échange
* [ ] UI dédiée (vue catalogue)

---

## ⏳ V8 – Formations & Tutoriels *(À faire)*

### Objectifs :

* Uploads vidéos / docs par formateurs
* Organisation en modules / cours
* Système de suivi (progression)

### Tâches :

* [ ] Modèle `Formation`, `Section`, `Ressource`
* [ ] Backend pour l’accès et la progression
* [ ] UI type LMS : liste, progression, vidéos

---

## ⏳ V9 – Découverte Sociale *(À faire)*

### Objectifs :

* Système de like / vues / commentaires
* Feed d'écoute : "tendances", "nouveautés", "suivis"
* Page publique par artiste

### Tâches :

* [ ] Modèle `Like`, `View`, `Comment`
* [ ] Algorithme simple de recommandation
* [ ] UI : feed scrollable, profils publics

---

## ⏳ V10 – Bibliothèque Personnelle *(À faire)*

### Objectifs :

* Sauvegarde de ressources dans "favoris"
* Organisation par dossiers (ex. "mes presets")
* Marque-pages et notes personnelles

### Tâches :

* [ ] Modèle `Favorite`, `Collection`
* [ ] Routes pour ajouter/enlever
* [ ] Interface utilisateur à onglets ou filtres

---

## ⏳ V11 – Gestion Comptes & Rôles *(À faire)*

### Objectifs :

* Rôles : admin, formateur, utilisateur
* Droits spécifiques par ressource
* Interface admin : modération, bannissement, analytics

### Tâches :

* [ ] Middleware Go `CheckRole(...)`
* [ ] Modèle `Role`, `Permission`
* [ ] Interface admin simple (UI tabulaire)

---

## ⏳ V12 – Application Standalone (Electron ou PWA) *(À faire)*

### Objectifs :

* Version installable en local
* Accès offline (caching PWA)
* Synchronisation dès reconnexion
* Accès plugins DAW distants (via AudioGridder)

### Tâches :

* [ ] Adaptation frontend en React PWA ou Electron
* [ ] Intégration AudioGridder + WebRTC ou tunnel SSH
* [ ] Packaging multiplateforme

---

## 🧩 Modules Transverses à Prévoir

| Module                             | Statut         |
| ---------------------------------- | -------------- |
| CI/CD GitHub Actions               | ⬜ À faire      |
| Tests (Go + Rust)                  | ⬜ À faire      |
| Nextcloud + ZFS + PostgreSQL (v10) | ⬜ À configurer |
| Gestion fine des erreurs API       | ⬜ À faire      |
| Design système des rôles           | ⬜ À structurer |

---