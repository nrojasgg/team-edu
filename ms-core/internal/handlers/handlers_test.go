package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ms-core/internal/core"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// testPlugin is a lightweight Plugin implementation for unit tests.
type testPlugin struct {
	name   string
	execFn func(ctx context.Context, data interface{}) (interface{}, error)
}

func (p *testPlugin) Name() string      { return p.name }
func (p *testPlugin) Version() string   { return "test" }
func (p *testPlugin) Initialize() error { return nil }
func (p *testPlugin) Shutdown() error   { return nil }
func (p *testPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	if p.execFn != nil {
		return p.execFn(ctx, data)
	}
	return map[string]interface{}{"result": "ok"}, nil
}

// echoPlugin returns whatever action it receives as the data field.
func echoPlugin(name string) *testPlugin {
	return &testPlugin{
		name: name,
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			return data, nil
		},
	}
}

// buildPM creates a PluginManager with the supplied plugin registered.
func buildPM(plugin *testPlugin) *core.PluginManager {
	pm := core.NewPluginManager()
	pm.RegisterPlugin(plugin) //nolint:errcheck
	return pm
}

// doRequest fires a Fiber request through app.Test and decodes the JSON body.
func doRequest(t *testing.T, app *fiber.App, method, path string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// non-JSON response (e.g. empty)
		return resp.StatusCode, nil
	}
	return resp.StatusCode, result
}

// ---------------------------------------------------------------------------
// plugin_errors tests
// ---------------------------------------------------------------------------

func TestPluginExecError_NotFound(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		pm := core.NewPluginManager() // empty – no plugins
		_, err := pm.ExecutePlugin("nonexistent", nil)
		if err != nil {
			return pluginExecError(c, err)
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestPluginExecError_Disabled(t *testing.T) {
	pm := buildPM(echoPlugin("dis"))
	pm.DisablePlugin("dis") //nolint:errcheck

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		_, err := pm.ExecutePlugin("dis", nil)
		if err != nil {
			return pluginExecError(c, err)
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// BFI44Handler tests
// ---------------------------------------------------------------------------

func setupBFI44App() (*fiber.App, *BFI44Handler) {
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()

	app.Post("/tests", h.CreateBFI44Test)
	app.Get("/sections/:section_id/tests", h.GetBFI44Tests)
	app.Get("/tests/:id", h.GetBFI44Test)
	app.Post("/tests/:test_id/responses", h.SubmitBFI44Response)
	app.Get("/tests/:test_id/responses", h.GetBFI44Responses)
	app.Get("/tests/:test_id/results", h.GetBFI44Results)
	app.Get("/section-responses", h.GetSectionBFI44Responses)
	app.Put("/tests/:id", h.UpdateBFI44Test)
	app.Delete("/tests/:id", h.DeleteBFI44Test)
	app.Get("/sections/:section_id/students/:student_id/bfi44", h.GetStudentBFI44)

	return app, h
}

func TestBFI44_CreateTest_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, body := doRequest(t, app, "POST", "/tests", map[string]interface{}{"title": "BFI Test"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body["success"])
	}
}

func TestBFI44_CreateTest_BadBody(t *testing.T) {
	app, _ := setupBFI44App()
	req := httptest.NewRequest("POST", "/tests", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestBFI44_GetTests_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, body := doRequest(t, app, "GET", "/sections/sec-1/tests", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}
}

func TestBFI44_GetTests_MissingSectionID(t *testing.T) {
	// Route doesn't have section_id param – test query param fallback: empty query
	app2 := fiber.New()
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app2.Get("/tests", h.GetBFI44Tests)
	status, _ := doRequest(t, app2, "GET", "/tests", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestBFI44_GetTest_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/tests/abc123", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_GetTest_MissingID(t *testing.T) {
	// Route with :test_id param but no value set → handled differently
	app2 := fiber.New()
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app2.Get("/tests/:test_id", h.GetBFI44Test)
	status, _ := doRequest(t, app2, "GET", "/tests/", nil)
	// Fiber returns 404 when :test_id is empty segment
	if status != http.StatusNotFound && status != http.StatusBadRequest {
		t.Errorf("expected 400 or 404, got %d", status)
	}
}

func TestBFI44_SubmitResponse_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "POST", "/tests/tid-1/responses", map[string]interface{}{"answers": []int{1, 2, 3}})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_SubmitResponse_MissingTestID(t *testing.T) {
	app2 := fiber.New()
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app2.Post("/responses", h.SubmitBFI44Response)
	status, _ := doRequest(t, app2, "POST", "/responses", map[string]interface{}{})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestBFI44_GetResponses_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/tests/tid-1/responses", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_GetResults_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/tests/tid-1/results", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_GetSectionResponses_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/section-responses?section_id=sec-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_GetSectionResponses_MissingSectionID(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/section-responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestBFI44_UpdateTest_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "PUT", "/tests/tid-1", map[string]interface{}{"title": "Updated"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_UpdateTest_MissingID(t *testing.T) {
	app2 := fiber.New()
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app2.Put("/tests", h.UpdateBFI44Test)
	status, _ := doRequest(t, app2, "PUT", "/tests", map[string]interface{}{})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestBFI44_DeleteTest_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "DELETE", "/tests/tid-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestBFI44_GetStudentBFI44_Success(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/sections/sec-1/students/stu-1/bfi44", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

// ---------------------------------------------------------------------------
// MBTIHandler tests
// ---------------------------------------------------------------------------

func setupMBTIApp() *fiber.App {
	pm := buildPM(echoPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()

	app.Post("/tests", h.CreateMBTITest)
	app.Get("/sections/:section_id/tests", h.GetMBTITests)
	app.Get("/tests/:id", h.GetMBTITest)
	app.Post("/tests/:test_id/responses", h.SubmitMBTIResponse)
	app.Get("/tests/:test_id/responses", h.GetMBTIResponses)
	app.Get("/tests/:test_id/results", h.GetMBTIResults)
	app.Get("/section-responses", h.GetSectionMBTIResponses)
	app.Put("/tests/:id", h.UpdateMBTITest)
	app.Delete("/tests/:id", h.DeleteMBTITest)
	app.Get("/sections/:section_id/students/:student_id", h.GetStudentMBTI)

	return app
}

func TestMBTI_CreateTest(t *testing.T) {
	app := setupMBTIApp()
	status, _ := doRequest(t, app, "POST", "/tests", map[string]interface{}{"title": "MBTI"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestMBTI_GetTests(t *testing.T) {
	app := setupMBTIApp()
	status, _ := doRequest(t, app, "GET", "/sections/s1/tests", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestMBTI_GetTest(t *testing.T) {
	app := setupMBTIApp()
	status, _ := doRequest(t, app, "GET", "/tests/tid1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestMBTI_SubmitResponse(t *testing.T) {
	app := setupMBTIApp()
	status, _ := doRequest(t, app, "POST", "/tests/tid1/responses", map[string]interface{}{})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestMBTI_GetStudentMBTI_MissingParams(t *testing.T) {
	pm := buildPM(echoPlugin("mbti"))
	h := NewMBTIHandler(pm)

	// Missing both params
	app2 := fiber.New()
	app2.Get("/nope", h.GetStudentMBTI)
	status, _ := doRequest(t, app2, "GET", "/nope", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

// ---------------------------------------------------------------------------
// EnneagramHandler tests
// ---------------------------------------------------------------------------

func setupEnneagramApp() *fiber.App {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()

	app.Post("/tests", h.CreateEnneagramTest)
	app.Get("/sections/:section_id/tests", h.GetEnneagramTests)
	app.Get("/tests/:id", h.GetEnneagramTest)
	app.Post("/tests/:test_id/responses", h.SubmitEnneagramResponse)
	app.Get("/tests/:test_id/responses", h.GetEnneagramResponses)
	app.Get("/tests/:test_id/results", h.GetEnneagramResults)
	app.Get("/tests/:test_id/section-results", h.GetEnneagramSectionResults)
	app.Get("/sections/:section_id/eneatypes", h.GetStudentEneatypes)
	app.Put("/tests/:id", h.UpdateEnneagramTest)
	app.Delete("/tests/:id", h.DeleteEnneagramTest)
	app.Post("/compatibility", h.CalculateGroupCompatibility)
	app.Get("/section-responses", h.GetSectionResponses)

	return app
}

func TestEnneagram_CreateTest(t *testing.T) {
	app := setupEnneagramApp()
	status, _ := doRequest(t, app, "POST", "/tests", map[string]interface{}{"title": "Ennea"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestEnneagram_GetSectionResults_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	app.Get("/results", h.GetEnneagramSectionResults)
	status, _ := doRequest(t, app, "GET", "/results", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestEnneagram_GetStudentEneatypes_MissingSectionID(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	app.Get("/eneatypes", h.GetStudentEneatypes)
	status, _ := doRequest(t, app, "GET", "/eneatypes", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestEnneagram_CalculateCompatibility_MissingEneatypes(t *testing.T) {
	app := setupEnneagramApp()
	status, _ := doRequest(t, app, "POST", "/compatibility", map[string]interface{}{"other": "data"})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestEnneagram_CalculateCompatibility_Success(t *testing.T) {
	app := setupEnneagramApp()
	status, _ := doRequest(t, app, "POST", "/compatibility", map[string]interface{}{"eneatypes": []int{1, 2, 3}})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

// ---------------------------------------------------------------------------
// GroupCompatibilityHandler tests
// ---------------------------------------------------------------------------

func setupGCApp() *fiber.App {
	pm := buildPM(echoPlugin("group_compatibility"))
	h := NewGroupCompatibilityHandler(pm)
	app := fiber.New()

	app.Post("/surveys", h.CreateSurvey)
	app.Get("/sections/:sectionId/surveys", h.GetSurveys)
	app.Get("/surveys/:id", h.GetSurvey)
	app.Post("/responses", h.SubmitResponse)
	app.Get("/surveys/:surveyId/responses", h.GetResponses)
	app.Get("/surveys/:surveyId/students/:studentId/response", h.GetStudentResponse)
	app.Put("/surveys/:id", h.UpdateSurvey)
	app.Delete("/surveys/:id", h.DeleteSurvey)
	app.Get("/sections/:sectionId/responses", h.GetSectionResponses)

	return app
}

func TestGC_CreateSurvey(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "POST", "/surveys", map[string]interface{}{"title": "GC Survey"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_GetSurveys(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "GET", "/sections/sec-1/surveys", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_GetSurvey(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_SubmitResponse(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "POST", "/responses", map[string]interface{}{"student_id": "s1"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_GetResponses(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/responses", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_UpdateSurvey(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "PUT", "/surveys/sid-1", map[string]interface{}{"title": "Updated"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_DeleteSurvey(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "DELETE", "/surveys/sid-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_GetSectionResponses(t *testing.T) {
	app := setupGCApp()
	status, _ := doRequest(t, app, "GET", "/sections/sec-1/responses", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_GetSectionResponses_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("group_compatibility"))
	h := NewGroupCompatibilityHandler(pm)
	app := fiber.New()
	app.Get("/sections-responses", h.GetSectionResponses)
	status, _ := doRequest(t, app, "GET", "/sections-responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

// ---------------------------------------------------------------------------
// PeerEvaluationHandler tests
// ---------------------------------------------------------------------------

func setupPEApp() *fiber.App {
	pm := buildPM(echoPlugin("peer_evaluation"))
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

	return app
}

func TestPE_CreateSurvey(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "POST", "/surveys", map[string]interface{}{"title": "PE Survey"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetSurveys_MissingSectionID(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/surveys", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 without section_id, got %d", status)
	}
}

func TestPE_GetSurveys_WithSectionID(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/surveys?section_id=sec-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetSurvey(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_SubmitResponse(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "POST", "/surveys/sid-1/responses", map[string]interface{}{})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetResponses(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/responses", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetResults(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/results", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetSectionGroupResponses(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/sections/s1/groups/g1/responses", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetSectionResponses(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/sections/s1/responses", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_UpdateSurvey(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "PUT", "/surveys/sid-1", map[string]interface{}{"title": "Updated"})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_DeleteSurvey(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "DELETE", "/surveys/sid-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestPE_GetSectionResponsesSummary(t *testing.T) {
	app := setupPEApp()
	status, _ := doRequest(t, app, "GET", "/sections/s1/responses-summary", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}
