package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/okinrev/veza-web-app/internal/api"
	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/websocket"
)

// getProjectRoot retourne le chemin absolu vers la racine du projet
func getProjectRoot() string {
	// Obtenir le chemin du répertoire de travail actuel
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Erreur lors de la récupération du répertoire de travail:", err)
	}
	// Remonter d'un niveau depuis le dossier backend
	return filepath.Dir(wd)
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Configuration
	cfg := config.New()

	// Mode Gin
	if cfg.Server.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database
	db, err := database.NewConnection(cfg.Database.URL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Migrations
	if err := database.RunMigrations(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Obtenir le chemin absolu du projet
	projectRoot := getProjectRoot()
	frontendPath := filepath.Join(projectRoot, "talas-frontend", "dist")

	log.Printf("📂 Chemin du nouveau frontend React: %s", frontendPath)

	// Configurer les routes
	router := gin.Default()

	// Middleware pour servir les fichiers statiques du build React
	router.Static("/assets", filepath.Join(frontendPath, "assets"))
	router.StaticFile("/favicon.ico", filepath.Join(frontendPath, "favicon.ico"))

	// Pour une SPA React, on sert toujours index.html pour les routes client-side
	router.NoRoute(func(c *gin.Context) {
		// Si c'est une requête API, renvoyer 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		
		// Pour toutes les autres routes, servir index.html (SPA routing)
		indexPath := filepath.Join(frontendPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			// Fallback: servir depuis le dev server si en développement
			c.JSON(http.StatusNotFound, gin.H{"error": "Frontend not built. Run 'npm run build' in talas-frontend/"})
		}
	})

	// Initialiser le gestionnaire WebSocket
	chatManager := websocket.NewChatManager(cfg.JWT.Secret)
	go chatManager.Run()

	// Route WebSocket pour le chat
	router.GET("/ws/chat", func(c *gin.Context) {
		chatManager.HandleWebSocket(c)
	})

	// Configurer les routes API
	api.SetupRoutes(router, db, cfg)

	// Démarrer le serveur
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// Middleware globaux
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     "talas-backend",
			"version":     "1.0.0",
			"timestamp":   time.Now().Unix(),
			"environment": cfg.Server.Environment,
		})
	})

	// Servir les fichiers statiques du build React
	frontendDistPath := "../talas-frontend/dist"
	router.Static("/assets", frontendDistPath+"/assets")
	router.StaticFile("/favicon.ico", frontendDistPath+"/favicon.ico")

	// Configuration SPA React - toutes les routes non-API servent index.html
	router.NoRoute(func(c *gin.Context) {
		// Si c'est une requête API, renvoyer 404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		
		// Pour toutes les autres routes, servir index.html (SPA routing)
		indexPath := frontendDistPath + "/index.html"
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			// Message plus clair en développement
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Frontend not built", 
				"message": "Run 'npm run build' in talas-frontend/ directory",
				"note": "Or use 'npm run dev' for development server on port 5173",
			})
		}
	})

	return router
}
