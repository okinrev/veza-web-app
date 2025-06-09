# Frontend Veza Web App

## Structure du projet

```
frontend/
├── public/                 # Fichiers statiques
│   ├── assets/            # Assets statiques (images, fonts, etc.)
│   └── index.html         # Point d'entrée de l'application
├── src/                   # Code source
│   ├── components/        # Composants réutilisables
│   │   ├── auth/         # Composants d'authentification
│   │   ├── chat/         # Composants de chat
│   │   ├── common/       # Composants communs (header, footer, etc.)
│   │   ├── listings/     # Composants de listings
│   │   ├── search/       # Composants de recherche
│   │   ├── shared/       # Composants de ressources partagées
│   │   ├── tracks/       # Composants de tracks
│   │   └── users/        # Composants utilisateurs
│   ├── pages/            # Pages de l'application
│   │   ├── admin/        # Pages d'administration
│   │   ├── auth/         # Pages d'authentification
│   │   ├── chat/         # Pages de chat
│   │   ├── listings/     # Pages de listings
│   │   ├── search/       # Pages de recherche
│   │   ├── shared/       # Pages de ressources partagées
│   │   ├── tracks/       # Pages de tracks
│   │   └── users/        # Pages utilisateurs
│   ├── styles/           # Styles CSS/SCSS
│   │   ├── components/   # Styles des composants
│   │   ├── pages/        # Styles des pages
│   │   └── main.scss     # Style principal
│   ├── scripts/          # Scripts JavaScript
│   │   ├── components/   # Scripts des composants
│   │   ├── pages/        # Scripts des pages
│   │   └── main.js       # Script principal
│   └── utils/            # Utilitaires
│       ├── api.js        # Fonctions d'API
│       ├── auth.js       # Fonctions d'authentification
│       └── websocket.js  # Gestion WebSocket
└── package.json          # Dépendances et scripts
```

## Organisation des composants

Chaque composant suit la structure suivante :
```
components/
└── component-name/
    ├── index.html        # Template HTML
    ├── style.scss        # Styles spécifiques
    └── script.js         # Logique JavaScript
```

## Organisation des pages

Chaque page suit la structure suivante :
```
pages/
└── page-name/
    ├── index.html        # Template HTML
    ├── style.scss        # Styles spécifiques
    └── script.js         # Logique JavaScript
```

## Scripts disponibles

- `npm start` : Démarre le serveur de développement
- `npm build` : Compile le projet pour la production
- `npm test` : Lance les tests
- `npm lint` : Vérifie le code avec ESLint
- `npm format` : Formate le code avec Prettier

## Technologies utilisées

- HTML5
- SCSS pour les styles
- JavaScript (ES6+)
- WebSocket pour le chat en temps réel
- Fetch API pour les requêtes HTTP 