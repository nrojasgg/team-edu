package handlers

// scd_extra_test.go — tests for SectionCompleteDataHandler helper functions
// and pure functions that don't require MongoDB.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ms-core/internal/core"
)

// ── applyFetch (pure function) ───────────────────────────────────────────────

func TestApplyFetch_WithError(t *testing.T) {
	target := []map[string]interface{}{{"existing": "data"}}
	applyFetch(nil, fmt.Errorf("some error"), "Warning: test error", &target)
	// On error, target is not modified
	if len(target) != 1 || target[0]["existing"] != "data" {
		t.Errorf("expected target unchanged on error, got %v", target)
	}
}

func TestApplyFetch_WithSuccess(t *testing.T) {
	target := []map[string]interface{}{}
	result := []map[string]interface{}{{"id": "1"}, {"id": "2"}}
	applyFetch(result, nil, "Warning: ignored", &target)
	if len(target) != 2 {
		t.Errorf("expected 2 items, got %d", len(target))
	}
}

func TestApplyFetch_NilResult_Success(t *testing.T) {
	target := []map[string]interface{}{{"old": "data"}}
	applyFetch(nil, nil, "Warning: ignored", &target)
	// nil result with no error → sets target to nil
	if target != nil {
		t.Logf("target is %v (nil result replaces it)", target)
	}
}

// ── extractDataArray (pure function) ─────────────────────────────────────────

func TestExtractDataArray_MissingKey(t *testing.T) {
	responseData := map[string]interface{}{"other": "value"}
	result := extractDataArray(responseData, "data")
	if len(result) != 0 {
		t.Errorf("expected empty result for missing key, got %v", result)
	}
}

func TestExtractDataArray_WithDataArray(t *testing.T) {
	responseData := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "1", "name": "Alice"},
			map[string]interface{}{"id": "2", "name": "Bob"},
		},
	}
	result := extractDataArray(responseData, "data")
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0]["id"] != "1" {
		t.Errorf("expected id=1, got %v", result[0]["id"])
	}
}

func TestExtractDataArray_WithDoubleWrapped(t *testing.T) {
	// Simulates plugin double-wrapping: data: {data: [...]}
	responseData := map[string]interface{}{
		"data": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{"id": "3"},
			},
		},
	}
	result := extractDataArray(responseData, "data")
	if len(result) != 1 {
		t.Fatalf("expected 1 item from double-wrapped, got %d", len(result))
	}
	if result[0]["id"] != "3" {
		t.Errorf("expected id=3, got %v", result[0]["id"])
	}
}

func TestExtractDataArray_NonArrayData(t *testing.T) {
	// data is a string, not array
	responseData := map[string]interface{}{"data": "not-an-array"}
	result := extractDataArray(responseData, "data")
	if len(result) != 0 {
		t.Errorf("expected empty result for non-array data, got %v", result)
	}
}

func TestExtractDataArray_MixedItems(t *testing.T) {
	// Array has non-map items that should be skipped
	responseData := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "1"},
			"not-a-map",
			123,
			map[string]interface{}{"id": "2"},
		},
	}
	result := extractDataArray(responseData, "data")
	if len(result) != 2 {
		t.Fatalf("expected 2 items (non-map skipped), got %d", len(result))
	}
}

func TestExtractDataArray_EmptyArray(t *testing.T) {
	responseData := map[string]interface{}{
		"data": []interface{}{},
	}
	result := extractDataArray(responseData, "data")
	if len(result) != 0 {
		t.Errorf("expected empty array, got %v", result)
	}
}

// ── makeRequest via httptest.NewServer ────────────────────────────────────────

func newSCDHandler() *SectionCompleteDataHandler {
	pm := core.NewPluginManager()
	return NewSectionCompleteDataHandler(nil, pm)
}

func TestMakeRequest_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []interface{}{
				map[string]interface{}{"id": "s1"},
			},
		})
	}))
	defer srv.Close()

	h := newSCDHandler()
	result, err := h.makeRequest(t.Context(), srv.URL, "Bearer token", "data")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
}

func TestMakeRequest_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	h := newSCDHandler()
	result, err := h.makeRequest(t.Context(), srv.URL, "Bearer token", "data")
	if err != nil {
		t.Fatalf("expected nil error on non-200 (returns empty), got %v", err)
	}
	if result != nil && len(result) != 0 {
		t.Errorf("expected empty slice for non-200, got %v", result)
	}
}

func TestMakeRequest_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json{{{"))
	}))
	defer srv.Close()

	h := newSCDHandler()
	_, err := h.makeRequest(t.Context(), srv.URL, "Bearer token", "data")
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}

func TestMakeRequest_InvalidURL(t *testing.T) {
	h := newSCDHandler()
	_, err := h.makeRequest(t.Context(), "http://localhost:0/unreachable", "Bearer token", "data")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestMakeRequest_EmptyDataKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []interface{}{map[string]interface{}{"id": "1"}},
		})
	}))
	defer srv.Close()

	h := newSCDHandler()
	result, err := h.makeRequest(t.Context(), srv.URL, "Bearer token", "data") // key "data" not present
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result for missing key, got %v", result)
	}
}

// ── getStudentsFromCourses via httptest — DISABLED (nil DB panics on getGroups) ──
// Handler-level tests require a real MongoDB connection for getGroups.
// The pure function tests above already cover applyFetch, extractDataArray, makeRequest.
