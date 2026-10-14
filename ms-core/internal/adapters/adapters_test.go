package adapters

import (
	"context"
	"testing"

	"ms-core/internal/core"
)

// ---------------------------------------------------------------------------
// mock Plugin for adapter tests
// ---------------------------------------------------------------------------

type mockPlugin struct {
	name    string
	version string
	execFn  func(ctx context.Context, data interface{}) (interface{}, error)
}

func (m *mockPlugin) Name() string      { return m.name }
func (m *mockPlugin) Version() string   { return m.version }
func (m *mockPlugin) Initialize() error { return nil }
func (m *mockPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	if m.execFn != nil {
		return m.execFn(ctx, data)
	}
	return nil, nil
}
func (m *mockPlugin) Shutdown() error { return nil }

// sectionCtx returns a context carrying a section_id value.
func sectionCtx(sectionID string) context.Context {
	return context.WithValue(context.Background(), core.ContextKeySectionID, sectionID)
}

// ---------------------------------------------------------------------------
// EnneagramAdapter tests
// ---------------------------------------------------------------------------

func TestEnneagramAdapter_ProviderID(t *testing.T) {
	a := NewEnneagramAffinityProvider(&mockPlugin{name: "enneagram", version: "1.0.0"})
	if a.ProviderID() != "enneagram" {
		t.Errorf("unexpected ProviderID: %s", a.ProviderID())
	}
}

func TestEnneagramAdapter_ProviderName(t *testing.T) {
	a := NewEnneagramAffinityProvider(&mockPlugin{name: "enneagram", version: "1.0.0"})
	if a.ProviderName() == "" {
		t.Error("expected non-empty ProviderName")
	}
}

func TestEnneagramAdapter_GetMetadata(t *testing.T) {
	a := NewEnneagramAffinityProvider(&mockPlugin{name: "enneagram", version: "1.0.0"})
	meta := a.GetMetadata()
	if meta.ID != "enneagram" {
		t.Errorf("unexpected metadata ID: %s", meta.ID)
	}
}

func TestEnneagramAdapter_SupportsContext_WithSectionID(t *testing.T) {
	a := NewEnneagramAffinityProvider(&mockPlugin{name: "enneagram", version: "1.0.0"})
	ctx := sectionCtx("section-1")
	if !a.SupportsContext(ctx) {
		t.Error("expected SupportsContext=true when section_id is present")
	}
}

func TestEnneagramAdapter_SupportsContext_NoSectionID(t *testing.T) {
	a := NewEnneagramAffinityProvider(&mockPlugin{name: "enneagram", version: "1.0.0"})
	if a.SupportsContext(context.Background()) {
		t.Error("expected SupportsContext=false when section_id is absent")
	}
}

func TestEnneagramAdapter_SupportsContext_FixedSectionMatch(t *testing.T) {
	a := NewEnneagramAdapter(&mockPlugin{name: "enneagram", version: "1.0.0"}, "sec-42")
	if !a.SupportsContext(sectionCtx("sec-42")) {
		t.Error("expected SupportsContext=true for matching fixed sectionID")
	}
	if a.SupportsContext(sectionCtx("sec-99")) {
		t.Error("expected SupportsContext=false for non-matching fixed sectionID")
	}
}

func TestEnneagramAdapter_CalculatePairwiseAffinity(t *testing.T) {
	calls := 0
	p := &mockPlugin{
		name:    "enneagram",
		version: "1.0.0",
		execFn: func(ctx context.Context, data interface{}) (interface{}, error) {
			req := data.(map[string]interface{})
			action, _ := req["action"].(string)
			calls++
			switch action {
			case "get_student_eneatypes":
				// adapter expects map[string]int directly
				return map[string]int{"A": 1, "B": 3}, nil
			case "get_pairwise_compatibility":
				return float64(0.75), nil
			}
			return nil, nil
		},
	}
	a := NewEnneagramAffinityProvider(p)
	ctx := sectionCtx("s1")
	score, err := a.CalculatePairwiseAffinity(ctx, "A", "B")
	if err != nil {
		t.Fatalf("CalculatePairwiseAffinity: %v", err)
	}
	if score != 0.75 {
		t.Errorf("expected 0.75, got %f", score)
	}
}

func TestEnneagramAdapter_CalculatePairwiseAffinity_NoSectionID(t *testing.T) {
	a := NewEnneagramAffinityProvider(&mockPlugin{name: "enneagram", version: "1.0.0"})
	_, err := a.CalculatePairwiseAffinity(context.Background(), "A", "B")
	if err == nil {
		t.Error("expected error without section_id in context")
	}
}

// ---------------------------------------------------------------------------
// MBTIAffinityProvider tests
// ---------------------------------------------------------------------------

func TestMBTIAdapter_ProviderID(t *testing.T) {
	a := NewMBTIAffinityProvider(&mockPlugin{name: "mbti", version: "1.0.0"})
	if a.ProviderID() != "mbti" {
		t.Errorf("unexpected ProviderID: %s", a.ProviderID())
	}
}

func TestMBTIAdapter_ProviderName(t *testing.T) {
	a := NewMBTIAffinityProvider(&mockPlugin{name: "mbti", version: "1.0.0"})
	if a.ProviderName() == "" {
		t.Error("expected non-empty ProviderName")
	}
}

func TestMBTIAdapter_GetMetadata(t *testing.T) {
	a := NewMBTIAffinityProvider(&mockPlugin{name: "mbti", version: "1.0.0"})
	meta := a.GetMetadata()
	if meta.ID != "mbti" {
		t.Errorf("unexpected metadata ID: %s", meta.ID)
	}
}

func TestMBTIAdapter_SupportsContext_WithSectionID(t *testing.T) {
	a := NewMBTIAffinityProvider(&mockPlugin{name: "mbti", version: "1.0.0"})
	if !a.SupportsContext(sectionCtx("sec-1")) {
		t.Error("expected SupportsContext=true")
	}
}

func TestMBTIAdapter_SupportsContext_NoSectionID(t *testing.T) {
	a := NewMBTIAffinityProvider(&mockPlugin{name: "mbti", version: "1.0.0"})
	if a.SupportsContext(context.Background()) {
		t.Error("expected SupportsContext=false without section_id")
	}
}

func TestMBTIAdapter_CalculatePairwiseAffinity(t *testing.T) {
	p := &mockPlugin{
		name:    "mbti",
		version: "1.0.0",
		execFn: func(ctx context.Context, data interface{}) (interface{}, error) {
			req := data.(map[string]interface{})
			action, _ := req["action"].(string)
			switch action {
			case "get_student_types":
				// adapter expects map[string]string directly
				return map[string]string{"A": "INTJ", "B": "ENFP"}, nil
			case "get_compatibility":
				return float64(0.6), nil
			}
			return nil, nil
		},
	}
	a := NewMBTIAffinityProvider(p)
	ctx := sectionCtx("s1")
	score, err := a.CalculatePairwiseAffinity(ctx, "A", "B")
	if err != nil {
		t.Fatalf("CalculatePairwiseAffinity: %v", err)
	}
	if score != 0.6 {
		t.Errorf("expected 0.6, got %f", score)
	}
}

func TestMBTIAdapter_CalculatePairwiseAffinity_NoSectionID(t *testing.T) {
	a := NewMBTIAffinityProvider(&mockPlugin{name: "mbti", version: "1.0.0"})
	_, err := a.CalculatePairwiseAffinity(context.Background(), "A", "B")
	if err == nil {
		t.Error("expected error without section_id in context")
	}
}

// ---------------------------------------------------------------------------
// SociogramAffinityProvider tests
// ---------------------------------------------------------------------------

func TestSociogramAdapter_ProviderID(t *testing.T) {
	a := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram", version: "1.0.0"})
	if a.ProviderID() != "sociogram" {
		t.Errorf("unexpected ProviderID: %s", a.ProviderID())
	}
}

func TestSociogramAdapter_ProviderName(t *testing.T) {
	a := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram", version: "1.0.0"})
	if a.ProviderName() == "" {
		t.Error("expected non-empty ProviderName")
	}
}

func TestSociogramAdapter_GetMetadata(t *testing.T) {
	a := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram", version: "1.0.0"})
	meta := a.GetMetadata()
	if meta.ID != "sociogram" {
		t.Errorf("unexpected metadata ID: %s", meta.ID)
	}
}

func TestSociogramAdapter_SupportsContext_WithSectionID(t *testing.T) {
	a := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram", version: "1.0.0"})
	if !a.SupportsContext(sectionCtx("s1")) {
		t.Error("expected SupportsContext=true")
	}
}

func TestSociogramAdapter_SupportsContext_NoSectionID(t *testing.T) {
	a := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram", version: "1.0.0"})
	if a.SupportsContext(context.Background()) {
		t.Error("expected SupportsContext=false without section_id")
	}
}

// ---------------------------------------------------------------------------
// pairChoiceScore / pairRejectionScore / clampNorm unit tests
// ---------------------------------------------------------------------------

func TestPairChoiceScore_MutualChoice(t *testing.T) {
	score, ok := pairChoiceScore(true, true)
	if !ok || score != 1.0 {
		t.Errorf("expected (1.0, true), got (%f, %v)", score, ok)
	}
}

func TestPairChoiceScore_UnilateralA(t *testing.T) {
	score, ok := pairChoiceScore(true, false)
	if !ok || score != 0.5 {
		t.Errorf("expected (0.5, true), got (%f, %v)", score, ok)
	}
}

func TestPairChoiceScore_UnilateralB(t *testing.T) {
	score, ok := pairChoiceScore(false, true)
	if !ok || score != 0.5 {
		t.Errorf("expected (0.5, true), got (%f, %v)", score, ok)
	}
}

func TestPairChoiceScore_NoChoice(t *testing.T) {
	score, ok := pairChoiceScore(false, false)
	if ok || score != 0 {
		t.Errorf("expected (0, false), got (%f, %v)", score, ok)
	}
}

func TestPairRejectionScore_MutualRejection(t *testing.T) {
	score, ok := pairRejectionScore(true, true)
	if !ok || score != -1.0 {
		t.Errorf("expected (-1.0, true), got (%f, %v)", score, ok)
	}
}

func TestPairRejectionScore_Unilateral(t *testing.T) {
	score, ok := pairRejectionScore(true, false)
	if !ok || score != -0.5 {
		t.Errorf("expected (-0.5, true), got (%f, %v)", score, ok)
	}
}

func TestPairRejectionScore_NoRejection(t *testing.T) {
	score, ok := pairRejectionScore(false, false)
	if ok || score != 0 {
		t.Errorf("expected (0, false), got (%f, %v)", score, ok)
	}
}

func TestClampNorm_AboveMax(t *testing.T) {
	if clampNorm(1.5) != 1.0 {
		t.Error("expected 1.0")
	}
}

func TestClampNorm_BelowMin(t *testing.T) {
	if clampNorm(-1.5) != -1.0 {
		t.Error("expected -1.0")
	}
}

func TestClampNorm_Within(t *testing.T) {
	if clampNorm(0.5) != 0.5 {
		t.Error("expected 0.5")
	}
}

// ---------------------------------------------------------------------------
// toSliceOfMaps / isResponseOne / fillBoolMap / ensureInnerMap
// ---------------------------------------------------------------------------

func TestToSliceOfMaps_Nil(t *testing.T) {
	result, ok := toSliceOfMaps(nil)
	if ok || result != nil {
		t.Error("expected (nil, false) for nil input")
	}
}

func TestToSliceOfMaps_Exact(t *testing.T) {
	input := []map[string]interface{}{{"a": 1}}
	result, ok := toSliceOfMaps(input)
	if !ok || len(result) != 1 {
		t.Errorf("expected (slice, true), got (%v, %v)", result, ok)
	}
}

func TestToSliceOfMaps_InterfaceSlice(t *testing.T) {
	input := []interface{}{map[string]interface{}{"x": "y"}}
	result, ok := toSliceOfMaps(input)
	if !ok || len(result) != 1 {
		t.Errorf("expected (slice, true), got (%v, %v)", result, ok)
	}
}

func TestToSliceOfMaps_EmptyInterfaceSlice(t *testing.T) {
	input := []interface{}{"not a map"}
	result, ok := toSliceOfMaps(input)
	if ok || len(result) != 0 {
		t.Errorf("expected (empty, false) when items aren't maps, got (%v, %v)", result, ok)
	}
}

func TestToSliceOfMaps_WrongType(t *testing.T) {
	_, ok := toSliceOfMaps("wrong type")
	if ok {
		t.Error("expected false for unsupported type")
	}
}

func TestIsResponseOne_Int(t *testing.T) {
	if !isResponseOne(1) {
		t.Error("expected true for int(1)")
	}
	if isResponseOne(2) {
		t.Error("expected false for int(2)")
	}
}

func TestIsResponseOne_Float64(t *testing.T) {
	if !isResponseOne(float64(1)) {
		t.Error("expected true for float64(1)")
	}
	if isResponseOne(float64(2)) {
		t.Error("expected false for float64(2)")
	}
}

func TestIsResponseOne_Int64(t *testing.T) {
	if !isResponseOne(int64(1)) {
		t.Error("expected true for int64(1)")
	}
}

func TestIsResponseOne_Nil(t *testing.T) {
	if isResponseOne(nil) {
		t.Error("expected false for nil")
	}
}

func TestFillBoolMap_NilAnswers(t *testing.T) {
	dst := map[string]bool{}
	fillBoolMap(dst, nil)
	if len(dst) != 0 {
		t.Error("expected empty map after fillBoolMap with nil")
	}
}

func TestFillBoolMap_ValidChoices(t *testing.T) {
	dst := map[string]bool{}
	answers := []interface{}{
		map[string]interface{}{"student_id": "B", "response": float64(1)},
		map[string]interface{}{"student_id": "C", "response": float64(0)},
	}
	fillBoolMap(dst, answers)
	if !dst["B"] {
		t.Error("expected B to be chosen")
	}
	if dst["C"] {
		t.Error("expected C not to be chosen (response=0)")
	}
}

func TestFillBoolMap_MapSlice(t *testing.T) {
	dst := map[string]bool{}
	answers := []map[string]interface{}{
		{"student_id": "X", "response": 1},
	}
	fillBoolMap(dst, answers)
	if !dst["X"] {
		t.Error("expected X to be set")
	}
}

func TestEnsureInnerMap_Creates(t *testing.T) {
	outer := map[string]map[string]bool{}
	ensureInnerMap(outer, "key1")
	if outer["key1"] == nil {
		t.Error("expected inner map to be created")
	}
}

func TestEnsureInnerMap_Idempotent(t *testing.T) {
	outer := map[string]map[string]bool{}
	outer["key1"] = map[string]bool{"x": true}
	ensureInnerMap(outer, "key1")
	if !outer["key1"]["x"] {
		t.Error("expected existing inner map not to be overwritten")
	}
}

func TestMapStringVal_Present(t *testing.T) {
	m := map[string]interface{}{"id": "abc"}
	if mapStringVal(m, "id") != "abc" {
		t.Error("expected 'abc'")
	}
}

func TestMapStringVal_Missing(t *testing.T) {
	m := map[string]interface{}{}
	if mapStringVal(m, "id") != "" {
		t.Error("expected empty string for missing key")
	}
}

// ---------------------------------------------------------------------------
// fillMatrixFromResponses
// ---------------------------------------------------------------------------

func TestFillMatrixFromResponses_Basic(t *testing.T) {
	matrix := &SociogramMatrix{
		Choices:             map[string]map[string]bool{},
		Rejections:          map[string]map[string]bool{},
		PerceivedChoices:    map[string]map[string]bool{},
		PerceivedRejections: map[string]map[string]bool{},
	}

	responses := []map[string]interface{}{
		{
			"student_id": "A",
			"question1":  []interface{}{map[string]interface{}{"student_id": "B", "response": float64(1)}},
			"question2":  []interface{}{},
			"question3":  []interface{}{},
			"question4":  []interface{}{},
		},
	}

	fillMatrixFromResponses(matrix, responses)

	if !matrix.Choices["A"]["B"] {
		t.Error("expected A → B choice to be set")
	}
}

func TestFillMatrixFromResponses_SkipEmptyStudentID(t *testing.T) {
	matrix := &SociogramMatrix{
		Choices:             map[string]map[string]bool{},
		Rejections:          map[string]map[string]bool{},
		PerceivedChoices:    map[string]map[string]bool{},
		PerceivedRejections: map[string]map[string]bool{},
	}

	responses := []map[string]interface{}{
		{"student_id": "", "question1": []interface{}{}},
	}

	fillMatrixFromResponses(matrix, responses)

	if len(matrix.Choices) != 0 {
		t.Error("expected no entries for empty student_id")
	}
}

// ---------------------------------------------------------------------------
// calculateRelationshipScore
// ---------------------------------------------------------------------------

func makeMatrix(choices, rejections, percChoices, percRejections map[string]map[string]bool) *SociogramMatrix {
	return &SociogramMatrix{
		Choices:             choices,
		Rejections:          rejections,
		PerceivedChoices:    percChoices,
		PerceivedRejections: percRejections,
	}
}

func TestCalculateRelationshipScore_MutualChoice(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram"})
	matrix := makeMatrix(
		map[string]map[string]bool{"A": {"B": true}, "B": {"A": true}},
		map[string]map[string]bool{},
		map[string]map[string]bool{},
		map[string]map[string]bool{},
	)
	score := p.calculateRelationshipScore(matrix, "A", "B")
	if score <= 0 {
		t.Errorf("mutual choice should give positive score, got %f", score)
	}
}

func TestCalculateRelationshipScore_MutualRejection(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram"})
	matrix := makeMatrix(
		map[string]map[string]bool{},
		map[string]map[string]bool{"A": {"B": true}, "B": {"A": true}},
		map[string]map[string]bool{},
		map[string]map[string]bool{},
	)
	score := p.calculateRelationshipScore(matrix, "A", "B")
	if score >= 0 {
		t.Errorf("mutual rejection should give negative score, got %f", score)
	}
}

func TestCalculateRelationshipScore_NoInteraction(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram"})
	matrix := makeMatrix(
		map[string]map[string]bool{},
		map[string]map[string]bool{},
		map[string]map[string]bool{},
		map[string]map[string]bool{},
	)
	score := p.calculateRelationshipScore(matrix, "A", "B")
	if score != 0.0 {
		t.Errorf("expected 0 for no interaction, got %f", score)
	}
}

func TestCalculateRelationshipScore_ConflictScenario(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram"})
	// A chooses B but B rejects A
	matrix := makeMatrix(
		map[string]map[string]bool{"A": {"B": true}},
		map[string]map[string]bool{"B": {"A": true}},
		map[string]map[string]bool{},
		map[string]map[string]bool{},
	)
	score := p.calculateRelationshipScore(matrix, "A", "B")
	// Conflict penalty reduces score
	if score >= 0.5 {
		t.Errorf("conflict scenario should have reduced score, got %f", score)
	}
}

// ---------------------------------------------------------------------------
// InvalidateCache
// ---------------------------------------------------------------------------

func TestSociogramAdapter_InvalidateCache(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram"})
	p.cache["sec1:survey1"] = &SociogramMatrix{}
	p.cache["sec2:survey1"] = &SociogramMatrix{}

	p.InvalidateCache("sec1")

	if _, exists := p.cache["sec1:survey1"]; exists {
		t.Error("expected cache entry for sec1 to be invalidated")
	}
	if _, exists := p.cache["sec2:survey1"]; !exists {
		t.Error("expected cache entry for sec2 to remain")
	}
}

// ---------------------------------------------------------------------------
// CalculatePairwiseAffinity with mock plugin
// ---------------------------------------------------------------------------

func TestSociogramAdapter_CalculatePairwiseAffinity_NoSectionID(t *testing.T) {
	a := NewSociogramAffinityProvider(&mockPlugin{name: "sociogram"})
	_, err := a.CalculatePairwiseAffinity(context.Background(), "A", "B")
	if err == nil {
		t.Error("expected error without section_id in context")
	}
}

func TestSociogramAdapter_CalculatePairwiseAffinity_NoSurveys(t *testing.T) {
	plugin := &mockPlugin{
		name: "sociogram",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	}
	a := NewSociogramAffinityProvider(plugin)
	score, err := a.CalculatePairwiseAffinity(sectionCtx("s1"), "A", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 0.0 {
		t.Errorf("expected 0 when no surveys, got %f", score)
	}
}

func TestSociogramAdapter_CalculatePairwiseAffinity_WithActiveSurvey(t *testing.T) {
	calls := 0
	plugin := &mockPlugin{
		name: "sociogram",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			req := data.(map[string]interface{})
			action, _ := req["action"].(string)
			calls++
			if action == "get_surveys" {
				return []map[string]interface{}{
					{"id": "survey-1", "is_active": true},
				}, nil
			}
			if action == "get_responses" {
				// A chose B, B chose A
				return []map[string]interface{}{
					{
						"student_id": "A",
						"question1":  []interface{}{map[string]interface{}{"student_id": "B", "response": float64(1)}},
						"question2":  []interface{}{},
						"question3":  []interface{}{},
						"question4":  []interface{}{},
					},
					{
						"student_id": "B",
						"question1":  []interface{}{map[string]interface{}{"student_id": "A", "response": float64(1)}},
						"question2":  []interface{}{},
						"question3":  []interface{}{},
						"question4":  []interface{}{},
					},
				}, nil
			}
			return nil, nil
		},
	}
	a := NewSociogramAffinityProvider(plugin)
	score, err := a.CalculatePairwiseAffinity(sectionCtx("s1"), "A", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score <= 0 {
		t.Errorf("expected positive affinity for mutual choice, got %f", score)
	}
	// Second call should use cache (no extra plugin calls)
	prevCalls := calls
	a.CalculatePairwiseAffinity(sectionCtx("s1"), "A", "B") //nolint:errcheck
	if calls != prevCalls {
		t.Error("expected second call to use cache (no extra plugin calls)")
	}
}

// ---------------------------------------------------------------------------
// selectSurveyID
// ---------------------------------------------------------------------------

func TestSelectSurveyID_PreferredMatch(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "s"})
	surveys := []map[string]interface{}{
		{"id": "s1", "is_active": true},
		{"id": "s2", "is_active": true},
	}
	id := p.selectSurveyID(surveys, "s2")
	if id != "s2" {
		t.Errorf("expected 's2', got '%s'", id)
	}
}

func TestSelectSurveyID_FallbackActive(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "s"})
	surveys := []map[string]interface{}{
		{"id": "s1", "is_active": false},
		{"id": "s2", "is_active": true},
	}
	id := p.selectSurveyID(surveys, "nonexistent")
	if id != "s2" {
		t.Errorf("expected 's2' (first active fallback), got '%s'", id)
	}
}

func TestSelectSurveyID_NoneActive(t *testing.T) {
	p := NewSociogramAffinityProvider(&mockPlugin{name: "s"})
	surveys := []map[string]interface{}{
		{"id": "s1", "is_active": false},
	}
	id := p.selectSurveyID(surveys, "")
	if id != "" {
		t.Errorf("expected empty string when no active surveys, got '%s'", id)
	}
}

// ---------------------------------------------------------------------------
// toInterfaceSlice
// ---------------------------------------------------------------------------

func TestToInterfaceSlice_Basic(t *testing.T) {
	src := []map[string]interface{}{{"k": "v"}}
	result := toInterfaceSlice(src)
	if len(result) != 1 {
		t.Errorf("expected len 1, got %d", len(result))
	}
}
