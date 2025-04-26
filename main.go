package main

import (
	"go_mailer/api"
	"go_mailer/config"
	"go_mailer/scheduler"
	"log"
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
	log.Printf("🚀 Starting Go Mailer Service - %s", time.Now().Format("2006-01-02 15:04:05"))

	// Load environment variables
	loadEnvFile()

	// Load configuration
	log.Println("⚙️ Loading application configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	log.Println("✅ Configuration loaded successfully")

	// Create a scheduler instance
	emailScheduler := scheduler.New(cfg)

	// Start the scheduler
	emailScheduler.Start()

	// Set up graceful shutdown
	setupGracefulShutdown(emailScheduler)

	// Path to the email template
	templatePath := "tamplets/email_template.html"
	log.Printf("📝 Using email template: %s", templatePath)

	// Schedule emails from Google Sheet immediately
	log.Println("🔄 Initiating first Google Sheet check...")
	err = api.ScheduleEmailsFromGoogleSheet(emailScheduler, templatePath)
	if err != nil {
		log.Printf("❌ Error during initial scheduling from Google Sheet: %v", err)
	}

	// Set up a ticker to check for new entries every 2 hours
	checkInterval := 2 * time.Second // For testing, use seconds
	log.Printf("⏰ Setting up automatic checks every %v", checkInterval)

	ticker := time.NewTicker(checkInterval)
	go func() {
		for t := range ticker.C {
			log.Printf("🔄 Scheduled check at %s - Checking Google Sheet for new emails...",
				t.Format("2006-01-02 15:04:05"))
			err := api.ScheduleEmailsFromGoogleSheet(emailScheduler, templatePath)
			if err != nil {
				log.Printf("❌ Error scheduling emails from Google Sheet: %v", err)
			}
		}
	}()

	// Wait for scheduler to run
	log.Println("✅ Application running. Press Ctrl+C to exit.")
	select {}
}

func setupGracefulShutdown(emailScheduler *scheduler.Scheduler) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("🛑 Shutdown signal received")
		emailScheduler.Stop()
		log.Println("👋 Application shutdown complete")
		os.Exit(0)
	}()
}

func loadEnvFile() {
	log.Println("🔑 Loading environment variables...")
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Println("⚠️ Warning: Error loading .env file:", errEnv)
	} else {
		log.Println("✅ Environment variables loaded successfully")
	}
}
