package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ms-core/internal/core"
	"ms-core/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	errTestIDRequired    = "test_id is required"
	errInvalidTestIDFmt  = "invalid test_id format"
	errSectionIDRequired = "section_id is required"
)

type MBTIPlugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *MBTIPlugin) Name() string {
	return p.name
}

func (p *MBTIPlugin) Version() string {
	return p.version
}

func (p *MBTIPlugin) Initialize() error {
	log.Printf("Initializing MBTI (Francis) plugin")
	p.collection = database.GetCollection("mbti_tests")

	if p.collection == nil {
		return fmt.Errorf("failed to get mbti_tests collection")
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

func (p *MBTIPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for MBTI plugin")
	}

	action, exists := request["action"]
	if !exists {
		action, exists = request["operation"]
		if !exists {
			return nil, fmt.Errorf("action is required")
		}
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
	case "get_student_mbti":
		return p.getStudentMBTI(ctx, request)
	case "calculate":
		return p.calculateFromRequest(ctx, request)
	case "get_compatibility":
		typeA, _ := request["type_a"].(string)
		typeB, _ := request["type_b"].(string)
		return p.GetMBTICompatibility(typeA, typeB), nil
	case "get_ocean_from_type":
		mbtiType, _ := request["mbti_type"].(string)
		if v, found := getOCEANFromMBTI(mbtiType); found {
			return v, nil
		}
		return nil, fmt.Errorf("mbti plugin: unknown MBTI type %q", mbtiType)
	case "get_metadata":
		return p.metadata, nil
	case "get_metadata_extended":
		return p.GetMetadataExtended(), nil
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}
}

func (p *MBTIPlugin) Shutdown() error {
	log.Printf("Shutting down MBTI (Francis) plugin")
	return nil
}

func (p *MBTIPlugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}

func (p *MBTIPlugin) GetMetadataExtended() *core.PluginMetadataExtended {
	questions := GetMBTIQuestions()
	fullQuestions := make([]core.FullQuestion, len(questions))
	for i, q := range questions {
		options := []core.QuestionOption{
			{Score: 0, Text: q.OptionA},
			{Score: 1, Text: q.OptionB},
		}
		fullQuestions[i] = core.FullQuestion{
			ID:       q.ID,
			Text:     q.Statement,
			Options:  options,
			Required: q.Required,
		}
	}
	return &core.PluginMetadataExtended{
		PluginMetadata: *p.metadata,
		FullQuestions:  fullQuestions,
	}
}
