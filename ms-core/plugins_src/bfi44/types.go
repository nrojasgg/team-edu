package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BFI44Test struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	SectionID   string             `bson:"section_id" json:"section_id"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	Questions   []BFI44Question    `bson:"questions" json:"questions"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	StartDate   *time.Time         `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time         `bson:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type BFI44Question struct {
	ID        int    `bson:"id" json:"id"`
	Text      string `bson:"text" json:"text"`
	Dimension string `bson:"dimension" json:"dimension"`
	IsReverse bool   `bson:"is_reverse" json:"is_reverse"`
	Required  bool   `bson:"required" json:"required"`
}

type BFI44Response struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TestID      primitive.ObjectID `bson:"test_id" json:"test_id"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	Answers     []BFI44Answer      `bson:"answers" json:"answers"`
	Results     BFI44Results       `bson:"results" json:"results"`
	IsComplete  bool               `bson:"is_complete" json:"is_complete"`
	SubmittedAt *time.Time         `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type BFI44Answer struct {
	QuestionID int `bson:"question_id" json:"question_id"`
	Score      int `bson:"score" json:"score"`
}

type BFI44Results struct {
	Extraversion      float64 `bson:"extraversion" json:"extraversion"`
	Agreeableness     float64 `bson:"agreeableness" json:"agreeableness"`
	Conscientiousness float64 `bson:"conscientiousness" json:"conscientiousness"`
	Neuroticism       float64 `bson:"neuroticism" json:"neuroticism"`
	Openness          float64 `bson:"openness" json:"openness"`

	OCEANVector OCEANVectorInline `bson:"ocean_vector" json:"ocean_vector"`

	CorrectedScores map[int]int `bson:"corrected_scores" json:"corrected_scores"`

	MissingItems map[string][]int `bson:"missing_items,omitempty" json:"missing_items,omitempty"`
	ImputedItems map[string][]int `bson:"imputed_items,omitempty" json:"imputed_items,omitempty"`
}

type OCEANVectorInline struct {
	Openness          float64 `bson:"openness" json:"openness"`
	Conscientiousness float64 `bson:"conscientiousness" json:"conscientiousness"`
	Extraversion      float64 `bson:"extraversion" json:"extraversion"`
	Agreeableness     float64 `bson:"agreeableness" json:"agreeableness"`
	Neuroticism       float64 `bson:"neuroticism" json:"neuroticism"`
}
