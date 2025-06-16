# 🧪 Test du Chat avec Token Valide

## 📋 Prérequis
- ✅ Backend Go en cours d'exécution sur port 8080
- ✅ Frontend React en cours d'exécution sur port 5174
- ✅ Serveur WebSocket Rust en cours d'exécution sur port 9001
- ✅ Token valide récupéré via curl

## 🔑 Token de Test Valide
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNSwidXNlcm5hbWUiOiJ0ZXN0Y2hhdCIsInJvbGUiOiJ1c2VyIiwiZXhwIjoxNzUwMDY0NDg1LCJpYXQiOjE3NTAwNjA4ODV9.44i0hQnfuUywY76dZ9vwEIQ6-Spqyk9xiwEapYlKc4o
```

## 👤 Utilisateur de Test
- **ID:** 15
- **Username:** testchat
- **Email:** testchat@example.com
- **Password:** testpassword123

## 🧪 Étapes de Test

### 1. Ouvrir le Frontend
```bash
# Le frontend devrait déjà être en cours d'exécution
# Ouvrir http://localhost:5174 dans le navigateur
```

### 2. Injecter le Token (Console du Navigateur)
```javascript
// Copier-coller ce code dans la console du navigateur (F12)
const testToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNSwidXNlcm5hbWUiOiJ0ZXN0Y2hhdCIsInJvbGUiOiJ1c2VyIiwiZXhwIjoxNzUwMDY0NDg1LCJpYXQiOjE3NTAwNjA4ODV9.44i0hQnfuUywY76dZ9vwEIQ6-Spqyk9xiwEapYlKc4o";

const testUser = {
  id: 15,
  username: "testchat",
  email: "testchat@example.com",
  role: "user"
};

localStorage.setItem('access_token', testToken);
localStorage.setItem('user', JSON.stringify(testUser));
console.log('✅ Token injecté, rechargez la page');
```

### 3. Naviguer vers le Chat
```
http://localhost:5174/chat
```

## ✅ Résultats Attendus

### APIs Testées avec Succès
- **`/api/v1/rooms`** → ✅ Retourne salon "general"
- **`/api/v1/users/except-me`** → ✅ Retourne liste des utilisateurs réels
- **`/api/v1/rooms/general/messages`** → ✅ Retourne messages du salon

### Frontend Corrigé
- ✅ Utilise `/api/v1/rooms` au lieu de `/chat/rooms`
- ✅ Gère le format de réponse `{"success":true,"data":[...]}`
- ✅ Gestion des tokens expirés avec rafraîchissement automatique

### Fonctionnalités Attendues
1. **Affichage du salon "general"** depuis la base de données
2. **Liste des utilisateurs réels** pour les messages privés
3. **Connexion WebSocket** au serveur Rust (si token valide)
4. **Messages historiques** du salon general

## 🐛 Dépannage

### Si le token expire
```bash
# Récupérer un nouveau token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "testchat@example.com", "password": "testpassword123"}'
```

### Si WebSocket refuse la connexion
- Vérifier que le serveur Rust est en cours d'exécution
- Vérifier que le token est valide et non expiré
- Regarder les logs du serveur Rust pour les erreurs d'authentification

### Si les APIs retournent 401
- Le token a probablement expiré
- Se reconnecter pour obtenir un nouveau token
- Vérifier que le header Authorization est correct

## 📊 Logs à Surveiller

### Backend Go
```
INFO: [GET] /api/v1/rooms ::1 200
INFO: [GET] /api/v1/users/except-me ::1 200
```

### Serveur Rust WebSocket
```
INFO: 🔌 Connexion TCP entrante
INFO: ✅ JWT valide, utilisateur connecté
```

### Frontend (Console)
```
[Chat] Salons récupérés depuis la base: {...}
[Chat] Utilisateurs chargés pour DM: 14 utilisateurs
[Chat] WebSocket connecté au serveur Rust
``` 