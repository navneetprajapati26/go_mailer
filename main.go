package main

import (
	"flag"
	"go_mailer/api"
	"go_mailer/config"
	"go_mailer/scheduler"
	"go_mailer/template"
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
	// Define command-line flags
	fetchSheetData := flag.Bool("fetch-sheet", false, "Fetch data from Google Sheet API")
	flag.Parse()

	// If the fetch-sheet flag is provided, run the example and exit
	if *fetchSheetData {
		api.ExampleFetchGoogleSheetData()
		return
	}

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

	// Example usage
	recipientEmail := "recipient@example.com"
	subject := "Flutter Developer Application"

	// Example template data
	data := template.TemplateData{
		RecipientName:       "John Doe",
		CompanyName:         "TechCorp ( start at: " + time.Now().Format("2006-01-02 15:04:05") + "send at: " + time.Now().Add(5*time.Minute).Format("2006-01-02 15:04:05") + ")",
		SpecificArea:        "mobile app development",
		SpecificAchievement: "user-centric design",
		SpecificProject:     "the mobile banking platform",
		RelevantSkill:       "Flutter architecture",
		SenderName:          "HR Department",
	}

	// Schedule the email for 5 minutes from now
	scheduleEmail(emailScheduler, recipientEmail, subject, "tamplets/email_template.html", data, time.Now().Add(5*time.Minute))

	// Wait for scheduler to run
	log.Println("Application running. Press Ctrl+C to exit.")
	select {}
}

// scheduleEmail is a simple function to schedule an email with the given parameters
func scheduleEmail(s *scheduler.Scheduler, to, subject, templatePath string, data template.TemplateData, sendTime time.Time) {
	// Schedule the email
	jobID, err := s.ScheduleEmail(to, subject, templatePath, data, sendTime)

	if err != nil {
		log.Printf("Failed to schedule email: %v", err)
		return
	}

	log.Printf("Email scheduled with ID: %s to be sent at: %s",
		jobID, sendTime.Format(time.RFC1123))
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
