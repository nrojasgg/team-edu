package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnneagramTest struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty"          json:"id"`
	Title       string              `bson:"title"                  json:"title"`
	Description string              `bson:"description"            json:"description"`
	SectionID   string              `bson:"section_id"             json:"section_id"`
	CreatedBy   string              `bson:"created_by"             json:"created_by"`
	Questions   []EnneagramQuestion `bson:"questions"              json:"questions"`
	IsActive    bool                `bson:"is_active"              json:"is_active"`
	StartDate   *time.Time          `bson:"start_date,omitempty"   json:"start_date,omitempty"`
	EndDate     *time.Time          `bson:"end_date,omitempty"     json:"end_date,omitempty"`
	CreatedAt   time.Time           `bson:"created_at"             json:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at"             json:"updated_at"`
}

type EnneagramQuestion struct {
	ID       int    `bson:"id"       json:"id"`
	Text     string `bson:"text"     json:"text"`
	Eneatype int    `bson:"eneatype" json:"eneatype"`
	Required bool   `bson:"required" json:"required"`
}

type EnneagramResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
	TestID      primitive.ObjectID `bson:"test_id"              json:"test_id"`
	StudentID   string             `bson:"student_id"           json:"student_id"`
	Answers     []EnneagramAnswer  `bson:"answers"              json:"answers"`
	Results     *EnneagramResults  `bson:"results,omitempty"    json:"results,omitempty"`
	IsComplete  bool               `bson:"is_complete"          json:"is_complete"`
	SubmittedAt *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"           json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"           json:"updated_at"`
}

type EnneagramAnswer struct {
	QuestionID int `bson:"question_id" json:"question_id"`
	Score      int `bson:"score"       json:"score"`
}

type EnneagramResults struct {
	E1           float64            `bson:"e1"            json:"e1"`
	E2           float64            `bson:"e2"            json:"e2"`
	E3           float64            `bson:"e3"            json:"e3"`
	E4           float64            `bson:"e4"            json:"e4"`
	E5           float64            `bson:"e5"            json:"e5"`
	E6           float64            `bson:"e6"            json:"e6"`
	E7           float64            `bson:"e7"            json:"e7"`
	E8           float64            `bson:"e8"            json:"e8"`
	E9           float64            `bson:"e9"            json:"e9"`
	DominantType int                `bson:"dominant_type" json:"dominant_type"`
	MaxScore     float64            `bson:"max_score"     json:"max_score"`
	OCEANVector  map[string]float64 `bson:"ocean_vector,omitempty" json:"ocean_vector,omitempty"`
}
