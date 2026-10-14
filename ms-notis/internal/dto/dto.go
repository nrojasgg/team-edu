package dto

// ApiResponse represents a successful API response
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// PasswordResetRequest represents a request to send a password reset email
type PasswordResetRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Name       string `json:"name" validate:"required"`
	ResetToken string `json:"reset_token" validate:"required"`
	ResetURL   string `json:"reset_url" validate:"required,url"`
}

// TestReminderRequest represents a request to send a test reminder email
type TestReminderRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	TestName string `json:"test_name" validate:"required"`
	TestURL  string `json:"test_url" validate:"required,url"`
	DueDate  string `json:"due_date,omitempty"`
}
