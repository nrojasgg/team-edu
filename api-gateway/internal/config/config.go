package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GatewayPort           string
	RegistroCurricularURL string
	CoursesURL            string
	CoreURL               string
	NotificationsURL      string
	JWTSecret             string
}

func LoadConfig() *Config {
	// Cargar archivo .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		GatewayPort:           getEnv("GATEWAY_PORT", "3000"),
		RegistroCurricularURL: getEnv("REGISTRO_CURRICULAR_URL", "http://localhost:3001"),
		CoursesURL:            getEnv("COURSES_URL", "http://localhost:3002"),
		CoreURL:               getEnv("CORE_URL", "http://localhost:3003"),
		NotificationsURL:      getEnv("NOTIFICATIONS_URL", "http://localhost:3004"),
		JWTSecret:             getEnv("JWT_SECRET", "default-secret-key"),
	}

	log.Printf("Config loaded: Gateway Port=%s, Registro Curricular URL=%s, Courses URL=%s, Core URL=%s, Notifications URL=%s",
		config.GatewayPort, config.RegistroCurricularURL, config.CoursesURL, config.CoreURL, config.NotificationsURL)

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
