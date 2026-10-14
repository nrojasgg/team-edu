package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	errTestIDRequired   = "test_id is required"
	errInvalidTestIDFmt = "invalid test_id format"
)

func (p *BFI44Plugin) createTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
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

	test := BFI44Test{
		Title:       title,
		Description: description,
		SectionID:   sectionID,
		CreatedBy:   createdBy,
		Questions:   getBFI44Questions(),
		IsActive:    true,
		StartDate:   parseOptionalDate(request, "start_date"),
		EndDate:     parseOptionalDate(request, "end_date"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := p.collection.InsertOne(ctx, test)
	if err != nil {
		return nil, err
	}

	test.ID = result.InsertedID.(primitive.ObjectID)
	return test, nil
}

func (p *BFI44Plugin) getTests(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID, ok := request["section_id"].(string)
	if !ok || sectionID == "" {
		return nil, fmt.Errorf("section_id is required")
	}

	cursor, err := p.collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tests []BFI44Test
	if err = cursor.All(ctx, &tests); err != nil {
		return nil, err
	}

	return tests, nil
}

func (p *BFI44Plugin) getTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	testIDStr := request["test_id"].(string)
	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, err
	}

	var test BFI44Test
	err = p.collection.FindOne(ctx, bson.M{"_id": testID}).Decode(&test)
	if err != nil {
		return nil, err
	}

	return test, nil
}

func (p *BFI44Plugin) updateTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	testIDStr, ok := request["test_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errTestIDRequired)
	}

	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt)
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if title, exists := request["title"].(string); exists {
		update["$set"].(bson.M)["title"] = title
	}
	if description, exists := request["description"].(string); exists {
		update["$set"].(bson.M)["description"] = description
	}
	if isActive, exists := request["is_active"].(bool); exists {
		update["$set"].(bson.M)["is_active"] = isActive
	}

	result, err := p.collection.UpdateOne(ctx, bson.M{"_id": testID}, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *BFI44Plugin) deleteTest(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	testIDStr, ok := request["test_id"].(string)
	if !ok {
		return nil, fmt.Errorf(errTestIDRequired)
	}

	testID, err := primitive.ObjectIDFromHex(testIDStr)
	if err != nil {
		return nil, fmt.Errorf(errInvalidTestIDFmt)
	}

	result, err := p.collection.DeleteOne(ctx, bson.M{"_id": testID})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseOptionalDate(request map[string]interface{}, key string) *time.Time {
	str, ok := request[key].(string)
	if !ok || str == "" {
		return nil
	}
	for _, format := range []string{time.RFC3339, "2006-01-02T15:04:05"} {
		if t, err := time.Parse(format, str); err == nil {
			return &t
		}
	}
	return nil
}
