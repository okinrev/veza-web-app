#!/bin/bash

# Couleurs pour les messages
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${BLUE}🧪 Test des endpoints reconnectés${NC}"

# 1. Test de santé
echo -e "\n${BLUE}1. Test de santé${NC}"
curl -s "$BASE_URL/health" | jq .

# 2. Test d'authentification
echo -e "\n${BLUE}2. Test d'authentification${NC}"
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' | jq -r '.data.access_token')

if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
    echo -e "${GREEN}✅ Authentification réussie${NC}"
    
    # 3. Test des routes utilisateurs
    echo -e "\n${BLUE}3. Test des routes utilisateurs${NC}"
    curl -s -X GET "$BASE_URL/api/v1/users/me" \
      -H "Authorization: Bearer $TOKEN" | jq .
    
    # 4. Test des routes tracks
    echo -e "\n${BLUE}4. Test des routes tracks${NC}"
    curl -s -X GET "$BASE_URL/api/v1/tracks" \
      -H "Authorization: Bearer $TOKEN" | jq .
    
    # 5. Test des routes shared resources
    echo -e "\n${BLUE}5. Test des routes shared resources${NC}"
    curl -s -X GET "$BASE_URL/api/v1/shared_ressources" \
      -H "Authorization: Bearer $TOKEN" | jq .
    
    # 6. Test des routes listings
    echo -e "\n${BLUE}6. Test des routes listings${NC}"
    curl -s -X GET "$BASE_URL/api/v1/listings" \
      -H "Authorization: Bearer $TOKEN" | jq .
    
    # 7. Test des routes chat
    echo -e "\n${BLUE}7. Test des routes chat${NC}"
    curl -s -X GET "$BASE_URL/api/v1/chat/rooms" \
      -H "Authorization: Bearer $TOKEN" | jq .
    
    # 8. Test des routes de recherche
    echo -e "\n${BLUE}8. Test des routes de recherche${NC}"
    curl -s -X GET "$BASE_URL/api/v1/tags" | jq .
    curl -s -X GET "$BASE_URL/api/v1/suggestions" | jq .
    
    echo -e "\n${GREEN}✅ Tous les tests terminés${NC}"
else
    echo -e "${RED}❌ Échec de l'authentification${NC}"
fi
