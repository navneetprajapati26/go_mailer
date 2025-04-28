package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	SenderEmail      string
	Password         string
	SMTPHost         string
	SMTPPort         string
	GOOGEL_SHEET_API string
	InputTimezone    string // Timezone for input times (e.g., "Asia/Kolkata")
	ServerTimezone   string // Timezone where server is running (e.g., "Asia/Singapore")
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	senderEmail := os.Getenv("SENDER_MAIL_ID")
	password := os.Getenv("PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	googelSheetApi := os.Getenv("GOOGEL_SHEET_API")

	// Set defaults if not provided
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}

	// Validate required fields
	if senderEmail == "" || password == "" {
		return nil, fmt.Errorf("SENDER_MAIL_ID and PASSWORD environment variables must be set")
	}

	return &Config{
		SenderEmail:      senderEmail,
		Password:         password,
		SMTPHost:         smtpHost,
		SMTPPort:         smtpPort,
		GOOGEL_SHEET_API: googelSheetApi,
	}, nil
}

// SMTPAddress returns the full SMTP server address (host:port)
func (c *Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%s", c.SMTPHost, c.SMTPPort)
}
