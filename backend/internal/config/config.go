package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
	
	// File upload settings
	MaxFileSize    int64
	UploadPath     string
	AllowedFormats []string
	
	// CORS settings
	AllowedOrigins []string
	
	// Rate limiting
	RateLimitEnabled bool
	RateLimitRPM     int
}

// New creates a new configuration with values from environment variables
func New() *Config {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/veza_db?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-256-bit-secret"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		
		// File upload defaults
		MaxFileSize:    10 * 1024 * 1024, // 10MB
		UploadPath:     getEnv("UPLOAD_PATH", "./static/uploads"),
		AllowedFormats: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".mp3", ".wav", ".mp4"},
		
		// CORS defaults
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		},
		
		// Rate limiting defaults
		RateLimitEnabled: getEnv("RATE_LIMIT_ENABLED", "true") == "true",
		RateLimitRPM:     60,
	}

	// Validate required fields
	if cfg.JWTSecret == "your-256-bit-secret" && cfg.Environment == "production" {
		panic("JWT_SECRET must be set for production environment")
	}

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}