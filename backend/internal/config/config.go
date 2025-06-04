package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
}

type ServerConfig struct {
	Port string
	Host string
}

func Load() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://veza:password@localhost:5432/veza_db?sslmode=disable"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default-secret-change-in-production"),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "localhost"),
		},
	}

	if config.JWT.Secret == "default-secret-change-in-production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
