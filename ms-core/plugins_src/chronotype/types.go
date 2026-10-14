package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChronotypeTest struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title       string               `bson:"title" json:"title"`
	Description string               `bson:"description" json:"description"`
	SectionID   string               `bson:"section_id" json:"section_id"`
	CreatedBy   string               `bson:"created_by" json:"created_by"`
	Questions   []ChronotypeQuestion `bson:"questions" json:"questions"`
	IsActive    bool                 `bson:"is_active" json:"is_active"`
	StartDate   *time.Time           `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time           `bson:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`
}

type ChronotypeQuestion struct {
	ID       int                `bson:"id" json:"id"`
	Text     string             `bson:"text" json:"text"`
	Options  []ChronotypeOption `bson:"options" json:"options"`
	Required bool               `bson:"required" json:"required"`
}

type ChronotypeOption struct {
	Score int    `bson:"score" json:"score"`
	Text  string `bson:"text" json:"text"`
}

type ChronotypeResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TestID      primitive.ObjectID `bson:"test_id" json:"test_id"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	Answers     []ChronotypeAnswer `bson:"answers" json:"answers"`
	Results     ChronotypeResults  `bson:"results" json:"results"`
	IsComplete  bool               `bson:"is_complete" json:"is_complete"`
	SubmittedAt *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type ChronotypeAnswer struct {
	QuestionID int `bson:"question_id" json:"question_id"`
	Score      int `bson:"score" json:"score"`
}

type ChronotypeResults struct {
	TotalScore int    `bson:"total_score" json:"total_score"`
	Category   string `bson:"category" json:"category"`
}
