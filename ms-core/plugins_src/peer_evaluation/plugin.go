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

type PeerEvaluationPlugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *PeerEvaluationPlugin) Name() string {
	return p.name
}

func (p *PeerEvaluationPlugin) Version() string {
	return p.version
}

func (p *PeerEvaluationPlugin) Initialize() error {
	fmt.Printf("DEBUG: Initializing peer evaluation plugin\n")
	p.collection = database.GetCollection("peer_evaluation_surveys")

	if p.collection == nil {
		fmt.Printf("DEBUG: Failed to get collection\n")
		return fmt.Errorf("failed to get collection")
	}

	fmt.Printf("DEBUG: Got collection successfully\n")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "section_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	responsesCollection := database.GetCollection("peer_evaluation_responses")
	if responsesCollection != nil {
		_, err = responsesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{{Key: "survey_id", Value: 1}},
		})
		if err != nil {
			return err
		}

		_, err = responsesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "survey_id", Value: 1},
				{Key: "evaluator_id", Value: 1},
				{Key: "evaluated_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		})
	}

	return err
}

func (p *PeerEvaluationPlugin) Shutdown() error {
	fmt.Printf("DEBUG: Shutting down peer evaluation plugin\n")
	return nil
}

func (p *PeerEvaluationPlugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}

func (p *PeerEvaluationPlugin) GetMetadataExtended() *core.PluginMetadataExtended {
	likertOptions := []core.QuestionOption{
		{Score: 1, Text: "Totalmente en desacuerdo"},
		{Score: 2, Text: "En desacuerdo"},
		{Score: 3, Text: "Neutral"},
		{Score: 4, Text: "De acuerdo"},
		{Score: 5, Text: "Totalmente de acuerdo"},
	}
	fullQuestions := make([]core.FullQuestion, 0, len(p.metadata.Preguntas))
	for id, text := range p.metadata.Preguntas {
		var qID int
		fmt.Sscanf(id, "%d", &qID)
		fullQuestions = append(fullQuestions, core.FullQuestion{
			ID:       qID,
			Text:     text,
			Options:  likertOptions,
			Required: true,
		})
	}
	return &core.PluginMetadataExtended{
		PluginMetadata: *p.metadata,
		FullQuestions:  fullQuestions,
	}
}

func (p *PeerEvaluationPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for peer evaluation plugin")
	}

	action, exists := request["action"]
	if !exists {
		return nil, fmt.Errorf("action is required")
	}

	switch action {
	case "create_survey":
		return p.createSurvey(ctx, request)
	case "get_surveys":
		return p.getSurveys(ctx, request)
	case "get_survey":
		return p.getSurvey(ctx, request)
	case "get_survey_peers":
		return p.getSurveyPeers(ctx, request)
	case "update_survey":
		return p.updateSurvey(ctx, request)
	case "delete_survey":
		return p.deleteSurvey(ctx, request)
	case "submit_response":
		return p.submitResponse(ctx, request)
	case "get_responses":
		return p.getResponses(ctx, request)
	case "get_results":
		return p.getResults(ctx, request)
	case "get_section_responses":
		return p.getSectionResponses(ctx, request)
	case "get_section_responses_summary":
		return p.getSectionResponsesSummary(ctx, request)
	case "get_section_group_responses":
		return p.getSectionGroupResponses(ctx, request)
	case "export_to_excel":
		return p.exportToExcel(ctx, request)
	case "get_metadata":
		return p.metadata, nil
	case "get_metadata_extended":
		return p.GetMetadataExtended(), nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
