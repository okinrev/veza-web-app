#!/bin/bash

echo "🧪 Test du Chat avec Données Réelles"
echo "=================================="

# Vérifier les salons en base
echo "📋 Salons en base de données :"
psql -d veza_db -c "SELECT id, name, is_private, created_at FROM rooms ORDER BY id;" -t

echo ""

# Vérifier les utilisateurs en base
echo "👥 Utilisateurs en base de données :"
psql -d veza_db -c "SELECT id, username, first_name, last_name, email FROM users ORDER BY id;" -t

echo ""

# Vérifier les messages récents
echo "💬 Messages récents :"
psql -d veza_db -c "SELECT id, from_user, to_user, room, content, timestamp FROM messages ORDER BY timestamp DESC LIMIT 10;" -t

echo ""
echo "🚀 Maintenant, ouvrez http://localhost:5174/chat et vérifiez :"
echo "   ✅ Les salons 'general' et 'afterworks' sont visibles"
echo "   ✅ Les vrais utilisateurs apparaissent dans les messages privés"
echo "   ✅ L'historique des messages est correct"
echo "   ✅ Logs dans la console : '[Chat] Salons récupérés depuis la base'" 