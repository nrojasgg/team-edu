package main

import (
	"time"

	"ms-core/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MBTIDimension string

const (
	DimensionEI MBTIDimension = "EI"
	DimensionSN MBTIDimension = "SN"
	DimensionTF MBTIDimension = "TF"
	DimensionJP MBTIDimension = "JP"
)

type MBTIQuestion struct {
	ID        int           `bson:"id" json:"id"`
	Dimension MBTIDimension `bson:"dimension" json:"dimension"`
	Statement string        `bson:"statement" json:"statement"`
	OptionA   string        `bson:"option_a" json:"option_a"`
	OptionB   string        `bson:"option_b" json:"option_b"`
	ValueA    string        `bson:"value_a" json:"value_a"`
	ValueB    string        `bson:"value_b" json:"value_b"`
	Required  bool          `bson:"required" json:"required"`
}

type MBTIAnswer struct {
	QuestionID int    `bson:"question_id" json:"question_id"`
	Selection  string `bson:"selection" json:"selection"`
}

type MBTIDimensionResult struct {
	Name       MBTIDimension `bson:"name" json:"name"`
	ScoreA     int           `bson:"score_a" json:"score_a"`
	ScoreB     int           `bson:"score_b" json:"score_b"`
	Preference string        `bson:"preference" json:"preference"`
	PCI        float64       `bson:"pci" json:"pci"`
}

type MBTIResult struct {
	MBTIType    string                `bson:"mbti_type" json:"mbti_type"`
	Dimensions  []MBTIDimensionResult `bson:"dimensions" json:"dimensions"`
	TotalScore  int                   `bson:"total_score" json:"total_score"`
	OCEANVector domain.OCEANVector    `bson:"ocean_vector" json:"ocean_vector"`
}

type MBTITest struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	SectionID   string             `bson:"section_id" json:"section_id"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	Questions   []MBTIQuestion     `bson:"questions" json:"questions"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	StartDate   *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type MBTIResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TestID      primitive.ObjectID `bson:"test_id" json:"test_id"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	Answers     []MBTIAnswer       `bson:"answers" json:"answers"`
	Results     MBTIResult         `bson:"results" json:"results"`
	IsComplete  bool               `bson:"is_complete" json:"is_complete"`
	SubmittedAt *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
