package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	CORSOrigin         string
	ApiGatewayURL      string
}

func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://user:password@localhost/registro_curricular?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key"),
		Port:               getEnv("PORT", "3000"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:2999/auth/callback"),
		CORSOrigin:         getEnv("CORS_ORIGIN", "http://localhost:2999"),
		ApiGatewayURL:      getEnv("API_GATEWAY_URL", "http://localhost:3000"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
