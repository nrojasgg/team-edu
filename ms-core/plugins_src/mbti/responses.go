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

func parseAnswersFromRequest(request map[string]interface{}) ([]MBTIAnswer, error) {
	answersData, exists := request["answers"]
	if !exists {
		return nil, fmt.Errorf("answers are required")
	}
	answersList, ok := answersData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("answers must be an array")
	}
	answers := make([]MBTIAnswer, 0, len(answersList))
	for _, answerData := range answersList {
		answerMap, ok := answerData.(map[string]interface{})
		if !ok {
			continue
		}
		var questionID int
		if qid, ok := answerMap["question_id"].(float64); ok {
			questionID = int(qid)
		} else if qid, ok := answerMap["question_id"].(int); ok {
			questionID = qid
		}
		selection, _ := answerMap["selection"].(string)
		answers = append(answers, MBTIAnswer{QuestionID: questionID, Selection: selection})
	}
	return answers, nil
}

func (p *MBTIPlugin) submitResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("mbti_responses")

	testIDStr, ok := request["test_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errTestIDRequired)
	}

	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid test_id: %v", err)
	}

	answers, err := parseAnswersFromRequest(request)
	if err != nil {
		return nil, err
	}

	results := CalculateMBTI(answers)

	studentID := ""
	if sid, exists := request["student_id"]; exists {
		if s, ok := sid.(string); ok {
			studentID = s
		}
	}

	now := time.Now()
	response := MBTIResponse{
		TestID:      testID,
		StudentID:   studentID,
		Answers:     answers,
		Results:     results,
		IsComplete:  true,
		SubmittedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := responsesCollection.InsertOne(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("failed to save response: %v", err)
	}

	response.ID = result.InsertedID.(primitive.ObjectID)
	return response, nil
}

func (p *MBTIPlugin) getResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("mbti_responses")

	testIDStr, ok := request["test_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errTestIDRequired)
	}

	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"test_id": testID}

	if studentID, exists := request["student_id"]; exists {
		filter["student_id"] = studentID
	}

	cursor, err := responsesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []MBTIResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

func (p *MBTIPlugin) fetchSectionTestIDs(ctx context.Context, sectionID string) ([]primitive.ObjectID, error) {
	cursor, err := p.collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tests []MBTITest
	if err = cursor.All(ctx, &tests); err != nil {
		return nil, err
	}

	ids := make([]primitive.ObjectID, len(tests))
	for i, test := range tests {
		ids[i] = test.ID
	}
	return ids, nil
}

func (p *MBTIPlugin) getSectionResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("mbti_responses")

	sectionID, ok := request["section_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	testIDs, err := p.fetchSectionTestIDs(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if len(testIDs) == 0 {
		return []MBTIResponse{}, nil
	}

	findOpts := options.Find().
		SetProjection(bson.M{"answers": 0}).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := responsesCollection.Find(ctx, bson.M{
		"test_id":     bson.M{"$in": testIDs},
		"is_complete": true,
	}, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []MBTIResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	deduped := make([]MBTIResponse, 0, len(responses))
	for _, r := range responses {
		if !seen[r.StudentID] {
			seen[r.StudentID] = true
			deduped = append(deduped, r)
		}
	}

	return deduped, nil
}

func (p *MBTIPlugin) getResults(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("mbti_responses")

	testIDStr, ok := request["test_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errTestIDRequired)
	}

	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, err
	}

	cursor, err := responsesCollection.Find(ctx, bson.M{
		"test_id":     testID,
		"is_complete": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []MBTIResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

func (p *MBTIPlugin) getStudentMBTI(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("mbti_responses")

	sectionID, ok := request["section_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	studentID, ok := request["student_id"].(string)
	if !ok {
		return nil, fmt.Errorf("student_id is required")
	}

	testIDs, err := p.fetchSectionTestIDs(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	if len(testIDs) == 0 {
		return nil, nil
	}

	var response MBTIResponse
	err = responsesCollection.FindOne(ctx, bson.M{
		"test_id":     bson.M{"$in": testIDs},
		"student_id":  studentID,
		"is_complete": true,
	}).Decode(&response)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (p *MBTIPlugin) calculateFromRequest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	answers, err := parseAnswersFromRequest(request)
	if err != nil {
		return nil, err
	}
	return CalculateMBTI(answers), nil
}
