#!/bin/bash

# scripts/dev-coordinated.sh
# Script pour démarrer l'environnement de développement complet

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Démarrage de l'environnement Talas coordonné${NC}"
echo "=============================================="

# Variables
BACKEND_PORT=8080
FRONTEND_PORT=5173
PROJECT_ROOT=$(pwd)

# Fonction pour nettoyer les processus à l'arrêt
cleanup() {
    echo -e "\n${YELLOW}🛑 Arrêt de l'environnement...${NC}"
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        echo "🔧 Backend arrêté"
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
        echo "⚛️ Frontend arrêté"
    fi
    exit 0
}

# Trap pour nettoyer à l'arrêt
trap cleanup SIGINT SIGTERM

# Vérifier que nous sommes dans le bon répertoire
if [ ! -f "backend/cmd/server/main.go" ] || [ ! -f "talas-frontend/package.json" ]; then
    echo -e "${RED}❌ Erreur: Lancez ce script depuis la racine du projet Talas${NC}"
    echo "Structure attendue:"
    echo "  backend/cmd/server/main.go"
    echo "  talas-frontend/package.json"
    exit 1
fi

# Fonction pour vérifier si un port est disponible
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️ Port $port déjà utilisé${NC}"
        return 1
    fi
    return 0
}

# Vérifier les ports
echo "🔍 Vérification des ports..."
if ! check_port $BACKEND_PORT; then
    echo -e "${RED}❌ Port $BACKEND_PORT (backend) déjà utilisé${NC}"
    echo "Tuez le processus avec: lsof -ti:$BACKEND_PORT | xargs kill"
    exit 1
fi

if ! check_port $FRONTEND_PORT; then
    echo -e "${YELLOW}⚠️ Port $FRONTEND_PORT (frontend) déjà utilisé, Vite trouvera un autre port${NC}"
fi

# Préparer l'environnement backend
echo -e "\n${BLUE}🔧 Préparation du backend...${NC}"
cd backend

# Vérifier les dépendances Go
if ! go mod verify >/dev/null 2>&1; then
    echo "📦 Installation des dépendances Go..."
    go mod tidy
fi

# Créer le fichier .env s'il n'existe pas
if [ ! -f .env ]; then
    echo "📝 Création du fichier .env backend..."
    cat > .env << 'EOF'
# Configuration Backend Talas
ENVIRONMENT=development
PORT=8080
GIN_MODE=debug

# Database
DATABASE_URL=postgres://localhost/talas_dev?sslmode=disable

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION=24h

# Upload
UPLOAD_DIR=./static
MAX_FILE_SIZE=104857600

# WebSocket
WS_PING_PERIOD=54s
WS_PONG_WAIT=60s
EOF
    echo -e "${GREEN}✅ Fichier .env créé${NC}"
fi

# Démarrer le backend
echo "🔧 Démarrage du backend Go..."
go run cmd/server/main.go &
BACKEND_PID=$!

# Attendre que le backend soit prêt
echo "⏳ Attente du démarrage du backend..."
for i in {1..10}; do
    if curl -s http://localhost:$BACKEND_PORT/health >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Backend prêt sur http://localhost:$BACKEND_PORT${NC}"
        break
    fi
    
    if [ $i -eq 10 ]; then
        echo -e "${RED}❌ Timeout: Backend ne répond pas${NC}"
        cleanup
        exit 1
    fi
    
    sleep 2
    echo "  Tentative $i/10..."
done

# Préparer l'environnement frontend
cd "$PROJECT_ROOT"
echo -e "\n${BLUE}⚛️ Préparation du frontend...${NC}"
cd talas-frontend

# Vérifier les dépendances npm
if [ ! -d "node_modules" ]; then
    echo "📦 Installation des dépendances npm..."
    npm install
fi

# Créer le fichier .env.local s'il n'existe pas
if [ ! -f .env.local ]; then
    echo "📝 Création du fichier .env.local frontend..."
    cat > .env.local << 'EOF'
# Configuration Frontend Talas
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_ENVIRONMENT=development
VITE_DEBUG=true
EOF
    echo -e "${GREEN}✅ Fichier .env.local créé${NC}"
fi

# Démarrer le frontend
echo "⚛️ Démarrage du frontend React..."
npm run dev &
FRONTEND_PID=$!

# Attendre que le frontend soit prêt
echo "⏳ Attente du démarrage du frontend..."
sleep 5

# Test de connectivité
echo -e "\n${BLUE}🔍 Test de connectivité...${NC}"

# Test backend
if curl -s http://localhost:$BACKEND_PORT/health | grep -q "ok"; then
    echo -e "${GREEN}✅ Backend: http://localhost:$BACKEND_PORT${NC}"
else
    echo -e "${RED}❌ Backend non accessible${NC}"
fi

# Test frontend (approximatif)
if curl -s http://localhost:$FRONTEND_PORT >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Frontend: http://localhost:$FRONTEND_PORT${NC}"
else
    echo -e "${YELLOW}⚠️ Frontend peut prendre quelques secondes supplémentaires${NC}"
fi

# Afficher les informations de développement
echo -e "\n${GREEN}🎉 Environnement prêt !${NC}"
echo "=============================="
echo -e "${BLUE}🔧 Backend (Go):${NC}      http://localhost:$BACKEND_PORT"
echo -e "${BLUE}⚛️ Frontend (React):${NC}  http://localhost:$FRONTEND_PORT"
echo -e "${BLUE}📊 Health Check:${NC}     http://localhost:$BACKEND_PORT/health"
echo -e "${BLUE}📡 WebSocket:${NC}        ws://localhost:$BACKEND_PORT/ws"
echo ""
echo -e "${YELLOW}💡 Conseils:${NC}"
echo "  • Le frontend Vite se recharge automatiquement"
echo "  • Le backend Go redémarre manuellement (Ctrl+C puis relancer)"
echo "  • Vérifiez les logs ci-dessous pour les erreurs"
echo "  • Les erreurs CORS indiquent des problèmes de configuration"
echo ""
echo -e "${GREEN}🛑 Pour arrêter: Ctrl+C${NC}"
echo ""

# Afficher les logs en temps réel
echo -e "${BLUE}📊 Logs en temps réel:${NC}"
echo "===================="

# Attendre indéfiniment (les logs apparaissent automatiquement)
wait

# Cleanup automatique à la fin
cleanup