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
	// Load environment variables
	loadEnvFile()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a scheduler instance
	emailScheduler := scheduler.New(cfg)

	// Start the scheduler
	emailScheduler.Start()

	// Set up graceful shutdown
	setupGracefulShutdown(emailScheduler)

	// Path to the email template
	templatePath := "tamplets/email_template.html"

	// Schedule emails from Google Sheet immediately
	log.Println("Scheduling emails from Google Sheet data...")
	err = api.ScheduleEmailsFromGoogleSheet(emailScheduler, templatePath)
	if err != nil {
		log.Printf("Error scheduling emails from Google Sheet: %v", err)
	}

	// Set up a ticker to check for new entries every 2 hours
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for range ticker.C {
			log.Println("Checking Google Sheet for new emails to schedule...")
			err := api.ScheduleEmailsFromGoogleSheet(emailScheduler, templatePath)
			if err != nil {
				log.Printf("Error scheduling emails from Google Sheet: %v", err)
			}
		}
	}()

	// Wait for scheduler to run
	log.Println("Application running. Checking Google Sheet every 2 hours. Press Ctrl+C to exit.")
	select {}
}

func setupGracefulShutdown(emailScheduler *scheduler.Scheduler) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutdown signal received")
		emailScheduler.Stop()
		os.Exit(0)
	}()
}

func loadEnvFile() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Println("Warning: Error loading .env file:", errEnv)
	}
}
