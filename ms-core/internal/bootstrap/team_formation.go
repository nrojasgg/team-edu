package bootstrap

import (
	"fmt"
	"log"

	"ms-core/internal/adapters"
	"ms-core/internal/config"
	"ms-core/internal/core"
)

// TeamFormationDependencies contiene todas las dependencias inicializadas
type TeamFormationDependencies struct {
	Engine   *core.TeamFormationEngine
	Registry *core.ProviderRegistry
	// Plugin es el subproceso team_formation cargado via hashicorp/go-plugin.
	// Puede ser nil si el binario no está en plugins_dist/ al arrancar.
	Plugin core.Plugin
}

// InitializeTeamFormation inicializa el motor de formación de equipos con los
// AffinityProviders disponibles. Los plugins se recuperan del PluginManager,
// por lo que no hay dependencia directa con el paquete plugins.
//
// Si un plugin no está cargado se loguea una advertencia y se continúa;
// el motor funcionará con los providers que sí estén disponibles.
func InitializeTeamFormation(cfg *config.Config, pm *core.PluginManager) (*TeamFormationDependencies, error) {
	engineConfig := &core.TeamFormationConfig{
		Weights:       cfg.TeamFormation.ProviderWeights,
		Algorithm:     cfg.TeamFormation.Algorithm,
		DefaultWeight: cfg.TeamFormation.DefaultWeight,
		GeneticParams: &core.GeneticAlgorithmParams{
			PopulationSize: cfg.TeamFormation.GeneticPopulationSize,
			Generations:    cfg.TeamFormation.GeneticGenerations,
			MutationRate:   cfg.TeamFormation.GeneticMutationRate,
			ElitismRate:    0.1,
		},
	}

	registry := core.NewProviderRegistry(&cfg.TeamFormation)
	engine := core.NewTeamFormationEngine(engineConfig)

	registerPluginProvider(pm, registry, "enneagram", func(p core.Plugin) core.AffinityProvider {
		return adapters.NewEnneagramAffinityProvider(p)
	})
	registerPluginProvider(pm, registry, "group_compatibility", func(p core.Plugin) core.AffinityProvider {
		return adapters.NewSociogramAffinityProvider(p)
	})
	registerPluginProvider(pm, registry, "mbti", func(p core.Plugin) core.AffinityProvider {
		return adapters.NewMBTIAffinityProvider(p)
	})

	// Registrar en el engine solo los providers elegibles según la configuración
	eligibleProviders := registry.GetWeightedEligibleProviders()
	if len(eligibleProviders) == 0 {
		return nil, fmt.Errorf("bootstrap: no eligible affinity providers available for team formation")
	}

	for _, wp := range eligibleProviders {
		if regErr := engine.RegisterProvider(wp.Provider); regErr != nil {
			log.Printf("bootstrap: failed to register provider %s in engine: %v",
				wp.Provider.ProviderID(), regErr)
		}
	}

	summary := registry.Summary()
	log.Printf("Team Formation Engine initialized:")
	log.Printf("  - Algorithm: %s", summary["algorithm"])
	log.Printf("  - Total providers: %d", summary["total_providers"])
	log.Printf("  - Eligible providers: %v", summary["eligible_providers"])
	log.Printf("  - Excluded providers: %v", summary["excluded_providers"])

	deps := &TeamFormationDependencies{
		Engine:   engine,
		Registry: registry,
	}

	// Intentar adjuntar el plugin team_formation si está cargado.
	// Su ausencia no es fatal: el handler hace fallback al engine integrado.
	if tfPlugin, pluginErr := pm.GetPlugin("team_formation"); pluginErr == nil {
		deps.Plugin = tfPlugin
		log.Printf("bootstrap: team_formation plugin attached (subprocess mode)")
	} else {
		log.Printf("bootstrap: team_formation plugin not found, using embedded engine (%v)", pluginErr)
	}

	return deps, nil
}

// registerPluginProvider attempts to get a named plugin from pm, wraps it with factory,
// and registers the resulting AffinityProvider in registry. Errors are logged as warnings.
func registerPluginProvider(
	pm *core.PluginManager,
	registry *core.ProviderRegistry,
	pluginName string,
	factory func(core.Plugin) core.AffinityProvider,
) {
	plugin, err := pm.GetPlugin(pluginName)
	if err != nil {
		log.Printf("bootstrap: %s plugin not loaded, skipping provider: %v", pluginName, err)
		return
	}
	provider := factory(plugin)
	if regErr := registry.Register(provider); regErr != nil {
		log.Printf("bootstrap: failed to register %s provider: %v", pluginName, regErr)
	}
}

// LogConfiguration imprime la configuración actual del motor
func LogConfiguration(cfg *config.TeamFormationConfig) {
	log.Println("=== Team Formation Configuration ===")
	log.Printf("Algorithm: %s", cfg.Algorithm)
	log.Printf("Default Weight: %.2f", cfg.DefaultWeight)
	log.Println("Provider Weights:")
	for id, weight := range cfg.ProviderWeights {
		log.Printf("  - %s: %.2f (%.0f%%)", id, weight, weight*100)
	}
	log.Printf("Excluded Providers: %v", cfg.ExcludedProviders)
	log.Println("=====================================")
}
