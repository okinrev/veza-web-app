package main

import (
	"log"
	"net/http"
	"os"
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

	// Configurer les routes
	router := gin.Default()

	// Middleware pour servir les fichiers statiques
	router.Static("/assets", "./frontend/public/assets")
	router.StaticFile("/favicon.ico", "./frontend/public/favicon.ico")

	// Routes pour les pages HTML
	router.GET("/", func(c *gin.Context) {
		c.File("./frontend/public/index.html")
	})

	router.GET("/login", func(c *gin.Context) {
		c.File("./frontend/public/login.html")
	})

	router.GET("/register", func(c *gin.Context) {
		c.File("./frontend/public/register.html")
	})

	router.GET("/chat", func(c *gin.Context) {
		c.File("./frontend/public/chat.html")
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
