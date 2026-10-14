package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OCEANVector represents the Big Five personality traits
type OCEANVector struct {
	Openness          float64 `bson:"o" json:"openness"`
	Conscientiousness float64 `bson:"c" json:"conscientiousness"`
	Extraversion      float64 `bson:"e" json:"extraversion"`
	Agreeableness     float64 `bson:"a" json:"agreeableness"`
	Neuroticism       float64 `bson:"n" json:"neuroticism"`
}

// StudentPsychometrics stores pre-calculated OCEAN vectors for students
// This collection serves as a cache to avoid recalculating personality vectors
// on every team formation request
type StudentPsychometrics struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	StudentID  string                 `bson:"student_id" json:"student_id"`
	SectionID  string                 `bson:"section_id" json:"section_id"`
	OCEAN      OCEANVector            `bson:"ocean" json:"ocean"`
	SourceType string                 `bson:"source_type" json:"source_type"` // "enneagram", "mbti", "manual"
	SourceData map[string]interface{} `bson:"source_data" json:"source_data"`
	Sociometry map[string]float64     `bson:"sociometry,omitempty" json:"sociometry,omitempty"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time              `bson:"updated_at" json:"updated_at"`
}

// GetStudentPsychometrics retrieves psychometric data for a student in a section
func GetStudentPsychometrics(ctx context.Context, studentID, sectionID string) (*StudentPsychometrics, error) {
	collection := GetCollection("student_psychometrics")

	filter := bson.M{
		"student_id": studentID,
		"section_id": sectionID,
	}

	var result StudentPsychometrics
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no psychometric data found for student %s in section %s", studentID, sectionID)
		}
		return nil, err
	}

	return &result, nil
}

// GetSectionPsychometrics retrieves all psychometric data for students in a section
func GetSectionPsychometrics(ctx context.Context, sectionID string) ([]StudentPsychometrics, error) {
	collection := GetCollection("student_psychometrics")

	filter := bson.M{"section_id": sectionID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []StudentPsychometrics
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// UpsertStudentPsychometrics creates or updates psychometric data
func UpsertStudentPsychometrics(ctx context.Context, data *StudentPsychometrics) error {
	collection := GetCollection("student_psychometrics")

	data.UpdatedAt = time.Now()
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}

	filter := bson.M{
		"student_id": data.StudentID,
		"section_id": data.SectionID,
	}

	update := bson.M{
		"$set": data,
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// BulkUpsertPsychometrics efficiently inserts or updates multiple psychometric records
func BulkUpsertPsychometrics(ctx context.Context, data []StudentPsychometrics) error {
	if len(data) == 0 {
		return nil
	}

	collection := GetCollection("student_psychometrics")

	models := make([]mongo.WriteModel, len(data))
	for i, record := range data {
		record.UpdatedAt = time.Now()
		if record.CreatedAt.IsZero() {
			record.CreatedAt = time.Now()
		}

		filter := bson.M{
			"student_id": record.StudentID,
			"section_id": record.SectionID,
		}

		update := bson.M{
			"$set": record,
		}

		models[i] = mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, models, opts)
	return err
}

// DeleteStudentPsychometrics removes psychometric data for a student in a section
func DeleteStudentPsychometrics(ctx context.Context, studentID, sectionID string) error {
	collection := GetCollection("student_psychometrics")

	filter := bson.M{
		"student_id": studentID,
		"section_id": sectionID,
	}

	_, err := collection.DeleteOne(ctx, filter)
	return err
}

// CreatePsychometricsIndexes creates necessary indexes for student_psychometrics collection
func CreatePsychometricsIndexes(ctx context.Context) error {
	collection := GetCollection("student_psychometrics")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "student_id", Value: 1},
				{Key: "section_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "section_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "source_type", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}
