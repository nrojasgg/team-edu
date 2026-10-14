package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"ms-core/internal/config"
	"ms-core/internal/core"
	"ms-core/internal/core/domain"
	"ms-core/internal/database"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GroupHandler struct {
	db            *mongo.Database
	pluginManager *core.PluginManager
}

func NewGroupHandler(db *mongo.Database, pm *core.PluginManager) *GroupHandler {
	return &GroupHandler{db: db, pluginManager: pm}
}

// Student representa un estudiante del servicio de courses
type Student struct {
	ID       int64  `json:"id"`
	Nombres  string `json:"nombres"`
	Paterno  string `json:"paterno"`
	Materno  string `json:"materno"`
	Email    string `json:"email"`
	Eneatype string `json:"eneatype,omitempty"`
}

// getStudentsFromCourses obtiene estudiantes de una sección desde el servicio de courses
func (h *GroupHandler) getStudentsFromCourses(sectionID int64, authHeader string) ([]Student, error) {
	cfg := config.LoadConfig()

	// 1. Obtener IDs de estudiantes de la sección
	studentsURL := fmt.Sprintf("%s/api/v1/sections/%d/students", cfg.CoursesServiceURL, sectionID)
	studentsReq, err := http.NewRequest("GET", studentsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create students request: %v", err)
	}
	studentsReq.Header.Set("Authorization", authHeader)
	studentsReq.Header.Set(headerContentType, mimeJSON)

	client := &http.Client{}
	studentsResp, err := client.Do(studentsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students: %v", err)
	}
	defer studentsResp.Body.Close()

	if studentsResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("courses service returned status %d", studentsResp.StatusCode)
	}

	var studentsData struct {
		IDs []int64 `json:"ids"`
	}
	if err := json.NewDecoder(studentsResp.Body).Decode(&studentsData); err != nil {
		return nil, fmt.Errorf("failed to decode students response: %v", err)
	}

	if len(studentsData.IDs) == 0 {
		return []Student{}, nil
	}

	// 2. Obtener detalles de usuarios en batch
	batchBody := map[string]interface{}{
		"ids": studentsData.IDs,
	}
	batchBodyBytes, err := json.Marshal(batchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode batch request: %v", err)
	}

	batchURL := fmt.Sprintf("%s/api/v1/users/batch", cfg.AuthServiceURL)
	batchReq, err := http.NewRequest("POST", batchURL, bytes.NewReader(batchBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %v", err)
	}
	batchReq.Header.Set("Authorization", authHeader)
	batchReq.Header.Set(headerContentType, mimeJSON)

	batchResp, err := client.Do(batchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user details: %v", err)
	}
	defer batchResp.Body.Close()

	if batchResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", batchResp.StatusCode)
	}

	var batchData struct {
		Users []struct {
			ID      int64  `json:"id"`
			Nombres string `json:"nombres"`
			Paterno string `json:"paterno"`
			Materno string `json:"materno"`
			Email   string `json:"email"`
		} `json:"users"`
	}
	if err := json.NewDecoder(batchResp.Body).Decode(&batchData); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %v", err)
	}

	// Mapear a []Student
	students := make([]Student, 0, len(batchData.Users))
	for _, user := range batchData.Users {
		students = append(students, Student{
			ID:      user.ID,
			Nombres: user.Nombres,
			Paterno: user.Paterno,
			Materno: user.Materno,
			Email:   user.Email,
		})
	}

	return students, nil
}

// Group representa un Grupo de estudiantes
// CriterionDetail holds the per-attribute breakdown that is stored in MongoDB
// and returned to the frontend for detailed score visualisation.
type CriterionDetail struct {
	// Score is the absolute-error fitness score: 1 - |Target - RealValue|
	Score float64 `json:"score" bson:"score"`
	// RealValue is the actual normalised diversity / metric value for the group (e.g. Blau index).
	RealValue float64 `json:"real_value" bson:"real_value"`
	// Target is the desired value configured by the professor for this criterion.
	Target float64 `json:"target" bson:"target"`
	// Weight is the normalised weight used in the fitness calculation for this criterion.
	Weight float64 `json:"weight" bson:"weight"`
}

type Group struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SectionID      int64              `json:"section_id" bson:"section_id"`
	GroupNumber    int                `json:"group_number" bson:"group_number"`
	GroupName      string             `json:"group_name" bson:"group_name"`
	StudentID      int64              `json:"student_id" bson:"student_id"`
	DominantType   string             `json:"dominant_type,omitempty" bson:"dominant_type,omitempty"`
	AffinityScore  *float64           `json:"affinity_score,omitempty" bson:"affinity_score,omitempty"`
	ProviderScores map[string]float64 `json:"provider_scores,omitempty" bson:"provider_scores,omitempty"`
	Algorithm      string             `json:"algorithm,omitempty" bson:"algorithm,omitempty"`
	Weights        map[string]float64 `json:"weights,omitempty" bson:"weights,omitempty"`
	Targets        map[string]float64 `json:"targets,omitempty" bson:"targets,omitempty"`
	// CriterionBreakdown stores the detailed per-criterion analysis used to compute
	// the group's fitness score. Each key matches a criterion name (e.g. "sex", "career_code").
	CriterionBreakdown map[string]CriterionDetail `json:"criterion_breakdown,omitempty" bson:"criterion_breakdown,omitempty"`
	CreatedAt          time.Time                  `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time                  `json:"updated_at" bson:"updated_at"`
}

// GetGroupsBySectionID obtiene todos los grupos de una sección
func (h *GroupHandler) GetGroupsBySectionID(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch groups",
		})
	}
	defer cursor.Close(ctx)

	var groups []Group
	if err = cursor.All(ctx, &groups); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode groups",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    groups,
	})
}

// CreateOrUpdateGroup crea o actualiza la asignación de un estudiante a un grupo
func (h *GroupHandler) CreateOrUpdateGroup(c *fiber.Ctx) error {
	var input struct {
		SectionID          int64                      `json:"section_id"`
		GroupNumber        int                        `json:"group_number"`
		GroupName          string                     `json:"group_name"`
		StudentID          int64                      `json:"student_id"`
		DominantType       string                     `json:"dominant_type,omitempty"`
		AffinityScore      *float64                   `json:"affinity_score,omitempty"`
		ProviderScores     map[string]float64         `json:"provider_scores,omitempty"`
		Algorithm          string                     `json:"algorithm,omitempty"`
		Weights            map[string]float64         `json:"weights,omitempty"`
		Targets            map[string]float64         `json:"targets,omitempty"`
		CriterionBreakdown map[string]CriterionDetail `json:"criterion_breakdown,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Buscar si ya existe un registro para este estudiante en esta sección
	filter := bson.M{
		"section_id": input.SectionID,
		"student_id": input.StudentID,
	}

	now := time.Now()
	setFields := bson.M{
		"section_id":    input.SectionID,
		"group_number":  input.GroupNumber,
		"group_name":    input.GroupName,
		"student_id":    input.StudentID,
		"dominant_type": input.DominantType,
		"updated_at":    now,
	}
	if input.AffinityScore != nil {
		setFields["affinity_score"] = *input.AffinityScore
	}
	if len(input.ProviderScores) > 0 {
		setFields["provider_scores"] = input.ProviderScores
	}
	if input.Algorithm != "" {
		setFields["algorithm"] = input.Algorithm
	}
	if len(input.Weights) > 0 {
		setFields["weights"] = input.Weights
	}
	if len(input.Targets) > 0 {
		setFields["targets"] = input.Targets
	}
	if len(input.CriterionBreakdown) > 0 {
		setFields["criterion_breakdown"] = input.CriterionBreakdown
	}
	update := bson.M{
		"$set": setFields,
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save group",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Group assignment saved successfully",
		"data": fiber.Map{
			"upserted_id": result.UpsertedID,
			"modified":    result.ModifiedCount,
			"upserted":    result.UpsertedCount,
		},
	})
}

// CreateOrUpdateGroupsBulk guarda múltiples asignaciones de estudiantes a grupos en una sola operación
func (h *GroupHandler) CreateOrUpdateGroupsBulk(c *fiber.Ctx) error {
	var input struct {
		SectionID   int64 `json:"section_id"`
		Assignments []struct {
			GroupNumber        int                        `json:"group_number"`
			GroupName          string                     `json:"group_name"`
			StudentID          int64                      `json:"student_id"`
			DominantType       string                     `json:"dominant_type,omitempty"`
			AffinityScore      *float64                   `json:"affinity_score,omitempty"`
			ProviderScores     map[string]float64         `json:"provider_scores,omitempty"`
			Algorithm          string                     `json:"algorithm,omitempty"`
			Weights            map[string]float64         `json:"weights,omitempty"`
			Targets            map[string]float64         `json:"targets,omitempty"`
			CriterionBreakdown map[string]CriterionDetail `json:"criterion_breakdown,omitempty"`
		} `json:"assignments"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(input.Assignments) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "assignments array is required and must not be empty",
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now()
	writes := make([]mongo.WriteModel, 0, len(input.Assignments))

	for _, a := range input.Assignments {
		if a.StudentID == 0 {
			continue
		}
		filter := bson.M{
			"section_id": input.SectionID,
			"student_id": a.StudentID,
		}
		setFields := bson.M{
			"section_id":    input.SectionID,
			"group_number":  a.GroupNumber,
			"group_name":    a.GroupName,
			"student_id":    a.StudentID,
			"dominant_type": a.DominantType,
			"updated_at":    now,
		}
		if a.AffinityScore != nil {
			setFields["affinity_score"] = *a.AffinityScore
		}
		if len(a.ProviderScores) > 0 {
			setFields["provider_scores"] = a.ProviderScores
		}
		if a.Algorithm != "" {
			setFields["algorithm"] = a.Algorithm
		}
		if len(a.Weights) > 0 {
			setFields["weights"] = a.Weights
		}
		if len(a.Targets) > 0 {
			setFields["targets"] = a.Targets
		}
		if len(a.CriterionBreakdown) > 0 {
			setFields["criterion_breakdown"] = a.CriterionBreakdown
		}

		update := bson.M{
			"$set": setFields,
			"$setOnInsert": bson.M{
				"created_at": now,
			},
		}

		writeModel := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		writes = append(writes, writeModel)
	}

	if len(writes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No valid assignments to save",
		})
	}

	res, err := collection.BulkWrite(ctx, writes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to bulk save groups: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Successfully processed %d group assignments", len(writes)),
		"data": fiber.Map{
			"matched":  res.MatchedCount,
			"modified": res.ModifiedCount,
			"upserted": res.UpsertedCount,
		},
	})
}

// RemoveStudentFromGroup elimina a un estudiante de su grupo
func (h *GroupHandler) RemoveStudentFromGroup(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	studentID, err := strconv.ParseInt(c.Params("studentId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid student ID",
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"section_id": sectionID,
		"student_id": studentID,
	}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove student from group",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Student not found in any group",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Student removed from group successfully",
	})
}

// PatchGroupScores actualiza affinity_score y provider_scores para todos los documentos
// de un group_number dentro de una sección. No modifica la composición del grupo.
// Es llamado después de un drag & drop para persistir los scores recalculados.
func (h *GroupHandler) PatchGroupScores(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	var input struct {
		Groups []struct {
			GroupID            string                     `json:"group_id"`
			AffinityScore      float64                    `json:"affinity_score"`
			ProviderScores     map[string]float64         `json:"provider_scores"`
			CriterionBreakdown map[string]CriterionDetail `json:"criterion_breakdown,omitempty"`
		} `json:"groups"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(input.Groups) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "groups array is required and must not be empty",
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	for _, g := range input.Groups {
		groupNumber, err := strconv.Atoi(g.GroupID)
		if err != nil {
			// group_id puede ser un número como string; si no es parseable lo saltamos
			continue
		}

		filter := bson.M{
			"section_id":   sectionID,
			"group_number": groupNumber,
		}
		setDoc := bson.M{
			"affinity_score":  g.AffinityScore,
			"provider_scores": g.ProviderScores,
			"updated_at":      now,
		}
		if len(g.CriterionBreakdown) > 0 {
			setDoc["criterion_breakdown"] = g.CriterionBreakdown
		}
		update := bson.M{
			"$set": setDoc,
		}

		if _, err := collection.UpdateMany(ctx, filter, update); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to update scores for group %s: %v", g.GroupID, err),
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Group scores updated successfully",
	})
}

// DeleteGroupsBySection elimina todos los grupos de una sección
func (h *GroupHandler) DeleteGroupsBySection(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.DeleteMany(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete groups",
		})
	}

	return c.JSON(fiber.Map{
		"success":       true,
		"message":       "Groups deleted successfully",
		"deleted_count": result.DeletedCount,
	})
}

// GroupMatesResponse representa la respuesta con los compañeros de grupo
type GroupMatesResponse struct {
	GroupName   string  `json:"group_name"`
	GroupNumber int     `json:"group_number"`
	Students    []int64 `json:"students"`
}

// GetStudentGroupMates obtiene los compañeros de grupo del estudiante autenticado
func (h *GroupHandler) GetStudentGroupMates(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	studentID, err := resolveUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var studentGroup Group
	err = collection.FindOne(ctx, bson.M{
		"section_id": sectionID,
		"student_id": studentID,
	}).Decode(&studentGroup)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Student not assigned to any group in this section",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch student group",
		})
	}

	cursor, err := collection.Find(ctx, bson.M{
		"section_id":   sectionID,
		"group_number": studentGroup.GroupNumber,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch group members",
		})
	}
	defer cursor.Close(ctx)

	var groupMembers []Group
	if err = cursor.All(ctx, &groupMembers); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode group members",
		})
	}

	var studentIDs []int64
	for _, member := range groupMembers {
		studentIDs = append(studentIDs, member.StudentID)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": GroupMatesResponse{
			GroupName:   studentGroup.GroupName,
			GroupNumber: studentGroup.GroupNumber,
			Students:    studentIDs,
		},
	})
}

// resolveUserID extracts the student ID from JWT locals, trying "userID" then "claims".
func resolveUserID(c *fiber.Ctx) (int64, error) {
	userID := c.Locals("userID")
	if userID == nil {
		if claims, ok := c.Locals("claims").(map[string]interface{}); ok {
			if id, exists := claims["id"]; exists {
				userID = id
			}
		}
	}
	if userID == nil {
		return 0, fmt.Errorf("User ID not found in token")
	}
	switch v := userID.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("Invalid user ID format")
	}
}

// classifyStudentsByEnneagram separates students into enneagram-typed groups
// and those with no enneatype.
func classifyStudentsByEnneagram(students []bson.M) (map[string][]bson.M, []bson.M) {
	enneagramGroups := make(map[string][]bson.M)
	var studentsWithoutEnneatype []bson.M
	for _, student := range students {
		eneatype, ok := student["eneatype"].(string)
		if !ok || eneatype == "" {
			studentsWithoutEnneatype = append(studentsWithoutEnneatype, student)
		} else {
			enneagramGroups[eneatype] = append(enneagramGroups[eneatype], student)
		}
	}
	return enneagramGroups, studentsWithoutEnneatype
}

// insertEnneagramGroups inserts one group document per enneagram type and returns
// the next available group number.
func insertEnneagramGroups(ctx context.Context, collection *mongo.Collection, sectionID int64, groups map[string][]bson.M, groupNumber int) (int, error) {
	for eneatype, groupStudents := range groups {
		if len(groupStudents) == 0 {
			continue
		}
		groupName := fmt.Sprintf("Grupo %s", eneatype)
		if err := insertStudentsIntoGroup(ctx, collection, sectionID, groupNumber, groupName, eneatype, groupStudents); err != nil {
			return groupNumber, err
		}
		groupNumber++
	}
	return groupNumber, nil
}

// FormGroupsByEnneagram forma grupos basados en tipos de eneagrama
func (h *GroupHandler) FormGroupsByEnneagram(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err = collection.DeleteMany(ctx, bson.M{"section_id": sectionID}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear existing groups",
		})
	}

	studentsCollection := h.db.Collection(database.StudentsCollection)
	cursor, err := studentsCollection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get students",
		})
	}
	defer cursor.Close(ctx)

	var students []bson.M
	if err = cursor.All(ctx, &students); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode students",
		})
	}

	enneagramGroups, studentsWithoutEnneatype := classifyStudentsByEnneagram(students)

	groupNumber, err := insertEnneagramGroups(ctx, collection, sectionID, enneagramGroups, 1)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if len(studentsWithoutEnneatype) > 0 {
		if err := insertStudentsIntoGroup(ctx, collection, sectionID, groupNumber, "Grupo Sin Enneatipo", "", studentsWithoutEnneatype); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Grupos formados exitosamente basados en tipos de eneagrama",
	})
}

// insertStudentsIntoGroup inserts one MongoDB Group document per student in the given list.
func insertStudentsIntoGroup(ctx context.Context, collection *mongo.Collection, sectionID int64, groupNumber int, groupName, dominantType string, students []bson.M) error {
	for _, student := range students {
		studentID, ok := student["id"].(int64)
		if !ok {
			continue
		}
		group := Group{
			SectionID:    sectionID,
			GroupNumber:  groupNumber,
			GroupName:    groupName,
			StudentID:    studentID,
			DominantType: dominantType,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if _, err := collection.InsertOne(ctx, group); err != nil {
			return fmt.Errorf("Failed to create group")
		}
	}
	return nil
}

// FormRandomGroups forma grupos de manera aleatoria
func (h *GroupHandler) FormRandomGroups(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header required",
		})
	}

	// Parse request body for students per group
	var requestBody struct {
		StudentsPerGroup int `json:"students_per_group"`
	}
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	studentsPerGroup := requestBody.StudentsPerGroup
	if studentsPerGroup <= 0 {
		studentsPerGroup = 4 // default
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Primero, eliminar todos los grupos existentes de la sección
	_, err = collection.DeleteMany(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to clear existing groups",
		})
	}

	// Obtener estudiantes de la sección desde el servicio de courses
	students, err := h.getStudentsFromCourses(sectionID, authHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get students",
		})
	}

	if len(students) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No students found in section",
		})
	}

	// Mezclar estudiantes aleatoriamente
	shuffledStudents := make([]Student, len(students))
	copy(shuffledStudents, students)

	rand.Shuffle(len(shuffledStudents), func(i, j int) {
		shuffledStudents[i], shuffledStudents[j] = shuffledStudents[j], shuffledStudents[i]
	})

	// Determinar número óptimo de grupos basado en studentsPerGroup
	totalGroups := (len(shuffledStudents) + studentsPerGroup - 1) / studentsPerGroup

	// Crear grupos
	for groupNum := 1; groupNum <= totalGroups; groupNum++ {
		groupName := fmt.Sprintf("Grupo %d", groupNum)
		startIdx := (groupNum - 1) * studentsPerGroup
		endIdx := groupNum * studentsPerGroup
		if endIdx > len(shuffledStudents) {
			endIdx = len(shuffledStudents)
		}

		groupStudents := shuffledStudents[startIdx:endIdx]

		// Crear documentos de grupo para cada estudiante
		for _, student := range groupStudents {
			group := Group{
				SectionID:   sectionID,
				GroupNumber: groupNum,
				GroupName:   groupName,
				StudentID:   student.ID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			_, err = collection.InsertOne(ctx, group)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create group",
				})
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Grupos formados exitosamente de manera aleatoria",
	})
}

// ExportGroupsToExcel exporta los grupos de una sección a un archivo Excel
func (h *GroupHandler) ExportGroupsToExcel(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("sectionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": errInvalidSectionID,
		})
	}

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization header",
		})
	}

	cfg := config.LoadConfig()
	client := &http.Client{}

	studentsData, err := fetchStudentsComplete(client, cfg.APIGatewayURL, sectionID, authHeader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	collection := h.db.Collection(database.GroupsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"section_id": sectionID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch groups",
		})
	}
	defer cursor.Close(ctx)

	var groups []Group
	if err = cursor.All(ctx, &groups); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode groups",
		})
	}

	type GroupInfo struct {
		GroupName   string
		GroupNumber int
	}
	groupMap := make(map[int64]GroupInfo)
	for _, g := range groups {
		groupMap[g.StudentID] = GroupInfo{GroupName: g.GroupName, GroupNumber: g.GroupNumber}
	}

	f := excelize.NewFile()

	// Build psychometric data maps directly from MongoDB (all three sources)
	sectionIDStr := strconv.FormatInt(sectionID, 10)
	pMaps := buildPsychDataMap(h.db, sectionIDStr)

	nameOf := func(s excelStudent) string {
		return fmt.Sprintf("%s %s %s", s.Nombres, s.Paterno, s.Materno)
	}

	// ── Hoja 1: Datos Generales ──────────────────────────────────────────
	{
		sh := "Datos Generales"
		idx, _ := f.NewSheet(sh)
		f.SetActiveSheet(idx)
		writeExcelHeaders(f, sh, []string{
			"ID", "Nombre Alumno", "RUT", "Correo", "Grupo",
		})
		for i, student := range studentsData {
			r := i + 2
			groupName := ""
			if info, ok := groupMap[student.ID]; ok {
				groupName = info.GroupName
			}
			f.SetCellValue(sh, col(1, r), student.ID)
			f.SetCellValue(sh, col(2, r), nameOf(student))
			f.SetCellValue(sh, col(3, r), student.Rut)
			f.SetCellValue(sh, col(4, r), student.Email)
			f.SetCellValue(sh, col(5, r), groupName)
		}
		f.SetColWidth(sh, "A", "A", 14)
		f.SetColWidth(sh, "B", "B", 32)
		f.SetColWidth(sh, "C", "C", 14)
		f.SetColWidth(sh, "D", "D", 30)
		f.SetColWidth(sh, "E", "E", 20)
	}

	// ── Hoja 2: Datos Demográficos ───────────────────────────────────────
	{
		sh := "Datos Demográficos"
		f.NewSheet(sh) //nolint:errcheck
		writeExcelHeaders(f, sh, []string{"Nombre Alumno", "Correo", "RUT", "Carrera", "Sexo"})
		for i, student := range studentsData {
			r := i + 2
			f.SetCellValue(sh, col(1, r), nameOf(student))
			f.SetCellValue(sh, col(2, r), student.Email)
			f.SetCellValue(sh, col(3, r), student.Rut)
			f.SetCellValue(sh, col(4, r), student.CareerCode)
			f.SetCellValue(sh, col(5, r), student.Sexo)
		}
		f.SetColWidth(sh, "A", "A", 28)
		f.SetColWidth(sh, "B", "B", 28)
		f.SetColWidth(sh, "C", "C", 12)
		f.SetColWidth(sh, "D", "D", 16)
		f.SetColWidth(sh, "E", "E", 8)
	}

	// ── Hoja 3: FFM – BFI-44 ─────────────────────────────────────────────
	{
		sh := "FFM (BFI-44)"
		f.NewSheet(sh) //nolint:errcheck
		writeExcelHeaders(f, sh, []string{
			"Nombre Alumno",
			"O (Apertura)", "C (Responsab.)", "E (Sociabilidad)", "A (Amabilidad)", "N (Est. Emocional)",
		})
		r := 2
		for _, student := range studentsData {
			if psych, ok := pMaps.BFI44[fmt.Sprintf("%d", student.ID)]; ok {
				f.SetCellValue(sh, col(1, r), nameOf(student))
				f.SetCellValue(sh, col(2, r), psych.OCEAN.Openness)
				f.SetCellValue(sh, col(3, r), psych.OCEAN.Conscientiousness)
				f.SetCellValue(sh, col(4, r), psych.OCEAN.Extraversion)
				f.SetCellValue(sh, col(5, r), psych.OCEAN.Agreeableness)
				f.SetCellValue(sh, col(6, r), psych.OCEAN.Neuroticism)
				r++
			}
		}
		f.SetColWidth(sh, "A", "A", 28)
		f.SetColWidth(sh, "B", "F", 14)
	}

	// ── Hoja 4: Eneagrama ────────────────────────────────────────────────
	{
		sh := "Eneagrama"
		f.NewSheet(sh) //nolint:errcheck
		writeExcelHeaders(f, sh, []string{
			"Nombre Alumno", "Eneatipo",
			"E1", "E2", "E3", "E4", "E5", "E6", "E7", "E8", "E9",
			"O (Apertura)", "C (Responsab.)", "E (Sociabilidad)", "A (Amabilidad)", "N (Est. Emocional)",
		})
		r := 2
		for _, student := range studentsData {
			if psych, ok := pMaps.Enneagram[fmt.Sprintf("%d", student.ID)]; ok {
				typeLabel := ""
				if psych.Eneatype > 0 {
					typeLabel = fmt.Sprintf("E%d", psych.Eneatype)
				}
				f.SetCellValue(sh, col(1, r), nameOf(student))
				f.SetCellValue(sh, col(2, r), typeLabel)
				f.SetCellValue(sh, col(3, r), psych.E1)
				f.SetCellValue(sh, col(4, r), psych.E2)
				f.SetCellValue(sh, col(5, r), psych.E3)
				f.SetCellValue(sh, col(6, r), psych.E4)
				f.SetCellValue(sh, col(7, r), psych.E5)
				f.SetCellValue(sh, col(8, r), psych.E6)
				f.SetCellValue(sh, col(9, r), psych.E7)
				f.SetCellValue(sh, col(10, r), psych.E8)
				f.SetCellValue(sh, col(11, r), psych.E9)
				f.SetCellValue(sh, col(12, r), psych.OCEAN.Openness)
				f.SetCellValue(sh, col(13, r), psych.OCEAN.Conscientiousness)
				f.SetCellValue(sh, col(14, r), psych.OCEAN.Extraversion)
				f.SetCellValue(sh, col(15, r), psych.OCEAN.Agreeableness)
				f.SetCellValue(sh, col(16, r), psych.OCEAN.Neuroticism)
				r++
			}
		}
		f.SetColWidth(sh, "A", "A", 28)
		f.SetColWidth(sh, "B", "B", 10)
		f.SetColWidth(sh, "C", "K", 7)
		f.SetColWidth(sh, "L", "P", 14)
	}

	// ── Hoja 5: MBTI ─────────────────────────────────────────────────────
	{
		sh := "MBTI"
		f.NewSheet(sh) //nolint:errcheck
		writeExcelHeaders(f, sh, []string{
			"Nombre Alumno", "Tipo MBTI",
			"O (Apertura)", "C (Responsab.)", "E (Sociabilidad)", "A (Amabilidad)", "N (Est. Emocional)",
		})
		r := 2
		for _, student := range studentsData {
			if psych, ok := pMaps.MBTI[fmt.Sprintf("%d", student.ID)]; ok {
				f.SetCellValue(sh, col(1, r), nameOf(student))
				f.SetCellValue(sh, col(2, r), psych.MBTIType)
				f.SetCellValue(sh, col(3, r), psych.OCEAN.Openness)
				f.SetCellValue(sh, col(4, r), psych.OCEAN.Conscientiousness)
				f.SetCellValue(sh, col(5, r), psych.OCEAN.Extraversion)
				f.SetCellValue(sh, col(6, r), psych.OCEAN.Agreeableness)
				f.SetCellValue(sh, col(7, r), psych.OCEAN.Neuroticism)
				r++
			}
		}
		f.SetColWidth(sh, "A", "A", 28)
		f.SetColWidth(sh, "B", "B", 12)
		f.SetColWidth(sh, "C", "G", 14)
	}

	// ── Hoja 6: Afinidad Social ───────────────────────────────────────────
	{
		sh := "Afinidad Social"
		f.NewSheet(sh) //nolint:errcheck
		writeExcelHeaders(f, sh, []string{
			"Nombre Alumno", "Promedio afinidad recibida", "Promedio afinidad enviada",
		})
		f.SetCellValue(sh, "A2", "Datos no disponibles en esta versión")
		f.SetColWidth(sh, "A", "C", 30) //nolint:errcheck
	}

	// ── Hoja 7: Historial de Formación ───────────────────────────────────
	{
		sh := "Historial de Formación"
		f.NewSheet(sh) //nolint:errcheck
		writeExcelHeaders(f, sh, []string{
			"Fecha Formación", "Algoritmo", "Proveedores Usados", "Pesos", "Targets",
			"Score Global", "Tiempo (ms)",
			"Equipo", "Miembros", "Score Equipo", "Scores por Proveedor", "Factible",
		})

		histCtx, histCancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer histCancel()

		hCursor, hErr := h.db.Collection(database.GroupsHistoryCollection).Find(
			histCtx,
			bson.M{"section_id": sectionIDStr},
			options.Find().SetSort(bson.D{{Key: "formed_at", Value: -1}}),
		)
		if hErr == nil {
			defer hCursor.Close(histCtx)
			var records []GroupHistoryRecord
			if hCursor.All(histCtx, &records) == nil {
				r := 2
				for _, rec := range records {
					// Format helper strings
					weightsStr := ""
					for k, v := range rec.Weights {
						if weightsStr != "" {
							weightsStr += ", "
						}
						weightsStr += fmt.Sprintf("%s:%.2f", k, v)
					}
					targetsStr := ""
					for k, v := range rec.Targets {
						if targetsStr != "" {
							targetsStr += ", "
						}
						targetsStr += fmt.Sprintf("%s:%.2f", k, v)
					}
					providersStr := ""
					for i, p := range rec.ProvidersUsed {
						if i > 0 {
							providersStr += ", "
						}
						providersStr += p
					}
					dateStr := rec.FormedAt.Format("02/01/2006 15:04:05")

					for ti, team := range rec.Teams {
						membersStr := ""
						for mi, m := range team.MemberIDs {
							if mi > 0 {
								membersStr += ", "
							}
							membersStr += m
						}
						provScoresStr := ""
						for k, v := range team.ProviderScores {
							if provScoresStr != "" {
								provScoresStr += ", "
							}
							provScoresStr += fmt.Sprintf("%s:%.3f", k, v)
						}
						feasible := "Sí"
						if !team.IsFeasible {
							feasible = "No"
						}

						// Only write metadata columns for the first team row of each record
						if ti == 0 {
							f.SetCellValue(sh, col(1, r), dateStr)
							f.SetCellValue(sh, col(2, r), rec.Algorithm)
							f.SetCellValue(sh, col(3, r), providersStr)
							f.SetCellValue(sh, col(4, r), weightsStr)
							f.SetCellValue(sh, col(5, r), targetsStr)
							f.SetCellValue(sh, col(6, r), rec.GlobalScore)
							f.SetCellValue(sh, col(7, r), rec.ExecutionMs)
						}
						f.SetCellValue(sh, col(8, r), team.TeamID)
						f.SetCellValue(sh, col(9, r), membersStr)
						f.SetCellValue(sh, col(10, r), team.AffinityScore)
						f.SetCellValue(sh, col(11, r), provScoresStr)
						f.SetCellValue(sh, col(12, r), feasible)
						r++
					}
				}
			}
		}

		f.SetColWidth(sh, "A", "A", 20) // Fecha
		f.SetColWidth(sh, "B", "B", 18) // Algoritmo
		f.SetColWidth(sh, "C", "C", 24) // Proveedores
		f.SetColWidth(sh, "D", "D", 20) // Pesos
		f.SetColWidth(sh, "E", "E", 20) // Targets
		f.SetColWidth(sh, "F", "F", 13) // Score Global
		f.SetColWidth(sh, "G", "G", 11) // Tiempo ms
		f.SetColWidth(sh, "H", "H", 10) // Equipo
		f.SetColWidth(sh, "I", "I", 40) // Miembros
		f.SetColWidth(sh, "J", "J", 13) // Score Equipo
		f.SetColWidth(sh, "K", "K", 28) // Scores Proveedor
		f.SetColWidth(sh, "L", "L", 10) // Factible
	}

	f.DeleteSheet("Sheet1")

	buf, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate Excel file",
		})
	}

	c.Set(headerContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=grupos-estudiantes-seccion-%d.xlsx", sectionID))
	c.Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	return c.Send(buf.Bytes())
}

// excelStudent holds data for a single student row in the Excel export.
type excelStudent struct {
	ID         int64
	Nombres    string
	Paterno    string
	Materno    string
	Email      string
	Rut        string
	CareerCode string
	Eneatype   interface{}
	Sexo       string
}

// studentPsychData holds psychometric results obtained from MongoDB directly.
type studentPsychData struct {
	Source   string // "enneagram", "bfi44", "mbti"
	OCEAN    domain.OCEANVector
	Eneatype int    // dominant enneagram type (0 if none)
	MBTIType string // e.g. "INTJ" (empty if none)
	// Enneagram per-type scores
	E1, E2, E3, E4, E5, E6, E7, E8, E9 float64
}

// psychMapsResult holds per-source response maps plus a merged (priority) map.
type psychMapsResult struct {
	Merged    map[string]studentPsychData // priority: enneagram > bfi44 > mbti
	Enneagram map[string]studentPsychData // all enneagram responses
	BFI44     map[string]studentPsychData // all bfi44 responses
	MBTI      map[string]studentPsychData // all mbti responses
}

// fetchStudentsComplete retrieves the full student list from the API Gateway.
func fetchStudentsComplete(client *http.Client, gatewayURL string, sectionID int64, authHeader string) ([]excelStudent, error) {
	url := fmt.Sprintf("%s/api/v1/sections/%d/students-complete", gatewayURL, sectionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create students request: %v", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set(headerContentType, mimeJSON)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch students: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Gateway returned status %d", resp.StatusCode)
	}

	var data struct {
		Success  bool `json:"success"`
		Students []struct {
			ID         int64       `json:"id"`
			Nombres    string      `json:"nombres"`
			Paterno    string      `json:"paterno"`
			Materno    string      `json:"materno"`
			Email      string      `json:"email"`
			Rut        string      `json:"rut"`
			CareerCode string      `json:"career_code"`
			Eneatype   interface{} `json:"eneatype"`
			Sexo       string      `json:"sexo"`
		} `json:"students"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("Failed to decode students response: %v", err)
	}

	result := make([]excelStudent, len(data.Students))
	for i, s := range data.Students {
		result[i] = excelStudent{ID: s.ID, Nombres: s.Nombres, Paterno: s.Paterno, Materno: s.Materno, Email: s.Email, Rut: s.Rut, CareerCode: s.CareerCode, Eneatype: s.Eneatype, Sexo: s.Sexo}
	}
	return result, nil
}

// fetchSectionDetails retrieves course name, section code and semester info (best-effort).
func fetchSectionDetails(client *http.Client, gatewayURL string, sectionID int64, authHeader string) (courseName, sectionCode, semester string) {
	url := fmt.Sprintf("%s/api/v1/sections/%d/complete", gatewayURL, sectionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set(headerContentType, mimeJSON)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var info struct {
		Course       map[string]interface{} `json:"course"`
		Section      map[string]interface{} `json:"section"`
		SemesterInfo string                 `json:"semester_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return
	}

	if info.Course != nil {
		if name, ok := info.Course["name"].(string); ok {
			courseName = name
		}
	}
	if info.Section != nil {
		if name, ok := info.Section["name"].(string); ok {
			sectionCode = name
		} else if code, ok := info.Section["code"].(string); ok {
			sectionCode = code
		}
	}
	semester = info.SemesterInfo
	return
}

// buildPsychDataMap queries MongoDB directly for enneagram, BFI-44 and MBTI responses
// for a given section, returning a map[studentIDStr]studentPsychData.
// It prefers enneagram data; if none, tries BFI-44, then MBTI.
func buildPsychDataMap(db *mongo.Database, sectionIDStr string) psychMapsResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := psychMapsResult{
		Merged:    map[string]studentPsychData{},
		Enneagram: map[string]studentPsychData{},
		BFI44:     map[string]studentPsychData{},
		MBTI:      map[string]studentPsychData{},
	}

	// ── helper: collect test IDs for a given tests-collection + section_id ──
	getTestIDs := func(collName string) []primitive.ObjectID {
		cursor, err := db.Collection(collName).Find(ctx, bson.M{"section_id": sectionIDStr}, options.Find().SetProjection(bson.M{"_id": 1}))
		if err != nil {
			return nil
		}
		defer cursor.Close(ctx)
		var docs []struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		cursor.All(ctx, &docs) //nolint:errcheck
		ids := make([]primitive.ObjectID, len(docs))
		for i, d := range docs {
			ids[i] = d.ID
		}
		return ids
	}

	// ── Enneagram ──────────────────────────────────────────────────────────
	if ids := getTestIDs("enneagram_tests"); len(ids) > 0 {
		type ennResp struct {
			StudentID string `bson:"student_id"`
			Results   *struct {
				DominantType int                `bson:"dominant_type"`
				OCEANVector  map[string]float64 `bson:"ocean_vector"`
				E1           float64            `bson:"e1"`
				E2           float64            `bson:"e2"`
				E3           float64            `bson:"e3"`
				E4           float64            `bson:"e4"`
				E5           float64            `bson:"e5"`
				E6           float64            `bson:"e6"`
				E7           float64            `bson:"e7"`
				E8           float64            `bson:"e8"`
				E9           float64            `bson:"e9"`
			} `bson:"results"`
			IsComplete bool `bson:"is_complete"`
		}
		cursor, err := db.Collection("enneagram_responses").Find(ctx,
			bson.M{"test_id": bson.M{"$in": ids}, "is_complete": true})
		if err == nil {
			defer cursor.Close(ctx)
			var rows []ennResp
			if cursor.All(ctx, &rows) == nil {
				for _, r := range rows {
					if r.Results == nil {
						continue
					}
					ov := r.Results.OCEANVector
					entry := studentPsychData{
						Source:   "enneagram",
						Eneatype: r.Results.DominantType,
						OCEAN: domain.OCEANVector{
							Openness:          ov["openness"],
							Conscientiousness: ov["conscientiousness"],
							Extraversion:      ov["extraversion"],
							Agreeableness:     ov["agreeableness"],
							Neuroticism:       ov["neuroticism"],
						},
						E1: r.Results.E1, E2: r.Results.E2, E3: r.Results.E3,
						E4: r.Results.E4, E5: r.Results.E5, E6: r.Results.E6,
						E7: r.Results.E7, E8: r.Results.E8, E9: r.Results.E9,
					}
					result.Enneagram[r.StudentID] = entry
					result.Merged[r.StudentID] = entry
				}
			}
		}
	}

	// ── BFI-44 (only for students without enneagram data) ──────────────────
	if ids := getTestIDs("bfi44_tests"); len(ids) > 0 {
		type bfi44Resp struct {
			StudentID string `bson:"student_id"`
			Results   *struct {
				OCEANVector struct {
					Openness          float64 `bson:"openness"`
					Conscientiousness float64 `bson:"conscientiousness"`
					Extraversion      float64 `bson:"extraversion"`
					Agreeableness     float64 `bson:"agreeableness"`
					Neuroticism       float64 `bson:"neuroticism"`
				} `bson:"ocean_vector"`
			} `bson:"results"`
			IsComplete bool `bson:"is_complete"`
		}
		cursor, err := db.Collection("bfi44_responses").Find(ctx,
			bson.M{"test_id": bson.M{"$in": ids}, "is_complete": true})
		if err == nil {
			defer cursor.Close(ctx)
			var rows []bfi44Resp
			if cursor.All(ctx, &rows) == nil {
				for _, r := range rows {
					if r.Results == nil {
						continue
					}
					ov := r.Results.OCEANVector
					bfi44Entry := studentPsychData{
						Source: "bfi44",
						OCEAN: domain.OCEANVector{
							Openness:          ov.Openness,
							Conscientiousness: ov.Conscientiousness,
							Extraversion:      ov.Extraversion,
							Agreeableness:     ov.Agreeableness,
							Neuroticism:       ov.Neuroticism,
						},
					}
					result.BFI44[r.StudentID] = bfi44Entry
					if _, exists := result.Merged[r.StudentID]; !exists {
						result.Merged[r.StudentID] = bfi44Entry
					}
				}
			}
		}
	}

	// ── MBTI (only for students without enneagram/bfi44 data) ─────────────
	if ids := getTestIDs("mbti_tests"); len(ids) > 0 {
		type mbtiResp struct {
			StudentID string `bson:"student_id"`
			Results   struct {
				MBTIType    string `bson:"mbti_type"`
				OCEANVector struct {
					Openness          float64 `bson:"openness"`
					Conscientiousness float64 `bson:"conscientiousness"`
					Extraversion      float64 `bson:"extraversion"`
					Agreeableness     float64 `bson:"agreeableness"`
					Neuroticism       float64 `bson:"neuroticism"`
				} `bson:"ocean_vector"`
			} `bson:"results"`
			IsComplete bool `bson:"is_complete"`
		}
		cursor, err := db.Collection("mbti_responses").Find(ctx,
			bson.M{"test_id": bson.M{"$in": ids}, "is_complete": true})
		if err == nil {
			defer cursor.Close(ctx)
			var rows []mbtiResp
			if cursor.All(ctx, &rows) == nil {
				for _, r := range rows {
					ov := r.Results.OCEANVector
					mbtiEntry := studentPsychData{
						Source:   "mbti",
						MBTIType: r.Results.MBTIType,
						OCEAN: domain.OCEANVector{
							Openness:          ov.Openness,
							Conscientiousness: ov.Conscientiousness,
							Extraversion:      ov.Extraversion,
							Agreeableness:     ov.Agreeableness,
							Neuroticism:       ov.Neuroticism,
						},
					}
					result.MBTI[r.StudentID] = mbtiEntry
					if _, exists := result.Merged[r.StudentID]; !exists {
						result.Merged[r.StudentID] = mbtiEntry
					}
				}
			}
		}
	}

	return result
}

// buildOCEANFromEneatype converts an eneatype value to an OCEAN vector via the enneagram plugin.
func (h *GroupHandler) buildOCEANFromEneatype(eneatype interface{}) domain.OCEANVector {
	if eneatype == nil || h.pluginManager == nil {
		return domain.OCEANVector{}
	}
	enneagramInt, found := toInt(eneatype)
	if !found {
		return domain.OCEANVector{}
	}
	raw, err := h.pluginManager.ExecutePlugin("enneagram", map[string]interface{}{
		"action":   "get_ocean_from_type",
		"eneatype": float64(enneagramInt),
	})
	if err != nil {
		return domain.OCEANVector{}
	}
	var ocean domain.OCEANVector
	if b, err := json.Marshal(raw); err == nil {
		json.Unmarshal(b, &ocean) //nolint:errcheck
	}
	return ocean
}

// toInt converts a float64, string, or int to int.
func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case float64:
		return int(val), true
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i, true
		}
	case int:
		return val, true
	}
	return 0, false
}

// writeExcelHeaders writes bold header row with blue background.
func writeExcelHeaders(f *excelize.File, sheetName string, headers []string) {
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"3182CE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, style)
	}
}

// StudentRowData bundles the per-student data needed to write one Excel row.
type StudentRowData struct {
	CourseName  string
	SectionCode string
	Semester    string
	Student     excelStudent
	GroupName   string
	Psych       studentPsychData
	HasPsych    bool
}

// col returns the Excel cell address for column index (1-based) and the given row.
func col(colIdx, row int) string {
	cell, _ := excelize.CoordinatesToCellName(colIdx, row)
	return cell
}

// writeExcelStudentRow writes one student's data at the given row number.
// Layout:
//
//	A  Nombre Curso      B  Sección         C  Semestre
//	D  Nombre Alumno     E  Correo           F  RUT
//	G  Carrera           H  Sexo             I  Grupo
//	J  Fuente Test       K  Eneatipo/Tipo
//	L  O  M  C  N  E  O  A  P  N
//	Q  E1 … Y  E9
func writeExcelStudentRow(f *excelize.File, sheetName string, row int, d StudentRowData) {
	fullName := fmt.Sprintf("%s %s %s", d.Student.Nombres, d.Student.Paterno, d.Student.Materno)
	// Basic info
	f.SetCellValue(sheetName, col(1, row), d.CourseName)
	f.SetCellValue(sheetName, col(2, row), d.SectionCode)
	f.SetCellValue(sheetName, col(3, row), d.Semester)
	f.SetCellValue(sheetName, col(4, row), fullName)
	f.SetCellValue(sheetName, col(5, row), d.Student.Email)
	f.SetCellValue(sheetName, col(6, row), d.Student.Rut)
	f.SetCellValue(sheetName, col(7, row), d.Student.CareerCode)
	f.SetCellValue(sheetName, col(8, row), d.Student.Sexo)
	f.SetCellValue(sheetName, col(9, row), d.GroupName)
	// Psychometric data (columns J-Y)
	if d.HasPsych {
		// Source + identifier
		f.SetCellValue(sheetName, col(10, row), d.Psych.Source)
		identifier := ""
		switch d.Psych.Source {
		case "enneagram":
			if d.Psych.Eneatype > 0 {
				identifier = fmt.Sprintf("E%d", d.Psych.Eneatype)
			}
		case "mbti":
			identifier = d.Psych.MBTIType
		}
		f.SetCellValue(sheetName, col(11, row), identifier)
		// OCEAN vector
		f.SetCellValue(sheetName, col(12, row), d.Psych.OCEAN.Openness)
		f.SetCellValue(sheetName, col(13, row), d.Psych.OCEAN.Conscientiousness)
		f.SetCellValue(sheetName, col(14, row), d.Psych.OCEAN.Extraversion)
		f.SetCellValue(sheetName, col(15, row), d.Psych.OCEAN.Agreeableness)
		f.SetCellValue(sheetName, col(16, row), d.Psych.OCEAN.Neuroticism)
		// Enneagram per-type scores (Q-Y, only if source is enneagram)
		if d.Psych.Source == "enneagram" {
			f.SetCellValue(sheetName, col(17, row), d.Psych.E1)
			f.SetCellValue(sheetName, col(18, row), d.Psych.E2)
			f.SetCellValue(sheetName, col(19, row), d.Psych.E3)
			f.SetCellValue(sheetName, col(20, row), d.Psych.E4)
			f.SetCellValue(sheetName, col(21, row), d.Psych.E5)
			f.SetCellValue(sheetName, col(22, row), d.Psych.E6)
			f.SetCellValue(sheetName, col(23, row), d.Psych.E7)
			f.SetCellValue(sheetName, col(24, row), d.Psych.E8)
			f.SetCellValue(sheetName, col(25, row), d.Psych.E9)
		}
	}
}

// setExcelColumnWidths sets predefined column widths for the export sheet.
func setExcelColumnWidths(f *excelize.File, sheetName string) {
	f.SetColWidth(sheetName, "A", "A", 30) // Nombre Curso
	f.SetColWidth(sheetName, "B", "B", 10) // Sección
	f.SetColWidth(sheetName, "C", "C", 15) // Semestre
	f.SetColWidth(sheetName, "D", "D", 28) // Nombre Alumno
	f.SetColWidth(sheetName, "E", "E", 28) // Correo
	f.SetColWidth(sheetName, "F", "F", 12) // RUT
	f.SetColWidth(sheetName, "G", "G", 16) // Carrera
	f.SetColWidth(sheetName, "H", "H", 8)  // Sexo
	f.SetColWidth(sheetName, "I", "I", 20) // Grupo
	f.SetColWidth(sheetName, "J", "J", 12) // Fuente Test
	f.SetColWidth(sheetName, "K", "K", 14) // Eneatipo/Tipo
	f.SetColWidth(sheetName, "L", "P", 10) // OCEAN (5 cols)
	f.SetColWidth(sheetName, "Q", "Y", 6)  // E1-E9
}
