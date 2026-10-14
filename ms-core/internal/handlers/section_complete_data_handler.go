package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"ms-core/internal/config"
	"ms-core/internal/core"
	"ms-core/internal/database"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SectionCompleteDataHandler struct {
	db            *mongo.Database
	pluginManager *core.PluginManager
}

func NewSectionCompleteDataHandler(db *mongo.Database, pluginManager *core.PluginManager) *SectionCompleteDataHandler {
	return &SectionCompleteDataHandler{
		db:            db,
		pluginManager: pluginManager,
	}
}

// SectionGroupData returns only students and groups (lightweight)
type SectionGroupData struct {
	Students []map[string]interface{} `json:"students"`
	Groups   []map[string]interface{} `json:"groups"`
}

// SectionSurveyData returns only questionnaires (without responses)
type SectionSurveyData struct {
	EnneagramQuestionnaires []map[string]interface{} `json:"enneagram_questionnaires"`
	BFI44Questionnaires     []map[string]interface{} `json:"bfi44_questionnaires"`
	MBTIQuestionnaires      []map[string]interface{} `json:"mbti_questionnaires"`
	CompatibilitySurveys    []map[string]interface{} `json:"compatibility_surveys"`
	DemographicTests        []map[string]interface{} `json:"demographic_tests"`
}

// CompleteSectionGroupsData is the consolidated response
type CompleteSectionGroupsData struct {
	Students                []map[string]interface{} `json:"students"`
	EnneagramQuestionnaires []map[string]interface{} `json:"enneagram_questionnaires"`
	BFI44Questionnaires     []map[string]interface{} `json:"bfi44_questionnaires"`
	MBTIQuestionnaires      []map[string]interface{} `json:"mbti_questionnaires"`
	CompatibilitySurveys    []map[string]interface{} `json:"compatibility_surveys"`
	DemographicTests        []map[string]interface{} `json:"demographic_tests"`
	Groups                  []map[string]interface{} `json:"groups"`
	EnneagramResponses      []map[string]interface{} `json:"enneagram_responses"`
	BFI44Responses          []map[string]interface{} `json:"bfi44_responses"`
	MBTIResponses           []map[string]interface{} `json:"mbti_responses"`
	CompatibilityResponses  []map[string]interface{} `json:"compatibility_responses"`
	DemographicResponses    []map[string]interface{} `json:"demographic_responses"`
	LastGroupHistory        *GroupHistoryRecord      `json:"last_group_history,omitempty"`
	Success                 bool                     `json:"success"`
}

// GetCompleteSectionGroupsData consolidates all 6 API calls into a single response
// applyFetch logs warnMsg if err is non-nil; otherwise it stores result in *target.
func applyFetch(result []map[string]interface{}, err error, warnMsg string, target *[]map[string]interface{}) {
	if err != nil {
		fmt.Printf("%s: %v\n", warnMsg, err)
		return
	}
	*target = result
}

func (h *SectionCompleteDataHandler) GetCompleteSectionGroupsData(c *fiber.Ctx) error {
	sectionIDStr := c.Params("section_id")
	sectionID, err := strconv.ParseInt(sectionIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid section ID",
		})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authorization header required",
		})
	}

	cfg := config.LoadConfig()
	ctx := context.Background()

	response := CompleteSectionGroupsData{
		Students:                []map[string]interface{}{},
		EnneagramQuestionnaires: []map[string]interface{}{},
		BFI44Questionnaires:     []map[string]interface{}{},
		MBTIQuestionnaires:      []map[string]interface{}{},
		CompatibilitySurveys:    []map[string]interface{}{},
		DemographicTests:        []map[string]interface{}{},
		Groups:                  []map[string]interface{}{},
		EnneagramResponses:      []map[string]interface{}{},
		BFI44Responses:          []map[string]interface{}{},
		MBTIResponses:           []map[string]interface{}{},
		CompatibilityResponses:  []map[string]interface{}{},
		DemographicResponses:    []map[string]interface{}{},
		Success:                 true,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getStudentsFromCourses(ctx, sectionID, authHeader)
		applyFetch(r, e, "Warning: Failed to fetch students", &response.Students)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getEnneagramQuestionnairesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch enneagram questionnaires", &response.EnneagramQuestionnaires)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getCompatibilitySurveysFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch compatibility surveys", &response.CompatibilitySurveys)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getGroups(ctx, sectionID)
		applyFetch(r, e, "Warning: Failed to fetch groups", &response.Groups)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getEnneagramResponsesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch enneagram responses", &response.EnneagramResponses)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getCompatibilityResponsesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch compatibility responses", &response.CompatibilityResponses)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getBFI44QuestionnairesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch BFI-44 questionnaires", &response.BFI44Questionnaires)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getBFI44ResponsesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch BFI-44 responses", &response.BFI44Responses)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getMBTIQuestionnairesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch MBTI questionnaires", &response.MBTIQuestionnaires)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getMBTIResponsesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch MBTI responses", &response.MBTIResponses)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getDemographicTestsFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch demographic tests", &response.DemographicTests)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getDemographicResponsesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch demographic responses", &response.DemographicResponses)
	}()

	if h.db != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			history, err := h.getLastGroupHistory(ctx, sectionIDStr)
			if err != nil {
				fmt.Printf("Warning: Failed to fetch last group history: %v\n", err)
			} else {
				response.LastGroupHistory = history
			}
		}()
	}

	wg.Wait()

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetSectionGroupData returns only students and groups (lightweight, no responses)
func (h *SectionCompleteDataHandler) GetSectionGroupData(c *fiber.Ctx) error {
	sectionIDStr := c.Params("section_id")
	sectionID, err := strconv.ParseInt(sectionIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid section ID",
		})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authorization header required",
		})
	}

	ctx := context.Background()

	response := SectionGroupData{
		Students: []map[string]interface{}{},
		Groups:   []map[string]interface{}{},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getStudentsFromCourses(ctx, sectionID, authHeader)
		applyFetch(r, e, "Warning: Failed to fetch students", &response.Students)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getGroups(ctx, sectionID)
		applyFetch(r, e, "Warning: Failed to fetch groups", &response.Groups)
	}()

	wg.Wait()

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetSectionSurveyData returns only questionnaires (without responses)
func (h *SectionCompleteDataHandler) GetSectionSurveyData(c *fiber.Ctx) error {
	sectionIDStr := c.Params("section_id")

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authorization header required",
		})
	}

	cfg := config.LoadConfig()
	ctx := context.Background()

	response := SectionSurveyData{
		EnneagramQuestionnaires: []map[string]interface{}{},
		BFI44Questionnaires:     []map[string]interface{}{},
		MBTIQuestionnaires:      []map[string]interface{}{},
		CompatibilitySurveys:    []map[string]interface{}{},
		DemographicTests:        []map[string]interface{}{},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getEnneagramQuestionnairesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch enneagram questionnaires", &response.EnneagramQuestionnaires)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getBFI44QuestionnairesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch BFI-44 questionnaires", &response.BFI44Questionnaires)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getMBTIQuestionnairesFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch MBTI questionnaires", &response.MBTIQuestionnaires)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getCompatibilitySurveysFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch compatibility surveys", &response.CompatibilitySurveys)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r, e := h.getDemographicTestsFromAPI(ctx, sectionIDStr, authHeader, cfg)
		applyFetch(r, e, "Warning: Failed to fetch demographic tests", &response.DemographicTests)
	}()

	wg.Wait()

	return c.Status(fiber.StatusOK).JSON(response)
}

// getLastGroupHistory retrieves the most recent GroupHistoryRecord for a section.
func (h *SectionCompleteDataHandler) getLastGroupHistory(ctx context.Context, sectionID string) (*GroupHistoryRecord, error) {
	coll := h.db.Collection(database.GroupsHistoryCollection)
	opts := options.FindOne().SetSort(bson.D{{Key: "formed_at", Value: -1}})
	var record GroupHistoryRecord
	err := coll.FindOne(ctx, bson.M{"section_id": sectionID}, opts).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query last group history: %w", err)
	}
	return &record, nil
}

// getStudentsFromCourses calls the courses microservice to get students
func (h *SectionCompleteDataHandler) getStudentsFromCourses(ctx context.Context, sectionID int64, authHeader string) ([]map[string]interface{}, error) {
	cfg := config.LoadConfig()

	// Cambiado a APIGatewayURL porque ms-courses no tiene /students-complete directamente.
	// El gateway es quien consolida la información de estudiantes.
	studentsURL := fmt.Sprintf("%s/api/v1/sections/%d/students-complete", cfg.APIGatewayURL, sectionID)
	req, err := http.NewRequestWithContext(ctx, "GET", studentsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create students request: %v", err)
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("courses service (via gateway) returned status %d", resp.StatusCode)
	}

	var responseData struct {
		Students []map[string]interface{} `json:"students"`
		Success  bool                     `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode students response: %v", err)
	}

	if !responseData.Success || responseData.Students == nil {
		return []map[string]interface{}{}, nil
	}

	return responseData.Students, nil
}

// getEnneagramQuestionnairesFromAPI fetches questionnaires by calling the enneagram handler
func (h *SectionCompleteDataHandler) getEnneagramQuestionnairesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/enneagram/questionnaires?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getCompatibilitySurveysFromAPI fetches surveys by calling the group-compatibility handler
func (h *SectionCompleteDataHandler) getCompatibilitySurveysFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/group-compatibility/surveys/section/%s", cfg.Port, sectionID), authHeader, "data")
}

// getEnneagramResponsesFromAPI fetches responses by calling the enneagram handler
func (h *SectionCompleteDataHandler) getEnneagramResponsesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/enneagram/responses?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getCompatibilityResponsesFromAPI fetches responses by calling the group-compatibility handler
func (h *SectionCompleteDataHandler) getCompatibilityResponsesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/group-compatibility/responses/section/%s", cfg.Port, sectionID), authHeader, "data")
}

// getBFI44QuestionnairesFromAPI fetches BFI-44 questionnaires
func (h *SectionCompleteDataHandler) getBFI44QuestionnairesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/bfi44/questionnaires?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getBFI44ResponsesFromAPI fetches BFI-44 responses for a section
func (h *SectionCompleteDataHandler) getBFI44ResponsesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/bfi44/responses?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getMBTIQuestionnairesFromAPI fetches MBTI questionnaires
func (h *SectionCompleteDataHandler) getMBTIQuestionnairesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/mbti/questionnaires?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getMBTIResponsesFromAPI fetches MBTI responses for a section
func (h *SectionCompleteDataHandler) getMBTIResponsesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/mbti/responses?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getDemographicTestsFromAPI fetches demographic tests for a section
func (h *SectionCompleteDataHandler) getDemographicTestsFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/demographic/tests?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// getDemographicResponsesFromAPI fetches demographic responses for a section
func (h *SectionCompleteDataHandler) getDemographicResponsesFromAPI(ctx context.Context, sectionID string, authHeader string, cfg *config.Config) ([]map[string]interface{}, error) {
	return h.makeRequest(ctx, fmt.Sprintf("http://127.0.0.1:%s/api/v1/demographic/responses?section_id=%s", cfg.Port, sectionID), authHeader, "data")
}

// makeRequest is a helper function to make HTTP requests and extract data
func (h *SectionCompleteDataHandler) makeRequest(ctx context.Context, url string, authHeader string, dataKey string) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []map[string]interface{}{}, nil
	}

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return extractDataArray(responseData, dataKey), nil
}

// extractDataArray extracts a []map[string]interface{} from a response envelope,
// unwrapping up to two levels of {"success":true,"data":...} wrapping.
func extractDataArray(responseData map[string]interface{}, dataKey string) []map[string]interface{} {
	data, ok := responseData[dataKey]
	if !ok {
		return []map[string]interface{}{}
	}
	// Unwrap one extra level if the plugin double-wrapped the response.
	if innerMap, ok := data.(map[string]interface{}); ok {
		if inner, exists := innerMap["data"]; exists {
			data = inner
		}
	}
	dataArray, ok := data.([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}
	result := make([]map[string]interface{}, 0, len(dataArray))
	for _, item := range dataArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			result = append(result, itemMap)
		}
	}
	return result
}

// getGroups fetches groups from MongoDB
func (h *SectionCompleteDataHandler) getGroups(ctx context.Context, sectionID int64) ([]map[string]interface{}, error) {
	collection := h.db.Collection("groups")

	filter := bson.M{
		"section_id": sectionID,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %v", err)
	}
	defer cursor.Close(ctx)

	var groups []map[string]interface{}
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, fmt.Errorf("failed to decode groups: %v", err)
	}

	if groups == nil {
		groups = []map[string]interface{}{}
	}

	return groups, nil
}
