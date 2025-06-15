#!/bin/bash

# Script de démarrage pour le développement Talas
# Frontend React + Backend Go

echo "🚀 Démarrage de l'environnement de développement Talas"
echo ""

# Vérifier si le frontend existe
if [ ! -d "talas-frontend" ]; then
    echo "❌ Le dossier talas-frontend n'existe pas"
    exit 1
fi

# Vérifier si le backend existe
if [ ! -d "backend" ]; then
    echo "❌ Le dossier backend n'existe pas"
    exit 1
fi

# Fonction pour tuer les processus en arrière-plan
cleanup() {
    echo ""
    echo "🛑 Arrêt des serveurs..."
    kill $(jobs -p) 2>/dev/null
    exit
}

# Piéger Ctrl+C
trap cleanup SIGINT

echo "📦 Installation des dépendances du frontend..."
cd talas-frontend
npm install --silent
cd ..

echo ""
echo "🔥 Démarrage du serveur de développement frontend (React + Vite)..."
echo "   📍 URL: http://localhost:5173"
cd talas-frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "🔧 Démarrage du serveur backend (Go + Gin)..."
echo "   📍 URL: http://localhost:8080"
echo "   📍 API: http://localhost:8080/api/v1"
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

echo ""
echo "✅ Serveurs démarrés !"
echo ""
echo "🌐 Frontend: http://localhost:5173"
echo "🔌 Backend:  http://localhost:8080"
echo "📡 API:      http://localhost:8080/api/v1"
echo ""
echo "💡 Le frontend communique avec le backend via l'API"
echo "💡 Les routes frontend sont gérées par React Router"
echo ""
echo "Appuyez sur Ctrl+C pour arrêter les serveurs"

# Attendre que les processus se terminent
wait 