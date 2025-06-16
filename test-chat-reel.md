# Test Chat avec Vraies Données

## Changements effectués

### 1. Récupération des salons depuis la base de données
- **Ancienne méthode** : Salons hardcodés (`general`, `random`)
- **Nouvelle méthode** : Récupération via `/api/v1/chat/rooms`
- **Format adapté** : Compatible avec le schéma de la base PostgreSQL

### 2. Récupération des utilisateurs depuis la base de données
- **Ancienne méthode** : Utilisateurs fictifs (`alice`, `bob`, `charlie`)
- **Nouvelle méthode** : Récupération via `/api/v1/users/except-me`
- **Format adapté** : Compatible avec le schéma users de PostgreSQL

## Tests à effectuer

### Test 1 : Vérification des salons
1. Ouvrir http://localhost:5174/chat
2. Vérifier que les salons affichés correspondent à ceux en base :
   - `general` ✓
   - `afterworks` ✓
3. Vérifier dans les logs de la console : `[Chat] Salons récupérés depuis la base:`

### Test 2 : Vérification des utilisateurs pour DM
1. Passer en mode "Messages privés"  
2. Vérifier que les utilisateurs affichés correspondent à ceux en base (sauf utilisateur connecté)
3. Vérifier dans les logs : `[Chat] Utilisateurs récupérés depuis la base:`

### Test 3 : Historique des messages
1. Rejoindre le salon `afterworks`
2. Vérifier que l'historique correspond aux messages en base
3. Vérifier dans les logs : `Historique salon reçu:`

## Logs attendus dans la console

```
[Chat] Chargement des salons depuis la base de données...
[Chat] Salons récupérés depuis la base: {rooms: [{id: 1, name: "general"}, {id: 2, name: "afterworks"}]}
[Chat] Salons chargés: 2 salons depuis la base
[Chat] Chargement des utilisateurs depuis la base de données...
[Chat] Utilisateurs récupérés depuis la base: {data: [...]}
[Chat] Utilisateurs chargés pour DM: X utilisateurs depuis la base
```

## Dépannage

### Si aucun salon ne s'affiche
- Vérifier que l'API `/api/v1/chat/rooms` fonctionne
- Vérifier le token d'authentification
- Mode fallback : salon `general` uniquement

### Si aucun utilisateur ne s'affiche
- Vérifier que l'API `/api/v1/users/except-me` fonctionne  
- Vérifier l'authentification

### Commandes de vérification

```bash
# Vérifier les salons en base
psql -d veza_db -c "SELECT * FROM rooms;"

# Vérifier les utilisateurs en base  
psql -d veza_db -c "SELECT id, username, first_name, last_name FROM users LIMIT 10;"

# Vérifier les messages en base
psql -d veza_db -c "SELECT id, from_user, room, content FROM messages ORDER BY timestamp DESC LIMIT 10;"
``` 