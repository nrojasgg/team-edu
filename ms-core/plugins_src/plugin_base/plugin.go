package main

import (
	"context"
	"fmt"
	"time"

	"ms-core/internal/core"
	"ms-core/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MyPlugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *MyPlugin) Name() string { return p.name }

func (p *MyPlugin) Version() string { return p.version }

func (p *MyPlugin) Initialize() error {
	fmt.Printf("DEBUG: Initializing %s plugin\n", p.name)

	p.collection = database.GetCollection("my_plugin_tests")
	if p.collection == nil {
		return fmt.Errorf("failed to get my_plugin_tests collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "section_id", Value: 1}},
	})
	if err != nil {
		return fmt.Errorf("failed to create section_id index: %w", err)
	}

	responsesCol := database.GetCollection("my_plugin_responses")
	if responsesCol != nil {
		_, err = responsesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "test_id", Value: 1}, {Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		})
		if err != nil {
			fmt.Printf("DEBUG: Warning creating unique index: %v\n", err)
		}
	}

	fmt.Printf("DEBUG: %s plugin initialized successfully\n", p.name)
	return nil
}

func (p *MyPlugin) Shutdown() error {
	fmt.Printf("DEBUG: Shutting down %s plugin\n", p.name)
	return nil
}

func (p *MyPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for %s plugin", p.name)
	}

	action, exists := request["action"]
	if !exists {
		return nil, fmt.Errorf("action is required")
	}

	switch action {
	case "get_metadata":
		return p.metadata, nil

	case "create_survey", "create_test":
		return p.createTest(ctx, request)
	case "get_surveys", "get_tests":
		return p.getTests(ctx, request)
	case "get_survey", "get_test":
		return p.getTest(ctx, request)
	case "update_survey", "update_test":
		return p.updateTest(ctx, request)
	case "delete_survey", "delete_test":
		return p.deleteTest(ctx, request)

	case "submit_response":
		return p.submitResponse(ctx, request)
	case "get_responses":
		return p.getResponses(ctx, request)
	case "get_student_response":
		return p.getStudentResponse(ctx, request)
	case "get_section_responses":
		return p.getSectionResponses(ctx, request)

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (p *MyPlugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}
