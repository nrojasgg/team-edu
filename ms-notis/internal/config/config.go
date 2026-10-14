package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string
	SMTPFromName  string
	FrontendURL   string
	AppEnv        string
	LogLevel      string
	CORSOrigin    string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		Port:          getEnv("PORT", "3004"),
		SMTPHost:      getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail: getEnv("SMTP_FROM_EMAIL", "noreply@TeamEdu.cl"),
		SMTPFromName:  getEnv("SMTP_FROM_NAME", "TeamEdu"),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:2999"),
		AppEnv:        getEnv("APP_ENV", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:2999"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
