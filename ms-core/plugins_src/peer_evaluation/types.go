package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupSnapshot struct {
	GroupID     string   `bson:"group_id" json:"group_id"`
	GroupName   string   `bson:"group_name" json:"group_name"`
	GroupNumber int      `bson:"group_number" json:"group_number"`
	StudentIDs  []string `bson:"student_ids" json:"student_ids"`
}

type PeerEvaluationSurvey struct {
	ID             primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
	Title          string                   `bson:"title" json:"title"`
	Description    string                   `bson:"description" json:"description"`
	SectionID      string                   `bson:"section_id" json:"section_id"`
	CreatedBy      string                   `bson:"created_by" json:"created_by"`
	Questions      []PeerEvaluationQuestion `bson:"questions" json:"questions"`
	IsActive       bool                     `bson:"is_active" json:"is_active"`
	StartDate      *time.Time               `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate        *time.Time               `bson:"end_date,omitempty" json:"end_date,omitempty"`
	GroupSnapshots []GroupSnapshot          `bson:"group_snapshots,omitempty" json:"group_snapshots,omitempty"`
	CreatedAt      time.Time                `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time                `bson:"updated_at" json:"updated_at"`
}

type PeerEvaluationQuestion struct {
	ID       int    `bson:"id" json:"id"`
	Text     string `bson:"text" json:"text"`
	Category string `bson:"category" json:"category"`
	Required bool   `bson:"required" json:"required"`
}

type PeerEvaluationResponse struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	SurveyID    primitive.ObjectID     `bson:"survey_id" json:"survey_id"`
	EvaluatorID string                 `bson:"evaluator_id" json:"evaluator_id"`
	EvaluatedID string                 `bson:"evaluated_id" json:"evaluated_id"`
	GroupID     string                 `bson:"group_id" json:"group_id"`
	Responses   []PeerEvaluationAnswer `bson:"responses" json:"responses"`
	IsComplete  bool                   `bson:"is_complete" json:"is_complete"`
	SubmittedAt *time.Time             `bson:"submitted_at,omitempty" json:"submitted_at,omitempty"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

type PeerEvaluationAnswer struct {
	QuestionID int `bson:"question_id" json:"question_id"`
	Score      int `bson:"score" json:"score"`
}

type StudentSummaryData struct {
	RUT       string
	Name      string
	Lastname  string
	GroupName string
	Scores    [5][]int
	Count     int
}

type SectionExcelInfo struct {
	CourseName  string
	SectionCode string
	Semester    string
}

type ExcelStyles struct {
	Header int
	Cell   int
}

type ResponseSummary struct {
	StudentID       string `bson:"student_id"`
	GroupID         string `bson:"group_id"`
	NumberResponses int    `bson:"number_responses"`
}
