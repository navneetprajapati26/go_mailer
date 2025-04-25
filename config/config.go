package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	SenderEmail string
	Password    string
	SMTPHost    string
	SMTPPort    string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	senderEmail := os.Getenv("SENDER_MAIL_ID")
	password := os.Getenv("PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

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
		SenderEmail: senderEmail,
		Password:    password,
		SMTPHost:    smtpHost,
		SMTPPort:    smtpPort,
	}, nil
}

// SMTPAddress returns the full SMTP server address (host:port)
func (c *Config) SMTPAddress() string {
	return fmt.Sprintf("%s:%s", c.SMTPHost, c.SMTPPort)
}
