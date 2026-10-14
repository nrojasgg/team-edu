package main

import (
	"time"

	"ms-core/internal/core"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MyTest struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	SectionID   string             `bson:"section_id" json:"section_id"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	Questions   []MyQuestion       `bson:"questions" json:"questions"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	StartDate   *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type MyQuestion struct {
	ID        int    `bson:"id" json:"id"`
	Text      string `bson:"text" json:"text"`
	Dimension string `bson:"dimension" json:"dimension"`
	IsReverse bool   `bson:"is_reverse" json:"is_reverse"`
	Required  bool   `bson:"required" json:"required"`
}

type MyResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TestID      primitive.ObjectID `bson:"test_id" json:"test_id"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	Answers     []MyAnswer         `bson:"answers" json:"answers"`
	Results     MyResults          `bson:"results" json:"results"`
	IsComplete  bool               `bson:"is_complete" json:"is_complete"`
	SubmittedAt *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type MyAnswer struct {
	QuestionID int `bson:"question_id" json:"question_id"`
	Score      int `bson:"score" json:"score"`
}

type MyResults struct {
	Extraversion      float64          `bson:"extraversion" json:"extraversion"`
	Agreeableness     float64          `bson:"agreeableness" json:"agreeableness"`
	Conscientiousness float64          `bson:"conscientiousness" json:"conscientiousness"`
	Neuroticism       float64          `bson:"neuroticism" json:"neuroticism"`
	Openness          float64          `bson:"openness" json:"openness"`
	OCEANVector       core.OCEANVector `bson:"ocean_vector" json:"ocean_vector"`
	Category          string           `bson:"category" json:"category"`
}
