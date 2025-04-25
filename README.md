# Go Mailer

A simple, dynamic email templating and sending system built with Go.

## Features

- Dynamic HTML email templates with variable substitution
- Email scheduling capability
- Clean architecture with separation of concerns
- Configuration management via environment variables

## Setup

1. Clone the repository
2. Create a `.env` file with the following content:

```
# Email Configuration
SENDER_MAIL_ID=your_email@gmail.com
PASSWORD=your_app_password
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

**Note:** For Gmail, you need to use an App Password:
1. Enable 2-Step Verification in your Google Account
2. Create an App Password at https://myaccount.google.com/apppasswords
3. Use that App Password here instead of your regular Gmail password

## Usage

### Running the Application

```bash
go run main.go
```

### Scheduling an Email

To schedule an email, use the `scheduleEmail` function:

```go
// Create template data
data := template.TemplateData{
    RecipientName:       "John Doe",
    CompanyName:         "TechCorp",
    SpecificArea:        "mobile app development",
    SpecificAchievement: "user-centric design",
    SpecificProject:     "the mobile banking platform", 
    RelevantSkill:       "Flutter architecture",
    SenderName:          "HR Department",
}

// Schedule the email
// Parameters: scheduler, recipient email, subject, template path, template data, send time
scheduleEmail(emailScheduler, "recipient@example.com", "Subject Line", 
              "tamplets/email_template.html", data, time.Now().Add(5*time.Minute))
```

This will schedule the email to be sent at the specified time (in this example, 5 minutes from now).

### Creating Your Own Email Templates

1. Create an HTML template file in the `tamplets` directory
2. Use Go template syntax for dynamic content:
   - `{{.RecipientName}}` - The name of the recipient
   - `{{.CompanyName}}` - The company name
   - etc.

## Project Structure

- `config/` - Configuration management
- `mailer/` - Email sending functionality
- `template/` - HTML template processing
- `tamplets/` - HTML email templates
- `scheduler/` - Email scheduling system
- `main.go` - Application entry point 