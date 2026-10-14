package config

import (
	"os"
	"strconv"
	"strings"
)

// Config configuración general del microservicio
type Config struct {
	Port              string
	MongoURI          string
	JWTSecret         string
	AuthServiceURL    string
	CoursesServiceURL string
	APIGatewayURL     string
	Environment       string

	// TeamFormation configuración del motor de formación de equipos
	TeamFormation TeamFormationConfig
}

// TeamFormationConfig configuración específica para el motor de formación de equipos
type TeamFormationConfig struct {
	// Algorithm: "greedy" o "genetic"
	Algorithm string

	// ProviderWeights pesos para cada AffinityProvider
	// Ejemplo: {"sociogram": 0.7, "enneagram": 0.3}
	ProviderWeights map[string]float64

	// DefaultWeight peso por defecto para providers sin configuración
	DefaultWeight float64

	// GeneticPopulationSize tamaño de población para algoritmo genético
	GeneticPopulationSize int

	// GeneticGenerations número de generaciones para algoritmo genético
	GeneticGenerations int

	// GeneticMutationRate tasa de mutación (0.0 - 1.0)
	GeneticMutationRate float64

	// ExcludedProviders lista de providers que NO deben usarse en formación
	// Por defecto: ["peer_evaluation"]
	ExcludedProviders []string
}

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "3003"),
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017/ms_core"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:3001"),
		CoursesServiceURL: getEnv("COURSES_SERVICE_URL", "http://localhost:3002"),
		APIGatewayURL:     getEnv("API_GATEWAY_URL", "http://localhost:3000"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		TeamFormation:     loadTeamFormationConfig(),
	}
}

// loadTeamFormationConfig carga la configuración de formación de equipos
func loadTeamFormationConfig() TeamFormationConfig {
	config := TeamFormationConfig{
		Algorithm:             getEnv("TEAM_FORMATION_ALGORITHM", "genetic"),
		DefaultWeight:         getEnvFloat("TEAM_FORMATION_DEFAULT_WEIGHT", 1.0),
		GeneticPopulationSize: getEnvInt("TEAM_FORMATION_GENETIC_POPULATION", 100),
		GeneticGenerations:    getEnvInt("TEAM_FORMATION_GENETIC_GENERATIONS", 50),
		GeneticMutationRate:   getEnvFloat("TEAM_FORMATION_GENETIC_MUTATION_RATE", 0.1),
		ProviderWeights:       make(map[string]float64),
		ExcludedProviders:     []string{"peer_evaluation"}, // Peer evaluation NO participa en formación
	}

	// Cargar pesos de providers desde variables de entorno
	// Formato: TEAM_FORMATION_WEIGHT_SOCIOGRAM=0.7
	//          TEAM_FORMATION_WEIGHT_ENNEAGRAM=0.3
	//          TEAM_FORMATION_WEIGHT_MBTI=0.0
	sociogramWeight := getEnvFloat("TEAM_FORMATION_WEIGHT_SOCIOGRAM", 0.7)
	enneagramWeight := getEnvFloat("TEAM_FORMATION_WEIGHT_ENNEAGRAM", 0.3)
	mbtiWeight := getEnvFloat("TEAM_FORMATION_WEIGHT_MBTI", 0.0)

	config.ProviderWeights["sociogram"] = sociogramWeight
	config.ProviderWeights["enneagram"] = enneagramWeight
	config.ProviderWeights["mbti"] = mbtiWeight

	// Cargar providers excluidos adicionales
	if excluded := getEnv("TEAM_FORMATION_EXCLUDED_PROVIDERS", ""); excluded != "" {
		for _, p := range strings.Split(excluded, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				config.ExcludedProviders = append(config.ExcludedProviders, p)
			}
		}
	}

	return config
}

// IsProviderExcluded verifica si un provider está excluido de la formación
func (c *TeamFormationConfig) IsProviderExcluded(providerID string) bool {
	for _, excluded := range c.ExcludedProviders {
		if excluded == providerID {
			return true
		}
	}
	return false
}

// GetProviderWeight obtiene el peso de un provider (o el default si no está configurado)
func (c *TeamFormationConfig) GetProviderWeight(providerID string) float64 {
	if weight, ok := c.ProviderWeights[providerID]; ok {
		return weight
	}
	return c.DefaultWeight
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
