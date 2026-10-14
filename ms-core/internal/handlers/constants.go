package handlers

// HTTP header constants
const (
	headerContentType = "Content-Type"
	mimeJSON          = "application/json"
)

// Common error messages shared across handlers
const (
	errInvalidRequestBody = "Invalid request body"
	errTestIDRequired     = "test_id is required"
	errSurveyIDRequired   = "Survey ID is required"
	errSectionIDRequired  = "section_id is required"
	errInvalidSectionID   = "Invalid section ID"
	errInvalidRequest     = "Invalid request"
)
