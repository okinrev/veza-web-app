# Test Rapide - Chat React ↔ Serveur Rust

## Étapes de test

### 1. Backend (PostgreSQL + Serveur Rust)
Puisque vous dites que le backend est bon, le serveur Rust devrait écouter sur `ws://localhost:9001`.

### 2. Frontend React
1. **Se connecter** avec un utilisateur valide dans l'application
2. **Ouvrir la console navigateur** (F12)
3. **Aller sur la page Chat** depuis le dashboard
4. **Vérifier les logs** dans la console

### 3. Logs attendus si tout fonctionne

#### Console navigateur (React)
```
[Chat] État d'authentification: { hasUser: true, hasToken: true, isAuthenticated: true, userId: 14 }
[Chat] Authentification OK, initialisation du chat...
[Rust Chat WebSocket] Token récupéré: Présent
[Rust Chat WebSocket] Auth state: { isAuthenticated: true, hasUser: true }
[Rust Chat WebSocket] URL de connexion: ws://localhost:9001/?token=***
[Rust Chat WebSocket] Connecté au serveur Rust
```

#### Console serveur Rust
```
🔌 Connexion TCP entrante
🔐 Authentification réussie user_id=14
✅ Connexion WS autorisée user_id=14
```

### 4. Test d'envoi de message

1. **Cliquer sur un salon** (ex: "general")
2. **Taper un message** et appuyer Entrée
3. **Vérifier** qu'il apparaît dans l'interface

#### Logs attendus pour l'envoi
```javascript
// Console React
[Rust Chat WebSocket] Envoi: {"type":"join","room":"general"}
[Rust Chat WebSocket] Message reçu: {"type":"join_ack","data":{"room":"general","status":"ok"}}
[Rust Chat WebSocket] Envoi: {"type":"message","room":"general","content":"Hello!"}
```

```bash
# Console Rust
👥 Rejoint la room general user_id=14
📨 Message room enregistré et diffusé
```

### 5. Problèmes courants

#### Token `undefined`
**Symptôme :** `ws://localhost:9001/?token=undefined`
**Solution :** 
- Vérifier que vous êtes bien connecté
- Rafraîchir la page
- Vérifier dans Application → Storage → localStorage que `authToken` existe

#### Connexion refusée
**Symptôme :** `Firefox can't establish a connection`
**Solution :**
- Vérifier que le serveur Rust tourne sur le port 9001
- Vérifier les logs du serveur Rust

#### JWT invalide (401)
**Symptôme :** `HTTP/1.1 401 Unauthorized`
**Solution :**
- Vérifier que `JWT_SECRET` est identique entre le backend Go et le serveur Rust
- Se reconnecter pour obtenir un nouveau token

### 6. Test manuel simple

Si le chat React ne fonctionne pas, testez directement avec un client WebSocket :

```javascript
// Dans la console navigateur
const token = localStorage.getItem('authToken');
const ws = new WebSocket(`ws://localhost:9001/?token=${token}`);
ws.onopen = () => console.log('Connected!');
ws.onmessage = (e) => console.log('Received:', e.data);
ws.send(JSON.stringify({type: 'join', room: 'general'}));
```

### 7. Débogage

Pour voir tous les détails :
```bash
# Terminal serveur Rust
RUST_LOG=debug cargo run

# Console navigateur
localStorage.setItem('debug', 'chat:*');
``` 