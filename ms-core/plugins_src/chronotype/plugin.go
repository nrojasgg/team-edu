package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"ms-core/internal/core"
	"ms-core/internal/database"
)

type ChronotypePlugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *ChronotypePlugin) Name() string {
	return p.name
}

func (p *ChronotypePlugin) Version() string {
	return p.version
}

func (p *ChronotypePlugin) Initialize() error {
	log.Printf("Initializing Chronotype plugin")
	p.collection = database.GetCollection("chronotype_tests")

	if p.collection == nil {
		return fmt.Errorf("failed to get collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "section_id", Value: 1}},
	})

	if err != nil {
		log.Printf("Error creating index: %v", err)
	}

	return err
}

func (p *ChronotypePlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for Chronotype plugin")
	}

	action, exists := request["action"]
	if !exists {
		return nil, fmt.Errorf("action is required")
	}

	switch action {
	case "create_test":
		return p.createTest(ctx, request)
	case "get_tests":
		return p.getTests(ctx, request)
	case "get_test":
		return p.getTest(ctx, request)
	case "update_test":
		return p.updateTest(ctx, request)
	case "delete_test":
		return p.deleteTest(ctx, request)
	case "submit_response":
		return p.submitResponse(ctx, request)
	case "get_responses":
		return p.getResponses(ctx, request)
	case "get_section_responses":
		return p.getSectionResponses(ctx, request)
	case "get_results":
		return p.getResults(ctx, request)
	case "get_student_chronotype":
		return p.getStudentResponse(ctx, request)
	case "get_metadata":
		return p.metadata, nil
	case "get_metadata_extended":
		return p.GetMetadataExtended(), nil
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}
}

func (p *ChronotypePlugin) Shutdown() error {
	log.Printf("Shutting down Chronotype plugin")
	return nil
}

func (p *ChronotypePlugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}

func (p *ChronotypePlugin) GetMetadataExtended() *core.PluginMetadataExtended {
	questions := getChronotypeQuestionsFull()
	fullQuestions := make([]core.FullQuestion, len(questions))
	for i, q := range questions {
		options := make([]core.QuestionOption, len(q.Options))
		for j, opt := range q.Options {
			options[j] = core.QuestionOption{Score: opt.Score, Text: opt.Text}
		}
		fullQuestions[i] = core.FullQuestion{
			ID:       q.ID,
			Text:     q.Text,
			Options:  options,
			Required: q.Required,
		}
	}
	return &core.PluginMetadataExtended{
		PluginMetadata: *p.metadata,
		FullQuestions:  fullQuestions,
	}
}
