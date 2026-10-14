package handlers

// psychometric_extra_test.go — covers remaining branches in psychometric_handler_shared.go
// and other psychometric handlers (Enneagram, BFI44, MBTI) that were missing error paths
// and the getResponses student_id query param branch.

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ── psychometricTestHandler shared function branches ──────────────────────────

// setupEnneagramAppForShared builds an Enneagram app with complete routes
// for testing psychometric_handler_shared functions.
func setupEnneagramAppForShared(plug *testPlugin) (*fiber.App, *EnneagramHandler) {
	pm := buildPM(plug)
	h := NewEnneagramHandler(pm)
	app := fiber.New()

	app.Post("/tests", h.CreateEnneagramTest)
	app.Get("/sections/:section_id/tests", h.GetEnneagramTests)
	app.Get("/tests/:id", h.GetEnneagramTest)
	app.Post("/tests/:test_id/responses", h.SubmitEnneagramResponse)
	app.Get("/tests/:test_id/responses", h.GetEnneagramResponses)
	app.Get("/tests/:test_id/results", h.GetEnneagramResults)
	app.Get("/section-responses", h.GetSectionResponses)
	app.Put("/tests/:id", h.UpdateEnneagramTest)
	app.Delete("/tests/:id", h.DeleteEnneagramTest)

	return app, h
}

// ── getResponses: student_id query param branch ───────────────────────────────

func TestEnneagram_GetResponses_WithStudentID(t *testing.T) {
	app, _ := setupEnneagramAppForShared(echoPlugin("enneagram"))
	// Send student_id query param to cover the optional branch in getResponses
	status, resp := doRequest(t, app, "GET", "/tests/tid-1/responses?student_id=stu-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	// Verify the echoPlugin received student_id in the request
	if resp != nil {
		if req, ok := resp["action"]; ok && req != "get_responses" {
			t.Errorf("expected action get_responses, got %v", req)
		}
	}
}

func TestBFI44_GetResponses_WithStudentID(t *testing.T) {
	app, _ := setupBFI44App()
	status, _ := doRequest(t, app, "GET", "/tests/tid-1/responses?student_id=stu-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestMBTI_GetResponses_WithStudentID(t *testing.T) {
	app := setupMBTIApp()
	status, _ := doRequest(t, app, "GET", "/tests/tid-1/responses?student_id=stu-1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

// ── getResults: plugin error path ────────────────────────────────────────────

func TestEnneagram_GetResults_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "GET", "/tests/tid-1/results", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in getResults")
	}
}

func TestBFI44_GetResults_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/tests/:test_id/results", h.GetBFI44Results)

	status, _ := doRequest(t, app, "GET", "/tests/tid-1/results", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetBFI44Results")
	}
}

func TestMBTI_GetResults_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/tests/:test_id/results", h.GetMBTIResults)

	status, _ := doRequest(t, app, "GET", "/tests/tid-1/results", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetMBTIResults")
	}
}

// ── deleteTest: plugin error path ────────────────────────────────────────────

func TestEnneagram_DeleteTest_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "DELETE", "/tests/tid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in deleteTest")
	}
}

func TestBFI44_DeleteTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Delete("/tests/:id", h.DeleteBFI44Test)

	status, _ := doRequest(t, app, "DELETE", "/tests/tid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in DeleteBFI44Test")
	}
}

func TestMBTI_DeleteTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Delete("/tests/:id", h.DeleteMBTITest)

	status, _ := doRequest(t, app, "DELETE", "/tests/tid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in DeleteMBTITest")
	}
}

// ── getResponses: plugin error path ──────────────────────────────────────────

func TestEnneagram_GetResponses_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "GET", "/tests/tid-1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in getResponses")
	}
}

func TestBFI44_GetResponses_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/tests/:test_id/responses", h.GetBFI44Responses)

	status, _ := doRequest(t, app, "GET", "/tests/tid-1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetBFI44Responses")
	}
}

func TestMBTI_GetResponses_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/tests/:test_id/responses", h.GetMBTIResponses)

	status, _ := doRequest(t, app, "GET", "/tests/tid-1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetMBTIResponses")
	}
}

// ── execute (shared): plugin error path ──────────────────────────────────────

func TestEnneagram_Execute_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	// createTest calls h.execute which calls pluginManager.ExecutePlugin
	status, _ := doRequest(t, app, "POST", "/tests", map[string]interface{}{"title": "T"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in execute(createTest)")
	}
}

func TestEnneagram_Update_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "PUT", "/tests/tid-1", map[string]interface{}{"title": "T"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in updateTest")
	}
}

// ── parseBody (shared): bad body path in psychometric ────────────────────────

func TestEnneagram_CreateTest_BadBody(t *testing.T) {
	app, _ := setupEnneagramAppForShared(echoPlugin("enneagram"))
	status, _ := doRequestRaw(t, app, "POST", "/tests", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestEnneagram_SubmitResponse_BadBody(t *testing.T) {
	app, _ := setupEnneagramAppForShared(echoPlugin("enneagram"))
	status, _ := doRequestRaw(t, app, "POST", "/tests/tid-1/responses", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body in submitResponse, got %d", status)
	}
}

func TestEnneagram_UpdateTest_BadBody(t *testing.T) {
	app, _ := setupEnneagramAppForShared(echoPlugin("enneagram"))
	status, _ := doRequestRaw(t, app, "PUT", "/tests/tid-1", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body in updateTest, got %d", status)
	}
}

// ── Enneagram-specific: getTests and getTest via section ──────────────────────

func TestEnneagram_GetTests_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "GET", "/sections/s1/tests", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in getTests")
	}
}

func TestEnneagram_GetTest_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "GET", "/tests/tid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in getTest")
	}
}

func TestEnneagram_GetSectionResponses_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	// GetSectionResponses requires section_id query param
	status, _ := doRequest(t, app, "GET", "/section-responses?section_id=s1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetSectionResponses")
	}
}

func TestEnneagram_SubmitResponse_PluginError(t *testing.T) {
	app, _ := setupEnneagramAppForShared(errPlugin("enneagram"))
	status, _ := doRequest(t, app, "POST", "/tests/tid-1/responses", map[string]interface{}{"answers": []int{1, 2}})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in submitResponse")
	}
}

// ── BFI44 plugin error paths (remaining) ─────────────────────────────────────

func TestBFI44_CreateTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Post("/tests", h.CreateBFI44Test)

	status, _ := doRequest(t, app, "POST", "/tests", map[string]interface{}{"title": "T"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in CreateBFI44Test")
	}
}

func TestBFI44_GetTests_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/sections/:section_id/tests", h.GetBFI44Tests)

	status, _ := doRequest(t, app, "GET", "/sections/s1/tests", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetBFI44Tests")
	}
}

func TestBFI44_GetTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/tests/:id", h.GetBFI44Test)

	status, _ := doRequest(t, app, "GET", "/tests/tid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetBFI44Test")
	}
}

func TestBFI44_SubmitResponse_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Post("/tests/:test_id/responses", h.SubmitBFI44Response)

	status, _ := doRequest(t, app, "POST", "/tests/tid-1/responses", map[string]interface{}{"answer": 1})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in SubmitBFI44Response")
	}
}

func TestBFI44_UpdateTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Put("/tests/:id", h.UpdateBFI44Test)

	status, _ := doRequest(t, app, "PUT", "/tests/tid-1", map[string]interface{}{"title": "U"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in UpdateBFI44Test")
	}
}

func TestBFI44_GetSectionResponses_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/section-responses", h.GetSectionBFI44Responses)

	status, _ := doRequest(t, app, "GET", "/section-responses?section_id=s1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetSectionBFI44Responses")
	}
}

// ── MBTI plugin error paths (remaining) ───────────────────────────────────────

func TestMBTI_CreateTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Post("/tests", h.CreateMBTITest)

	status, _ := doRequest(t, app, "POST", "/tests", map[string]interface{}{"title": "T"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in CreateMBTITest")
	}
}

func TestMBTI_GetTests_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/sections/:section_id/tests", h.GetMBTITests)

	status, _ := doRequest(t, app, "GET", "/sections/s1/tests", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetMBTITests")
	}
}

func TestMBTI_GetTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/tests/:id", h.GetMBTITest)

	status, _ := doRequest(t, app, "GET", "/tests/tid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetMBTITest")
	}
}

func TestMBTI_SubmitResponse_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Post("/tests/:test_id/responses", h.SubmitMBTIResponse)

	status, _ := doRequest(t, app, "POST", "/tests/tid-1/responses", map[string]interface{}{"answer": 1})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in SubmitMBTIResponse")
	}
}

func TestMBTI_UpdateTest_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Put("/tests/:id", h.UpdateMBTITest)

	status, _ := doRequest(t, app, "PUT", "/tests/tid-1", map[string]interface{}{"title": "U"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in UpdateMBTITest")
	}
}

func TestMBTI_GetSectionResponses_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/section-responses", h.GetSectionMBTIResponses)

	status, _ := doRequest(t, app, "GET", "/section-responses?section_id=s1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetSectionMBTIResponses")
	}
}

// ── group compatibility handler error paths ───────────────────────────────────

func setupGCAppWithPlugin(plug *testPlugin) *fiber.App {
	pm := buildPM(plug)
	h := NewGroupCompatibilityHandler(pm)
	app := fiber.New()

	app.Post("/surveys", h.CreateSurvey)
	app.Get("/sections/:sectionId/surveys", h.GetSurveys)
	app.Get("/surveys/:id", h.GetSurvey)
	app.Post("/responses", h.SubmitResponse)
	app.Get("/surveys/:surveyId/responses", h.GetResponses)
	app.Put("/surveys/:id", h.UpdateSurvey)
	app.Delete("/surveys/:id", h.DeleteSurvey)
	app.Get("/sections/:sectionId/responses", h.GetSectionResponses)

	return app
}

func TestGC_CreateSurvey_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "POST", "/surveys", map[string]interface{}{"title": "GC"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in CreateSurvey")
	}
}

func TestGC_GetSurveys_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "GET", "/sections/s1/surveys", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetSurveys")
	}
}

func TestGC_GetSurvey_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetSurvey")
	}
}

func TestGC_SubmitResponse_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "POST", "/responses", map[string]interface{}{})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in SubmitResponse")
	}
}

func TestGC_GetResponses_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetResponses")
	}
}

func TestGC_UpdateSurvey_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "PUT", "/surveys/sid-1", map[string]interface{}{"title": "U"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in UpdateSurvey")
	}
}

func TestGC_DeleteSurvey_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "DELETE", "/surveys/sid-1", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in DeleteSurvey")
	}
}

func TestGC_GetSectionResponses_PluginError(t *testing.T) {
	app := setupGCAppWithPlugin(errPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "GET", "/sections/s1/responses", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetSectionResponses")
	}
}

// TestGC_GetSectionResponses_MissingSectionID covers requiredParam error branch
func TestGC_GetSectionResponses_MissingSectionID(t *testing.T) {
	pm := buildPM(echoPlugin("group_compatibility"))
	h := NewGroupCompatibilityHandler(pm)
	app := fiber.New()
	// No :sectionId in the route, so requiredParam("sectionId") fails
	app.Get("/no-section/responses", h.GetSectionResponses)
	status, _ := doRequest(t, app, "GET", "/no-section/responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing sectionId, got %d", status)
	}
}

// TestGC_GetResponses_WithQueryParams covers the optional section_id and group_id branches
func TestGC_GetResponses_WithSectionAndGroupID(t *testing.T) {
	app := setupGCAppWithPlugin(echoPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/responses?section_id=s1&group_id=g1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestGC_GetResponses_WithSectionIDOnly(t *testing.T) {
	app := setupGCAppWithPlugin(echoPlugin("group_compatibility"))
	status, _ := doRequest(t, app, "GET", "/surveys/sid-1/responses?section_id=s1", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

// ── GetStudentBFI44 missing-params branch ─────────────────────────────────────

func TestBFI44_GetStudentBFI44_MissingParam(t *testing.T) {
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	// Route without :student_id param
	app.Get("/sections/:section_id/bfi44", h.GetStudentBFI44)
	status, _ := doRequest(t, app, "GET", "/sections/s1/bfi44", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 when student_id missing, got %d", status)
	}
}

func TestBFI44_GetStudentBFI44_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/sections/:section_id/students/:student_id/bfi44", h.GetStudentBFI44)
	status, _ := doRequest(t, app, "GET", "/sections/s1/students/stu-1/bfi44", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetStudentBFI44")
	}
}

// ── GetStudentMBTI missing-params branch ──────────────────────────────────────

func TestMBTI_GetStudentMBTI_MissingParam(t *testing.T) {
	pm := buildPM(echoPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	// Route without :student_id param
	app.Get("/sections/:section_id/mbti", h.GetStudentMBTI)
	status, _ := doRequest(t, app, "GET", "/sections/s1/mbti", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 when student_id missing, got %d", status)
	}
}

func TestMBTI_GetStudentMBTI_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/sections/:section_id/students/:student_id/mbti", h.GetStudentMBTI)
	status, _ := doRequest(t, app, "GET", "/sections/s1/students/stu-1/mbti", nil)
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in GetStudentMBTI")
	}
}

// ── Enneagram CalculateGroupCompatibility coverage ────────────────────────────

func TestEnneagram_CalculateGroupCompatibility_BadBody(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	app.Post("/calculate", h.CalculateGroupCompatibility)
	status, _ := doRequestRaw(t, app, "POST", "/calculate", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body in CalculateGroupCompatibility, got %d", status)
	}
}

func TestEnneagram_CalculateGroupCompatibility_PluginError(t *testing.T) {
	pm := buildPM(errPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	app.Post("/calculate", h.CalculateGroupCompatibility)
	status, _ := doRequest(t, app, "POST", "/calculate", map[string]interface{}{"section_id": "s1"})
	if status == http.StatusOK {
		t.Error("expected non-200 for plugin error in CalculateGroupCompatibility")
	}
}

func TestGC_CreateSurvey_BadBody(t *testing.T) {
	app := setupGCAppWithPlugin(echoPlugin("group_compatibility"))
	status, _ := doRequestRaw(t, app, "POST", "/surveys", "not-json", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestGC_SubmitResponse_BadBody(t *testing.T) {
	app := setupGCAppWithPlugin(echoPlugin("group_compatibility"))
	status, _ := doRequestRaw(t, app, "POST", "/responses", "bad", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

func TestGC_UpdateSurvey_BadBody(t *testing.T) {
	app := setupGCAppWithPlugin(echoPlugin("group_compatibility"))
	status, _ := doRequestRaw(t, app, "PUT", "/surveys/sid-1", "bad", "text/plain")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for bad body, got %d", status)
	}
}

// ── psychometricTestHandler getResults/deleteTest/getResponses missing test_id ─

func TestEnneagram_GetResults_MissingTestID(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	// Route without :test_id param → testID == "" → 400
	app.Get("/tests-noid/results", h.GetEnneagramResults)
	status, _ := doRequest(t, app, "GET", "/tests-noid/results", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing test_id in getResults, got %d", status)
	}
}

func TestEnneagram_DeleteTest_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	// Route without :id param → testID == "" → 400
	app.Delete("/tests-noid", h.DeleteEnneagramTest)
	status, _ := doRequest(t, app, "DELETE", "/tests-noid", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing id in deleteTest, got %d", status)
	}
}

func TestEnneagram_GetResponses_MissingTestID(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewEnneagramHandler(pm)
	app := fiber.New()
	// Route without :test_id param → testID == "" → 400
	app.Get("/tests-noid/responses", h.GetEnneagramResponses)
	status, _ := doRequest(t, app, "GET", "/tests-noid/responses", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing test_id in getResponses, got %d", status)
	}
}

func TestBFI44_GetResults_MissingTestID(t *testing.T) {
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Get("/tests-noid/results", h.GetBFI44Results)
	status, _ := doRequest(t, app, "GET", "/tests-noid/results", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing test_id in BFI44 getResults, got %d", status)
	}
}

func TestBFI44_DeleteTest_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("bfi44"))
	h := NewBFI44Handler(pm)
	app := fiber.New()
	app.Delete("/tests-noid", h.DeleteBFI44Test)
	status, _ := doRequest(t, app, "DELETE", "/tests-noid", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing id in BFI44 deleteTest, got %d", status)
	}
}

func TestMBTI_GetResults_MissingTestID(t *testing.T) {
	pm := buildPM(echoPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Get("/tests-noid/results", h.GetMBTIResults)
	status, _ := doRequest(t, app, "GET", "/tests-noid/results", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing test_id in MBTI getResults, got %d", status)
	}
}

func TestMBTI_DeleteTest_MissingID(t *testing.T) {
	pm := buildPM(echoPlugin("mbti"))
	h := NewMBTIHandler(pm)
	app := fiber.New()
	app.Delete("/tests-noid", h.DeleteMBTITest)
	status, _ := doRequest(t, app, "DELETE", "/tests-noid", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing id in MBTI deleteTest, got %d", status)
	}
}
