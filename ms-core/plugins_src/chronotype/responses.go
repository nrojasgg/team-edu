package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"ms-core/internal/database"
)

func (p *ChronotypePlugin) submitResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("chronotype_responses")

	testIDStr, ok := request["test_id"].(string)
	if !ok {
		return nil, fmt.Errorf("test_id is required")
	}

	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid test_id: %v", err)
	}

	var answers []ChronotypeAnswer
	if answersData, exists := request["answers"]; exists {
		answersList := answersData.([]interface{})
		for _, answerData := range answersList {
			answerMap := answerData.(map[string]interface{})
			answers = append(answers, ChronotypeAnswer{
				QuestionID: int(answerMap["question_id"].(float64)),
				Score:      int(answerMap["score"].(float64)),
			})
		}
	} else {
		return nil, fmt.Errorf("answers are required")
	}

	inputMap := make(map[int]int)
	for _, ans := range answers {
		inputMap[ans.QuestionID] = ans.Score
	}

	results, err := p.Calculate(inputMap)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate results: %v", err)
	}

	studentID := ""
	if sid, exists := request["student_id"]; exists {
		if studentIDStr, ok := sid.(string); ok {
			studentID = studentIDStr
		}
	}

	response := ChronotypeResponse{
		TestID:     testID,
		StudentID:  studentID,
		Answers:    answers,
		Results:    results,
		IsComplete: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	now := time.Now()
	response.SubmittedAt = &now

	result, err := responsesCollection.InsertOne(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("failed to save response: %v", err)
	}

	response.ID = result.InsertedID.(primitive.ObjectID)
	return response, nil
}

func (p *ChronotypePlugin) getResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("chronotype_responses")

	testIDStr := request["test_id"].(string)
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

	var responses []ChronotypeResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

func (p *ChronotypePlugin) getStudentResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("chronotype_responses")
	testsCollection := database.GetCollection("chronotype_tests")

	sectionID := request["section_id"].(string)
	studentID := request["student_id"].(string)

	testsCursor, err := testsCollection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, err
	}
	defer testsCursor.Close(ctx)

	var tests []ChronotypeTest
	if err = testsCursor.All(ctx, &tests); err != nil {
		return nil, err
	}

	testIDs := make([]primitive.ObjectID, len(tests))
	for i, test := range tests {
		testIDs[i] = test.ID
	}

	if len(testIDs) == 0 {
		return nil, nil
	}

	var response ChronotypeResponse
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

func (p *ChronotypePlugin) getSectionResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("chronotype_responses")
	testsCollection := database.GetCollection("chronotype_tests")

	sectionID := request["section_id"].(string)

	testsCursor, err := testsCollection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, err
	}
	defer testsCursor.Close(ctx)

	var tests []ChronotypeTest
	if err = testsCursor.All(ctx, &tests); err != nil {
		return nil, err
	}

	testIDs := make([]primitive.ObjectID, len(tests))
	for i, test := range tests {
		testIDs[i] = test.ID
	}

	if len(testIDs) == 0 {
		return []ChronotypeResponse{}, nil
	}

	cursor, err := responsesCollection.Find(ctx, bson.M{
		"test_id":     bson.M{"$in": testIDs},
		"is_complete": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []ChronotypeResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

func (p *ChronotypePlugin) getResults(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("chronotype_responses")

	testIDStr := request["test_id"].(string)
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

	var responses []ChronotypeResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}
