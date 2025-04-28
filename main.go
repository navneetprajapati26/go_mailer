package main

import (
	"go_mailer/api"
	"go_mailer/config"
	"go_mailer/logger"
	"go_mailer/scheduler"
	"go_mailer/template"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	// Set up initial log message with timestamp
	logger.Info("🚀 Starting Go Mailer Service - %s", time.Now().Format("2006-01-02 15:04:05"))

	// Load environment variables
	loadEnvFile()

	// Load configuration
	logger.Info("⚙️ Loading application configuration...")
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("❌ Failed to load configuration: %v", err)
	}
	logger.Info("✅ Configuration loaded successfully")

	// Create a scheduler instance
	emailScheduler := scheduler.New(cfg)

	// Start the scheduler
	emailScheduler.Start()

	// Set up graceful shutdown
	setupGracefulShutdown(emailScheduler)

	// Schedule emails from Google Sheet immediately
	logger.Info("🔄 Initiating first Google Sheet check...")
	err = api.ScheduleEmailsFromGoogleSheet(emailScheduler, cfg)
	if err != nil {
		logger.Error("❌ Error during initial scheduling from Google Sheet: %v", err)
	}

	// Set up a ticker to check for new entries every 2 hours
	checkInterval := 2 * time.Hour // For testing, use seconds
	logger.Info("⏰ Setting up automatic checks every %v", checkInterval)

	ticker := time.NewTicker(checkInterval)
	go func() {
		for t := range ticker.C {
			logger.Info("🔄 Scheduled check at %s - Checking Google Sheet for new emails...",
				t.Format("2006-01-02 15:04:05"))
			err := api.ScheduleEmailsFromGoogleSheet(emailScheduler, cfg)
			if err != nil {
				logger.Error("❌ Error scheduling emails from Google Sheet: %v", err)
			}
		}
	}()

	// Wait for scheduler to run
	logger.Info("✅ Application running. Press Ctrl+C to exit.")
	logger.Info("🔍 Available templates:")
	logger.Info("   - Default: %s", template.DefaultEmailTemplate)
	logger.Info("   - Casual: %s", template.CasualEmailTemplate)
	logger.Info("   - Minimal: %s", template.MinimalEmailTemplate)
	select {}
}

func setupGracefulShutdown(emailScheduler *scheduler.Scheduler) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("🛑 Shutdown signal received")
		emailScheduler.Stop()
		logger.Info("👋 Application shutdown complete")
		os.Exit(0)
	}()
}

func loadEnvFile() {
	logger.Info("🔑 Loading environment variables...")
	errEnv := godotenv.Load()
	if errEnv != nil {
		logger.Warning("⚠️ Warning: Error loading .env file: %v", errEnv)
	} else {
		logger.Info("✅ Environment variables loaded successfully")
	}
}
