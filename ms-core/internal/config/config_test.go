package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Unset env vars to get defaults
	os.Unsetenv("PORT")
	os.Unsetenv("MONGO_URI")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("ENVIRONMENT")

	cfg := LoadConfig()

	if cfg.Port != "3003" {
		t.Errorf("expected Port 3003, got %s", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected Environment development, got %s", cfg.Environment)
	}
	if cfg.TeamFormation.Algorithm != "genetic" {
		t.Errorf("expected Algorithm genetic, got %s", cfg.TeamFormation.Algorithm)
	}
	if cfg.TeamFormation.DefaultWeight != 1.0 {
		t.Errorf("expected DefaultWeight 1.0, got %f", cfg.TeamFormation.DefaultWeight)
	}
	if cfg.TeamFormation.GeneticPopulationSize != 100 {
		t.Errorf("expected GeneticPopulationSize 100, got %d", cfg.TeamFormation.GeneticPopulationSize)
	}
	if cfg.TeamFormation.GeneticGenerations != 50 {
		t.Errorf("expected GeneticGenerations 50, got %d", cfg.TeamFormation.GeneticGenerations)
	}
	if cfg.TeamFormation.GeneticMutationRate != 0.1 {
		t.Errorf("expected GeneticMutationRate 0.1, got %f", cfg.TeamFormation.GeneticMutationRate)
	}
}

func TestLoadConfig_FromEnv(t *testing.T) {
	os.Setenv("PORT", "8080")
	os.Setenv("MONGO_URI", "mongodb://testhost:27017/testdb")
	os.Setenv("JWT_SECRET", "mysecret")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("TEAM_FORMATION_ALGORITHM", "greedy")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("MONGO_URI")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("TEAM_FORMATION_ALGORITHM")
	}()

	cfg := LoadConfig()

	if cfg.Port != "8080" {
		t.Errorf("expected Port 8080, got %s", cfg.Port)
	}
	if cfg.MongoURI != "mongodb://testhost:27017/testdb" {
		t.Errorf("unexpected MongoURI: %s", cfg.MongoURI)
	}
	if cfg.JWTSecret != "mysecret" {
		t.Errorf("unexpected JWTSecret: %s", cfg.JWTSecret)
	}
	if cfg.Environment != "production" {
		t.Errorf("unexpected Environment: %s", cfg.Environment)
	}
	if cfg.TeamFormation.Algorithm != "greedy" {
		t.Errorf("unexpected Algorithm: %s", cfg.TeamFormation.Algorithm)
	}
}

func TestLoadConfig_ProviderWeights(t *testing.T) {
	os.Setenv("TEAM_FORMATION_WEIGHT_SOCIOGRAM", "0.5")
	os.Setenv("TEAM_FORMATION_WEIGHT_ENNEAGRAM", "0.3")
	os.Setenv("TEAM_FORMATION_WEIGHT_MBTI", "0.2")
	defer func() {
		os.Unsetenv("TEAM_FORMATION_WEIGHT_SOCIOGRAM")
		os.Unsetenv("TEAM_FORMATION_WEIGHT_ENNEAGRAM")
		os.Unsetenv("TEAM_FORMATION_WEIGHT_MBTI")
	}()

	cfg := LoadConfig()

	if cfg.TeamFormation.ProviderWeights["sociogram"] != 0.5 {
		t.Errorf("expected sociogram weight 0.5, got %f", cfg.TeamFormation.ProviderWeights["sociogram"])
	}
	if cfg.TeamFormation.ProviderWeights["enneagram"] != 0.3 {
		t.Errorf("expected enneagram weight 0.3, got %f", cfg.TeamFormation.ProviderWeights["enneagram"])
	}
	if cfg.TeamFormation.ProviderWeights["mbti"] != 0.2 {
		t.Errorf("expected mbti weight 0.2, got %f", cfg.TeamFormation.ProviderWeights["mbti"])
	}
}

func TestLoadConfig_ExcludedProviders(t *testing.T) {
	os.Setenv("TEAM_FORMATION_EXCLUDED_PROVIDERS", "enneagram, mbti")
	defer os.Unsetenv("TEAM_FORMATION_EXCLUDED_PROVIDERS")

	cfg := LoadConfig()

	// Should still contain peer_evaluation plus the ones from env
	hasEnneagram := false
	hasMBTI := false
	for _, p := range cfg.TeamFormation.ExcludedProviders {
		if p == "enneagram" {
			hasEnneagram = true
		}
		if p == "mbti" {
			hasMBTI = true
		}
	}
	if !hasEnneagram {
		t.Error("expected enneagram in ExcludedProviders")
	}
	if !hasMBTI {
		t.Error("expected mbti in ExcludedProviders")
	}
}

func TestTeamFormationConfig_IsProviderExcluded(t *testing.T) {
	cfg := &TeamFormationConfig{
		ExcludedProviders: []string{"peer_evaluation", "enneagram"},
	}

	if !cfg.IsProviderExcluded("peer_evaluation") {
		t.Error("expected peer_evaluation to be excluded")
	}
	if !cfg.IsProviderExcluded("enneagram") {
		t.Error("expected enneagram to be excluded")
	}
	if cfg.IsProviderExcluded("sociogram") {
		t.Error("expected sociogram not to be excluded")
	}
	if cfg.IsProviderExcluded("mbti") {
		t.Error("expected mbti not to be excluded")
	}
}

func TestTeamFormationConfig_GetProviderWeight(t *testing.T) {
	cfg := &TeamFormationConfig{
		ProviderWeights: map[string]float64{
			"sociogram": 0.7,
			"enneagram": 0.3,
		},
		DefaultWeight: 1.0,
	}

	if cfg.GetProviderWeight("sociogram") != 0.7 {
		t.Errorf("expected sociogram weight 0.7, got %f", cfg.GetProviderWeight("sociogram"))
	}
	if cfg.GetProviderWeight("enneagram") != 0.3 {
		t.Errorf("expected enneagram weight 0.3, got %f", cfg.GetProviderWeight("enneagram"))
	}
	// Provider not in map → should return DefaultWeight
	if cfg.GetProviderWeight("mbti") != 1.0 {
		t.Errorf("expected mbti default weight 1.0, got %f", cfg.GetProviderWeight("mbti"))
	}
}

func TestGetEnvFloat_InvalidValue(t *testing.T) {
	os.Setenv("TEAM_FORMATION_GENETIC_MUTATION_RATE", "not_a_float")
	defer os.Unsetenv("TEAM_FORMATION_GENETIC_MUTATION_RATE")

	cfg := LoadConfig()
	// Invalid float should fall back to default (0.1)
	if cfg.TeamFormation.GeneticMutationRate != 0.1 {
		t.Errorf("expected fallback 0.1, got %f", cfg.TeamFormation.GeneticMutationRate)
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	os.Setenv("TEAM_FORMATION_GENETIC_POPULATION", "not_an_int")
	defer os.Unsetenv("TEAM_FORMATION_GENETIC_POPULATION")

	cfg := LoadConfig()
	// Invalid int should fall back to default (100)
	if cfg.TeamFormation.GeneticPopulationSize != 100 {
		t.Errorf("expected fallback 100, got %d", cfg.TeamFormation.GeneticPopulationSize)
	}
}
