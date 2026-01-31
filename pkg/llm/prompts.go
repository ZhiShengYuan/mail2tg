package llm

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/kexi/mail-to-tg/pkg/models"
)

// EmailSummaryPrompt is the template for email summarization
const EmailSummaryPrompt = `You are an AI that summarizes emails and extracts key information.

Email:
From: {{.FromAddress}}{{if .FromName}} ({{.FromName}}){{end}}
Subject: {{.Subject}}
Content:
{{.Body}}

Provide JSON with:
{
  "summary": "2-3 sentence summary focusing on most important info",
  "extracted_data": {
    "verification_codes": ["123456"],
    "amounts": ["$49.99"],
    "due_dates": ["2024-02-15"],
    "action_items": ["Click verification link"],
    "tracking_numbers": ["1Z999AA10123456784"]
  }
}

Focus on actionable information. Use empty arrays if nothing to extract.
Be concise and highlight the most important information.
For verification codes, extract any numeric or alphanumeric codes.
For amounts, include currency symbols.
For dates, use ISO format (YYYY-MM-DD) when possible.
For action items, be specific about what the user needs to do.`

// EmailPromptData contains data for template rendering
type EmailPromptData struct {
	FromAddress string
	FromName    string
	Subject     string
	Body        string
}

// BuildEmailPrompt builds the email summarization prompt from an email message
func BuildEmailPrompt(email *models.EmailMessage) (string, error) {
	data := EmailPromptData{
		FromAddress: email.FromAddress,
		Subject:     "No subject",
		Body:        "",
	}

	if email.FromName != nil && *email.FromName != "" {
		data.FromName = *email.FromName
	}

	if email.Subject != nil && *email.Subject != "" {
		data.Subject = *email.Subject
	}

	// Prefer text body, fallback to HTML body
	if email.TextBody != nil && *email.TextBody != "" {
		data.Body = *email.TextBody
	} else if email.HTMLBody != nil && *email.HTMLBody != "" {
		// For HTML, we'll use it as-is (the LLM can handle HTML)
		// In a production system, you might want to strip HTML tags
		data.Body = *email.HTMLBody
	}

	// Limit body length to prevent excessive token usage (max ~4000 chars)
	if len(data.Body) > 4000 {
		data.Body = data.Body[:4000] + "... (truncated)"
	}

	tmpl, err := template.New("prompt").Parse(EmailSummaryPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
