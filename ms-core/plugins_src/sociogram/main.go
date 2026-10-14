// Sociogram subprocess plugin for ms-core.
// Compile with:
//
//	go build -o ../../plugins_dist/sociogram .
//
// Launched as a subprocess by the host via hashicorp/go-plugin.
package main

import (
	"context"
	"fmt"
	"time"

	"ms-core/internal/database"
	"ms-core/plugins_src/shared"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---------------------------------------------------------------------------
// Plugin subprocess entry point
// ---------------------------------------------------------------------------

// main starts the plugin's net/rpc server when launched as a subprocess by the host.
// All environment variables (MONGO_URI, etc.) are inherited from the parent process.
func main() {
	shared.Serve(NewGroupCompatibilityPlugin())
}

// GroupCompatibilityPlugin implementa el plugin de compatibilidad grupal
type GroupCompatibilityPlugin struct {
	name       string
	version    string
	collection *mongo.Collection
}

// CompatibilityQuestion representa una pregunta de compatibilidad
type CompatibilityQuestion struct {
	ID       int    `bson:"id" json:"id"`
	Text     string `bson:"text" json:"text"`
	Category string `bson:"category" json:"category"`
	Required bool   `bson:"required" json:"required"`
}

// QuestionResponse representa la respuesta a una pregunta específica
// Mapea estudiantes con su respuesta (sí/no)
type QuestionResponse struct {
	QuestionID int             `bson:"question_id" json:"question_id"`
	Answers    []StudentAnswer `bson:"answers" json:"answers"`
}

// StudentAnswer mapea un estudiante con su respuesta sociométrica
type StudentAnswer struct {
	StudentID string `bson:"student_id" json:"student_id"`
	Response  int    `bson:"response" json:"response"` // 1 = Elección (Choice), 0 = Indiferencia, -1 = Rechazo (Rejection)
}

// GroupCompatibilityStudentResponse representa la respuesta de un estudiante al cuestionario de compatibilidad
type GroupCompatibilityStudentResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SurveyID    primitive.ObjectID `bson:"survey_id" json:"survey_id"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	SectionID   string             `bson:"section_id" json:"section_id"`
	SurveyType  string             `bson:"survey_type" json:"survey_type"`               // "full_section" o "group_members"
	GroupID     string             `bson:"group_id,omitempty" json:"group_id,omitempty"` // Solo si survey_type es "group_members"
	Question1   []StudentAnswer    `bson:"question1" json:"question1"`                   // ¿A quién elegirías?
	Question2   []StudentAnswer    `bson:"question2" json:"question2"`                   // ¿Quién te elegirías?
	Question3   []StudentAnswer    `bson:"question3" json:"question3"`                   // ¿A quién NO elegirías?
	Question4   []StudentAnswer    `bson:"question4" json:"question4"`                   // ¿Quién NO te elegirría?
	IsComplete  bool               `bson:"is_complete" json:"is_complete"`
	SubmittedAt *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// GroupCompatibilitySurvey representa una encuesta de compatibilidad grupal
type GroupCompatibilitySurvey struct {
	ID          primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	Title       string                  `bson:"title" json:"title"`
	Description string                  `bson:"description" json:"description"`
	SectionID   string                  `bson:"section_id" json:"section_id"`
	CreatedBy   string                  `bson:"created_by" json:"created_by"`
	SurveyType  string                  `bson:"survey_type" json:"survey_type"` // "full_section" o "group_members"
	Questions   []CompatibilityQuestion `bson:"questions" json:"questions"`
	IsActive    bool                    `bson:"is_active" json:"is_active"`
	StartDate   *time.Time              `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time              `bson:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt   time.Time               `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time               `bson:"updated_at" json:"updated_at"`
}

// NewGroupCompatibilityPlugin crea una nueva instancia del plugin
func NewGroupCompatibilityPlugin() *GroupCompatibilityPlugin {
	return &GroupCompatibilityPlugin{
		name:    "group_compatibility",
		version: "1.0.0",
	}
}

// Name retorna el nombre del plugin
func (p *GroupCompatibilityPlugin) Name() string {
	return p.name
}

// Version retorna la versión del plugin
func (p *GroupCompatibilityPlugin) Version() string {
	return p.version
}

// Initialize inicializa el plugin
func (p *GroupCompatibilityPlugin) Initialize() error {
	fmt.Printf("DEBUG: Initializing group compatibility plugin\n")
	p.collection = database.GetCollection("group_compatibility_surveys")

	if p.collection == nil {
		fmt.Printf("DEBUG: Failed to get collection\n")
		return fmt.Errorf("failed to get collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Crear índices necesarios
	_, err := p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "section_id", Value: 1}},
	})
	if err != nil {
		fmt.Printf("DEBUG: Error creating section_id index: %v\n", err)
		return err
	}

	_, err = p.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "is_active", Value: 1}},
	})
	if err != nil {
		fmt.Printf("DEBUG: Error creating is_active index: %v\n", err)
		return err
	}

	fmt.Printf("DEBUG: Group compatibility plugin initialized successfully\n")
	return nil
}

// Execute ejecuta operaciones del plugin
func (p *GroupCompatibilityPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	request, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format for group compatibility plugin")
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
	case "submit_response":
		return p.submitResponse(ctx, request)
	case "get_responses":
		return p.getResponses(ctx, request)
	case "get_section_responses":
		return p.getSectionResponses(ctx, request)
	case "get_student_response":
		return p.getStudentResponse(ctx, request)
	case "update_survey":
		return p.updateSurvey(ctx, request)
	case "delete_survey":
		return p.deleteSurvey(ctx, request)
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}
}

// createSurvey crea una nueva encuesta de compatibilidad
func (p *GroupCompatibilityPlugin) createSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	// Preguntas predefinidas
	questions := []CompatibilityQuestion{
		{
			ID:       1,
			Text:     "¿A quién(es) elegirías para trabajar en equipo?",
			Category: "positive_choice",
			Required: true,
		},
		{
			ID:       2,
			Text:     "¿Quién(es) crees que te elegirían para trabajar en equipo?",
			Category: "positive_choice",
			Required: true,
		},
		{
			ID:       3,
			Text:     "¿A quién(es) NO elegirías para trabajar en equipo?",
			Category: "negative_choice",
			Required: true,
		},
		{
			ID:       4,
			Text:     "¿Quién(es) crees que NO te elegirían para trabajar en equipo?",
			Category: "negative_choice",
			Required: true,
		},
	}

	survey := GroupCompatibilitySurvey{
		Title:       request["title"].(string),
		Description: request["description"].(string),
		SectionID:   request["section_id"].(string),
		CreatedBy:   request["created_by"].(string),
		SurveyType:  request["survey_type"].(string), // "full_section" o "group_members"
		Questions:   questions,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Establecer fechas por defecto si no están presentes
	defaultStartDate, _ := time.Parse(time.RFC3339Nano, "2025-08-31T08:00:00.000Z")
	defaultEndDate, _ := time.Parse(time.RFC3339Nano, "2025-10-14T23:59:00.000Z")
	survey.StartDate = &defaultStartDate
	survey.EndDate = &defaultEndDate

	// Agregar fechas si están presentes
	if startDateStr, exists := request["start_date"].(string); exists && startDateStr != "" {
		parsedDate, err := time.Parse(time.RFC3339Nano, startDateStr)
		if err == nil {
			survey.StartDate = &parsedDate
			fmt.Printf("DEBUG [group_compatibility.go]: Successfully parsed start_date '%s'\n", startDateStr)
		} else {
			fmt.Printf("DEBUG [group_compatibility.go]: Failed to parse start_date '%s': %v\n", startDateStr, err)
		}
	}

	if endDateStr, exists := request["end_date"].(string); exists && endDateStr != "" {
		parsedDate, err := time.Parse(time.RFC3339Nano, endDateStr)
		if err == nil {
			survey.EndDate = &parsedDate
			fmt.Printf("DEBUG [group_compatibility.go]: Successfully parsed end_date '%s'\n", endDateStr)
		} else {
			fmt.Printf("DEBUG [group_compatibility.go]: Failed to parse end_date '%s': %v\n", endDateStr, err)
		}
	}

	result, err := p.collection.InsertOne(ctx, survey)
	if err != nil {
		fmt.Printf("DEBUG: Error inserting survey: %v\n", err)
		return nil, err
	}

	survey.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Printf("DEBUG: Survey created with ID: %v\n", survey.ID)
	return survey, nil
}

// getSurveys obtiene las encuestas de una sección
func (p *GroupCompatibilityPlugin) getSurveys(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	sectionID := request["section_id"].(string)

	filter := bson.M{"section_id": sectionID}

	cursor, err := p.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var surveys []GroupCompatibilitySurvey
	if err = cursor.All(ctx, &surveys); err != nil {
		return nil, err
	}

	if surveys == nil {
		surveys = []GroupCompatibilitySurvey{}
	}

	// Return full survey data instead of DTO
	return surveys, nil
}

// getSurvey obtiene una encuesta específica
func (p *GroupCompatibilityPlugin) getSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr := request["survey_id"].(string)
	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, err
	}

	var survey GroupCompatibilitySurvey
	err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey)
	if err != nil {
		return nil, err
	}

	// Return full survey data instead of DTO
	return survey, nil
}

// submitResponse envía una respuesta al cuestionario de compatibilidad
func (p *GroupCompatibilityPlugin) submitResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("group_compatibility_responses")

	surveyIDStr := request["survey_id"].(string)
	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, err
	}

	studentID := request["student_id"].(string)
	sectionID := request["section_id"].(string)
	surveyType := request["survey_type"].(string)

	// Procesar respuestas para cada pregunta
	question1Data := request["question1"].([]interface{})
	question1 := p.parseStudentAnswers(question1Data)

	question2Data := request["question2"].([]interface{})
	question2 := p.parseStudentAnswers(question2Data)

	question3Data := request["question3"].([]interface{})
	question3 := p.parseStudentAnswers(question3Data)

	question4Data := request["question4"].([]interface{})
	question4 := p.parseStudentAnswers(question4Data)

	response := GroupCompatibilityStudentResponse{
		SurveyID:   surveyID,
		StudentID:  studentID,
		SectionID:  sectionID,
		SurveyType: surveyType,
		Question1:  question1,
		Question2:  question2,
		Question3:  question3,
		Question4:  question4,
		IsComplete: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Agregar group_id si está en el request (para survey_type "group_members")
	if groupID, exists := request["group_id"].(string); exists {
		response.GroupID = groupID
	}

	// Marcar como enviada
	now := time.Now()
	response.SubmittedAt = &now

	result, err := responsesCollection.InsertOne(ctx, response)
	if err != nil {
		fmt.Printf("DEBUG: Error inserting response: %v\n", err)
		return nil, err
	}

	response.ID = result.InsertedID.(primitive.ObjectID)
	fmt.Printf("DEBUG: Response submitted for student %s\n", studentID)
	return response, nil
}

// parseStudentAnswers convierte los datos de respuesta en formato StudentAnswer
// Response values: 1 (Elección), 0 (Indiferencia), -1 (Rechazo)
func (p *GroupCompatibilityPlugin) parseStudentAnswers(data []interface{}) []StudentAnswer {
	var answers []StudentAnswer

	for _, item := range data {
		itemMap := item.(map[string]interface{})

		// Handle response as either float64 (JSON number) or int
		var response int
		switch v := itemMap["response"].(type) {
		case float64:
			response = int(v)
		case int:
			response = v
		default:
			response = 0 // Default to indifferent if conversion fails
		}

		answer := StudentAnswer{
			StudentID: itemMap["student_id"].(string),
			Response:  response,
		}
		answers = append(answers, answer)
	}

	return answers
}

// getResponses obtiene las respuestas de una encuesta
func (p *GroupCompatibilityPlugin) getResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("group_compatibility_responses")

	surveyIDStr := request["survey_id"].(string)
	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"survey_id": surveyID}

	// Filtros adicionales opcionales
	if sectionID, exists := request["section_id"].(string); exists && sectionID != "" {
		filter["section_id"] = sectionID
	}

	if groupID, exists := request["group_id"].(string); exists && groupID != "" {
		filter["group_id"] = groupID
	}

	cursor, err := responsesCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []GroupCompatibilityStudentResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	if responses == nil {
		responses = []GroupCompatibilityStudentResponse{}
	}

	return responses, nil
}

// getSectionResponses obtiene todas las respuestas de compatibilidad de una sección
func (p *GroupCompatibilityPlugin) getSectionResponses(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("group_compatibility_responses")
	surveysCollection := database.GetCollection("group_compatibility_surveys")

	sectionID := request["section_id"]

	// Convertir section_id a string
	var sectionIDStr string
	if sectionIDFloat, ok := sectionID.(float64); ok {
		sectionIDStr = fmt.Sprintf("%.0f", sectionIDFloat)
	} else if sectionIDStrVal, ok := sectionID.(string); ok {
		sectionIDStr = sectionIDStrVal
	} else {
		return nil, fmt.Errorf("invalid section_id type")
	}

	// Primero obtener las encuestas de esta sección
	surveyFilter := bson.M{"section_id": sectionIDStr}

	surveysCursor, err := surveysCollection.Find(ctx, surveyFilter)
	if err != nil {
		return nil, err
	}
	defer surveysCursor.Close(ctx)

	var surveys []GroupCompatibilitySurvey
	if err = surveysCursor.All(ctx, &surveys); err != nil {
		return nil, err
	}

	// Obtener los IDs de las encuestas
	surveyIDs := make([]primitive.ObjectID, len(surveys))
	for i, survey := range surveys {
		surveyIDs[i] = survey.ID
	}

	if len(surveyIDs) == 0 {
		return []GroupCompatibilityStudentResponse{}, nil
	}

	// Obtener las respuestas de todas las encuestas de la sección
	responsesFilter := bson.M{
		"survey_id": bson.M{"$in": surveyIDs},
	}

	cursor, err := responsesCollection.Find(ctx, responsesFilter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []GroupCompatibilityStudentResponse
	if err = cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	if responses == nil {
		responses = []GroupCompatibilityStudentResponse{}
	}

	return responses, nil
}

// getStudentResponse obtiene la respuesta de un estudiante específico
func (p *GroupCompatibilityPlugin) getStudentResponse(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	responsesCollection := database.GetCollection("group_compatibility_responses")

	surveyIDStr := request["survey_id"].(string)
	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, err
	}

	studentID := request["student_id"].(string)

	var response GroupCompatibilityStudentResponse
	err = responsesCollection.FindOne(ctx, bson.M{
		"survey_id":  surveyID,
		"student_id": studentID,
	}).Decode(&response)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return response, nil
}

// updateSurvey actualiza una encuesta existente
func (p *GroupCompatibilityPlugin) updateSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr := request["survey_id"].(string)
	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, err
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

	result, err := p.collection.UpdateByID(ctx, surveyID, update)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("survey not found")
	}

	// Obtener la encuesta actualizada
	var survey GroupCompatibilitySurvey
	err = p.collection.FindOne(ctx, bson.M{"_id": surveyID}).Decode(&survey)
	if err != nil {
		return nil, err
	}

	return survey, nil
}

// deleteSurvey elimina una encuesta y sus respuestas asociadas
func (p *GroupCompatibilityPlugin) deleteSurvey(ctx context.Context, request map[string]interface{}) (interface{}, error) {
	surveyIDStr := request["survey_id"].(string)
	surveyID, err := primitive.ObjectIDFromHex(surveyIDStr)
	if err != nil {
		return nil, err
	}

	// Eliminar la encuesta
	result, err := p.collection.DeleteOne(ctx, bson.M{"_id": surveyID})
	if err != nil {
		return nil, err
	}

	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("survey not found")
	}

	// Eliminar respuestas asociadas
	responsesCollection := database.GetCollection("group_compatibility_responses")
	_, err = responsesCollection.DeleteMany(ctx, bson.M{"survey_id": surveyID})
	if err != nil {
		return nil, err
	}

	return bson.M{"deleted": true}, nil
}

// Shutdown cierra el plugin
func (p *GroupCompatibilityPlugin) Shutdown() error {
	return nil
}
