package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"ms-core/internal/core"
)

// EnneagramAdapter adapta un core.Plugin de tipo enneagram a la interfaz AffinityProvider.
// Trabaja exclusivamente contra la interfaz core.Plugin para no depender del paquete plugins.
type EnneagramAdapter struct {
	plugin    core.Plugin
	sectionID string
}

// NewEnneagramAdapter crea un EnneagramAdapter con un sectionID fijo.
func NewEnneagramAdapter(plugin core.Plugin, sectionID string) *EnneagramAdapter {
	return &EnneagramAdapter{
		plugin:    plugin,
		sectionID: sectionID,
	}
}

// NewEnneagramAffinityProvider crea un EnneagramAdapter sin sectionID fijo;
// el sectionID será extraído del contexto en cada llamada.
func NewEnneagramAffinityProvider(plugin core.Plugin) core.AffinityProvider {
	return &EnneagramAdapter{
		plugin:    plugin,
		sectionID: "",
	}
}

// ProviderID implementa AffinityProvider.
func (a *EnneagramAdapter) ProviderID() string {
	return "enneagram"
}

// ProviderName implementa AffinityProvider.
func (a *EnneagramAdapter) ProviderName() string {
	return "Enneagram Personality Test"
}

// CalculatePairwiseAffinity calcula la afinidad entre dos personas usando la
// matriz de compatibilidad de eneatipos. La compatibilidad se delega al plugin
// via "get_pairwise_compatibility" para que toda la lógica de eneagrama resida
// en el binario del plugin.
func (a *EnneagramAdapter) CalculatePairwiseAffinity(ctx context.Context, personA, personB string) (float64, error) {
	typeA, err := a.getEneatype(ctx, personA)
	if err != nil {
		return 0.5, err
	}

	typeB, err := a.getEneatype(ctx, personB)
	if err != nil {
		return 0.5, err
	}

	// Delegate to plugin so that the compatibility logic lives only in the plugin binary.
	raw, err := a.plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_pairwise_compatibility",
		"eneatype_a": float64(typeA),
		"eneatype_b": float64(typeB),
	})
	if err != nil {
		return 0.5, fmt.Errorf("enneagram_adapter: get_pairwise_compatibility failed: %w", err)
	}
	score, ok := raw.(float64)
	if !ok {
		return 0.5, fmt.Errorf("enneagram_adapter: unexpected result type from get_pairwise_compatibility: %T", raw)
	}
	return score, nil
}

// SupportsContext implementa AffinityProvider.
func (a *EnneagramAdapter) SupportsContext(ctx context.Context) bool {
	sectionID, ok := ctx.Value(core.ContextKeySectionID).(string)
	if !ok || sectionID == "" {
		return false
	}
	// Si el adaptador tiene sectionID fijo, verificar coincidencia.
	if a.sectionID != "" {
		return a.sectionID == sectionID
	}
	return true
}

// GetMetadata implementa AffinityProvider.
func (a *EnneagramAdapter) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{
		ID:          "enneagram",
		Name:        "Enneagram Personality Test",
		Description: "9-type personality system with compatibility matrix",
		Version:     "1.0.0",
		Category:    "personality",
	}
}

// PopulateOCEANVector rellena el vector OCEAN de un miembro del equipo
// basándose en su eneatipo dominante. El mapeo OCEAN se obtiene del plugin
// para que la tabla no resida en el binario del orquestador.
func (a *EnneagramAdapter) PopulateOCEANVector(ctx context.Context, member *core.TeamMember) error {
	eneatype, err := a.getEneatype(ctx, member.ID)
	if err != nil {
		return err
	}

	oceanVec, err := a.getOCEANFromPlugin(ctx, eneatype)
	if err != nil {
		return err
	}

	member.Psychometrics = oceanVec
	if member.RawType == nil {
		member.RawType = make(map[string]interface{})
	}
	member.RawType["enneagram"] = eneatype
	member.RawType["source"] = "enneagram_test"

	return nil
}

// PopulateSectionOCEANVectors rellena los vectores OCEAN de todos los miembros
// de una sección en una sola llamada al plugin.
func (a *EnneagramAdapter) PopulateSectionOCEANVectors(ctx context.Context, members []*core.TeamMember) error {
	eneatypes, err := a.getAllEneatypes(ctx)
	if err != nil {
		return err
	}

	for _, member := range members {
		eneatype, exists := eneatypes[member.ID]
		if !exists {
			continue
		}

		oceanVec, err := a.getOCEANFromPlugin(ctx, eneatype)
		if err != nil {
			continue
		}

		member.Psychometrics = oceanVec
		if member.RawType == nil {
			member.RawType = make(map[string]interface{})
		}
		member.RawType["enneagram"] = eneatype
		member.RawType["source"] = "enneagram_test"
	}

	return nil
}

// getOCEANFromPlugin asks the plugin for the OCEAN vector of a given eneatype.
// The plugin response (enneagramOCEAN) has JSON keys matching core.OCEANVector
// so a json round-trip produces the correct struct.
func (a *EnneagramAdapter) getOCEANFromPlugin(ctx context.Context, eneatype int) (*core.OCEANVector, error) {
	raw, err := a.plugin.Execute(ctx, map[string]interface{}{
		"action":   "get_ocean_from_type",
		"eneatype": float64(eneatype),
	})
	if err != nil {
		return nil, fmt.Errorf("enneagram_adapter: get_ocean_from_type failed: %w", err)
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("enneagram_adapter: failed to marshal OCEAN response: %w", err)
	}

	var v core.OCEANVector
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("enneagram_adapter: failed to unmarshal OCEAN response: %w", err)
	}

	return &v, nil
}

// ---------------------------------------------------------------------------
// helpers internos
// ---------------------------------------------------------------------------

func (a *EnneagramAdapter) resolveSectionID(ctx context.Context) (string, error) {
	if a.sectionID != "" {
		return a.sectionID, nil
	}
	if ctxSectionID, ok := ctx.Value(core.ContextKeySectionID).(string); ok && ctxSectionID != "" {
		return ctxSectionID, nil
	}
	return "", fmt.Errorf("enneagram_adapter: section_id not available in adapter or context")
}

func (a *EnneagramAdapter) getAllEneatypes(ctx context.Context) (map[string]int, error) {
	sectionID, err := a.resolveSectionID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := a.plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_student_eneatypes",
		"section_id": sectionID,
	})
	if err != nil {
		return nil, fmt.Errorf("enneagram_adapter: failed to get eneatypes: %w", err)
	}

	eneatypes, ok := result.(map[string]int)
	if !ok {
		return nil, fmt.Errorf("enneagram_adapter: unexpected result type from get_student_eneatypes: %T", result)
	}

	return eneatypes, nil
}

func (a *EnneagramAdapter) getEneatype(ctx context.Context, studentID string) (int, error) {
	eneatypes, err := a.getAllEneatypes(ctx)
	if err != nil {
		return 0, err
	}

	eneatype, exists := eneatypes[studentID]
	if !exists {
		return 0, fmt.Errorf("enneagram_adapter: no enneagram data for student %s", studentID)
	}

	return eneatype, nil
}

// studentIDToString convierte un ID de cualquier tipo a string.
func studentIDToString(studentID interface{}) string {
	switch v := studentID.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}
