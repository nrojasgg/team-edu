package main

import (
	"context"
	"fmt"
	"time"

	"ms-core/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func (p *GroupCompatibilityPlugin) createSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	title, ok := request["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title is required and must be a string")
	}

	description, _ := request["description"].(string)

	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf("section_id is required and must be a string")
	}

	createdBy, ok := request["created_by"].(string)
	if !ok || createdBy == "" {
		return nil, fmt.Errorf("created_by is required and must be a string")
	}

	surveyType, _ := request["survey_type"].(string)
	if surveyType == "" {
		surveyType = "full_section"
	}

	survey := GroupCompatibilitySurvey{
		Title:       title,
		Description: description,
		SectionID:   sectionID,
		CreatedBy:   createdBy,
		SurveyType:  surveyType,
		Questions:   defaultSociometricQuestions,
		IsActive:    true,
		StartDate:   parseOptionalDate(request, "start_date"),
		EndDate:     parseOptionalDate(request, "end_date"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	insertResult, err := p.collection.InsertOne(ctx, survey)
	if err != nil {
		return nil, fmt.Errorf("failed to create survey: %v", err)
	}

	survey.ID = insertResult.InsertedID.(primitive.ObjectID)
	return survey, nil
}

func (p *GroupCompatibilityPlugin) getSurveys(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf("section_id is required")
	}

	cursor, err := p.collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, fmt.Errorf("failed to get surveys: %v", err)
	}
	defer cursor.Close(ctx)

	var surveys []GroupCompatibilitySurvey
	if err = cursor.All(ctx, &surveys); err != nil {
		return nil, fmt.Errorf("failed to decode surveys: %v", err)
	}

	if surveys == nil {
		surveys = []GroupCompatibilitySurvey{}
	}
	return surveys, nil
}

func (p *GroupCompatibilityPlugin) getSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr, ok := request["survey_id"].(string)
	if !ok || surveyIDStr == "" {
		return nil, fmt.Errorf(errSurveyIDRequired)
	}

	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidSurveyIDFmt, err)
	}

	var survey GroupCompatibilitySurvey
	if err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey); err != nil {
		return nil, fmt.Errorf("survey not found: %v", err)
	}
	return survey, nil
}

func (p *GroupCompatibilityPlugin) updateSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
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

	result, err := p.collection.UpdateOne(ctx, bson.M{"_id": surveyID}, bson.M{"$set": updateData})
	if err != nil {
		return nil, fmt.Errorf("failed to update survey: %v", err)
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("survey not found")
	}

	var updated GroupCompatibilitySurvey
	if err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&updated); err != nil {
		return nil, fmt.Errorf("failed to fetch updated survey: %v", err)
	}
	return updated, nil
}

func (p *GroupCompatibilityPlugin) deleteSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
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

	responsesCol := database.GetCollection("group_compatibility_responses")
	if responsesCol != nil {
		_, err = responsesCol.DeleteMany(ctx, bson.M{"survey_id": surveyID})
		if err != nil {
			fmt.Printf("DEBUG: Warning deleting associated responses: %v\n", err)
		}
	}

	return map[string]interface{}{"deleted": true, "message": "Survey deleted successfully"}, nil
}
