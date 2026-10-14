package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"ms-notis/internal/config"
	"ms-notis/internal/models"
	"path/filepath"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailService interface {
	SendPasswordResetEmail(to, name, resetToken, resetURL string) error
	SendTestReminderEmail(to, name, testName, testURL, dueDate string) error
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

// SendPasswordResetEmail sends a password reset email
func (s *emailService) SendPasswordResetEmail(to, name, resetToken, resetURL string) error {
	log.Printf("📧 SendPasswordResetEmail called for: %s", to)

	// Build complete reset URL with token
	completeResetURL := fmt.Sprintf("%s?token=%s", resetURL, resetToken)
	log.Printf("Reset URL: %s", completeResetURL)

	data := models.PasswordResetData{
		Name:      name,
		ResetURL:  completeResetURL,
		ExpiresIn: "1 hora",
	}

	subject := "Restablecer tu Contraseña - TeamEdu"
	body, err := s.renderTemplate("password_reset.html", data)
	if err != nil {
		log.Printf("❌ Error rendering template: %v", err)
		if s.cfg.AppEnv == "development" {
			log.Printf("\n--- [MODO DESARROLLO] TEMPLATE ERROR ---")
			log.Printf("Para: %s", to)
			log.Printf("Reset URL: %s", completeResetURL)
			log.Printf("------------------------------------------\n")
			return nil
		}
		return fmt.Errorf("error rendering template: %w", err)
	}

	return s.sendEmail(to, subject, body, true)
}

// SendTestReminderEmail sends a test reminder email
func (s *emailService) SendTestReminderEmail(to, name, testName, testURL, dueDate string) error {
	data := models.TestReminderData{
		Name:     name,
		TestName: testName,
		TestURL:  testURL,
		DueDate:  dueDate,
	}

	subject := fmt.Sprintf("Recordatorio: %s - TeamEdu", testName)
	body, err := s.renderTemplate("test_reminder.html", data)
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}

	return s.sendEmail(to, subject, body, true)
}

// renderTemplate renders an HTML email template
func (s *emailService) renderTemplate(templateName string, data interface{}) (string, error) {
	templatePath := filepath.Join("templates", templateName)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buf.String(), nil
}

// sendEmail sends an email via SMTP
func (s *emailService) sendEmail(to, subject, body string, isHTML bool) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.cfg.SMTPFromName, s.cfg.SMTPFromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}
	m.SetBody(contentType, body)

	port, err := strconv.Atoi(s.cfg.SMTPPort)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}

	d := gomail.NewDialer(s.cfg.SMTPHost, port, s.cfg.SMTPUsername, s.cfg.SMTPPassword)

	// For development, allow insecure connections
	if s.cfg.AppEnv == "development" {
		// d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ Error enviando email a %s: %v", to, err)

		// En desarrollo, si falla el SMTP, imprimimos el contenido para que el desarrollador pueda seguir probando
		if s.cfg.AppEnv == "development" {
			log.Printf("\n--- [MODO DESARROLLO] EMAIL NO ENVIADO (SMTP) ---")
			log.Printf("Para: %s", to)
			log.Printf("Asunto: %s", subject)
			log.Printf("--------------------------------------------------\n")
			return nil // No devolvemos error para no bloquear el flujo del frontend
		}

		return fmt.Errorf("error sending email: %w", err)
	}

	log.Printf("✅ Email sent successfully to %s\n", to)
	return nil
}
