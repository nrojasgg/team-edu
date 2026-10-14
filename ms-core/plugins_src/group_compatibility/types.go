package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SociometricQuestion struct {
	ID   int    `bson:"id"   json:"id"`
	Text string `bson:"text" json:"text"`
}

var defaultSociometricQuestions = []SociometricQuestion{
	{ID: 1, Text: "¿A quién(es) elegirías para trabajar en equipo?"},
	{ID: 2, Text: "¿Quién(es) crees que te elegirían para trabajar en equipo?"},
	{ID: 3, Text: "¿A quién(es) NO elegirías para trabajar en equipo?"},
	{ID: 4, Text: "¿Quién(es) crees que NO te elegirían para trabajar en equipo?"},
}

type GroupCompatibilitySurvey struct {
	ID          primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	Title       string                `bson:"title" json:"title"`
	Description string                `bson:"description" json:"description"`
	SectionID   string                `bson:"section_id" json:"section_id"`
	CreatedBy   string                `bson:"created_by" json:"created_by"`
	SurveyType  string                `bson:"survey_type" json:"survey_type"`
	Questions   []SociometricQuestion `bson:"questions" json:"questions"`
	IsActive    bool                  `bson:"is_active" json:"is_active"`
	StartDate   *time.Time            `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time            `bson:"end_date,omitempty" json:"end_date,omitempty"`
	CreatedAt   time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time             `bson:"updated_at" json:"updated_at"`
}

type QuestionAnswer struct {
	StudentID string `bson:"student_id" json:"student_id"`
	Response  int    `bson:"response" json:"response"`
}

type GroupCompatibilityResponse struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SurveyID    primitive.ObjectID `bson:"survey_id" json:"survey_id"`
	StudentID   string             `bson:"student_id" json:"student_id"`
	SectionID   string             `bson:"section_id" json:"section_id"`
	SurveyType  string             `bson:"survey_type" json:"survey_type"`
	Question1   []QuestionAnswer   `bson:"question1" json:"question1"`
	Question2   []QuestionAnswer   `bson:"question2" json:"question2"`
	Question3   []QuestionAnswer   `bson:"question3" json:"question3"`
	Question4   []QuestionAnswer   `bson:"question4" json:"question4"`
	SubmittedAt time.Time          `bson:"submitted_at" json:"submitted_at"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

const (
	errSurveyIDRequired        = "survey_id is required"
	errInvalidSurveyIDFmt      = "invalid survey_id: %v"
	errFailedGetRespCollection = "failed to get responses collection"
)
