package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (p *EnneagramPlugin) createTest(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	title, _ := req["title"].(string)
	description, _ := req["description"].(string)
	sectionID, _ := req["section_id"].(string)
	createdBy, _ := req["created_by"].(string)

	if title == "" || sectionID == "" {
		return nil, fmt.Errorf("enneagram plugin: title and section_id are required")
	}

	questions := p.parseQuestions(req["questions"])

	now := time.Now()
	test := EnneagramTest{
		Title:       title,
		Description: description,
		SectionID:   sectionID,
		CreatedBy:   createdBy,
		Questions:   questions,
		IsActive:    true,
		StartDate:   parseOptionalDate(req, "start_date"),
		EndDate:     parseOptionalDate(req, "end_date"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := p.collection.InsertOne(ctx, test)
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to create test: %w", err)
	}

	test.ID = result.InsertedID.(primitive.ObjectID)
	return test, nil
}

func (p *EnneagramPlugin) getTests(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	sectionID, _ := req["section_id"].(string)

	filter := bson.M{}
	if sectionID != "" {
		filter["section_id"] = sectionID
	}

	cursor, err := p.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to query tests: %w", err)
	}
	defer cursor.Close(ctx)

	var tests []EnneagramTest
	if err := cursor.All(ctx, &tests); err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to decode tests: %w", err)
	}
	if tests == nil {
		tests = []EnneagramTest{}
	}
	return tests, nil
}

func (p *EnneagramPlugin) getTest(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	testIDStr, _ := req["test_id"].(string)
	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt, err)
	}

	var test EnneagramTest
	if err := p.collection.FindOne(ctx, bson.M{"_id": testID}).Decode(&test); err != nil {
		return nil, fmt.Errorf("enneagram plugin: test not found: %w", err)
	}
	return test, nil
}

func (p *EnneagramPlugin) updateTest(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	testIDStr, _ := req["test_id"].(string)
	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt, err)
	}

	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if title, ok := req["title"].(string); ok && title != "" {
		update["$set"].(bson.M)["title"] = title
	}
	if desc, ok := req["description"].(string); ok {
		update["$set"].(bson.M)["description"] = desc
	}
	if isActive, ok := req["is_active"].(bool); ok {
		update["$set"].(bson.M)["is_active"] = isActive
	}
	if sd := parseOptionalDate(req, "start_date"); sd != nil {
		update["$set"].(bson.M)["start_date"] = sd
	}
	if ed := parseOptionalDate(req, "end_date"); ed != nil {
		update["$set"].(bson.M)["end_date"] = ed
	}

	if _, err := p.collection.UpdateOne(ctx, bson.M{"_id": testID}, update); err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to update test: %w", err)
	}

	return p.getTest(ctx, req)
}

func (p *EnneagramPlugin) deleteTest(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	testIDStr, _ := req["test_id"].(string)
	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt, err)
	}

	result, err := p.collection.DeleteOne(ctx, bson.M{"_id": testID})
	if err != nil {
		return nil, fmt.Errorf("enneagram plugin: failed to delete test: %w", err)
	}

	return map[string]interface{}{
		"deleted": result.DeletedCount > 0,
	}, nil
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

func (p *EnneagramPlugin) parseQuestions(raw interface{}) []EnneagramQuestion {
	if raw == nil {
		return nil
	}
	slice, ok := raw.([]interface{})
	if !ok {
		return nil
	}

	questions := make([]EnneagramQuestion, 0, len(slice))
	for _, item := range slice {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		q := EnneagramQuestion{}
		if id, ok := m["id"].(float64); ok {
			q.ID = int(id)
		}
		if text, ok := m["text"].(string); ok {
			q.Text = text
		}
		if et, ok := m["eneatype"].(float64); ok {
			q.Eneatype = int(et)
		}
		if req, ok := m["required"].(bool); ok {
			q.Required = req
		}
		questions = append(questions, q)
	}
	return questions
}

func (p *EnneagramPlugin) parseAnswers(raw interface{}) []EnneagramAnswer {
	if raw == nil {
		return nil
	}
	slice, ok := raw.([]interface{})
	if !ok {
		return nil
	}

	answers := make([]EnneagramAnswer, 0, len(slice))
	for _, item := range slice {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		a := EnneagramAnswer{}
		switch v := m["question_id"].(type) {
		case float64:
			a.QuestionID = int(v)
		case int:
			a.QuestionID = v
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				a.QuestionID = n
			}
		}
		switch v := m["score"].(type) {
		case float64:
			a.Score = int(v)
		case int:
			a.Score = v
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				a.Score = n
			}
		}
		answers = append(answers, a)
	}
	return answers
}

func calculateResults(answers []EnneagramAnswer) EnneagramResults {
	scores := [10]float64{}

	for _, a := range answers {
		if a.QuestionID >= 1 && a.QuestionID <= 9 {
			scores[a.QuestionID] += float64(a.Score)
		}
	}

	dominant := 1
	maxScore := scores[1]
	for i := 2; i <= 9; i++ {
		if scores[i] > maxScore {
			maxScore = scores[i]
			dominant = i
		}
	}

	var oceanVec map[string]float64
	if v, ok := getOCEANFromEnneagram(dominant); ok {
		oceanVec = map[string]float64{
			"openness":          v.Openness,
			"conscientiousness": v.Conscientiousness,
			"extraversion":      v.Extraversion,
			"agreeableness":     v.Agreeableness,
			"neuroticism":       v.Neuroticism,
		}
	}

	return EnneagramResults{
		E1:           scores[1],
		E2:           scores[2],
		E3:           scores[3],
		E4:           scores[4],
		E5:           scores[5],
		E6:           scores[6],
		E7:           scores[7],
		E8:           scores[8],
		E9:           scores[9],
		DominantType: dominant,
		MaxScore:     maxScore,
		OCEANVector:  oceanVec,
	}
}
