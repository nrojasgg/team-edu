package models

// EmailTemplate represents an email template with its data
type EmailTemplate struct {
	Subject string
	Body    string
	IsHTML  bool
}

// PasswordResetData represents the data for password reset email template
type PasswordResetData struct {
	Name      string
	ResetURL  string
	ExpiresIn string
}

// TestReminderData represents the data for test reminder email template
type TestReminderData struct {
	Name     string
	TestName string
	TestURL  string
	DueDate  string
}
