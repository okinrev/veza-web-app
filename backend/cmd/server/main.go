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
	frontendPath := filepath.Join(projectRoot, "frontend")

	log.Printf("📂 Chemin du frontend: %s", frontendPath)

	// Configurer les routes
	router := gin.Default()

	// Middleware pour servir les fichiers statiques
	router.Static("/assets", filepath.Join(frontendPath, "assets"))
	router.StaticFile("/favicon.ico", filepath.Join(frontendPath, "favicon.ico"))

	// Routes pour les pages HTML
	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "login.html"))
	})

	router.GET("/login", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "login.html"))
	})

	router.GET("/register", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "register.html"))
	})

	router.GET("/chat", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "chat.html"))
	})

	router.GET("/track", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "track.html"))
	})

	router.GET("/dashboard", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "dashboard.html"))
	})

	router.GET("/hub", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "hub.html"))
	})

	router.GET("/search", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "search.html"))
	})

	router.GET("/listings", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "listings.html"))
	})

	router.GET("/user_products", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "user_products.html"))
	})

	router.GET("/room", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "room.html"))
	})

	router.GET("/users", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "users.html"))
	})

	router.GET("/message", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "message.html"))
	})

	router.GET("/shared_ressources", func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "shared_ressources.html"))
	})

	// Redirection des routes sans extension vers les fichiers HTML
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		log.Printf("📂 Tentative d'accès à: %s", path)

		if !strings.HasSuffix(path, ".html") {
			htmlPath := filepath.Join(frontendPath, path+".html")
			log.Printf("🔍 Recherche du fichier: %s", htmlPath)

			if _, err := os.Stat(htmlPath); err == nil {
				log.Printf("✅ Fichier trouvé, envoi de: %s", htmlPath)
				c.File(htmlPath)
				return
			} else {
				log.Printf("❌ Fichier non trouvé: %s", htmlPath)
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
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

	// Servir les fichiers statiques
	router.Static("/public", "../frontend/public")

	// Routes principales
	router.StaticFile("/", "../frontend/public/login.html")
	router.StaticFile("/login", "../frontend/public/login.html")
	router.StaticFile("/register", "../frontend/public/register.html")
	router.StaticFile("/dashboard", "../frontend/public/dashboard.html")
	router.StaticFile("/shared_ressources", "../frontend/public/shared_ressources.html")
	router.StaticFile("/chat", "../frontend/public/chat.html")
	router.StaticFile("/profile", "../frontend/public/profile.html")
	router.StaticFile("/settings", "../frontend/public/settings.html")

	// Redirection des routes sans extension vers les fichiers HTML
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		log.Printf("📂 Tentative d'accès à: %s", path)

		if !strings.HasSuffix(path, ".html") {
			htmlPath := "../frontend/public" + path + ".html"
			log.Printf("🔍 Recherche du fichier: %s", htmlPath)

			if _, err := os.Stat(htmlPath); err == nil {
				log.Printf("✅ Fichier trouvé, envoi de: %s", htmlPath)
				c.File(htmlPath)
				return
			} else {
				log.Printf("❌ Fichier non trouvé: %s", htmlPath)
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})

	return router
}
