package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ms-core/internal/config"
	"ms-core/internal/database"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (p *PeerEvaluationPlugin) submitResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	evaluatorID, ok := request["evaluator_id"].(string)
	if !ok || evaluatorID == "" {
		return nil, fmt.Errorf("evaluator_id is required")
	}

	evaluatedID, ok := request["evaluated_id"].(string)
	if !ok || evaluatedID == "" {
		return nil, fmt.Errorf("evaluated_id is required")
	}

	groupID, ok := request["group_id"].(string)
	if !ok || groupID == "" {
		return nil, fmt.Errorf("group_id is required")
	}

	answers, err := parsePeerResponses(request)
	if err != nil {
		return nil, err
	}

	var survey PeerEvaluationSurvey
	if err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey); err != nil {
		return nil, fmt.Errorf(errFailedGetSurveyFmt, err)
	}

	if !validateGroupMembership(survey.GroupSnapshots, groupID, evaluatorID, evaluatedID) {
		return nil, fmt.Errorf("evaluator and evaluated must be in the same group according to the survey snapshot")
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	if responsesCollection == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	now := time.Now()

	filter := bson.M{
		"survey_id":    surveyID,
		"evaluator_id": evaluatorID,
		"evaluated_id": evaluatedID,
	}
	update := bson.M{
		"$set": bson.M{
			"group_id":     groupID,
			"responses":    answers,
			"is_complete":  true,
			"submitted_at": now,
			"updated_at":   now,
		},
		"$setOnInsert": bson.M{
			"survey_id":    surveyID,
			"evaluator_id": evaluatorID,
			"evaluated_id": evaluatedID,
			"created_at":   now,
		},
	}
	upsertOpts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var saved PeerEvaluationResponse
	if err = responsesCollection.FindOneAndUpdate(ctx, filter, update, upsertOpts).Decode(&saved); err != nil {
		return nil, fmt.Errorf("failed to submit response: %v", err)
	}

	return saved, nil
}

func (p *PeerEvaluationPlugin) getResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	evaluatorID, ok := request["evaluator_id"].(string)
	if !ok {
		evaluatorID = ""
	}

	evaluatedIDsStr, ok := request["evaluated_ids"].(string)
	if !ok {
		evaluatedIDsStr = ""
	}

	groupID, ok := request["group_id"].(string)
	if !ok {
		groupID = ""
	}

	filter := bson.M{"survey_id": surveyID}
	if evaluatorID != "" {
		filter["evaluator_id"] = evaluatorID
	}
	if evaluatedIDsStr != "" {
		evaluatedIDs := strings.Split(evaluatedIDsStr, ",")
		filter["evaluated_id"] = bson.M{"$in": evaluatedIDs}
	}
	if groupID != "" {
		filter["group_id"] = groupID
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	if responsesCollection == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	cursor, err := responsesCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf(errFailedGetResponsesFmt, err)
	}
	defer cursor.Close(ctx)

	var responses []PeerEvaluationResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf(errFailedDecodeResponsesFmt, err)
	}

	return map[string]interface{}{
		"success": true,
		"data":    responses,
	}, nil
}

func averageScores(questionScores map[int][]float64) []float64 {
	averages := make([]float64, 5)
	for qid := 1; qid <= 5; qid++ {
		scores := questionScores[qid]
		if len(scores) == 0 {
			continue
		}
		sum := 0.0
		for _, s := range scores {
			sum += s
		}
		averages[qid-1] = sum / float64(len(scores))
	}
	return averages
}

func (p *PeerEvaluationPlugin) getResults(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	if responsesCollection == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	cursor, err := responsesCollection.Find(ctx, bson.M{"survey_id": surveyID})
	if err != nil {
		return nil, fmt.Errorf(errFailedGetResponsesFmt, err)
	}
	defer cursor.Close(ctx)

	var responses []PeerEvaluationResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf(errFailedDecodeResponsesFmt, err)
	}

	resultsMap := make(map[string]map[int][]float64)
	for _, response := range responses {
		if _, exists := resultsMap[response.EvaluatedID]; !exists {
			resultsMap[response.EvaluatedID] = make(map[int][]float64)
		}
		for _, ans := range response.Responses {
			resultsMap[response.EvaluatedID][ans.QuestionID] = append(
				resultsMap[response.EvaluatedID][ans.QuestionID], float64(ans.Score))
		}
	}

	var results []map[string]interface{}
	for studentID, questionScores := range resultsMap {
		results = append(results, map[string]interface{}{
			"student_id": studentID,
			"averages":   averageScores(questionScores),
		})
	}

	return map[string]interface{}{"success": true, "data": results}, nil
}

func (p *PeerEvaluationPlugin) getSectionResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	surveyIDs, err := fetchSectionSurveyIDs(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if len(surveyIDs) == 0 {
		return []interface{}{}, nil
	}

	return fetchResponsesBySurveyIDs(ctx, surveyIDs)
}

func (p *PeerEvaluationPlugin) getSectionResponsesSummary(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	surveyIDs, err := fetchSectionSurveyIDs(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if len(surveyIDs) == 0 {
		return []interface{}{}, nil
	}

	responses, err := fetchResponsesBySurveyIDs(ctx, surveyIDs)
	if err != nil {
		return nil, err
	}

	type ResponseSummary struct {
		StudentID       string `bson:"student_id"`
		GroupID         string `bson:"group_id"`
		NumberResponses int    `bson:"number_responses"`
	}

	summaryMap := make(map[string]*ResponseSummary)
	for _, resp := range responses {
		evaluatorID, _ := resp["evaluator_id"].(string)
		groupID, _ := resp["group_id"].(string)
		responsesArray, _ := resp["responses"].(primitive.A)
		questionCount := len(responsesArray)
		key := evaluatorID + ":" + groupID
		if summary, exists := summaryMap[key]; exists {
			summary.NumberResponses += questionCount
		} else {
			summaryMap[key] = &ResponseSummary{
				StudentID:       evaluatorID,
				GroupID:         groupID,
				NumberResponses: questionCount,
			}
		}
	}

	result := make([]ResponseSummary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		result = append(result, *summary)
	}
	return result, nil
}

func (p *PeerEvaluationPlugin) getSectionGroupResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	groupID, ok := request["group_id"].(string)
	if !ok || groupID == "" {
		return nil, fmt.Errorf("group_id is required")
	}

	surveyIDs, err := fetchSectionSurveyIDs(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if len(surveyIDs) == 0 {
		return []interface{}{}, nil
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	filter := bson.M{
		"survey_id": bson.M{"$in": surveyIDs},
		"group_id":  groupID,
	}

	cursor, err := responsesCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf(errFailedGetResponsesFmt, err)
	}
	defer cursor.Close(ctx)

	var responses []bson.M
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf(errFailedDecodeResponsesFmt, err)
	}
	if responses == nil {
		responses = []bson.M{}
	}
	return responses, nil
}

func fetchStudentMap(ctx context.Context, responses []PeerEvaluationResponse) map[string]map[string]interface{} {
	studentIDsSet := make(map[string]bool)
	for _, resp := range responses {
		studentIDsSet[resp.EvaluatorID] = true
		studentIDsSet[resp.EvaluatedID] = true
	}

	var studentIDsInt64 []int64
	for idStr := range studentIDsSet {
		var idInt64 int64
		if _, err := fmt.Sscanf(idStr, "%d", &idInt64); err == nil {
			studentIDsInt64 = append(studentIDsInt64, idInt64)
		}
	}

	studentMap := make(map[string]map[string]interface{})
	if len(studentIDsInt64) == 0 {
		return studentMap
	}

	studentsCollection := database.GetCollection("students")
	if studentsCollection == nil {
		return studentMap
	}

	studentCursor, err := studentsCollection.Find(ctx, bson.M{"_id": bson.M{"$in": studentIDsInt64}})
	if err != nil {
		return studentMap
	}
	defer studentCursor.Close(ctx)

	var students []bson.M
	if err = studentCursor.All(ctx, &students); err == nil {
		for _, student := range students {
			if id, ok := student["_id"].(int64); ok {
				studentMap[fmt.Sprintf("%d", id)] = student
			}
		}
	}
	return studentMap
}

func fetchSectionInfo(sectionID, authHeader string) (courseName, sectionCode, semester string) {
	if sectionID == "" || authHeader == "" {
		return
	}
	cfg := config.LoadConfig()
	url := fmt.Sprintf("%s/api/v1/sections/%s/complete", cfg.APIGatewayURL, sectionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var info struct {
		Course       map[string]interface{} `json:"course"`
		Section      map[string]interface{} `json:"section"`
		SemesterInfo string                 `json:"semester_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return
	}
	if info.Course != nil {
		courseName, _ = info.Course["name"].(string)
	}
	if info.Section != nil {
		if name, ok := info.Section["name"].(string); ok {
			sectionCode = name
		} else if code, ok := info.Section["code"].(string); ok {
			sectionCode = code
		}
	}
	semester = info.SemesterInfo
	return
}

func excelHeaderStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func excelCellStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "D3D3D3", Style: 1},
			{Type: "top", Color: "D3D3D3", Style: 1},
			{Type: "bottom", Color: "D3D3D3", Style: 1},
			{Type: "right", Color: "D3D3D3", Style: 1},
		},
	})
	return style
}

func writeSheetHeaders(f *excelize.File, sheetName string, headers []string, style int) {
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, style)
	}
}

func setSheetColWidths(f *excelize.File, sheetName string, widths []float64) {
	for i, w := range widths {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, colName, colName, w)
	}
}

func groupNameForID(snapshots []GroupSnapshot, groupID string) string {
	for _, s := range snapshots {
		if s.GroupID == groupID {
			return s.GroupName
		}
	}
	return groupID
}

func getStringValue(data map[string]interface{}, key string, defaultValue string) string {
	if data == nil {
		return defaultValue
	}
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}

func writeDetailSheet(f *excelize.File, sheetName string, styles ExcelStyles,
	responses []PeerEvaluationResponse, survey PeerEvaluationSurvey,
	studentMap map[string]map[string]interface{}, info SectionExcelInfo) {

	headers := []string{
		"Nombre Curso", "Secci\xf3n", "Semestre", "ID_Encuesta", "Grupo ID",
		"Nombre Evaluador", "Rut Evaluador", "Nombre Evaluado", "Rut Evaluado",
		"Criterio 1", "Criterio 2", "Criterio 3", "Criterio 4", "Criterio 5", "Feedback",
	}
	writeSheetHeaders(f, sheetName, headers, styles.Header)
	setSheetColWidths(f, sheetName, []float64{25, 12, 15, 25, 15, 25, 12, 25, 12, 12, 12, 12, 12, 12, 40})

	row := 2
	for _, resp := range responses {
		evatorData := studentMap[resp.EvaluatorID]
		evatedData := studentMap[resp.EvaluatedID]
		evatorFull := getStringValue(evatorData, "name", "N/A") + " " + getStringValue(evatorData, "lastname", "N/A")
		evatedFull := getStringValue(evatedData, "name", "N/A") + " " + getStringValue(evatedData, "lastname", "N/A")
		grpName := groupNameForID(survey.GroupSnapshots, resp.GroupID)

		scores := make([]int, 5)
		for _, answer := range resp.Responses {
			if answer.QuestionID >= 1 && answer.QuestionID <= 5 {
				scores[answer.QuestionID-1] = answer.Score
			}
		}

		values := []interface{}{
			info.CourseName, info.SectionCode, info.Semester, survey.ID.Hex(), grpName,
			evatorFull, getStringValue(evatorData, "rut", "N/A"),
			evatedFull, getStringValue(evatedData, "rut", "N/A"),
			scores[0], scores[1], scores[2], scores[3], scores[4], "",
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			f.SetCellValue(sheetName, cell, v)
			f.SetCellStyle(sheetName, cell, cell, styles.Cell)
		}
		row++
	}
}

func buildStudentSummaryMap(responses []PeerEvaluationResponse, survey PeerEvaluationSurvey,
	studentMap map[string]map[string]interface{}) map[string]*StudentSummaryData {

	summaryMap := make(map[string]*StudentSummaryData)
	for _, resp := range responses {
		if _, exists := summaryMap[resp.EvaluatedID]; !exists {
			evatedData := studentMap[resp.EvaluatedID]
			summaryMap[resp.EvaluatedID] = &StudentSummaryData{
				RUT:       getStringValue(evatedData, "rut", resp.EvaluatedID),
				Name:      getStringValue(evatedData, "name", "N/A"),
				Lastname:  getStringValue(evatedData, "lastname", "N/A"),
				GroupName: groupNameForID(survey.GroupSnapshots, resp.GroupID),
			}
		}
		summary := summaryMap[resp.EvaluatedID]
		for _, answer := range resp.Responses {
			if answer.QuestionID >= 1 && answer.QuestionID <= 5 {
				summary.Scores[answer.QuestionID-1] = append(summary.Scores[answer.QuestionID-1], answer.Score)
			}
		}
		summary.Count++
	}
	return summaryMap
}

func calculateStudentAvgScore(scores [5][]int) float64 {
	total, valid := 0.0, 0
	for i := 0; i < 5; i++ {
		if len(scores[i]) == 0 {
			continue
		}
		sum := 0
		for _, score := range scores[i] {
			sum += score
		}
		total += float64(sum) / float64(len(scores[i]))
		valid++
	}
	if valid > 0 {
		return total / float64(valid)
	}
	return 0
}

func lookupCareer(data map[string]interface{}) string {
	if c, ok := data["career_code"].(string); ok {
		return c
	}
	if c, ok := data["career"].(string); ok {
		return c
	}
	return ""
}

func writeSummarySheet(f *excelize.File, sheetName string, styles ExcelStyles,
	summaryMap map[string]*StudentSummaryData, survey PeerEvaluationSurvey,
	studentMap map[string]map[string]interface{}, info SectionExcelInfo) {

	headers := []string{
		"Nombre Curso", "Secci\xf3n", "Semestre", "ID_Encuesta",
		"Nombre Alumno", "Correo", "Rut", "Carrera", "Grupo",
		"promedio final", "cantidad de evaluaciones",
	}
	writeSheetHeaders(f, sheetName, headers, styles.Header)
	setSheetColWidths(f, sheetName, []float64{25, 12, 15, 25, 30, 30, 15, 15, 15, 15, 20})

	summaryRow := 2
	for evaluatedID, s := range summaryMap {
		evatedData := studentMap[evaluatedID]
		values := []interface{}{
			info.CourseName, info.SectionCode, info.Semester, survey.ID.Hex(),
			s.Name + " " + s.Lastname,
			getStringValue(evatedData, "email", "N/A"),
			s.RUT, lookupCareer(evatedData), s.GroupName,
			calculateStudentAvgScore(s.Scores), s.Count,
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, summaryRow)
			f.SetCellValue(sheetName, cell, v)
			f.SetCellStyle(sheetName, cell, cell, styles.Cell)
		}
		summaryRow++
	}
}

func (p *PeerEvaluationPlugin) exportToExcel(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	var survey PeerEvaluationSurvey
	if err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey); err != nil {
		return nil, fmt.Errorf(errFailedGetSurveyFmt, err)
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	if responsesCollection == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	cursor, err := responsesCollection.Find(ctx, bson.M{"survey_id": surveyID})
	if err != nil {
		return nil, fmt.Errorf(errFailedGetResponsesFmt, err)
	}
	defer cursor.Close(ctx)

	var responses []PeerEvaluationResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf(errFailedDecodeResponsesFmt, err)
	}

	studentMap := fetchStudentMap(ctx, responses)

	authHeader, _ := request["auth_header"].(string)
	courseName, sectionCode, semester := fetchSectionInfo(survey.SectionID, authHeader)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing Excel file: %v\n", err)
		}
	}()

	hStyle := excelHeaderStyle(f)
	cStyle := excelCellStyle(f)

	detailSheet := "Detalle de Respuestas"
	detailIdx, err := f.NewSheet(detailSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %v", err)
	}
	f.SetActiveSheet(detailIdx)
	f.DeleteSheet("Sheet1")
	sheetStyles := ExcelStyles{Header: hStyle, Cell: cStyle}
	sectionInfo := SectionExcelInfo{CourseName: courseName, SectionCode: sectionCode, Semester: semester}
	writeDetailSheet(f, detailSheet, sheetStyles, responses, survey, studentMap, sectionInfo)

	summarySheet := "Resumen y Promedios"
	if summaryIdx, err := f.NewSheet(summarySheet); err == nil {
		f.SetActiveSheet(summaryIdx)
		summaryMap := buildStudentSummaryMap(responses, survey, studentMap)
		writeSummarySheet(f, summarySheet, sheetStyles, summaryMap, survey, studentMap, sectionInfo)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}

	return map[string]interface{}{
		"success":  true,
		"filename": fmt.Sprintf("evaluacion_pares_%s.xlsx", survey.Title),
		"data":     buffer.Bytes(),
	}, nil
}
