package handlers

// pe_extra_test.go — covers remaining branches in peer_evaluation_handler.go
// and plugin_handler_shared.go that weren't covered by handlers_test.go.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ms-core/internal/core"

	"github.com/gofiber/fiber/v2"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// errPlugin creates a testPlugin that always returns an error.
func errPlugin(name string) *testPlugin {
	return &testPlugin{
		name: name,
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return nil, errors.New("plugin execution failed")
		},
	}
}

// buildPEAppWithPlugin creates a Fiber app for PeerEvaluationHandler
// with a custom plugin.
func buildPEAppWithPlugin(plugin *testPlugin) *fiber.App {
	pm := buildPM(plugin)
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()

	app.Post("/surveys", h.CreatePeerEvaluationSurvey)
	app.Get("/surveys", h.GetPeerEvaluationSurveys)
	app.Get("/surveys/:id", h.GetPeerEvaluationSurvey)
	app.Post("/peers", h.GetSurveyPeers)
	app.Post("/surveys/:survey_id/responses", h.SubmitPeerEvaluationResponse)
	app.Get("/surveys/:survey_id/responses", h.GetPeerEvaluationResponses)
	app.Get("/surveys/:survey_id/results", h.GetPeerEvaluationResults)
	app.Get("/sections/:section_id/groups/:group_id/responses", h.GetSectionGroupResponses)
	app.Get("/sections/:section_id/responses", h.GetSectionResponses)
	app.Put("/surveys/:id", h.UpdatePeerEvaluationSurvey)
	app.Delete("/surveys/:id", h.DeletePeerEvaluationSurvey)
	app.Get("/sections/:section_id/responses-summary", h.GetSectionResponsesSummary)
	app.Post("/surveys/:survey_id/export", h.ExportPeerEvaluationToExcel)

	return app
}

// ── CreatePeerEvaluationSurvey ────────────────────────────────────────────────

func TestPE_CreateSurvey_BadBody(t *testing.T) {
	app := buildPEAppWithPlugin(echoPlugin("peer_evaluation"))
	// Send non-JSON body to trigger parseRequestBody error
	status, _ := doRequestRaw(t, app, "POST", "/surveys", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestPE_CreateSurvey_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "POST", "/surveys", map[string]interface{}{"title": "test"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetSurveyPeers ────────────────────────────────────────────────────────────

func TestPE_GetSurveyPeers_BadBody(t *testing.T) {
	app := buildPEAppWithPlugin(echoPlugin("peer_evaluation"))
	status, _ := doRequestRaw(t, app, "POST", "/peers", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestPE_GetSurveyPeers_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "POST", "/peers", map[string]interface{}{"survey_id": "s1"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetPeerEvaluationSurveys (missing section_id is already covered) ──────────

func TestPE_GetSurveys_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/surveys?section_id=s1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetPeerEvaluationSurvey ───────────────────────────────────────────────────

func TestPE_GetSurvey_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── SubmitPeerEvaluationResponse ──────────────────────────────────────────────

func TestPE_SubmitResponse_BadBody(t *testing.T) {
	app := buildPEAppWithPlugin(echoPlugin("peer_evaluation"))
	status, _ := doRequestRaw(t, app, "POST", "/surveys/sid-1/responses", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestPE_SubmitResponse_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "POST", "/surveys/sid-1/responses", map[string]interface{}{})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetPeerEvaluationResponses ────────────────────────────────────────────────

func TestPE_GetResponses_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetPeerEvaluationResults ──────────────────────────────────────────────────

func TestPE_GetResults_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/results", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetSectionGroupResponses ──────────────────────────────────────────────────

func TestPE_GetSectionGroupResponses_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/sections/s1/groups/g1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetSectionResponses ───────────────────────────────────────────────────────

func TestPE_GetSectionResponses_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/sections/s1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── UpdatePeerEvaluationSurvey ────────────────────────────────────────────────

func TestPE_UpdateSurvey_BadBody(t *testing.T) {
	app := buildPEAppWithPlugin(echoPlugin("peer_evaluation"))
	status, _ := doRequestRaw(t, app, "PUT", "/surveys/sid-1", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestPE_UpdateSurvey_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "PUT", "/surveys/sid-1", map[string]interface{}{"title": "Updated"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── DeletePeerEvaluationSurvey ────────────────────────────────────────────────

func TestPE_DeleteSurvey_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "DELETE", "/surveys/sid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── GetSectionResponsesSummary ────────────────────────────────────────────────

func TestPE_GetSectionResponsesSummary_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "GET", "/sections/s1/responses-summary", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error")
	}
}

// ── ExportPeerEvaluationToExcel ───────────────────────────────────────────────

func TestPE_ExportToExcel_Success(t *testing.T) {
	// Plugin returns a result with filename + []byte data
	plug := &testPlugin{
		name: "peer_evaluation",
		execFn: func(_ context.Context, _ interface{}) (interface{}, error) {
			return map[string]interface{}{
				"filename": "export.xlsx",
				"data":     []byte("fake-excel-bytes"),
			}, nil
		},
	}
	app := buildPEAppWithPlugin(plug)
	status, _ := doRequest(t, app, "POST", "/surveys/sid-1/export", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200 for valid Excel export, got %d", status)
	}
}

func TestPE_ExportToExcel_PluginError(t *testing.T) {
	app := buildPEAppWithPlugin(errPlugin("peer_evaluation"))
	status, _ := doRequest(t, app, "POST", "/surveys/sid-1/export", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error during export")
	}
}

// ── plugin_handler_shared.go ──────────────────────────────────────────────────

// TestParseRequestBody_BadContentType tests parseRequestBody with non-JSON input
func TestParseRequestBody_BadContentType(t *testing.T) {
	pm := core.NewPluginManager()
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		_, err := parseRequestBody(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"ok": true})
	})
	_ = h
	status, _ := doRequestRaw(t, app, "POST", "/test", "not-json-at-all", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

// TestExecutePluginData_PluginError tests executePluginData when plugin fails
func TestExecutePluginData_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("test-plugin"))
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return executePluginData(c, pm, "test-plugin", map[string]interface{}{"action": "test"})
	})
	req, _ := makeTestRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Error("expected non-200 for plugin error in executePluginData")
	}
}

// TestExecutePluginRaw_Success tests executePluginRaw returning a result
func TestExecutePluginRaw_Success(t *testing.T) {
	pm := buildPM(echoPlugin("test-plugin"))
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		result, err := executePluginRaw(c, pm, "test-plugin", map[string]interface{}{"action": "ping"})
		if err != nil {
			return err
		}
		return c.JSON(result)
	})
	req, _ := makeTestRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestExecutePluginRaw_PluginError tests executePluginRaw when plugin fails.
// pluginExecError writes the HTTP response directly and returns nil,
// so executePluginRaw returns (nil, nil) — the response is already written.
func TestExecutePluginRaw_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("test-plugin"))
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		result, err := executePluginRaw(c, pm, "test-plugin", map[string]interface{}{"action": "ping"})
		if err != nil {
			return err
		}
		// result is nil when plugin fails (response already written by pluginExecError)
		if result == nil {
			return nil
		}
		return c.JSON(result)
	})
	req, _ := makeTestRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	// pluginExecError already wrote a 500 (or 404/503) before returning nil
	if resp.StatusCode < 400 {
		t.Errorf("expected error status >=400 for plugin error in executePluginRaw, got %d", resp.StatusCode)
	}
}

// TestRequiredParam_Missing tests requiredParam with missing parameter
func TestRequiredParam_Missing(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		_, err := requiredParam(c, "missing_key", "missing_key is required")
		if err != nil {
			return err
		}
		return c.SendStatus(http.StatusOK)
	})
	req, _ := makeTestRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// TestRequiredQuery_Present tests requiredQuery with query param present
func TestRequiredQuery_Present(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		v, err := requiredQuery(c, "key", "key is required")
		if err != nil {
			return err
		}
		return c.SendString(v)
	})
	req, _ := makeTestRequest("GET", "/test?key=hello", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ── helpers for PE tests ──────────────────────────────────────────────────────

// doRequestRaw fires a Fiber request with a raw string body and custom content-type.
func doRequestRaw(t *testing.T, app *fiber.App, method, path, body, contentType string) (int, map[string]interface{}) {
	t.Helper()
	bodyBytes := []byte(body)
	req, err := makeTestRequestRaw(method, path, bodyBytes, contentType)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	bodyData, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyData, &result)
	return resp.StatusCode, result
}

func makeTestRequestRaw(method, path string, body []byte, contentType string) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func makeTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

var _ = fmt.Sprintf // silence unused import

// ── PE missing param paths (requiredParam error branch) ───────────────────────

// Each PE handler that uses requiredParam needs a route WITHOUT the param
// to trigger the error branch.

func TestPE_GetSurvey_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	// Register route without the :id param so requiredParam returns error
	app.Get("/surveys-noId", h.GetPeerEvaluationSurvey)
	status, _ := doRequest(t, app, "GET", "/surveys-noId", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing survey id, got %d", status)
	}
}

func TestPE_GetResponses_MissingSurveyID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Get("/surveys-noId/responses", h.GetPeerEvaluationResponses)
	status, _ := doRequest(t, app, "GET", "/surveys-noId/responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing survey_id, got %d", status)
	}
}

func TestPE_GetResults_MissingSurveyID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Get("/surveys-noId/results", h.GetPeerEvaluationResults)
	status, _ := doRequest(t, app, "GET", "/surveys-noId/results", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing survey_id, got %d", status)
	}
}

func TestPE_GetSectionGroupResponses_MissingSectionID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	// No :section_id param in route → requiredParam fails
	app.Get("/nosection/groups/:group_id/responses", h.GetSectionGroupResponses)
	status, _ := doRequest(t, app, "GET", "/nosection/groups/g1/responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing section_id, got %d", status)
	}
}

func TestPE_GetSectionGroupResponses_MissingGroupID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	// :section_id present but no :group_id param → second requiredParam fails
	app.Get("/sections/:section_id/nogroup/responses", h.GetSectionGroupResponses)
	status, _ := doRequest(t, app, "GET", "/sections/s1/nogroup/responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing group_id, got %d", status)
	}
}

func TestPE_GetSectionResponses_MissingSectionID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Get("/nosection/responses", h.GetSectionResponses)
	status, _ := doRequest(t, app, "GET", "/nosection/responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing section_id, got %d", status)
	}
}

func TestPE_UpdateSurvey_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Put("/surveys-noid", h.UpdatePeerEvaluationSurvey)
	status, _ := doRequest(t, app, "PUT", "/surveys-noid", map[string]interface{}{"t": "v"})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing survey id in update, got %d", status)
	}
}

func TestPE_DeleteSurvey_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Delete("/surveys-noid", h.DeletePeerEvaluationSurvey)
	status, _ := doRequest(t, app, "DELETE", "/surveys-noid", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing survey id in delete, got %d", status)
	}
}

func TestPE_GetSectionResponsesSummary_MissingSectionID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Get("/nosection/responses-summary", h.GetSectionResponsesSummary)
	status, _ := doRequest(t, app, "GET", "/nosection/responses-summary", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing section_id in GetSectionResponsesSummary, got %d", status)
	}
}

func TestPE_SubmitResponse_MissingSurveyID(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	app.Post("/responses", h.SubmitPeerEvaluationResponse)
	status, _ := doRequest(t, app, "POST", "/responses", map[string]interface{}{})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing survey_id in SubmitResponse, got %d", status)
	}
}

// TestPE_GetSurveys_MissingQuery covers requiredQuery error in GetPeerEvaluationSurveys
func TestPE_GetSurveys_MissingQuery(t *testing.T) {
	app := buildPEAppWithPlugin(echoPlugin("peer_evaluation"))
	// No section_id query param — requiredQuery should fail
	status, _ := doRequest(t, app, "GET", "/surveys", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing section_id query param, got %d", status)
	}
}

// ── requiredQuery missing branch ─────────────────────────────────────────────

func TestRequiredQuery_Missing(t *testing.T) {
	pm := buildPM(echoPlugin("peer_evaluation"))
	h := NewPeerEvaluationHandler(pm)
	app := fiber.New()
	// No :section_id in query
	app.Get("/surveys/no-query", h.GetPeerEvaluationSurveys)
	status, _ := doRequest(t, app, "GET", "/surveys/no-query", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}
