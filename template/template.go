package template

import (
	"bytes"
	"html/template"
	"os"
)

// TemplateData holds the data to be injected into the email template
type TemplateData struct {
	RecipientName       string
	CompanyName         string
	SpecificArea        string
	SpecificAchievement string
	SpecificProject     string
	RelevantSkill       string
	SenderName          string
}

// Process reads an HTML template file and replaces placeholder values with actual data
func Process(templatePath string, data TemplateData) (string, error) {
	// First read the template file
	htmlContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	htmlString := string(htmlContent)

	// Create a template with placeholder replacements
	tmpl, err := template.New("email").Parse(htmlString)
	if err != nil {
		return "", err
	}

	// Execute the template with the provided data
	var processedHTML bytes.Buffer
	if err := tmpl.Execute(&processedHTML, data); err != nil {
		return "", err
	}

	return processedHTML.String(), nil
}
