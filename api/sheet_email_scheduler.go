package api

import (
	"go_mailer/scheduler"
	"go_mailer/template"
	"log"
	"time"
)

// EmailCompletionCallback is a function that gets called when an email is sent successfully
type EmailCompletionCallback func(jobID string, email string) error

// ScheduleEmailsFromGoogleSheet fetches data from Google Sheet and schedules emails for entries
// where SendStatus is false
func ScheduleEmailsFromGoogleSheet(emailScheduler *scheduler.Scheduler, templatePath string) error {
	// Fetch data from Google Sheet API
	response, err := FetchGoogleSheetData()
	if err != nil {
		return err
	}

	// Check if the request was successful
	if response.Status != "success" {
		log.Println("API returned non-success status:", response.Status)
		return nil
	}

	log.Printf("Fetched %d records from Google Sheet\n", len(response.Data))

	// Process each record
	for _, record := range response.Data {
		// Only schedule emails for records with SendStatus = false
		if !record.SendStatus {
			// Create template data for the email
			data := template.TemplateData{
				RecipientName:       record.EmployeeName,
				CompanyName:         record.CompanyName,
				SpecificArea:        "your area of expertise",
				SpecificAchievement: "your achievements",
				SpecificProject:     "your projects",
				RelevantSkill:       record.Roll, // Using the role as a relevant skill
				SenderName:          "HR Department",
			}

			// Determine when to send the email
			var sendTime time.Time
			if time.Now().After(record.SendAt) {
				// If SendAt time is in the past, schedule for immediate sending (2 minutes from now)
				sendTime = time.Now().Add(2 * time.Minute)
			} else {
				sendTime = record.SendAt
			}

			// Schedule the email
			subject := "Regarding " + record.Roll + " Position at " + record.CompanyName
			scheduleEmailWithCallback(emailScheduler, record.Email, subject, templatePath, data, sendTime)
		} else {
			log.Printf("Skipping record for %s as it already has SendStatus=true\n", record.Email)
		}
	}

	return nil
}

// scheduleEmailWithCallback schedules an email and sets up a callback function
// that will be called when the email is sent successfully
func scheduleEmailWithCallback(
	s *scheduler.Scheduler,
	to, subject, templatePath string,
	data template.TemplateData,
	sendTime time.Time,
) string {
	// Schedule the email
	jobID, err := s.ScheduleEmail(to, subject, templatePath, data, sendTime)

	if err != nil {
		log.Printf("Failed to schedule email: %v", err)
		return ""
	}

	log.Printf("Email scheduled with ID: %s to be sent at: %s to: %s",
		jobID, sendTime.Format(time.RFC1123), to)

	// Register the callback function
	s.RegisterCallback(jobID, func(successful bool) {
		if successful {
			// If email was sent successfully, update the Google Sheet
			err := UpdateSendStatus(to, true)
			if err != nil {
				log.Printf("Failed to update send status for %s: %v", to, err)
			} else {
				log.Printf("Successfully updated send status for %s", to)
			}
		}
	})

	return jobID
}
