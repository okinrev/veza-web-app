# 🖥️ Version 12 — Application Desktop Standalone (Tauri)

🎯 Objectif : Offrir une version autonome de l’application Talas, installable sur Windows, macOS et Linux, intégrant le frontend React et communicant avec les modules backend/API existants.

---

## 🧩 Fonctionnalités à Implémenter

### 🧳 Packaging en application native
- Utilisation de **Tauri** pour :
  - Emballer le frontend React
  - Fournir une interface desktop légère
  - Accéder aux ressources locales de l’utilisateur (fichiers, audio)

### 🔌 Intégration avec l’API backend
- Communication avec l’API REST existante (Go)
- Authentification JWT persistée en local
- Support de la lecture audio, chat WebSocket, fichiers favoris…

### 🔐 Sécurité & isolation
- Sandboxing des accès au système de fichiers
- Déclaration stricte des permissions dans `tauri.conf.json`
- Communication sécurisée entre Tauri et backend distant

### 🛠️ Fonctionnalités additionnelles Desktop (optionnel)
- Drop zone : drag & drop de fichiers audio à streamer
- Stockage local de certains fichiers (offline mode ?)
- Notifications desktop

---

## 🔧 Stack Technique

| Composant       | Technologie                            |
|------------------|----------------------------------------|
| Enveloppe Desktop | [Tauri](https://tauri.app/) (Rust + JS)|
| Frontend         | React (déjà utilisé)                   |
| Backend distant  | Go API (déjà développé)                |
| Streaming        | Rust (module en gRPC/HTTP local/distant) |

---

## 📁 Structure du Dossier

standalone/
├── src-tauri/
│ ├── tauri.conf.json
│ ├── icons/
│ └── build.rs (optionnel)
├── package.json
├── src/
│ └── main.jsx (React app déjà existante)

---

## ⚙️ Installation et Build

### 📦 Prérequis
- Node.js
- Cargo (Rust)
- Tauri CLI

### 🔧 Initialisation
```bash
npm install -D @tauri-apps/cli
cargo install tauri-cli
npm run tauri init
```
### 🚀 Lancer l’app en dev

```bash
npm run tauri dev
```

### 🏁 Build final (Windows, Linux, macOS)

```bash
npm run tauri build
```

---

## ✅ Checklist de Validation

* [ ] Application installable sur Windows, macOS, Linux
* [ ] Connexion stable avec l’API distante (auth, data)
* [ ] Frontend React embarqué fonctionne identiquement à la version web
* [ ] Sécurité configurée dans `tauri.conf.json` (CSP, sandbox)
* [ ] Accès aux fichiers du système contrôlé et restreint
* [ ] L’application tourne sans connexion (mode partiel ou read-only)

---

## 💡 Bonus possibles

* Mode offline pour lecture locale de fichiers favoris
* Envoi direct de fichiers depuis le poste local
* Intégration AudioGridder simplifiée en mode local

---

# 🧱 **Plan de Développement – Version 12 (Tauri Desktop)**

## 🔹 **V12.1 – Initialisation du projet Tauri avec React**

| Étapes                                                                    | Objectif |
| ------------------------------------------------------------------------- | -------- |
| 🧪 Créer un dossier `standalone/` contenant le projet Tauri               |          |
| 🔧 `npm create tauri-app` ou `tauri init` dans le projet React existant   |          |
| ⚙️ Configurer `tauri.conf.json` (nom, permissions minimales, URL API)     |          |
| ✅ Vérifier que `npm run tauri dev` fonctionne et charge le frontend React |          |

---

## 🔹 **V12.2 – Intégration de l’API distante et persistance du JWT**

| Tâches                                                                   | Détails |
| ------------------------------------------------------------------------ | ------- |
| 🔑 Authentification persistée via `localStorage` ou `tauri::fs::app_dir` |         |
| 🔌 Tester tous les appels API (REST, WebSocket) en mode desktop          |         |
| 🔐 Gérer les erreurs d’auth (token expiré, reconnection)                 |         |
| 🚫 Interdire hardcoded URLs : utiliser `.env` + fallback Tauri config    |         |

---

## 🔹 **V12.3 – Accès aux fichiers locaux et sécurité**

| Élément                                                                          | Détail |
| -------------------------------------------------------------------------------- | ------ |
| 📂 Implémenter `tauri::dialog::open()` pour importer fichiers audio              |        |
| 🔐 Restreindre les accès au système de fichiers dans `tauri.conf.json`           |        |
| 🚨 CSP : Activer une politique de contenu stricte (`default-src`, `connect-src`) |        |
| 🛡️ Activer sandbox, désactiver les APIs inutiles (clipboard, shell…)            |        |

---

## 🔹 **V12.4 – Build multiplateforme et packaging**

| Plateformes | Commandes                                     |
| ----------- | --------------------------------------------- |
| 🪟 Windows  | `npm run tauri build` → `.msi`                |
| 🍎 macOS    | `npm run tauri build` → `.dmg`                |
| 🐧 Linux    | `npm run tauri build` → `.AppImage` ou `.deb` |

> 📦 Tu peux aussi utiliser [tauri-action](https://github.com/tauri-apps/tauri-action) sur GitHub CI pour automatiser les builds.

---

## 🔹 **V12.5 – Bonus optionnels (offline, drop, notifications)**

| Fonction                       | Stack                                                 |
| ------------------------------ | ----------------------------------------------------- |
| 📥 Drag & drop fichiers        | `tauri::drag_drop_api` + callback React               |
| 📴 Mode offline lecture locale | Cache local de fichiers `appDataPath + /cache/audio/` |
| 🔔 Notifications système       | `tauri::notification::notify`                         |

---

# 🔍 **Checklist Finale**

| Composant                                          | Statut      |
| -------------------------------------------------- | ----------- |
| 🎯 Initialisation Tauri/React                      | ✅           |
| 🌐 Connexion à l’API REST/WebSocket                | ✅           |
| 🔐 Sécurité Tauri                                  | ✅           |
| 🏁 Builds multiplateformes                         | ✅           |
| 🎧 Player audio React fonctionnel                  | ✅           |
| 🖥️ Fonctionnalités Desktop natives (optionnelles) | ⚙️ En cours |

---

## 📁 Exemple de `tauri.conf.json` minimal

```json
{
  "tauri": {
    "bundle": {
      "identifier": "fr.talas.app",
      "icon": ["icons/icon.png"]
    },
    "security": {
      "csp": "default-src 'self'; connect-src https://api.talas.fr ws://chat.talas.fr"
    },
    "allowlist": {
      "fs": {
        "all": false,
        "read": true,
        "write": false
      },
      "shell": false,
      "clipboard": false
    }
  }
}
```

---
