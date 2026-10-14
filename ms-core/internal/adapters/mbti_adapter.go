package adapters

import (
	"context"
	"fmt"

	"ms-core/internal/core"
)

// MBTIAffinityProvider adapta un core.Plugin de tipo MBTI a la interfaz AffinityProvider.
// Trabaja exclusivamente contra la interfaz core.Plugin para no depender del paquete plugins.
type MBTIAffinityProvider struct {
	plugin core.Plugin
}

// NewMBTIAffinityProvider crea un nuevo adaptador para MBTI.
func NewMBTIAffinityProvider(plugin core.Plugin) core.AffinityProvider {
	return &MBTIAffinityProvider{
		plugin: plugin,
	}
}

// ProviderID implementa AffinityProvider.
func (p *MBTIAffinityProvider) ProviderID() string {
	return "mbti"
}

// ProviderName implementa AffinityProvider.
func (p *MBTIAffinityProvider) ProviderName() string {
	return "Myers-Briggs Type Indicator"
}

// CalculatePairwiseAffinity calcula la afinidad entre dos personas usando sus
// tipos MBTI. Los tipos se obtienen del plugin via Execute con la acción
// "get_student_types". La compatibilidad se delega al plugin via "get_compatibility"
// para que toda la lógica MBTI resida en el binario del plugin.
func (p *MBTIAffinityProvider) CalculatePairwiseAffinity(ctx context.Context, personA, personB string) (float64, error) {
	typeA, err := p.getMBTIType(ctx, personA)
	if err != nil {
		return 0.5, err
	}

	typeB, err := p.getMBTIType(ctx, personB)
	if err != nil {
		return 0.5, err
	}

	// Delegate to plugin so that the compatibility logic lives only in the plugin binary.
	raw, err := p.plugin.Execute(ctx, map[string]interface{}{
		"action": "get_compatibility",
		"type_a": typeA,
		"type_b": typeB,
	})
	if err != nil {
		return 0.5, fmt.Errorf("mbti_adapter: get_compatibility failed: %w", err)
	}
	score, ok := raw.(float64)
	if !ok {
		return 0.5, fmt.Errorf("mbti_adapter: unexpected result type from get_compatibility: %T", raw)
	}
	return score, nil
}

// SupportsContext implementa AffinityProvider.
// Requiere que el contexto tenga un section_id válido.
func (p *MBTIAffinityProvider) SupportsContext(ctx context.Context) bool {
	sectionID, ok := ctx.Value(core.ContextKeySectionID).(string)
	return ok && sectionID != ""
}

// GetMetadata implementa AffinityProvider.
func (p *MBTIAffinityProvider) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{
		ID:          "mbti",
		Name:        "Myers-Briggs Type Indicator",
		Description: "Calcula compatibilidad basada en tipos de personalidad MBTI",
		Version:     "1.0.0",
		Category:    "personality",
	}
}

// ---------------------------------------------------------------------------
// helpers internos
// ---------------------------------------------------------------------------

func (p *MBTIAffinityProvider) resolveSectionID(ctx context.Context) (string, error) {
	sectionID, ok := ctx.Value(core.ContextKeySectionID).(string)
	if !ok || sectionID == "" {
		return "", fmt.Errorf("mbti_adapter: section_id not available in context")
	}
	return sectionID, nil
}

func (p *MBTIAffinityProvider) getAllMBTITypes(ctx context.Context) (map[string]string, error) {
	sectionID, err := p.resolveSectionID(ctx)
	if err != nil {
		return nil, err
	}

	result, err := p.plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_student_types",
		"section_id": sectionID,
	})
	if err != nil {
		return nil, fmt.Errorf("mbti_adapter: failed to get student types: %w", err)
	}

	types, ok := result.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("mbti_adapter: unexpected result type from get_student_types: %T", result)
	}

	return types, nil
}

func (p *MBTIAffinityProvider) getMBTIType(ctx context.Context, studentID string) (string, error) {
	types, err := p.getAllMBTITypes(ctx)
	if err != nil {
		return "", err
	}

	mbtiType, exists := types[studentID]
	if !exists {
		return "", fmt.Errorf("mbti_adapter: no MBTI data for student %s", studentID)
	}

	return mbtiType, nil
}
