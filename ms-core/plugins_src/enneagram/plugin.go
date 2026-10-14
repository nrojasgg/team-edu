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
	errInvalidTestIDFmt  = "enneagram plugin: invalid test_id: %w"
	errSectionIDRequired = "enneagram plugin: section_id is required"
)

type EnneagramPlugin struct {
	name       string
	version    string
	collection *mongo.Collection
	metadata   *core.PluginMetadata
}

func (p *EnneagramPlugin) Name() string    { return p.name }
func (p *EnneagramPlugin) Version() string { return p.version }

func (p *EnneagramPlugin) Initialize() error {
	p.collection = database.GetCollection("enneagram_tests")
	if p.collection == nil {
		return fmt.Errorf("enneagram plugin: failed to get MongoDB collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{bson.E{Key: "section_id", Value: 1}},
	})
	if err != nil {
		log.Printf("enneagram plugin: warning — could not create index: %v", err)
	}

	log.Printf("enneagram plugin: initialized (collection ready)")
	return nil
}

func (p *EnneagramPlugin) Shutdown() error {
	log.Printf("enneagram plugin: shutdown")
	return nil
}

func (p *EnneagramPlugin) GetMetadata() *core.PluginMetadata {
	return p.metadata
}

func (p *EnneagramPlugin) GetMetadataExtended() *core.PluginMetadataExtended {
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

func (p *EnneagramPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("enneagram plugin: data must be map[string]interface{}")
	}

	action, _ := request["action"].(string)

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
	case "get_results":
		return p.getResults(ctx, request)
	case "get_section_responses":
		return p.getSectionResponses(ctx, request)
	case "get_section_results":
		return p.getSectionResults(ctx, request)
	case "get_student_eneatypes":
		return p.getStudentEneatypes(ctx, request)
	case "calculate_group_compatibility":
		return p.calculateGroupCompatibility(ctx, request)
	case "get_pairwise_compatibility":
		e1Raw, _ := request["eneatype_a"].(float64)
		e2Raw, _ := request["eneatype_b"].(float64)
		return float64(getCompatibility(int(e1Raw), int(e2Raw))), nil
	case "get_ocean_from_type":
		eneatypeRaw, _ := request["eneatype"].(float64)
		v, ok := getOCEANFromEnneagram(int(eneatypeRaw))
		if !ok {
			return nil, fmt.Errorf("enneagram plugin: unknown eneatype %d", int(eneatypeRaw))
		}
		return v, nil
	case "get_metadata":
		return p.metadata, nil
	case "get_metadata_extended":
		return p.GetMetadataExtended(), nil
	default:
		return nil, fmt.Errorf("enneagram plugin: unknown action %q", action)
	}
}
