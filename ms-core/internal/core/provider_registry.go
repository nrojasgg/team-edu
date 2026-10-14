package core

import (
	"fmt"
	"log"
	"sync"

	"ms-core/internal/config"
)

// ProviderRegistry mantiene un registro de todos los AffinityProviders disponibles
// y gestiona cuáles son elegibles para la formación de equipos
type ProviderRegistry struct {
	providers map[string]AffinityProvider
	config    *config.TeamFormationConfig
	mu        sync.RWMutex
}

// NewProviderRegistry crea una nueva instancia del registro
func NewProviderRegistry(cfg *config.TeamFormationConfig) *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]AffinityProvider),
		config:    cfg,
	}
}

// Register registra un nuevo AffinityProvider
func (r *ProviderRegistry) Register(provider AffinityProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	id := provider.ProviderID()
	if id == "" {
		return fmt.Errorf("provider ID cannot be empty")
	}

	if _, exists := r.providers[id]; exists {
		return fmt.Errorf("provider already registered: %s", id)
	}

	r.providers[id] = provider

	// Log si está excluido de formación
	if r.config != nil && r.config.IsProviderExcluded(id) {
		log.Printf("AffinityProvider registered (EXCLUDED from team formation): %s", id)
	} else {
		log.Printf("AffinityProvider registered: %s", id)
	}

	return nil
}

// Get obtiene un provider por su ID
func (r *ProviderRegistry) Get(id string) (AffinityProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[id]
	return provider, exists
}

// GetEligibleProviders retorna solo los providers elegibles para formación de equipos
// (excluye los que están en la lista de exclusión como peer_evaluation)
func (r *ProviderRegistry) GetEligibleProviders() []AffinityProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var eligible []AffinityProvider
	for id, provider := range r.providers {
		if r.config == nil || !r.config.IsProviderExcluded(id) {
			eligible = append(eligible, provider)
		}
	}
	return eligible
}

// GetAllProviders retorna todos los providers registrados
func (r *ProviderRegistry) GetAllProviders() []AffinityProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]AffinityProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// GetProviderIDs retorna los IDs de todos los providers registrados
func (r *ProviderRegistry) GetProviderIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}

// IsExcluded verifica si un provider está excluido de la formación de equipos
func (r *ProviderRegistry) IsExcluded(providerID string) bool {
	if r.config == nil {
		return false
	}
	return r.config.IsProviderExcluded(providerID)
}

// HasProvider verifica si un provider está registrado
func (r *ProviderRegistry) HasProvider(providerID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.providers[providerID]
	return exists
}

// GetWeightedEligibleProviders retorna los providers elegibles con sus pesos
func (r *ProviderRegistry) GetWeightedEligibleProviders() []WeightedProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var weighted []WeightedProvider
	for id, provider := range r.providers {
		if r.config == nil || !r.config.IsProviderExcluded(id) {
			weight := r.config.DefaultWeight
			if r.config != nil {
				weight = r.config.GetProviderWeight(id)
			}
			weighted = append(weighted, WeightedProvider{
				Provider: provider,
				Weight:   weight,
			})
		}
	}
	return weighted
}

// Summary retorna un resumen del estado del registro
func (r *ProviderRegistry) Summary() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	eligible := make([]map[string]interface{}, 0)
	excluded := make([]string, 0)

	for id, provider := range r.providers {
		meta := provider.GetMetadata()
		if r.config != nil && r.config.IsProviderExcluded(id) {
			excluded = append(excluded, id)
		} else {
			weight := r.config.DefaultWeight
			if r.config != nil {
				weight = r.config.GetProviderWeight(id)
			}
			eligible = append(eligible, map[string]interface{}{
				"id":       id,
				"name":     meta.Name,
				"category": meta.Category,
				"weight":   weight,
			})
		}
	}

	return map[string]interface{}{
		"total_providers":    len(r.providers),
		"eligible_providers": eligible,
		"excluded_providers": excluded,
		"algorithm":          r.config.Algorithm,
	}
}
