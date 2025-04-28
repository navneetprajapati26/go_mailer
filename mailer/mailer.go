package mailer

import (
	"fmt"
	"go_mailer/config"
	"go_mailer/logger"
	"go_mailer/template"
	"net/smtp"
	"os"
)

// Mailer handles sending emails using templates
type Mailer struct {
	config *config.Config
}

// New creates a new Mailer instance
func New(cfg *config.Config) *Mailer {
	return &Mailer{
		config: cfg,
	}
}

// SendWithTemplate sends an email with dynamically populated HTML template
func (m *Mailer) SendWithTemplate(to string, subject string, htmlFilePath string, templateData template.TemplateData) error {
	// Process the template with the provided data
	processedHTML, err := template.Process(htmlFilePath, templateData)
	if err != nil {
		return fmt.Errorf("template processing error: %w", err)
	}

	// Create proper email with MIME headers
	header := make(map[string]string)
	header["From"] = m.config.SenderEmail
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""

	// Construct message with proper headers
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + processedHTML

	err = smtp.SendMail(
		m.config.SMTPAddress(),
		smtp.PlainAuth("", m.config.SenderEmail, m.config.Password, m.config.SMTPHost),
		m.config.SenderEmail,
		[]string{to},
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("smtp error: %w", err)
	}

	logger.Info("Email sent successfully to %s", to)
	return nil
}

// Send is kept for backward compatibility
func Send(to string, subject string, htmlFilePath string) {
	from := os.Getenv("SENDER_MAIL_ID")
	pass := os.Getenv("PASSWORD")

	logger.Info("From and Pass from ENV: %s %s", from, pass)

	htmlContent, err := os.ReadFile(htmlFilePath)
	if err != nil {
		logger.Error("error reading HTML file: %s", err)
		return
	}
	// Create proper email with MIME headers
	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""

	// Construct message with proper headers
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + string(htmlContent)

	err = smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(message))

	if err != nil {
		logger.Error("smtp error: %s", err)
		return
	}

	logger.Info("sent, visit http://foobarbazz.mailinator.com")
}
