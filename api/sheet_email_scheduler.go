package api

import (
	"go_mailer/config"
	"go_mailer/logger"
	"go_mailer/scheduler"
	"go_mailer/template"
	"strings"
	"time"
)

// EmailCompletionCallback is a function that gets called when an email is sent successfully
type EmailCompletionCallback func(jobID string, email string) error

// getTemplatePath returns the path to the email template based on the template name
func getTemplatePath(templateName string) string {
	// Print raw template name for debugging
	logger.Debug("🔍 Raw template name from API: '%s'", templateName)

	// Default to the standard template if no template is specified
	if templateName == "" {
		logger.Warning("⚠️ Empty template name, using default template")
		return template.DefaultEmailTemplate
	}

	// Convert to lowercase and trim spaces for case-insensitive matching
	templateNameLower := strings.ToLower(strings.TrimSpace(templateName))
	logger.Debug("🔄 Processed template name: '%s'", templateNameLower)

	var selectedTemplate string

	switch templateNameLower {
	case "normal":
		selectedTemplate = template.DefaultEmailTemplate
	case "casual":
		selectedTemplate = template.CasualEmailTemplate
	case "minimal":
		selectedTemplate = template.MinimalEmailTemplate
	default:
		logger.Warning("⚠️ Unknown template name: '%s', using default template", templateNameLower)
		selectedTemplate = template.DefaultEmailTemplate
	}

	logger.Debug("✅ Selected template path: %s", selectedTemplate)
	return selectedTemplate
}

// ScheduleEmailsFromGoogleSheet fetches data from Google Sheet and schedules emails for entries
// where SendStatus is false
func ScheduleEmailsFromGoogleSheet(emailScheduler *scheduler.Scheduler, cfg *config.Config) error {
	// Fetch data from Google Sheet API
	logger.Info("🔄 Fetching data from Google Sheet API...")
	response, err := FetchGoogleSheetData(cfg)
	if err != nil {
		logger.Error("❌ Error fetching data from Google Sheet API: %v", err)
		return err
	}

	// Check if the request was successful
	if response.Status != "success" {
		logger.Warning("⚠️ Google Sheet API returned non-success status: %s", response.Status)
		return nil
	}

	logger.Info("✅ Successfully fetched %d records from Google Sheet", len(response.Data))

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
	logger.Info("ℹ️ Found %d emails already scheduled and pending", pendingCount)

	// Track stats for logging
	skippedSent := 0
	skippedPending := 0
	scheduled := 0

	// Process each record
	for _, record := range response.Data {
		// Log the raw record for debugging
		logger.Debug("🔍 Processing record: %+v", record)

		// Skip if SendStatus is true (already sent)
		if record.SendStatus {
			skippedSent++
			continue
		}

		// Skip if email is already scheduled and pending
		if pendingEmails[record.Email] {
			logger.Info("⏭️ Skipping %s (%s at %s) - already scheduled and pending",
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

		// Get IST location
		ist := time.FixedZone("IST", 5*60*60+30*60)

		// Convert SendAtDate to IST
		sendAtDateIST := record.SendAtDate.In(ist)
		year, month, day := sendAtDateIST.Date()

		// Extract time components from SendAtTime
		hour, min, sec := record.SendAtTime.Clock()

		// Validate if SendAtTime is not the default value (1899-12-30)
		if record.SendAtTime.Year() == 1899 && record.SendAtTime.Month() == 12 && record.SendAtTime.Day() == 30 {
			// Use default time of 00:00:00 if SendAtTime is default
			hour, min, sec = 0, 0, 0
		}

		// Create the combined time in IST
		combinedSendTime := time.Date(year, month, day, hour, min, sec, 0, ist)

		// Determine when to send the email
		var sendTime time.Time
		if time.Now().In(ist).After(combinedSendTime) {
			// If combined time is in the past, schedule for immediate sending (1 minute from now)
			sendTime = time.Now().In(ist).Add(time.Minute)
			logger.Info("⏱️ Send time for %s is in the past (%s), rescheduling to %s", record.Email, combinedSendTime.Format("2006-01-02 15:04:05 MST"), sendTime.Format("2006-01-02 15:04:05 MST"))
		} else {
			sendTime = combinedSendTime
		}

		// Get the appropriate template path based on the template name in the record
		templatePath := getTemplatePath(record.TemplateName)
		logger.Debug("📄 Using template: %s for email to %s", templatePath, record.Email)

		// Schedule the email
		subject := "Regarding " + record.Roll + " Position at " + record.CompanyName
		jobID := scheduleEmailWithCallback(emailScheduler, record.Email, subject, templatePath, data, sendTime, cfg)
		if jobID != "" {
			scheduled++
			logger.Info("📅 Scheduled email to %s (%s) at %s IST - Subject: %s", record.Email, record.EmployeeName, sendTime, subject)
		}

		// Mark this email as pending to avoid scheduling it again in this batch
		pendingEmails[record.Email] = true
	}

	// Summary log
	logger.Info("📊 Summary: %d records processed, %d scheduled, %d skipped (already sent), %d skipped (already pending)", len(response.Data), scheduled, skippedSent, skippedPending)

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
		logger.Error("❌ Failed to schedule email to %s: %v", to, err)
		return ""
	}

	// Register the callback function
	s.RegisterCallback(jobID, func(successful bool) {
		if successful {
			// If email was sent successfully, update the Google Sheet
			logger.Info("✉️ Email sent successfully to %s, updating Google Sheet...", to)
			err := UpdateSendStatus(to, true, cfg)
			if err != nil {
				logger.Error("❌ Failed to update send status for %s: %v", to, err)
			} else {
				logger.Info("✅ Successfully updated send status for %s in Google Sheet", to)
			}
		} else {
			logger.Error("❌ Email to %s failed to send", to)
		}
	})

	return jobID
}
