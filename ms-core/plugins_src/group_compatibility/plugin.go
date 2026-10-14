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

type GroupCompatibilityPlugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *GroupCompatibilityPlugin) Name() string    { return p.name }
func (p *GroupCompatibilityPlugin) Version() string { return p.version }

func (p *GroupCompatibilityPlugin) Initialize() error {
	fmt.Printf("DEBUG: Initializing group_compatibility plugin\n")
	p.collection = database.GetCollection("group_compatibility_surveys")
	if p.collection == nil {
		return fmt.Errorf("failed to get group_compatibility_surveys collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "section_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	responsesCol := database.GetCollection("group_compatibility_responses")
	if responsesCol != nil {
		_, err = responsesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "survey_id", Value: 1}},
		})
		if err != nil {
			return err
		}
		_, err = responsesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "survey_id", Value: 1}, {Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		})
		if err != nil {
			fmt.Printf("DEBUG: Warning creating unique index on responses: %v\n", err)
		}
	}

	fmt.Printf("DEBUG: group_compatibility plugin initialized successfully\n")
	return nil
}

func (p *GroupCompatibilityPlugin) Shutdown() error {
	fmt.Printf("DEBUG: Shutting down group_compatibility plugin\n")
	return nil
}

func (p *GroupCompatibilityPlugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}

func (p *GroupCompatibilityPlugin) GetMetadataExtended() *core.PluginMetadataExtended {
	fullQuestions := make([]core.FullQuestion, len(defaultSociometricQuestions))
	for i, q := range defaultSociometricQuestions {
		fullQuestions[i] = core.FullQuestion{
			ID:       q.ID,
			Text:     q.Text,
			Options:  []core.QuestionOption{},
			Required: true,
		}
	}
	return &core.PluginMetadataExtended{
		PluginMetadata: *p.metadata,
		FullQuestions:  fullQuestions,
	}
}

func (p *GroupCompatibilityPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for group_compatibility plugin")
	}

	action, exists := request["action"]
	if !exists {
		return nil, fmt.Errorf("action is required")
	}

	switch action {
	case "create_survey", "create_test":
		return p.createSurvey(ctx, request)
	case "get_surveys", "get_tests":
		return p.getSurveys(ctx, request)
	case "get_survey", "get_test":
		return p.getSurvey(ctx, request)
	case "update_survey", "update_test":
		return p.updateSurvey(ctx, request)
	case "delete_survey", "delete_test":
		return p.deleteSurvey(ctx, request)
	case "submit_response":
		return p.submitResponse(ctx, request)
	case "get_responses":
		return p.getResponses(ctx, request)
	case "get_student_response":
		return p.getStudentResponse(ctx, request)
	case "get_section_responses":
		return p.getSectionResponses(ctx, request)
	case "get_metadata":
		return p.metadata, nil
	case "get_metadata_extended":
		return p.GetMetadataExtended(), nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
