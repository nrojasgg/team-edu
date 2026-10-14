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

type BFI44Plugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *BFI44Plugin) Name() string {
	return p.name
}

func (p *BFI44Plugin) Version() string {
	return p.version
}

func (p *BFI44Plugin) Initialize() error {
	log.Printf("Initializing BFI-44 plugin")
	p.collection = database.GetCollection("bfi44_tests")

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

func (p *BFI44Plugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for BFI-44 plugin")
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
	case "get_student_bfi44":
		return p.getStudentResponse(ctx, request)
	case "get_student_ocean":
		return p.getStudentOcean(ctx, request)
	case "get_metadata":
		return p.metadata, nil
	case "get_metadata_extended":
		return p.GetMetadataExtended(), nil
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}
}

func (p *BFI44Plugin) Shutdown() error {
	log.Printf("Shutting down BFI-44 plugin")
	return nil
}

func (p *BFI44Plugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}

func (p *BFI44Plugin) GetMetadataExtended() *core.PluginMetadataExtended {
	questions := getBFI44Questions()
	likertOptions := []core.QuestionOption{
		{Score: 1, Text: "Totalmente en desacuerdo"},
		{Score: 2, Text: "En desacuerdo"},
		{Score: 3, Text: "Neutral"},
		{Score: 4, Text: "De acuerdo"},
		{Score: 5, Text: "Totalmente de acuerdo"},
	}
	fullQuestions := make([]core.FullQuestion, len(questions))
	for i, q := range questions {
		fullQuestions[i] = core.FullQuestion{
			ID:       q.ID,
			Text:     q.Text,
			Options:  likertOptions,
			Required: q.Required,
		}
	}
	return &core.PluginMetadataExtended{
		PluginMetadata: *p.metadata,
		FullQuestions:  fullQuestions,
	}
}
