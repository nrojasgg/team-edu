package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	AuthServiceURL string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return &Config{
		Port:           getEnv("PORT", "3002"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:1234@localhost:5433/courses_db?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:3001"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
