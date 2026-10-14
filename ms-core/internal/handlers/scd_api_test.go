package handlers

// scd_api_test.go — tests for *FromAPI functions and getStudentsFromCourses
// by passing a config struct whose Port points to an httptest server.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ms-core/internal/config"
	"ms-core/internal/core"
)

// setupMockAPIServer creates an httptest server that returns a JSON
// envelope with a "data" array containing one item.
func setupMockAPIServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": []interface{}{
				map[string]interface{}{"id": "item1"},
			},
		})
	}))
}

// portFromServer extracts the port number from an httptest.Server.URL string.
// e.g. "http://127.0.0.1:54321" → "54321"
func portFromServer(server *httptest.Server) string {
	url := server.URL
	lastColon := strings.LastIndex(url, ":")
	if lastColon < 0 {
		return ""
	}
	return url[lastColon+1:]
}

// cfgWithPort builds a minimal config pointing Port at the given port string.
func cfgWithPort(port string) *config.Config {
	return &config.Config{
		Port:          port,
		JWTSecret:     "test-secret",
		APIGatewayURL: "http://localhost:3000",
	}
}

// newSCDHandlerForAPI creates a SectionCompleteDataHandler with a nil DB.
func newSCDHandlerForAPI() *SectionCompleteDataHandler {
	pm := core.NewPluginManager()
	return NewSectionCompleteDataHandler(nil, pm)
}

// ── getEnneagramQuestionnairesFromAPI ─────────────────────────────────────────

func TestGetEnneagramQuestionnairesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getEnneagramQuestionnairesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetCompatibilitySurveysFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getCompatibilitySurveysFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetEnneagramResponsesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getEnneagramResponsesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetCompatibilityResponsesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getCompatibilityResponsesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetBFI44QuestionnairesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getBFI44QuestionnairesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetBFI44ResponsesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getBFI44ResponsesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetMBTIQuestionnairesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getMBTIQuestionnairesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestGetMBTIResponsesFromAPI_Success(t *testing.T) {
	srv := setupMockAPIServer(t)
	defer srv.Close()

	h := newSCDHandlerForAPI()
	result, err := h.getMBTIResponsesFromAPI(context.Background(), "42", "Bearer token", cfgWithPort(portFromServer(srv)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

// ── getStudentsFromCourses ────────────────────────────────────────────────────

func TestGetStudentsFromCourses_WithMockGateway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"students": []interface{}{
				map[string]interface{}{"id": "s1", "name": "Alice"},
			},
		})
	}))
	defer srv.Close()

	t.Setenv("API_GATEWAY_URL", srv.URL)

	h := newSCDHandlerForAPI()
	result, err := h.getStudentsFromCourses(context.Background(), 42, "Bearer token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one student")
	}
}

func TestGetStudentsFromCourses_GatewayNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv("API_GATEWAY_URL", srv.URL)

	h := newSCDHandlerForAPI()
	_, err := h.getStudentsFromCourses(context.Background(), 42, "Bearer token")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestGetStudentsFromCourses_GatewayInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer srv.Close()

	t.Setenv("API_GATEWAY_URL", srv.URL)

	h := newSCDHandlerForAPI()
	_, err := h.getStudentsFromCourses(context.Background(), 42, "Bearer token")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetStudentsFromCourses_SuccessFalse(t *testing.T) {
	// Server returns success=false with nil students
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  false,
			"students": nil,
		})
	}))
	defer srv.Close()

	t.Setenv("API_GATEWAY_URL", srv.URL)

	h := newSCDHandlerForAPI()
	result, err := h.getStudentsFromCourses(context.Background(), 42, "Bearer token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// success=false returns empty slice, not an error
	if result == nil {
		t.Fatal("expected empty slice, not nil")
	}
}
