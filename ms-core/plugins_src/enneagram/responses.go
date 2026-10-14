package main

import (
	"context"
	"fmt"
	"time"

	"ms-core/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (p *EnneagramPlugin) submitResponse(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	testIDStr, _ := req["test_id"].(string)
	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt, err)
	}

	studentID, _ := req["student_id"].(string)
	if studentID == "" {
		return nil, fmt.Errorf("enneagram plugin: student_id is required")
	}

	answers := p.parseAnswers(req["answers"])
	results := calculateResults(answers)
	now := time.Now()

	response := EnneagramResponse{
		TestID:      testID,
		StudentID:   studentID,
		Answers:     answers,
		Results:     &results,
		IsComplete:  true,
		SubmittedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	responsesCol := database.GetCollection("enneagram_responses")
	result2, err := responsesCol.InsertOne(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to submit response: %w", err)
	}

	response.ID = result2.InsertedID.(primitive.ObjectID)
	return response, nil
}

func (p *EnneagramPlugin) getResponses(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	testIDStr, _ := req["test_id"].(string)
	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt, err)
	}

	studentID, _ := req["student_id"].(string)

	filter := bson.M{"test_id": testID}
	if studentID != "" {
		filter["student_id"] = studentID
	}

	responsesCol := database.GetCollection("enneagram_responses")
	cursor, err := responsesCol.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to query responses: %w", err)
	}
	defer cursor.Close(ctx)

	var responses []EnneagramResponse
	if err := cursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to decode responses: %w", err)
	}
	if responses == nil {
		responses = []EnneagramResponse{}
	}

	for i := range responses {
		r := &responses[i]
		if r.Results != nil && r.Results.OCEANVector == nil {
			if v, ok := getOCEANFromEnneagram(r.Results.DominantType); ok {
				r.Results.OCEANVector = map[string]float64{
					"openness":          v.Openness,
					"conscientiousness": v.Conscientiousness,
					"extraversion":      v.Extraversion,
					"agreeableness":     v.Agreeableness,
					"neuroticism":       v.Neuroticism,
				}
			}
		}
	}

	return responses, nil
}

func (p *EnneagramPlugin) getResults(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	studentID, _ := req["student_id"].(string)
	testIDStr, _ := req["test_id"].(string)

	filter := bson.M{"student_id": studentID, "is_complete": true}
	if testIDStr != "" {
		testID, err := primitive.ObjectIDFromHex(testIDStr)
		if err == nil {
			filter["test_id"] = testID
		}
	}

	responsesCol := database.GetCollection("enneagram_responses")
	var response EnneagramResponse
	if err := responsesCol.FindOne(ctx, filter).Decode(&response); err != nil {
		return nil, fmt.Errorf("enneagram plugin: results not found for student %s: %w", studentID, err)
	}
	return response.Results, nil
}

func (p *EnneagramPlugin) getSectionResponses(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	sectionID, _ := req["section_id"].(string)
	if sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	cursor, err := p.collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to query tests: %w", err)
	}
	defer cursor.Close(ctx)

	var tests []EnneagramTest
	if err := cursor.All(ctx, &tests); err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to decode tests: %w", err)
	}

	if len(tests) == 0 {
		return []EnneagramResponse{}, nil
	}

	testIDs := make([]primitive.ObjectID, len(tests))
	for i, t := range tests {
		testIDs[i] = t.ID
	}

	responsesCol := database.GetCollection("enneagram_responses")
	findOpts := options.Find().
		SetProjection(bson.M{"answers": 0}).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	resCursor, err := responsesCol.Find(ctx, bson.M{"test_id": bson.M{"$in": testIDs}}, findOpts)
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to query responses: %w", err)
	}
	defer resCursor.Close(ctx)

	var responses []EnneagramResponse
	if err := resCursor.All(ctx, &responses); err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to decode section responses: %w", err)
	}
	if responses == nil {
		responses = []EnneagramResponse{}
	}

	for i := range responses {
		r := &responses[i]
		if r.Results == nil {
			continue
		}
		if r.Results.OCEANVector != nil {
			continue
		}
		if v, ok := getOCEANFromEnneagram(r.Results.DominantType); ok {
			r.Results.OCEANVector = map[string]float64{
				"openness":          v.Openness,
				"conscientiousness": v.Conscientiousness,
				"extraversion":      v.Extraversion,
				"agreeableness":     v.Agreeableness,
				"neuroticism":       v.Neuroticism,
			}
		}
	}

	// Deduplicate by student_id, keeping the newest response
	seen := make(map[string]bool)
	deduped := make([]EnneagramResponse, 0, len(responses))
	for _, r := range responses {
		if !seen[r.StudentID] {
			seen[r.StudentID] = true
			deduped = append(deduped, r)
		}
	}

	return deduped, nil
}

func (p *EnneagramPlugin) getSectionResults(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	sectionID, _ := req["section_id"].(string)
	if sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	responsesRaw, err := p.getSectionResponses(ctx, req)
	if err != nil {
		return nil, err
	}

	responses, _ := responsesRaw.([]EnneagramResponse)

	results := make(map[string]*EnneagramResults, len(responses))
	for _, r := range responses {
		if r.IsComplete && r.Results != nil {
			results[r.StudentID] = r.Results
		}
	}
	return results, nil
}

func (p *EnneagramPlugin) getStudentEneatypes(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	sectionID, _ := req["section_id"].(string)
	if sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	responsesRaw, err := p.getSectionResponses(ctx, req)
	if err != nil {
		return nil, err
	}

	responses, _ := responsesRaw.([]EnneagramResponse)

	eneatypes := make(map[string]int, len(responses))
	for _, r := range responses {
		if r.IsComplete && r.Results != nil {
			eneatypes[r.StudentID] = r.Results.DominantType
		}
	}
	return eneatypes, nil
}

func (p *EnneagramPlugin) calculateGroupCompatibility(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	sectionID, _ := req["section_id"].(string)
	if sectionID == "" {
		return nil, fmt.Errorf(errSectionIDRequired)
	}

	eneatypesRaw, err := p.getStudentEneatypes(ctx, req)
	if err != nil {
		return nil, err
	}

	eneatypes, _ := eneatypesRaw.(map[string]int)
	if len(eneatypes) < 2 {
		return map[string]interface{}{
			"compatible_pairs": 0,
			"total_pairs":      0,
			"group_score":      0.0,
		}, nil
	}

	studentIDs := make([]string, 0, len(eneatypes))
	for sid := range eneatypes {
		studentIDs = append(studentIDs, sid)
	}

	totalPairs := 0
	compatiblePairs := 0

	for i := 0; i < len(studentIDs); i++ {
		for j := i + 1; j < len(studentIDs); j++ {
			e1 := eneatypes[studentIDs[i]]
			e2 := eneatypes[studentIDs[j]]
			totalPairs++
			if getCompatibility(e1, e2) == 1 {
				compatiblePairs++
			}
		}
	}

	groupScore := 0.0
	if totalPairs > 0 {
		groupScore = float64(compatiblePairs) / float64(totalPairs)
	}

	return map[string]interface{}{
		"compatible_pairs": compatiblePairs,
		"total_pairs":      totalPairs,
		"group_score":      groupScore,
	}, nil
}
