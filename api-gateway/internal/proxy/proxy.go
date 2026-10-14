package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// String constants – avoids duplicated literals throughout the file
// ---------------------------------------------------------------------------

const (
	// Log format strings
	fmtProxyingRequest  = "Proxying request from %s to %s"
	fmtErrCreatingReq   = "Error creating request: %v"
	fmtErrReadingResp   = "Error reading response: %v"
	fmtErrRespFromCore  = "Error response from ms-core: %s"
	fmtErrFetchingUsers = "Error fetching users: %v"
	fmtErrFetchingResps = "Error fetching responses: %v"

	// Error messages returned to clients
	errFailedCreateProxyRequest = "Failed to create proxy request"
	errFailedConnectToCore      = "Failed to connect to ms-core"
	errFailedCreateRequest      = "Failed to create request"
	errFailedReadRespCore       = "Failed to read response from ms-core"
	errFailedFetchRespsCore     = "Failed to fetch responses from ms-core"
	errAuthHeaderRequired       = "Authorization header required"
	errAuthTokenRequired        = "Authorization token is required"
	errQuestionnaireIDRequired  = "questionnaire_id is required"
	errSectionIDRequired        = "section_id is required"
	errSectionIDQueryRequired   = "section_id query parameter is required"

	// HTTP headers / MIME types
	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	mimeApplicationJSON = "application/json"

	// URL path fragments & format strings
	apiV1Prefix      = "/api/v1"
	fmtUsersBatchURL = "%s/api/v1/users/batch"
	fmtStudentIDQS   = "%s?student_id=%s"
)

type ProxyService struct {
	RegistroCurricularURL string
	CoursesURL            string
	CoreURL               string
	NotificationsURL      string
	HTTPClient            *http.Client
}

func NewProxyService(registroCurricularURL, coursesURL string) *ProxyService {
	return &ProxyService{
		RegistroCurricularURL: registroCurricularURL,
		CoursesURL:            coursesURL,
		CoreURL:               "",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// NewProxyServiceWithCoreURL permite inicializar el proxy con CoreURL y NotificationsURL
func NewProxyServiceWithCoreURL(registroCurricularURL, coursesURL, coreURL string) *ProxyService {
	return &ProxyService{
		RegistroCurricularURL: registroCurricularURL,
		CoursesURL:            coursesURL,
		CoreURL:               coreURL,
		NotificationsURL:      "",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// NewProxyServiceComplete permite inicializar el proxy con todos los URLs
func NewProxyServiceComplete(registroCurricularURL, coursesURL, coreURL, notificationsURL string) *ProxyService {
	return &ProxyService{
		RegistroCurricularURL: registroCurricularURL,
		CoursesURL:            coursesURL,
		CoreURL:               coreURL,
		NotificationsURL:      notificationsURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Low-level request helpers
// ---------------------------------------------------------------------------

// makeAuthRequest creates and executes an authenticated HTTP request.
// It sets Content-Type: application/json and the provided token, then returns
// the full response body, HTTP status code, and any error.
func (p *ProxyService) makeAuthRequest(method, url string, body io.Reader, token string) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, err
	}
	if token != "" {
		req.Header.Set(headerAuthorization, token)
	}
	req.Header.Set(headerContentType, mimeApplicationJSON)
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return bodyBytes, resp.StatusCode, nil
}

// fetchResourceList performs a GET request and decodes the body as a JSON array.
func (p *ProxyService) fetchResourceList(url, authHeader string) ([]interface{}, error) {
	respBytes, status, err := p.makeAuthRequest("GET", url, nil, authHeader)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("request to %s returned status %d", url, status)
	}
	var result []interface{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("decoding response from %s: %w", url, err)
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// User-batch helpers
// ---------------------------------------------------------------------------

// fetchUsersBatch calls the registry-curricular /users/batch endpoint and
// returns a map of user-id-string → user data map.
func (p *ProxyService) fetchUsersBatch(ids []int64, token string) (map[string]map[string]interface{}, error) {
	requestBody := map[string]interface{}{"ids": ids}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling batch request: %w", err)
	}
	url := fmt.Sprintf(fmtUsersBatchURL, p.RegistroCurricularURL)
	respBytes, status, err := p.makeAuthRequest("POST", url, bytes.NewReader(bodyBytes), token)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("users/batch returned status %d", status)
	}
	var usersData struct {
		Users []map[string]interface{} `json:"users"`
	}
	if err := json.Unmarshal(respBytes, &usersData); err != nil {
		return nil, fmt.Errorf("decoding users/batch response: %w", err)
	}
	return buildUsersMap(usersData.Users), nil
}

// buildUsersMap converts a slice of user maps into a string-keyed map (by "id").
func buildUsersMap(users []map[string]interface{}) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{}, len(users))
	for _, user := range users {
		if id, ok := user["id"]; ok {
			var idStr string
			switch v := id.(type) {
			case string:
				idStr = v
			case float64:
				idStr = fmt.Sprintf("%.0f", v)
			case int:
				idStr = fmt.Sprintf("%d", v)
			}
			if idStr != "" {
				m[idStr] = user
			}
		}
	}
	return m
}

// extractStudentIDsInt extracts unique student_id values (as int64) from a
// slice of response maps.
func extractStudentIDsInt(responses []map[string]interface{}) []int64 {
	seen := make(map[string]bool)
	ids := make([]int64, 0)
	for _, r := range responses {
		var sid string
		switch v := r["student_id"].(type) {
		case string:
			sid = v
		case float64:
			sid = fmt.Sprintf("%.0f", v)
		case int, int64:
			sid = fmt.Sprintf("%d", v)
		}
		if sid == "" || seen[sid] {
			continue
		}
		seen[sid] = true
		if id, err := strconv.ParseInt(sid, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// ---------------------------------------------------------------------------
// Student / response enrichment helpers
// ---------------------------------------------------------------------------

// buildFullName assembles a full name from nombres, paterno, materno fields.
func buildFullName(user map[string]interface{}) string {
	get := func(key string) string {
		if v, ok := user[key]; ok && v != nil {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	parts := []string{get("nombres"), get("paterno"), get("materno")}
	var segments []string
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}
	return strings.TrimSpace(strings.Join(segments, " "))
}

// buildSimplifiedStudentResponses merges questionnaire responses with user data
// into the simplified format expected by the enneagram and BFI-44 handlers.
func buildSimplifiedStudentResponses(
	responses []map[string]interface{},
	usersMap map[string]map[string]interface{},
) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(responses))
	for _, response := range responses {
		var sid string
		switch v := response["student_id"].(type) {
		case string:
			sid = v
		case float64:
			sid = fmt.Sprintf("%.0f", v)
		case int, int64:
			sid = fmt.Sprintf("%d", v)
		}
		if sid == "" {
			continue
		}
		user, exists := usersMap[sid]
		if !exists {
			continue
		}

		simplified := map[string]interface{}{
			"student_id":  sid,
			"is_complete": false,
			"email":       "",
			"name":        buildFullName(user),
		}
		if val, ok := response["is_complete"].(bool); ok {
			simplified["is_complete"] = val
		}
		if submittedAt, ok := response["submitted_at"]; ok {
			simplified["submitted_at"] = submittedAt
		}
		if results, ok := response["results"].(map[string]interface{}); ok {
			simplified["results"] = results
		}
		if e, ok := user["email"]; ok && e != nil {
			simplified["email"] = fmt.Sprintf("%v", e)
		}
		out = append(out, simplified)
	}
	return out
}

// buildEnrichedResponses adds student_name and student_email fields to each
// response entry (used by the MBTI handler).
func buildEnrichedResponses(
	responses []map[string]interface{},
	usersMap map[string]map[string]interface{},
) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(responses))
	for _, response := range responses {
		var sid string
		switch v := response["student_id"].(type) {
		case string:
			sid = v
		case float64:
			sid = fmt.Sprintf("%.0f", v)
		case int, int64:
			sid = fmt.Sprintf("%d", v)
		}
		if sid == "" {
			continue
		}
		user, exists := usersMap[sid]
		if !exists {
			continue
		}

		enriched := make(map[string]interface{}, len(response))
		for k, v := range response {
			enriched[k] = v
		}
		enriched["student_name"] = fmt.Sprintf("%v %v %v",
			user["nombres"], user["paterno"], user["materno"])
		enriched["student_email"] = user["email"]
		out = append(out, enriched)
	}
	return out
}

// simplifyQuestionnaires trims a raw questionnaire list to the basic display fields.
func simplifyQuestionnaires(data []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(data))
	for _, q := range data {
		out = append(out, map[string]interface{}{
			"id":          q["id"],
			"title":       q["title"],
			"description": q["description"],
			"start_date":  q["start_date"],
			"end_date":    q["end_date"],
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// Courses / section helpers
// ---------------------------------------------------------------------------

// buildSemesterMap converts a slice of semester maps into a float64-keyed lookup.
func buildSemesterMap(semesters []interface{}) map[float64]map[string]interface{} {
	m := make(map[float64]map[string]interface{})
	for _, sem := range semesters {
		if semMap, ok := sem.(map[string]interface{}); ok {
			if id, ok := semMap["id"].(float64); ok {
				m[id] = semMap
			}
		}
	}
	return m
}

// buildCourseWithSections injects a "sections" array (with semester tags) into a
// course map, given the full lists of course-offerings, sections and semester data.
func buildCourseWithSections(
	courseMap map[string]interface{},
	courseOfferings []interface{},
	sections []interface{},
	semesterMap map[float64]map[string]interface{},
) map[string]interface{} {
	courseID, _ := courseMap["id"].(float64)
	courseSections := make([]map[string]interface{}, 0)
	for _, co := range courseOfferings {
		coMap, ok := co.(map[string]interface{})
		if !ok {
			continue
		}
		cID, _ := coMap["course_id"].(float64)
		if cID != courseID {
			continue
		}
		coID, _ := coMap["id"].(float64)
		semesterID, _ := coMap["semester_id"].(float64)
		for _, section := range sections {
			sectionMap, ok := section.(map[string]interface{})
			if !ok {
				continue
			}
			sectionCOID, ok := sectionMap["course_offering_id"].(float64)
			if !ok || sectionCOID != coID {
				continue
			}
			sectionWithSemester := make(map[string]interface{}, len(sectionMap)+2)
			for k, v := range sectionMap {
				sectionWithSemester[k] = v
			}
			if semester := semesterMap[semesterID]; semester != nil {
				sectionWithSemester["semester"] = fmt.Sprintf("%v_%v", semester["term"], semester["year"])
				sectionWithSemester["semester_id"] = semesterID
			} else {
				sectionWithSemester["semester"] = "N/A"
			}
			courseSections = append(courseSections, sectionWithSemester)
		}
	}
	result := make(map[string]interface{}, len(courseMap)+1)
	for k, v := range courseMap {
		result[k] = v
	}
	result["sections"] = courseSections
	return result
}

// fetchSectionDetails fetches course-offering, course, and semester data linked
// to a section's course_offering_id.
func (p *ProxyService) fetchSectionDetails(
	courseOfferingID interface{},
	authHeader string,
) (courseOffering, course, semester map[string]interface{}) {
	coURL := fmt.Sprintf("%s%s/course-offerings/%v", p.CoursesURL, apiV1Prefix, courseOfferingID)
	coBytes, coStatus, coErr := p.makeAuthRequest("GET", coURL, nil, authHeader)
	if coErr != nil || coStatus != http.StatusOK {
		return nil, nil, nil
	}
	if err := json.Unmarshal(coBytes, &courseOffering); err != nil {
		log.Printf("Error decoding course offering: %v", err)
		return nil, nil, nil
	}
	if courseID, ok := courseOffering["course_id"]; ok && courseID != nil {
		courseURL := fmt.Sprintf("%s%s/courses/%v", p.CoursesURL, apiV1Prefix, courseID)
		if cBytes, cStatus, cErr := p.makeAuthRequest("GET", courseURL, nil, authHeader); cErr == nil && cStatus == http.StatusOK {
			if err := json.Unmarshal(cBytes, &course); err != nil {
				log.Printf("Error decoding course: %v", err)
			}
		}
	}
	if semesterID, ok := courseOffering["semester_id"]; ok && semesterID != nil {
		semURL := fmt.Sprintf("%s%s/semesters/%v", p.CoursesURL, apiV1Prefix, semesterID)
		if sBytes, sStatus, sErr := p.makeAuthRequest("GET", semURL, nil, authHeader); sErr == nil && sStatus == http.StatusOK {
			if err := json.Unmarshal(sBytes, &semester); err != nil {
				log.Printf("Error decoding semester: %v", err)
			}
		}
	}
	return courseOffering, course, semester
}

// fetchQuestionnaireResponses retrieves questionnaire response data from ms-core.
// Returns the data slice, the HTTP status (non-zero on server errors), and any error.
func (p *ProxyService) fetchQuestionnaireResponses(responsesURL, token string) ([]map[string]interface{}, int, error) {
	respBytes, status, err := p.makeAuthRequest("GET", responsesURL, nil, token)
	if err != nil {
		return nil, 0, err
	}
	if status != http.StatusOK {
		return nil, status, fmt.Errorf("ms-core returned status %d: %s", status, string(respBytes))
	}
	var responsesData struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &responsesData); err != nil {
		return nil, status, err
	}
	return responsesData.Data, status, nil
}

// fetchQuestionnairesFromCore retrieves and simplifies a questionnaire list from ms-core.
func (p *ProxyService) fetchQuestionnairesFromCore(questionnairesURL, token string, c *fiber.Ctx) error {
	respBytes, status, err := p.makeAuthRequest("GET", questionnairesURL, nil, token)
	if err != nil {
		log.Printf(fmtErrFetchingResps, err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to fetch questionnaires from ms-core"})
	}
	if status != http.StatusOK {
		log.Printf(fmtErrRespFromCore, string(respBytes))
		return c.Status(status).Send(respBytes)
	}
	var questionnairesData struct {
		Success bool                     `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBytes, &questionnairesData); err != nil {
		log.Printf("Error decoding questionnaires: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode questionnaires"})
	}
	return c.JSON(fiber.Map{"success": true, "data": simplifyQuestionnaires(questionnairesData.Data)})
}

// handleQuestionnaireResponses is the shared implementation for fetching and
// enriching questionnaire responses (enneagram and BFI-44 paths).
func (p *ProxyService) handleQuestionnaireResponses(c *fiber.Ctx, responsesURL, token string) error {
	responses, status, err := p.fetchQuestionnaireResponses(responsesURL, token)
	if err != nil {
		if status == 0 {
			log.Printf(fmtErrFetchingResps, err)
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": errFailedFetchRespsCore})
		}
		rawBytes, _, _ := p.makeAuthRequest("GET", responsesURL, nil, token)
		log.Printf(fmtErrRespFromCore, string(rawBytes))
		return c.Status(status).Send(rawBytes)
	}
	if len(responses) == 0 {
		return c.JSON(fiber.Map{"success": true, "data": []map[string]interface{}{}})
	}
	studentIDsInt := extractStudentIDsInt(responses)
	if len(studentIDsInt) == 0 {
		return c.JSON(fiber.Map{"success": true, "data": responses})
	}
	log.Printf("Fetching users for %d student IDs", len(studentIDsInt))
	usersMap, err := p.fetchUsersBatch(studentIDsInt, token)
	if err != nil {
		log.Printf(fmtErrFetchingUsers, err)
		return c.JSON(fiber.Map{"success": true, "data": responses})
	}
	return c.JSON(fiber.Map{"success": true, "data": buildSimplifiedStudentResponses(responses, usersMap)})
}

// proxyToCoreSection proxies a GET request to ms-core and streams the response.
func (p *ProxyService) proxyToCoreSection(c *fiber.Ctx, targetURL, token, logLabel string) error {
	log.Printf("Fetching %s from: %s", logLabel, targetURL)
	body, status, err := p.makeAuthRequest("GET", targetURL, nil, token)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateRequest})
	}
	if status != http.StatusOK {
		log.Printf(fmtErrRespFromCore, string(body))
		return c.Status(status).Send(body)
	}
	return c.Status(status).Send(body)
}

// ---------------------------------------------------------------------------
// Proxy handlers
// ---------------------------------------------------------------------------

// ProxyToCore redirige las solicitudes al microservicio ms-core
func (p *ProxyService) ProxyToCore(c *fiber.Ctx) error {
	originalPath := c.OriginalURL()
	targetPath := strings.TrimPrefix(originalPath, apiV1Prefix+"/core")
	if targetPath == "" {
		targetPath = "/"
	}
	// Mantener el prefijo /api/v1 para ms-core
	targetURL := fmt.Sprintf("%s%s%s", p.CoreURL, apiV1Prefix, targetPath)
	log.Printf(fmtProxyingRequest, originalPath, targetURL)
	var bodyReader io.Reader
	if c.Body() != nil {
		bodyReader = bytes.NewReader(c.Body())
	}
	req, err := http.NewRequest(c.Method(), targetURL, bodyReader)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": errFailedCreateProxyRequest,
		})
	}
	for key, values := range c.GetReqHeaders() {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error making request to ms-core: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": errFailedConnectToCore,
		})
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(fmtErrReadingResp, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": errFailedReadRespCore,
		})
	}
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}
	return c.Status(resp.StatusCode).Send(body)
}

// ProxyToRegistroCurricular redirige las solicitudes al microservicio de registro curricular
func (p *ProxyService) ProxyToRegistroCurricular(c *fiber.Ctx) error {
	originalPath := c.OriginalURL()
	targetPath := strings.TrimPrefix(originalPath, apiV1Prefix+"/registro")
	if targetPath == "" {
		targetPath = "/"
	}
	targetURL := fmt.Sprintf("%s%s", p.RegistroCurricularURL, targetPath)
	log.Printf(fmtProxyingRequest, originalPath, targetURL)

	var bodyReader io.Reader
	if c.Body() != nil {
		bodyReader = bytes.NewReader(c.Body())
	}
	req, err := http.NewRequest(c.Method(), targetURL, bodyReader)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": errFailedCreateProxyRequest,
		})
	}
	for key, values := range c.GetReqHeaders() {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error making request to microservice: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to connect to microservice",
		})
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(fmtErrReadingResp, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read microservice response",
		})
	}
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}
	return c.Status(resp.StatusCode).Send(body)
}

// ProxyToCourses redirige las solicitudes al microservicio de cursos
func (p *ProxyService) ProxyToCourses(c *fiber.Ctx) error {
	originalPath := c.OriginalURL()
	var targetPath string
	if strings.HasPrefix(originalPath, apiV1Prefix+"/courses") ||
		strings.HasPrefix(originalPath, apiV1Prefix+"/semesters") ||
		strings.HasPrefix(originalPath, apiV1Prefix+"/course-offerings") ||
		strings.HasPrefix(originalPath, apiV1Prefix+"/sections") ||
		strings.HasPrefix(originalPath, apiV1Prefix+"/student-sections") ||
		strings.HasPrefix(originalPath, apiV1Prefix+"/students") ||
		strings.HasPrefix(originalPath, apiV1Prefix+"/invitation-codes") {
		targetPath = strings.TrimPrefix(originalPath, apiV1Prefix)
	} else {
		targetPath = originalPath
	}
	if targetPath == "" {
		targetPath = "/"
	}
	targetURL := fmt.Sprintf("%s%s%s", p.CoursesURL, apiV1Prefix, targetPath)
	log.Printf(fmtProxyingRequest, originalPath, targetURL)

	var bodyReader io.Reader
	if c.Body() != nil {
		bodyReader = bytes.NewReader(c.Body())
	}
	req, err := http.NewRequest(c.Method(), targetURL, bodyReader)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": errFailedCreateRequest,
		})
	}
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error making request to courses microservice: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to connect to courses microservice",
		})
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(fmtErrReadingResp, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read courses microservice response",
		})
	}
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}
	return c.Status(resp.StatusCode).Send(body)
}

// fetchStudentDetails es un helper privado que obtiene información completa de estudiantes.
// Combina datos de ms-auth (información personal) y ms-core (eneatypes).
func (p *ProxyService) fetchStudentDetails(studentIDs []int64, sectionID string, token string) ([]map[string]interface{}, error) {
	batchBody := map[string]interface{}{"ids": studentIDs}
	batchBodyBytes, err := encodeJSON(batchBody)
	if err != nil {
		return nil, fmt.Errorf("error encoding batch request: %w", err)
	}

	batchURL := fmt.Sprintf(fmtUsersBatchURL, p.RegistroCurricularURL)
	batchRespBytes, batchStatus, err := p.makeAuthRequest("POST", batchURL, bytes.NewReader(batchBodyBytes), token)
	if err != nil {
		return nil, fmt.Errorf("error fetching batch users: %w", err)
	}
	if batchStatus != http.StatusOK {
		return nil, fmt.Errorf("batch endpoint returned status %d: %s", batchStatus, string(batchRespBytes))
	}

	var batchData map[string]interface{}
	if err := json.Unmarshal(batchRespBytes, &batchData); err != nil {
		return nil, fmt.Errorf("error decoding batch response: %w", err)
	}

	var users []interface{}
	if usersArray, ok := batchData["users"].([]interface{}); ok {
		users = usersArray
	} else if usersList, ok := batchData["data"].([]interface{}); ok {
		users = usersList
	} else {
		users = []interface{}{}
	}

	// Obtener eneatypes de estudiantes (si hay sectionID)
	eneatypesMap := make(map[string]interface{})
	if sectionID != "" && p.CoreURL != "" {
		eneatypesURL := fmt.Sprintf("%s%s/enneagram/sections/%s/student-eneatypes", p.CoreURL, apiV1Prefix, sectionID)
		if enBytes, enStatus, enErr := p.makeAuthRequest("GET", eneatypesURL, nil, token); enErr == nil && enStatus == http.StatusOK {
			var eneatypesData map[string]interface{}
			if json.Unmarshal(enBytes, &eneatypesData) == nil {
				if eneatypes, ok := eneatypesData["data"].(map[string]interface{}); ok {
					eneatypesMap = eneatypes
				}
			}
		}
	}

	transformedStudents := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userMap, ok := user.(map[string]interface{})
		if !ok {
			continue
		}
		student := map[string]interface{}{
			"id":          userMap["id"],
			"nombres":     userMap["nombres"],
			"paterno":     userMap["paterno"],
			"materno":     userMap["materno"],
			"email":       userMap["email"],
			"rut":         userMap["rut"],
			"career_code": userMap["career_code"],
			"picture_url": userMap["picture_url"],
			"sexo":        userMap["sexo"],
			"edad":        userMap["edad"],
			"nationality": userMap["nationality"],
		}
		if userID, ok := userMap["id"].(float64); ok {
			userIDStr := fmt.Sprintf("%.0f", userID)
			if eneatypeData, exists := eneatypesMap[userIDStr]; exists {
				// eneatypeData is the dominant_type int directly (float64 after JSON decode)
				student["eneatype"] = eneatypeData
			}
		}
		transformedStudents = append(transformedStudents, student)
	}
	return transformedStudents, nil
}

// GetSectionStudentsComplete combina tres llamadas en una sola respuesta:
// GET /sections/{id}/students → IDs, POST /users/batch → datos, GET /core/enneagram/... → eneatypes
func (p *ProxyService) GetSectionStudentsComplete(c *fiber.Ctx) error {
	sectionID := c.Params("id")
	authHeader := c.Get(headerAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthHeaderRequired})
	}

	studentsURL := fmt.Sprintf("%s%s/sections/%s/students", p.CoursesURL, apiV1Prefix, sectionID)
	studentsReq, err := http.NewRequest("GET", studentsURL, nil)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateRequest})
	}
	studentsReq.Header.Set(headerAuthorization, authHeader)
	studentsReq.Header.Set(headerContentType, mimeApplicationJSON)

	studentsResp, err := p.HTTPClient.Do(studentsReq)
	if err != nil {
		log.Printf("Error fetching students: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to fetch students"})
	}
	defer studentsResp.Body.Close()

	if studentsResp.StatusCode == http.StatusNotFound {
		return c.JSON(fiber.Map{"success": true, "students": []interface{}{}})
	}
	if studentsResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(studentsResp.Body)
		log.Printf("Students endpoint returned status %d: %s", studentsResp.StatusCode, string(body))
		return c.Status(studentsResp.StatusCode).JSON(fiber.Map{"error": "Failed to fetch students from courses service"})
	}

	var studentsData map[string]interface{}
	if err := decodeJSON(studentsResp.Body, &studentsData); err != nil {
		log.Printf("Error decoding students response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode students response"})
	}

	studentIDs, ok := studentsData["ids"].([]interface{})
	if !ok || len(studentIDs) == 0 {
		return c.JSON(fiber.Map{"success": true, "students": []interface{}{}})
	}

	studentIDsInt := make([]int64, 0, len(studentIDs))
	for _, id := range studentIDs {
		switch v := id.(type) {
		case float64:
			studentIDsInt = append(studentIDsInt, int64(v))
		case int64:
			studentIDsInt = append(studentIDsInt, v)
		case int:
			studentIDsInt = append(studentIDsInt, int64(v))
		}
	}

	transformedStudents, err := p.fetchStudentDetails(studentIDsInt, sectionID, authHeader)
	if err != nil {
		log.Printf("Error fetching student details: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch student details"})
	}
	return c.JSON(fiber.Map{"success": true, "students": transformedStudents})
}

// GetCoursesComplete combina múltiples llamadas para la página de cursos:
// GET /courses, GET /semesters/current, GET /semesters, GET /course-offerings, GET /sections
func (p *ProxyService) GetCoursesComplete(c *fiber.Ctx) error {
	authHeader := c.Get(headerAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthHeaderRequired})
	}

	courses, err := p.fetchResourceList(fmt.Sprintf("%s%s/courses", p.CoursesURL, apiV1Prefix), authHeader)
	if err != nil {
		log.Printf("Error fetching courses: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to fetch courses"})
	}

	var currentSemester map[string]interface{}
	if csBytes, csStatus, csErr := p.makeAuthRequest("GET",
		fmt.Sprintf("%s%s/semesters/current", p.CoursesURL, apiV1Prefix), nil, authHeader,
	); csErr == nil && csStatus == http.StatusOK {
		var semesterResponse map[string]interface{}
		if json.Unmarshal(csBytes, &semesterResponse) == nil {
			if data, ok := semesterResponse["data"].(map[string]interface{}); ok {
				currentSemester = data
			}
		}
	}

	semesters, _ := p.fetchResourceList(fmt.Sprintf("%s%s/semesters", p.CoursesURL, apiV1Prefix), authHeader)
	if semesters == nil {
		semesters = []interface{}{}
	}

	courseOfferings, _ := p.fetchResourceList(fmt.Sprintf("%s%s/course-offerings", p.CoursesURL, apiV1Prefix), authHeader)
	if courseOfferings == nil {
		courseOfferings = []interface{}{}
	}

	sections, _ := p.fetchResourceList(fmt.Sprintf("%s%s/sections", p.CoursesURL, apiV1Prefix), authHeader)
	if sections == nil {
		sections = []interface{}{}
	}

	semesterMap := buildSemesterMap(semesters)
	enrichedCourses := make([]interface{}, 0, len(courses))
	for _, course := range courses {
		if courseMap, ok := course.(map[string]interface{}); ok {
			enrichedCourses = append(enrichedCourses, buildCourseWithSections(courseMap, courseOfferings, sections, semesterMap))
		} else {
			enrichedCourses = append(enrichedCourses, course)
		}
	}

	return c.JSON(fiber.Map{
		"success":          true,
		"courses":          enrichedCourses,
		"current_semester": currentSemester,
		"semesters":        semesters,
		"course_offerings": courseOfferings,
		"sections":         sections,
	})
}

// GetSectionComplete combina múltiples llamadas para obtener información completa de una sección:
// GET /sections/{id}, GET /course-offerings/{id}, GET /courses/{id}, GET /semesters/{id}
func (p *ProxyService) GetSectionComplete(c *fiber.Ctx) error {
	sectionID := c.Params("id")
	authHeader := c.Get(headerAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthHeaderRequired})
	}

	sectionURL := fmt.Sprintf("%s%s/sections/%s", p.CoursesURL, apiV1Prefix, sectionID)
	sectionBytes, sectionStatus, err := p.makeAuthRequest("GET", sectionURL, nil, authHeader)
	if err != nil || sectionStatus != http.StatusOK {
		log.Printf("Error fetching section: %v, status: %d", err, sectionStatus)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Failed to fetch section"})
	}
	var section map[string]interface{}
	if err := json.Unmarshal(sectionBytes, &section); err != nil {
		log.Printf("Error decoding section: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode section"})
	}

	courseOfferingID := section["course_offering_id"]
	if courseOfferingID == nil {
		return c.JSON(fiber.Map{"success": true, "section": section})
	}

	courseOffering, course, semester := p.fetchSectionDetails(courseOfferingID, authHeader)
	response := fiber.Map{"success": true, "section": section}
	if courseOffering != nil {
		response["course_offering"] = courseOffering
	}
	if course != nil {
		response["course"] = course
	}
	if semester != nil {
		response["semester"] = semester
		if term, ok := semester["term"]; ok {
			if year, ok := semester["year"]; ok {
				response["semester_info"] = fmt.Sprintf("%v_%v", term, year)
			}
		}
	}
	return c.JSON(response)
}

// GetSectionGroupsData obtiene toda la información necesaria para la sección de grupos.
// La petición se redirige a ms-core, que consolida la respuesta en un solo payload.
func (p *ProxyService) GetSectionGroupsData(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	authHeader := c.Get(headerAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthHeaderRequired})
	}

	targetURL := fmt.Sprintf("%s%s/sections/%s/complete-groups-data", p.CoreURL, apiV1Prefix, sectionID)
	body, status, err := p.makeAuthRequest("GET", targetURL, nil, authHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateRequest})
	}
	c.Set(headerContentType, mimeApplicationJSON)
	return c.Status(status).Send(body)
}

// GetSectionGroupData proxies to ms-core section group-data endpoint
func (p *ProxyService) GetSectionGroupData(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	authHeader := c.Get(headerAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthHeaderRequired})
	}

	targetURL := fmt.Sprintf("%s%s/sections/%s/group-data", p.CoreURL, apiV1Prefix, sectionID)
	body, status, err := p.makeAuthRequest("GET", targetURL, nil, authHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateRequest})
	}
	c.Set(headerContentType, mimeApplicationJSON)
	return c.Status(status).Send(body)
}

// GetSectionSurveyData proxies to ms-core section survey-data endpoint
func (p *ProxyService) GetSectionSurveyData(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	authHeader := c.Get(headerAuthorization)
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthHeaderRequired})
	}

	targetURL := fmt.Sprintf("%s%s/sections/%s/survey-data", p.CoreURL, apiV1Prefix, sectionID)
	body, status, err := p.makeAuthRequest("GET", targetURL, nil, authHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateRequest})
	}
	c.Set(headerContentType, mimeApplicationJSON)
	return c.Status(status).Send(body)
}

// Helper functions para encoding/decoding JSON
func encodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decodeJSON(r io.Reader, v interface{}) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// HealthCheck verifica el estado de los microservicios
func (p *ProxyService) HealthCheck() error {
	// Verificar registro curricular
	resp, err := p.HTTPClient.Get(fmt.Sprintf("%s/health", p.RegistroCurricularURL))
	if err != nil {
		return fmt.Errorf("registro curricular: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registro curricular returned status %d", resp.StatusCode)
	}

	// Verificar courses
	resp, err = p.HTTPClient.Get(fmt.Sprintf("%s/health", p.CoursesURL))
	if err != nil {
		return fmt.Errorf("courses: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("courses returned status %d", resp.StatusCode)
	}

	return nil
}

// GetQuestionnaireResponses obtiene las respuestas de un cuestionario Enneagram con información de estudiantes
func (p *ProxyService) GetQuestionnaireResponses(c *fiber.Ctx) error {
	questionnaireID := c.Params("id")
	if questionnaireID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errQuestionnaireIDRequired})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	responsesURL := fmt.Sprintf("%s%s/plugin/enneagram/tests/%s/responses", p.CoreURL, apiV1Prefix, questionnaireID)

	if studentID := c.Query("student_id"); studentID != "" {
		responsesURL = fmt.Sprintf(fmtStudentIDQS, responsesURL, studentID)
	}
	log.Printf("Fetching responses from: %s", responsesURL)
	return p.handleQuestionnaireResponses(c, responsesURL, token)
}

// GetQuestionnaires obtiene cuestionarios Enneagram con formato simplificado (solo campos básicos)
func (p *ProxyService) GetQuestionnaires(c *fiber.Ctx) error {
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	sectionID := c.Query("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDQueryRequired})
	}
	url := fmt.Sprintf("%s%s/plugin/enneagram/tests?section_id=%s", p.CoreURL, apiV1Prefix, sectionID)
	log.Printf("Fetching questionnaires from: %s", url)
	return p.fetchQuestionnairesFromCore(url, token, c)
}

// GetBFI44Questionnaires obtiene cuestionarios BFI-44 con formato simplificado
func (p *ProxyService) GetBFI44Questionnaires(c *fiber.Ctx) error {
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	sectionID := c.Query("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDQueryRequired})
	}
	url := fmt.Sprintf("%s%s/plugin/bfi44/tests?section_id=%s", p.CoreURL, apiV1Prefix, sectionID)
	log.Printf("Fetching BFI-44 questionnaires from: %s", url)
	return p.fetchQuestionnairesFromCore(url, token, c)
}

// GetBFI44QuestionnaireResponses obtiene las respuestas de un cuestionario BFI-44 con información de estudiantes
func (p *ProxyService) GetBFI44QuestionnaireResponses(c *fiber.Ctx) error {
	questionnaireID := c.Params("id")
	if questionnaireID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errQuestionnaireIDRequired})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	responsesURL := fmt.Sprintf("%s%s/plugin/bfi44/tests/%s/responses", p.CoreURL, apiV1Prefix, questionnaireID)
	if studentID := c.Query("student_id"); studentID != "" {
		responsesURL = fmt.Sprintf(fmtStudentIDQS, responsesURL, studentID)
	}
	log.Printf("Fetching BFI-44 responses from: %s", responsesURL)
	return p.handleQuestionnaireResponses(c, responsesURL, token)
}

// GetMyGroupComplete obtiene el grupo del estudiante con información completa de los compañeros
// GET /core/groups/section/{sectionId}/my-group -> obtiene el grupo con IDs
// Luego enriquece los datos usando fetchStudentDetails
func (p *ProxyService) GetMyGroupComplete(c *fiber.Ctx) error {
	sectionID := c.Params("sectionId")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDRequired})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}

	// 1. Obtener el grupo desde ms-core
	myGroupURL := fmt.Sprintf("%s%s/groups/section/%s/my-group", p.CoreURL, apiV1Prefix, sectionID)
	log.Printf("Fetching my-group from: %s", myGroupURL)

	req, err := http.NewRequest("GET", myGroupURL, nil)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateRequest})
	}
	req.Header.Set(headerAuthorization, token)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error fetching my-group: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to fetch my-group from ms-core",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Error response from ms-core: %s", string(body))
		return c.Status(resp.StatusCode).Send(body)
	}

	var myGroupData struct {
		Success bool `json:"success"`
		Data    struct {
			GroupName   string  `json:"group_name"`
			GroupNumber int     `json:"group_number"`
			Students    []int64 `json:"students"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&myGroupData); err != nil {
		log.Printf("Error decoding my-group response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode my-group response",
		})
	}

	// Si no hay estudiantes en el grupo, devolver respuesta vacía
	if len(myGroupData.Data.Students) == 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"group_name":   myGroupData.Data.GroupName,
				"group_number": myGroupData.Data.GroupNumber,
				"students":     []interface{}{},
			},
		})
	}

	// 2. Obtener información completa de los estudiantes usando el helper
	studentsDetails, err := p.fetchStudentDetails(myGroupData.Data.Students, sectionID, token)
	if err != nil {
		log.Printf("Error fetching student details: %v", err)
		// Si falla, devolver solo con IDs
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"group_name":   myGroupData.Data.GroupName,
				"group_number": myGroupData.Data.GroupNumber,
				"students":     myGroupData.Data.Students,
			},
		})
	}

	// 3. Retornar respuesta enriquecida
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"group_name":   myGroupData.Data.GroupName,
			"group_number": myGroupData.Data.GroupNumber,
			"students":     studentsDetails,
		},
	})
}

// GetSectionResponses obtiene todas las respuestas de evaluación de pares para una sección
func (p *ProxyService) GetSectionResponses(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDRequired})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	url := fmt.Sprintf("%s%s/sections/%s/responses", p.CoreURL, apiV1Prefix, sectionID)
	log.Printf("Fetching all peer evaluation responses from: %s", url)
	return p.proxyToCoreSection(c, url, token, "peer evaluation responses")
}

// GetSectionResponsesSummary obtiene un resumen de respuestas agregadas por estudiante y grupo
func (p *ProxyService) GetSectionResponsesSummary(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDRequired})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	url := fmt.Sprintf("%s%s/sections/%s/responses/summary", p.CoreURL, apiV1Prefix, sectionID)
	log.Printf("Fetching peer evaluation responses summary from: %s", url)
	return p.proxyToCoreSection(c, url, token, "peer evaluation responses summary")
}

// GetSectionGroupResponses obtiene las respuestas de evaluación de pares para un grupo específico de una sección
func (p *ProxyService) GetSectionGroupResponses(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	groupID := c.Params("group_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDRequired})
	}
	if groupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "group_id is required"})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	url := fmt.Sprintf("%s%s/sections/%s/responses/%s", p.CoreURL, apiV1Prefix, sectionID, groupID)
	log.Printf("Fetching peer evaluation responses from: %s", url)
	return p.proxyToCoreSection(c, url, token, "peer evaluation group responses")
}

// ExportPeerEvaluationExcel redirige la solicitud de exportación a Excel a ms-core
func (p *ProxyService) ExportPeerEvaluationExcel(c *fiber.Ctx) error {
	surveyID := c.Params("surveyId")
	if surveyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "surveyId is required"})
	}
	token := c.Get(headerAuthorization)
	// En ms-core la ruta es /api/v1/surveys/:surveyId/export/excel
	exportURL := fmt.Sprintf("%s%s/surveys/%s/export/excel", p.CoreURL, apiV1Prefix, surveyID)
	log.Printf(fmtProxyingRequest, c.OriginalURL(), exportURL)

	req, err := http.NewRequest("GET", exportURL, nil)
	if err != nil {
		log.Printf(fmtErrCreatingReq, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedCreateProxyRequest})
	}
	if token != "" {
		req.Header.Set(headerAuthorization, token)
	}
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error making request to ms-core: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": errFailedConnectToCore})
	}
	defer resp.Body.Close()

	// Copiar todos los headers de la respuesta (incluyendo Content-Type y Content-Disposition)
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(fmtErrReadingResp, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errFailedReadRespCore})
	}
	return c.Status(resp.StatusCode).Send(body)
}

// GetMBTIQuestionnaires obtiene cuestionarios MBTI con formato simplificado
func (p *ProxyService) GetMBTIQuestionnaires(c *fiber.Ctx) error {
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	sectionID := c.Query("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errSectionIDQueryRequired})
	}
	url := fmt.Sprintf("%s%s/plugin/mbti/tests?section_id=%s", p.CoreURL, apiV1Prefix, sectionID)
	log.Printf("Fetching MBTI questionnaires from: %s", url)
	return p.fetchQuestionnairesFromCore(url, token, c)
}

// GetMBTIQuestionnaireResponses obtiene las respuestas de un cuestionario MBTI con información de estudiantes
func (p *ProxyService) GetMBTIQuestionnaireResponses(c *fiber.Ctx) error {
	questionnaireID := c.Params("id")
	if questionnaireID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errQuestionnaireIDRequired})
	}
	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}
	responsesURL := fmt.Sprintf("%s%s/plugin/mbti/tests/%s/responses", p.CoreURL, apiV1Prefix, questionnaireID)
	if studentID := c.Query("student_id"); studentID != "" {
		responsesURL = fmt.Sprintf(fmtStudentIDQS, responsesURL, studentID)
	}
	log.Printf("Fetching MBTI responses from: %s", responsesURL)

	responses, status, err := p.fetchQuestionnaireResponses(responsesURL, token)
	if err != nil {
		log.Printf(fmtErrFetchingResps, err)
		return c.Status(status).JSON(fiber.Map{"error": errFailedFetchRespsCore})
	}
	if len(responses) == 0 {
		return c.JSON(fiber.Map{"success": true, "data": []map[string]interface{}{}})
	}

	ids := extractStudentIDsInt(responses)
	if len(ids) == 0 {
		return c.JSON(fiber.Map{"success": true, "data": responses})
	}
	usersMap, err := p.fetchUsersBatch(ids, token)
	if err != nil {
		log.Printf(fmtErrFetchingUsers, err)
		return c.JSON(fiber.Map{"success": true, "data": responses})
	}
	return c.JSON(fiber.Map{"success": true, "data": buildEnrichedResponses(responses, usersMap)})
}

// GetStudentTestResponse verifica si un estudiante ya respondió un test específico
// Endpoint: GET /api/v1/core/plugin/:pluginName/tests/:testId/responses/student/:studentId
func (p *ProxyService) GetStudentTestResponse(c *fiber.Ctx) error {
	pluginName := c.Params("pluginName")
	testId := c.Params("testId")
	studentId := c.Params("studentId")

	if pluginName == "" || testId == "" || studentId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "pluginName, testId and studentId are required"})
	}

	token := c.Get(headerAuthorization)
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": errAuthTokenRequired})
	}

	// Llamar al endpoint de ms-core - usa query param para student_id (el standard del sistema)
	url := fmt.Sprintf("%s%s/plugin/%s/tests/%s/responses?student_id=%s", p.CoreURL, apiV1Prefix, pluginName, testId, studentId)
	log.Printf("Checking student test response from: %s", url)

	body, status, err := p.makeAuthRequest("GET", url, nil, token)
	if err != nil {
		log.Printf("Error checking student response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check student response"})
	}

	// Si el test no existe o no hay respuesta, ms-core podría devolver 404 o empty data
	if status == http.StatusNotFound {
		return c.JSON(fiber.Map{
			"success":       true,
			"data":          nil,
			"has_responded": false,
		})
	}

	if status != http.StatusOK {
		log.Printf("ms-core returned status %d: %s", status, string(body))
		return c.Status(status).Send(body)
	}

	// Parsear la respuesta
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		log.Printf("Error parsing response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse response"})
	}

	// Verificar si la respuesta está completa
	data := responseData["data"]
	hasResponded := false
	isComplete := false

	if data != nil {
		if respMap, ok := data.(map[string]interface{}); ok {
			if val, ok := respMap["is_complete"].(bool); ok && val {
				hasResponded = true
				isComplete = true
			}
		} else if respArr, ok := data.([]interface{}); ok && len(respArr) > 0 {
			// A veces devuelve un array con un elemento
			if respMap, ok := respArr[0].(map[string]interface{}); ok {
				if val, ok := respMap["is_complete"].(bool); ok && val {
					hasResponded = true
					isComplete = true
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"success":       true,
		"has_responded": hasResponded,
		"is_complete":   isComplete,
		"data":          data,
	})
}
