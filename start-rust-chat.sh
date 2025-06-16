#!/bin/bash

# Script pour démarrer le serveur de chat Rust
# Usage: ./start-rust-chat.sh

echo "🚀 Démarrage du serveur de chat Rust..."

# Vérifier que Rust est installé
if ! command -v cargo &> /dev/null; then
    echo "❌ Rust n'est pas installé. Installez-le avec:"
    echo "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
    exit 1
fi

# Vérifier que PostgreSQL est démarré
if ! pg_isready -q; then
    echo "❌ PostgreSQL n'est pas démarré."
    echo "Démarrez PostgreSQL avec: sudo systemctl start postgresql"
    exit 1
fi

# Se déplacer dans le répertoire du serveur Rust
cd backend/modules/chat_server

# Vérifier que le fichier .env existe
if [ ! -f .env ]; then
    echo "⚠️  Fichier .env manquant. Création d'un exemple..."
    cat > .env << EOF
# Configuration du serveur de chat Rust
DATABASE_URL=postgresql://postgres:password@localhost/veza_dev
JWT_SECRET=your_jwt_secret_here
WS_BIND_ADDR=127.0.0.1:9001
RUST_LOG=chat_server=debug
EOF
    echo "📝 Fichier .env créé. Modifiez-le avec vos paramètres et relancez."
    exit 1
fi

# Charger les variables d'environnement
export $(grep -v '^#' .env | xargs)

echo "📊 Configuration:"
echo "  - Base de données: $DATABASE_URL"
echo "  - WebSocket: ws://$WS_BIND_ADDR"
echo "  - Log level: $RUST_LOG"

# Démarrer le serveur
echo ""
echo "🔥 Démarrage du serveur WebSocket Rust..."
echo "   Connectez votre client React sur ws://localhost:9001"
echo "   Appuyez sur Ctrl+C pour arrêter"
echo ""

cargo run 