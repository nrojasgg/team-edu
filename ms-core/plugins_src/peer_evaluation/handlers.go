package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"ms-core/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	errSurveyIDRequired         = "survey_id is required"
	errInvalidSurveyIDFmt       = "invalid survey_id: %v"
	errSectionIDRequired        = "section_id is required"
	errFailedGetSurveysFmt      = "failed to get surveys: %v"
	errFailedDecodeSurveysFmt   = "failed to decode surveys: %v"
	errFailedGetSurveyFmt       = "failed to get survey: %v"
	errFailedGetRespCollection  = "failed to get responses collection"
	errFailedGetResponsesFmt    = "failed to get responses: %v"
	errFailedDecodeResponsesFmt = "failed to decode responses: %v"
)

func (p *PeerEvaluationPlugin) getDefaultQuestions() []PeerEvaluationQuestion {
	return []PeerEvaluationQuestion{
		{
			ID:       1,
			Text:     "Realiza las tareas que le son asignadas dentro del grupo en los plazos requeridos.",
			Category: "Tareas",
			Required: true,
		},
		{
			ID:       2,
			Text:     "Participa de forma activa en los espacios de encuentro del equipo, compartiendo la información, conocimientos y las experiencias.",
			Category: "Participación",
			Required: true,
		},
		{
			ID:       3,
			Text:     "Colabora en la definición, organización y distribución de las tareas del equipo.",
			Category: "Organización",
			Required: true,
		},
		{
			ID:       4,
			Text:     "Se orienta a la consecución de acuerdos y objetivos comunes y se compromete con ellos.",
			Category: "Compromiso",
			Required: true,
		},
		{
			ID:       5,
			Text:     "Toma en cuenta los puntos de vista de los demás y retroalimenta constructivamente.",
			Category: "Prosocialidad",
			Required: true,
		},
	}
}

func parseOptionalDate(request map[string]interface{}, key string) *time.Time {
	str, ok := request[key].(string)
	if !ok || str == "" {
		return nil
	}
	for _, format := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(format, str); err == nil {
			return &t
		}
	}
	return nil
}

func (p *PeerEvaluationPlugin) createSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	title, ok := request["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title is required and must be a string")
	}

	description, ok := request["description"].(string)
	if !ok {
		description = ""
	}

	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf("section_id is required and must be a string")
	}

	createdBy, ok := request["created_by"].(string)
	if !ok || createdBy == "" {
		return nil, fmt.Errorf("created_by is required and must be a string")
	}

	groupSnapshots, err := p.getCurrentGroupSnapshots(ctx, sectionID)
	if err != nil {
		fmt.Printf("DEBUG: Warning getting group snapshots: %v\n", err)
		groupSnapshots = []GroupSnapshot{}
	}

	survey := PeerEvaluationSurvey{
		Title:          title,
		Description:    description,
		SectionID:      sectionID,
		CreatedBy:      createdBy,
		Questions:      p.getDefaultQuestions(),
		IsActive:       true,
		StartDate:      parseOptionalDate(request, "start_date"),
		EndDate:        parseOptionalDate(request, "end_date"),
		GroupSnapshots: groupSnapshots,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := p.collection.InsertOne(ctx, survey)
	if err != nil {
		return nil, fmt.Errorf("failed to create survey: %v", err)
	}

	survey.ID = result.InsertedID.(primitive.ObjectID)

	return map[string]interface{}{
		"success": true,
		"data":    survey,
	}, nil
}

func (p *PeerEvaluationPlugin) getSurveys(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	filter := bson.M{"section_id": sectionID}

	cursor, err := p.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf(errFailedGetSurveysFmt, err)
	}
	defer cursor.Close(ctx)

	var surveys []PeerEvaluationSurvey
	if err = cursor.All(ctx, &surveys); err != nil {
		return nil, fmt.Errorf(errFailedDecodeSurveysFmt, err)
	}

	return map[string]interface{}{
		"success": true,
		"data":    surveys,
	}, nil
}

func (p *PeerEvaluationPlugin) getSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	var survey PeerEvaluationSurvey
	err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey)
	if err != nil {
		return nil, fmt.Errorf(errFailedGetSurveyFmt, err)
	}

	return map[string]interface{}{
		"success": true,
		"data":    survey,
	}, nil
}

func (p *PeerEvaluationPlugin) updateSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	updateData := bson.M{
		"updated_at": time.Now(),
	}

	if title, ok := request["title"].(string); ok && title != "" {
		updateData["title"] = title
	}

	if description, ok := request["description"].(string); ok {
		updateData["description"] = description
	}

	if isActive, ok := request["is_active"].(bool); ok {
		updateData["is_active"] = isActive
	}

	if sd := parseOptionalDate(request, "start_date"); sd != nil {
		updateData["start_date"] = *sd
	}
	if ed := parseOptionalDate(request, "end_date"); ed != nil {
		updateData["end_date"] = *ed
	}

	if snapshotsRaw, ok := request["group_snapshots"].([]interface{}); ok {
		snapshots := make([]GroupSnapshot, 0, len(snapshotsRaw))
		for _, sRaw := range snapshotsRaw {
			sMap, ok := sRaw.(map[string]interface{})
			if !ok {
				continue
			}
			snapshot := GroupSnapshot{}
			if gid, ok := sMap["group_id"].(string); ok {
				snapshot.GroupID = gid
			}
			if gname, ok := sMap["group_name"].(string); ok {
				snapshot.GroupName = gname
			}
			if gnum, ok := sMap["group_number"].(float64); ok {
				snapshot.GroupNumber = int(gnum)
			}
			if sids, ok := sMap["student_ids"].([]interface{}); ok {
				ids := make([]string, 0, len(sids))
				for _, sid := range sids {
					if sidStr, ok := sid.(string); ok {
						ids = append(ids, sidStr)
					}
				}
				snapshot.StudentIDs = ids
			}
			snapshots = append(snapshots, snapshot)
		}
		updateData["group_snapshots"] = snapshots
	}

	result, err := p.collection.UpdateOne(ctx, bson.M{"_id": surveyID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to update survey: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("survey not found")
	}

	var survey PeerEvaluationSurvey
	err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated survey: %v", err)
	}

	return map[string]interface{}{
		"success": true,
		"data":    survey,
	}, nil
}

func (p *PeerEvaluationPlugin) deleteSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	result, err := p.collection.DeleteOne(ctx, bson.M{"_id": surveyID})
	if err != nil {
		return nil, fmt.Errorf("failed to delete survey: %v", err)
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("survey not found")
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	if responsesCollection != nil {
		_, err = responsesCollection.DeleteMany(ctx, bson.M{"survey_id": surveyID})
		if err != nil {
			fmt.Printf("DEBUG: Warning deleting associated responses: %v\n", err)
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "Survey deleted successfully",
	}, nil
}

func (p *PeerEvaluationPlugin) getSurveyPeers(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	userIDStr, ok := request["user_id"].(string)
	if !ok || userIDStr == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	var survey PeerEvaluationSurvey
	err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey)
	if err != nil {
		return nil, fmt.Errorf(errFailedGetSurveyFmt, err)
	}

	for _, snapshot := range survey.GroupSnapshots {
		for _, studentID := range snapshot.StudentIDs {
			if studentID == userIDStr {
				return map[string]interface{}{
					"success":    true,
					"group_id":   snapshot.GroupID,
					"group_name": snapshot.GroupName,
					"peers":      snapshot.StudentIDs,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("user not found in any group snapshot for this survey")
}

func (p *PeerEvaluationPlugin) getCurrentGroupSnapshots(ctx context.Context, sectionID string) ([]GroupSnapshot, error) {
	groupsCollection := database.GetCollection("groups")
	if groupsCollection == nil {
		return []GroupSnapshot{}, fmt.Errorf("failed to get groups collection")
	}

	var sectionIDInt64 int64
	if _, err := fmt.Sscanf(sectionID, "%d", &sectionIDInt64); err != nil {
		return []GroupSnapshot{}, fmt.Errorf("invalid section_id format: %v", err)
	}

	cursor, err := groupsCollection.Find(ctx, bson.M{"section_id": sectionIDInt64})
	if err != nil {
		return []GroupSnapshot{}, fmt.Errorf("failed to find groups: %v", err)
	}
	defer cursor.Close(ctx)

	var groupDocs []bson.M
	if err = cursor.All(ctx, &groupDocs); err != nil {
		return []GroupSnapshot{}, fmt.Errorf("failed to decode groups: %v", err)
	}

	type GroupData struct {
		GroupID     string
		GroupName   string
		GroupNumber int
		StudentIDs  []string
	}

	groupDataMap := make(map[int]*GroupData)
	for _, groupDoc := range groupDocs {
		groupID, _ := groupDoc["_id"].(primitive.ObjectID)
		groupName, _ := groupDoc["group_name"].(string)
		groupNumber, _ := groupDoc["group_number"].(int32)
		studentIDInt64, _ := groupDoc["student_id"].(int64)
		studentID := fmt.Sprintf("%d", studentIDInt64)
		groupNum := int(groupNumber)
		if data, exists := groupDataMap[groupNum]; exists {
			data.StudentIDs = append(data.StudentIDs, studentID)
		} else {
			groupDataMap[groupNum] = &GroupData{
				GroupID:     groupID.Hex(),
				GroupName:   groupName,
				GroupNumber: groupNum,
				StudentIDs:  []string{studentID},
			}
		}
	}

	var keys []int
	for k := range groupDataMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	snapshots := make([]GroupSnapshot, 0, len(keys))
	for _, groupNum := range keys {
		d := groupDataMap[groupNum]
		snapshots = append(snapshots, GroupSnapshot{
			GroupID:     d.GroupID,
			GroupName:   d.GroupName,
			GroupNumber: d.GroupNumber,
			StudentIDs:  d.StudentIDs,
		})
	}
	return snapshots, nil
}

func fetchSectionSurveyIDs(ctx context.Context, sectionID string) ([]primitive.ObjectID, error) {
	surveysCollection := database.GetCollection("peer_evaluation_surveys")
	cursor, err := surveysCollection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, fmt.Errorf(errFailedGetSurveysFmt, err)
	}
	defer cursor.Close(ctx)

	var surveys []bson.M
	if err = cursor.All(ctx, &surveys); err != nil {
		return nil, fmt.Errorf(errFailedDecodeSurveysFmt, err)
	}

	ids := make([]primitive.ObjectID, 0, len(surveys))
	for _, s := range surveys {
		if id, ok := s["_id"].(primitive.ObjectID); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func fetchResponsesBySurveyIDs(ctx context.Context, surveyIDs []primitive.ObjectID) ([]bson.M, error) {
	responsesCollection := database.GetCollection("peer_evaluation_responses")
	cursor, err := responsesCollection.Find(ctx, bson.M{"survey_id": bson.M{"$in": surveyIDs}})
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

func validateGroupMembership(snapshots []GroupSnapshot, groupID, evaluatorID, evaluatedID string) bool {
	for _, snapshot := range snapshots {
		if snapshot.GroupID != groupID {
			continue
		}
		evaluatorFound, evaluatedFound := false, false
		for _, studentID := range snapshot.StudentIDs {
			if studentID == evaluatorID {
				evaluatorFound = true
			}
			if studentID == evaluatedID {
				evaluatedFound = true
			}
		}
		if evaluatorFound && evaluatedFound {
			return true
		}
	}
	return false
}

func parsePeerResponses(request map[string]interface{}) ([]PeerEvaluationAnswer, error) {
	responsesData, ok := request["responses"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("responses is required and must be an array")
	}
	answers := make([]PeerEvaluationAnswer, 0, len(responsesData))
	for _, r := range responsesData {
		respMap, ok := r.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response format")
		}
		questionIDFloat, ok := respMap["question_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("question_id must be a number")
		}
		scoreFloat, ok := respMap["score"].(float64)
		if !ok {
			return nil, fmt.Errorf("score must be a number")
		}
		score := int(scoreFloat)
		if score < 1 || score > 4 {
			return nil, fmt.Errorf("score must be between 1 and 4")
		}
		answers = append(answers, PeerEvaluationAnswer{
			QuestionID: int(questionIDFloat),
			Score:      score,
		})
	}
	if len(answers) != 5 {
		return nil, fmt.Errorf("all 5 questions must be answered")
	}
	return answers, nil
}
