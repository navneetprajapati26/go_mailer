package api

import (
	"go_mailer/config"
	"go_mailer/scheduler"
	"go_mailer/template"
	"log"
	"strings"
	"time"
)

// EmailCompletionCallback is a function that gets called when an email is sent successfully
type EmailCompletionCallback func(jobID string, email string) error

// getTemplatePath returns the path to the email template based on the template name
func getTemplatePath(templateName string) string {
	// Print raw template name for debugging
	log.Printf("🔍 Raw template name from API: '%s'", templateName)

	// Default to the standard template if no template is specified
	if templateName == "" {
		log.Printf("⚠️ Empty template name, using default template")
		return template.DefaultEmailTemplate
	}

	// Convert to lowercase and trim spaces for case-insensitive matching
	templateNameLower := strings.ToLower(strings.TrimSpace(templateName))
	log.Printf("🔄 Processed template name: '%s'", templateNameLower)

	var selectedTemplate string

	switch templateNameLower {
	case "normal":
		selectedTemplate = template.DefaultEmailTemplate
	case "casual":
		selectedTemplate = template.CasualEmailTemplate
	case "minimal":
		selectedTemplate = template.MinimalEmailTemplate
	default:
		log.Printf("⚠️ Unknown template name: '%s', using default template", templateNameLower)
		selectedTemplate = template.DefaultEmailTemplate
	}

	log.Printf("✅ Selected template path: %s", selectedTemplate)
	return selectedTemplate
}

// ScheduleEmailsFromGoogleSheet fetches data from Google Sheet and schedules emails for entries
// where SendStatus is false
func ScheduleEmailsFromGoogleSheet(emailScheduler *scheduler.Scheduler, cfg *config.Config) error {
	// Fetch data from Google Sheet API
	log.Println("🔄 Fetching data from Google Sheet API...")
	response, err := FetchGoogleSheetData(cfg)
	if err != nil {
		log.Printf("❌ Error fetching data from Google Sheet API: %v", err)
		return err
	}

	// Check if the request was successful
	if response.Status != "success" {
		log.Printf("⚠️ Google Sheet API returned non-success status: %s", response.Status)
		return nil
	}

	log.Printf("✅ Successfully fetched %d records from Google Sheet", len(response.Data))

	// Get all pending jobs to check if emails are already scheduled
	allJobs := emailScheduler.ListJobs()
	pendingEmails := make(map[string]bool)

	// Populate pendingEmails map with emails that are already scheduled but not sent
	pendingCount := 0
	for _, job := range allJobs {
		if job.Status == "pending" {
			pendingEmails[job.To] = true
			pendingCount++
		}
	}
	log.Printf("ℹ️ Found %d emails already scheduled and pending", pendingCount)

	// Track stats for logging
	skippedSent := 0
	skippedPending := 0
	scheduled := 0

	// Process each record
	for _, record := range response.Data {
		// Log the raw record for debugging
		log.Printf("🔍 Processing record: %+v", record)

		// Skip if SendStatus is true (already sent)
		if record.SendStatus {
			skippedSent++
			continue
		}

		// Skip if email is already scheduled and pending
		if pendingEmails[record.Email] {
			log.Printf("⏭️ Skipping %s (%s at %s) - already scheduled and pending",
				record.Email, record.EmployeeName, record.CompanyName)
			skippedPending++
			continue
		}

		// Create template data for the email with the updated structure
		data := template.TemplateData{
			RecipientName:   record.EmployeeName,
			CompanyName:     record.CompanyName,
			ApplyingForRoll: record.Roll,
		}

		// Combine date and time to create a full send time
		// Extract date components from SendAtDate
		year, month, day := record.SendAtDate.Date()

		// Extract time components from SendAtTime
		hour, min, sec := record.SendAtTime.Clock()

		// Create a new combined datetime
		combinedSendTime := time.Date(year, month, day, hour, min, sec, 0, time.Local)

		// Determine when to send the email
		var sendTime time.Time
		if time.Now().After(combinedSendTime) {
			// If combined time is in the past, schedule for immediate sending (1 minute from now)
			sendTime = time.Now().Add(1 * time.Minute)
			log.Printf("⏱️ Send time for %s is in the past (%s), rescheduling to %s",
				record.Email, combinedSendTime.Format("2006-01-02 15:04:05"),
				sendTime.Format("2006-01-02 15:04:05"))
		} else {
			sendTime = combinedSendTime
		}

		// Get the appropriate template path based on the template name in the record
		templatePath := getTemplatePath(record.TemplateName)
		log.Printf("📄 Using template: %s for email to %s", templatePath, record.Email)

		// Schedule the email
		subject := "Regarding " + record.Roll + " Position at " + record.CompanyName
		jobID := scheduleEmailWithCallback(emailScheduler, record.Email, subject, templatePath, data, sendTime, cfg)
		if jobID != "" {
			scheduled++
			log.Printf("📅 Scheduled email to %s (%s) at %s - Subject: %s",
				record.Email, record.EmployeeName, sendTime.Format("2006-01-02 15:04:05"), subject)
		}

		// Mark this email as pending to avoid scheduling it again in this batch
		pendingEmails[record.Email] = true
	}

	// Summary log
	log.Printf("📊 Summary: %d records processed, %d scheduled, %d skipped (already sent), %d skipped (already pending)",
		len(response.Data), scheduled, skippedSent, skippedPending)

	return nil
}

// scheduleEmailWithCallback schedules an email and sets up a callback function
// that will be called when the email is sent successfully
func scheduleEmailWithCallback(
	s *scheduler.Scheduler,
	to, subject, templatePath string,
	data template.TemplateData,
	sendTime time.Time,
	cfg *config.Config,
) string {
	// Schedule the email
	jobID, err := s.ScheduleEmail(to, subject, templatePath, data, sendTime)

	if err != nil {
		log.Printf("❌ Failed to schedule email to %s: %v", to, err)
		return ""
	}

	// Register the callback function
	s.RegisterCallback(jobID, func(successful bool) {
		if successful {
			// If email was sent successfully, update the Google Sheet
			log.Printf("✉️ Email sent successfully to %s, updating Google Sheet...", to)
			err := UpdateSendStatus(to, true, cfg)
			if err != nil {
				log.Printf("❌ Failed to update send status for %s: %v", to, err)
			} else {
				log.Printf("✅ Successfully updated send status for %s in Google Sheet", to)
			}
		} else {
			log.Printf("❌ Email to %s failed to send", to)
		}
	})

	return jobID
}
