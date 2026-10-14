package adapters

// adapters_extra_test.go — covers branches in enneagram_adapter.go and sociogram_adapter.go
// that were at 0% or low coverage in the previous measurement.

import (
	"context"
	"errors"
	"testing"

	"ms-core/internal/core"
)

// ────────────────────────────────────────────────────────────────────────────
// studentIDToString
// ────────────────────────────────────────────────────────────────────────────

func TestStudentIDToString_String(t *testing.T) {
	if got := studentIDToString("abc"); got != "abc" {
		t.Errorf("expected 'abc', got %q", got)
	}
}

func TestStudentIDToString_Int(t *testing.T) {
	if got := studentIDToString(42); got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
}

func TestStudentIDToString_Float64(t *testing.T) {
	if got := studentIDToString(float64(7)); got != "7" {
		t.Errorf("expected '7', got %q", got)
	}
}

func TestStudentIDToString_Default(t *testing.T) {
	// An unsupported type should use fmt.Sprintf
	result := studentIDToString(true)
	if result == "" {
		t.Error("expected non-empty for boolean")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// getOCEANFromPlugin
// ────────────────────────────────────────────────────────────────────────────

func TestGetOCEANFromPlugin_Success(t *testing.T) {
	plug := &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return map[string]interface{}{
				"openness":          0.5,
				"conscientiousness": 0.6,
				"extraversion":      0.4,
				"agreeableness":     0.7,
				"neuroticism":       0.3,
			}, nil
		},
	}
	a := NewEnneagramAdapter(plug, "section-1")
	vec, err := a.getOCEANFromPlugin(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vec == nil {
		t.Fatal("expected non-nil OCEANVector")
	}
	if vec.Openness != 0.5 {
		t.Errorf("expected Openness=0.5, got %f", vec.Openness)
	}
}

func TestGetOCEANFromPlugin_PluginError(t *testing.T) {
	plug := &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, errors.New("plugin error")
		},
	}
	a := NewEnneagramAdapter(plug, "section-1")
	_, err := a.getOCEANFromPlugin(context.Background(), 3)
	if err == nil {
		t.Error("expected error from plugin failure")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// PopulateOCEANVector
// ────────────────────────────────────────────────────────────────────────────

// mockPluginForPopulate returns eneatype map from get_student_eneatypes and ocean from get_ocean_from_type
func mockPluginForPopulate(studentID string, eneatype int) *mockPlugin {
	return &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			req, ok := data.(map[string]interface{})
			if !ok {
				return nil, errors.New("bad request")
			}
			action, _ := req["action"].(string)
			switch action {
			case "get_student_eneatypes":
				return map[string]int{studentID: eneatype}, nil
			case "get_ocean_from_type":
				return map[string]interface{}{
					"openness":          0.5,
					"conscientiousness": 0.6,
					"extraversion":      0.4,
					"agreeableness":     0.7,
					"neuroticism":       0.3,
				}, nil
			}
			return nil, nil
		},
	}
}

func TestPopulateOCEANVector_Success(t *testing.T) {
	plug := mockPluginForPopulate("stu-1", 3)
	a := NewEnneagramAdapter(plug, "section-1")
	member := &core.TeamMember{ID: "stu-1"}
	if err := a.PopulateOCEANVector(context.Background(), member); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.Psychometrics == nil {
		t.Error("expected Psychometrics to be populated")
	}
	if member.RawType["enneagram"] != 3 {
		t.Errorf("expected RawType['enneagram']=3, got %v", member.RawType["enneagram"])
	}
}

func TestPopulateOCEANVector_MemberNotFound(t *testing.T) {
	plug := mockPluginForPopulate("other-stu", 3)
	a := NewEnneagramAdapter(plug, "section-1")
	member := &core.TeamMember{ID: "stu-1"}
	err := a.PopulateOCEANVector(context.Background(), member)
	// should return error: no enneagram data for student stu-1
	if err == nil {
		t.Error("expected error for unknown student")
	}
}

func TestPopulateOCEANVector_GetEneatypeFails(t *testing.T) {
	plug := &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, errors.New("plugin down")
		},
	}
	a := NewEnneagramAdapter(plug, "section-1")
	member := &core.TeamMember{ID: "stu-1"}
	err := a.PopulateOCEANVector(context.Background(), member)
	if err == nil {
		t.Error("expected error when plugin fails")
	}
}

func TestPopulateOCEANVector_ExistingRawType(t *testing.T) {
	// When member.RawType is already initialized, it should be extended (not overwritten)
	plug := mockPluginForPopulate("stu-1", 3)
	a := NewEnneagramAdapter(plug, "section-1")
	member := &core.TeamMember{
		ID:      "stu-1",
		RawType: map[string]interface{}{"existing_key": "existing_val"},
	}
	if err := a.PopulateOCEANVector(context.Background(), member); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.RawType["existing_key"] != "existing_val" {
		t.Error("existing RawType key should be preserved")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// PopulateSectionOCEANVectors
// ────────────────────────────────────────────────────────────────────────────

func TestPopulateSectionOCEANVectors_Success(t *testing.T) {
	plug := &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			req, ok := data.(map[string]interface{})
			if !ok {
				return nil, errors.New("bad request")
			}
			action, _ := req["action"].(string)
			switch action {
			case "get_student_eneatypes":
				return map[string]int{"stu-1": 1, "stu-2": 5}, nil
			case "get_ocean_from_type":
				return map[string]interface{}{
					"openness":          0.5,
					"conscientiousness": 0.6,
					"extraversion":      0.4,
					"agreeableness":     0.7,
					"neuroticism":       0.3,
				}, nil
			}
			return nil, nil
		},
	}
	a := NewEnneagramAdapter(plug, "section-1")
	members := []*core.TeamMember{
		{ID: "stu-1"},
		{ID: "stu-2"},
		{ID: "stu-unknown"}, // not in eneatypes map → should be skipped
	}
	if err := a.PopulateSectionOCEANVectors(context.Background(), members); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if members[0].Psychometrics == nil {
		t.Error("expected stu-1 to be populated")
	}
	if members[1].Psychometrics == nil {
		t.Error("expected stu-2 to be populated")
	}
	if members[2].Psychometrics != nil {
		t.Error("expected stu-unknown to be skipped (nil Psychometrics)")
	}
}

func TestPopulateSectionOCEANVectors_GetAllEneatypesFails(t *testing.T) {
	plug := &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, errors.New("plugin down")
		},
	}
	a := NewEnneagramAdapter(plug, "section-1")
	err := a.PopulateSectionOCEANVectors(context.Background(), []*core.TeamMember{{ID: "stu-1"}})
	if err == nil {
		t.Error("expected error when getAllEneatypes fails")
	}
}

func TestPopulateSectionOCEANVectors_OCEANPluginError_Skips(t *testing.T) {
	// When getOCEANFromPlugin fails, the member is skipped (not an error for the whole function)
	callCount := 0
	plug := &mockPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			callCount++
			req, _ := data.(map[string]interface{})
			action, _ := req["action"].(string)
			if action == "get_student_eneatypes" {
				return map[string]int{"stu-1": 1}, nil
			}
			// get_ocean_from_type fails
			return nil, errors.New("ocean plugin error")
		},
	}
	a := NewEnneagramAdapter(plug, "section-1")
	members := []*core.TeamMember{{ID: "stu-1"}}
	err := a.PopulateSectionOCEANVectors(context.Background(), members)
	// error should be nil (member is just skipped)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if members[0].Psychometrics != nil {
		t.Error("expected Psychometrics to be nil since ocean failed")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Sociogram adapter fillBoolMap and buildMatrix missing branches
// ────────────────────────────────────────────────────────────────────────────

func TestFillBoolMap_Nil(t *testing.T) {
	dst := make(map[string]bool)
	fillBoolMap(dst, nil) // nil input → no changes
	if len(dst) != 0 {
		t.Error("expected empty map for nil input")
	}
}

func TestFillBoolMap_UnsupportedType(t *testing.T) {
	dst := make(map[string]bool)
	fillBoolMap(dst, "not a slice")
	if len(dst) != 0 {
		t.Error("expected empty map for unsupported type")
	}
}

func TestFillBoolMap_SliceOfMaps(t *testing.T) {
	dst := make(map[string]bool)
	answers := []map[string]interface{}{
		{"student_id": "stu-1", "response": 1},
		{"student_id": "stu-2", "response": 0},
	}
	fillBoolMap(dst, answers)
	if !dst["stu-1"] {
		t.Error("expected stu-1 to be true")
	}
	if dst["stu-2"] {
		t.Error("expected stu-2 to be false/not set")
	}
}

func TestFillBoolMap_SliceOfInterfaces_Response1AsFloat(t *testing.T) {
	dst := make(map[string]bool)
	answers := []interface{}{
		map[string]interface{}{"student_id": "stu-3", "response": float64(1)},
		map[string]interface{}{"no_student_id": "x"},            // missing student_id → skip
		map[string]interface{}{"student_id": "", "response": 1}, // empty studentID → skip
	}
	fillBoolMap(dst, answers)
	if !dst["stu-3"] {
		t.Error("expected stu-3 to be true")
	}
}

func TestFillBoolMap_ItemNotMap(t *testing.T) {
	dst := make(map[string]bool)
	answers := []interface{}{"not-a-map", 42}
	fillBoolMap(dst, answers) // should not panic
	if len(dst) != 0 {
		t.Error("expected empty map")
	}
}

func TestBuildMatrix_GetSurveysFails(t *testing.T) {
	plug := &mockPlugin{
		name: "group_compatibility",
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, errors.New("plugin error")
		},
	}
	sp := NewSociogramAffinityProvider(plug)
	_, err := sp.buildMatrix(context.Background(), "section-1", "")
	if err == nil {
		t.Error("expected error when surveys call fails")
	}
}

func TestBuildMatrix_EmptySurveys(t *testing.T) {
	plug := &mockPlugin{
		name: "group_compatibility",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			return []map[string]interface{}{}, nil // empty surveys
		},
	}
	sp := NewSociogramAffinityProvider(plug)
	matrix, err := sp.buildMatrix(context.Background(), "section-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matrix != nil {
		t.Error("expected nil matrix for empty surveys")
	}
}

func TestBuildMatrix_GetResponsesFails(t *testing.T) {
	callCount := 0
	plug := &mockPlugin{
		name: "group_compatibility",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			callCount++
			req, _ := data.(map[string]interface{})
			action, _ := req["action"].(string)
			if action == "get_surveys" {
				return []map[string]interface{}{
					{"id": "survey-1", "is_active": true},
				}, nil
			}
			// get_responses fails
			return nil, errors.New("responses error")
		},
	}
	sp := NewSociogramAffinityProvider(plug)
	_, err := sp.buildMatrix(context.Background(), "section-1", "survey-1")
	if err == nil {
		t.Error("expected error when get_responses fails")
	}
}

func TestBuildMatrix_EmptyResponses(t *testing.T) {
	plug := &mockPlugin{
		name: "group_compatibility",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			req, _ := data.(map[string]interface{})
			action, _ := req["action"].(string)
			if action == "get_surveys" {
				return []map[string]interface{}{
					{"id": "survey-1", "is_active": true},
				}, nil
			}
			// get_responses returns empty
			return []map[string]interface{}{}, nil
		},
	}
	sp := NewSociogramAffinityProvider(plug)
	matrix, err := sp.buildMatrix(context.Background(), "section-1", "survey-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matrix != nil {
		t.Error("expected nil matrix for empty responses")
	}
}

func TestGetOrBuildMatrix_UsesCachedMatrix(t *testing.T) {
	callCount := 0
	plug := &mockPlugin{
		name: "group_compatibility",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			callCount++
			req, _ := data.(map[string]interface{})
			action, _ := req["action"].(string)
			if action == "get_surveys" {
				return []map[string]interface{}{
					{"id": "survey-1", "is_active": true},
				}, nil
			}
			return []map[string]interface{}{
				{"student_id": "stu-1", "question1": []interface{}{
					map[string]interface{}{"student_id": "stu-2", "response": 1},
				}},
			}, nil
		},
	}
	sp := NewSociogramAffinityProvider(plug)

	// First call builds matrix
	m1, err := sp.getOrBuildMatrix(context.Background(), "section-1", "survey-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	initialCalls := callCount

	// Second call should use cached matrix
	m2, err := sp.getOrBuildMatrix(context.Background(), "section-1", "survey-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m1 != m2 {
		t.Error("expected same matrix pointer from cache")
	}
	if callCount != initialCalls {
		t.Error("expected no new plugin calls when using cache")
	}
}
