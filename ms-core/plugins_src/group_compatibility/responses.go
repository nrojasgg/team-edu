package main

import (
	"context"
	"fmt"
	"time"

	"ms-core/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func parseQuestionAnswers(raw interface{}) ([]QuestionAnswer, error) {
	arr, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("question answers must be an array")
	}
	answers := make([]QuestionAnswer, 0, len(arr))
	for _, item := range arr {
		ans, err := parseOneAnswer(item)
		if err != nil {
			return nil, err
		}
		answers = append(answers, ans)
	}
	return answers, nil
}

func parseOneAnswer(item interface{}) (QuestionAnswer, error) {
	m, ok := item.(map[string]interface{})
	if !ok {
		return QuestionAnswer{}, fmt.Errorf("each answer must be an object")
	}
	studentID, ok := m["student_id"].(string)
	if !ok {
		return QuestionAnswer{}, fmt.Errorf("answer student_id must be a string")
	}
	return QuestionAnswer{StudentID: studentID, Response: toIntResponse(m["response"])}, nil
}

func toIntResponse(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	default:
		return 0
	}
}

func (p *GroupCompatibilityPlugin) submitResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	studentID, ok := request["student_id"].(string)
	if !ok || studentID == "" {
		return nil, fmt.Errorf("student_id is required")
	}

	sectionID, _ := request["section_id"].(string)
	surveyType, _ := request["survey_type"].(string)

	q1, err := parseQuestionAnswers(request["question1"])
	if err != nil {
		return nil, fmt.Errorf("invalid question1: %v", err)
	}
	q2, err := parseQuestionAnswers(request["question2"])
	if err != nil {
		return nil, fmt.Errorf("invalid question2: %v", err)
	}
	q3, err := parseQuestionAnswers(request["question3"])
	if err != nil {
		return nil, fmt.Errorf("invalid question3: %v", err)
	}
	q4, err := parseQuestionAnswers(request["question4"])
	if err != nil {
		return nil, fmt.Errorf("invalid question4: %v", err)
	}

	now := time.Now()

	responsesCol := database.GetCollection("group_compatibility_responses")
	if responsesCol == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	filter := bson.M{"survey_id": surveyID, "student_id": studentID}
	update := bson.M{
		"$set": bson.M{
			"survey_id":    surveyID,
			"student_id":   studentID,
			"section_id":   sectionID,
			"survey_type":  surveyType,
			"question1":    q1,
			"question2":    q2,
			"question3":    q3,
			"question4":    q4,
			"submitted_at": now,
			"updated_at":   now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	upsertResult, err := responsesCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to submit response: %v", err)
	}

	var saved GroupCompatibilityResponse
	if err = responsesCol.FindOne(ctx, filter).Decode(&saved); err != nil {
		_ = upsertResult
		return map[string]interface{}{"submitted": true}, nil
	}
	return saved, nil
}

func (p *GroupCompatibilityPlugin) getResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	filter := bson.M{"survey_id": surveyID}

	if sectionID, ok := request["section_id"].(string); ok && sectionID != "" {
		filter["section_id"] = sectionID
	}

	responsesCol := database.GetCollection("group_compatibility_responses")
	if responsesCol == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	cursor, err := responsesCol.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %v", err)
	}
	defer cursor.Close(ctx)

	var responses []GroupCompatibilityResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf("failed to decode responses: %v", err)
	}

	if responses == nil {
		responses = []GroupCompatibilityResponse{}
	}
	return responses, nil
}

func (p *GroupCompatibilityPlugin) getStudentResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	studentID, ok := request["student_id"].(string)
	if !ok || studentID == "" {
		return nil, fmt.Errorf("student_id is required")
	}

	responsesCol := database.GetCollection("group_compatibility_responses")
	if responsesCol == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	var response GroupCompatibilityResponse
	err = responsesCol.FindOne(ctx, bson.M{"survey_id": surveyID, "student_id": studentID}).Decode(&response)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get response: %v", err)
	}
	return response, nil
}

func (p *GroupCompatibilityPlugin) getSectionResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf("section_id is required")
	}

	surveyIDs, err := p.collectSurveyIDs(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if len(surveyIDs) == 0 {
		return map[string]interface{}{"success": true, "data": []interface{}{}}, nil
	}

	responsesCol := database.GetCollection("group_compatibility_responses")
	if responsesCol == nil {
		return nil, fmt.Errorf(errFailedGetRespCollection)
	}

	respCursor, err := responsesCol.Find(ctx, bson.M{"survey_id": bson.M{"$in": surveyIDs}})
	if err != nil {
		return nil, fmt.Errorf("failed to get responses: %v", err)
	}
	defer respCursor.Close(ctx)

	var sectionResponses []GroupCompatibilityResponse
	if err = respCursor.All(ctx, &sectionResponses); err != nil {
		return nil, fmt.Errorf("failed to decode responses: %v", err)
	}

	if sectionResponses == nil {
		sectionResponses = []GroupCompatibilityResponse{}
	}
	return sectionResponses, nil
}

func (p *GroupCompatibilityPlugin) collectSurveyIDs(ctx context.Context, sectionID string) ([]primitive.ObjectID, error) {
	cursor, err := p.collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, fmt.Errorf("failed to get surveys: %v", err)
	}
	defer cursor.Close(ctx)

	var surveys []bson.M
	if err = cursor.All(ctx, &surveys); err != nil {
		return nil, fmt.Errorf("failed to decode surveys: %v", err)
	}

	ids := make([]primitive.ObjectID, 0, len(surveys))
	for _, s := range surveys {
		if id, ok := s["_id"].(primitive.ObjectID); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
