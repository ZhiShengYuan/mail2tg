package notifier

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/kexi/mail-to-tg/pkg/llm"
	"github.com/kexi/mail-to-tg/pkg/models"
	"gopkg.in/telebot.v3"
)

type Formatter struct {
	baseURL string
}

func NewFormatter(baseURL string) *Formatter {
	return &Formatter{baseURL: baseURL}
}

func (f *Formatter) FormatEmailNotification(email *models.EmailMessage) (string, *telebot.ReplyMarkup) {
	var message strings.Builder

	message.WriteString("<b>📧 New Email</b>\n\n")

	// From
	if email.FromName != nil && *email.FromName != "" {
		message.WriteString(fmt.Sprintf("<b>From:</b> %s &lt;%s&gt;\n",
			html.EscapeString(*email.FromName),
			html.EscapeString(email.FromAddress)))
	} else {
		message.WriteString(fmt.Sprintf("<b>From:</b> %s\n",
			html.EscapeString(email.FromAddress)))
	}

	// Subject
	subject := "No subject"
	if email.Subject != nil && *email.Subject != "" {
		subject = *email.Subject
	}
	message.WriteString(fmt.Sprintf("<b>Subject:</b> %s\n\n",
		html.EscapeString(subject)))

	// AI Summary (or fallback to preview)
	if email.AISummary != nil && *email.AISummary != "" {
		message.WriteString("<b>🤖 Summary:</b>\n")
		message.WriteString(html.EscapeString(*email.AISummary))
		message.WriteString("\n\n")

		// Show extracted verification codes prominently
		if email.AIExtractedData != nil {
			var extractedData map[string]interface{}
			if err := json.Unmarshal([]byte(*email.AIExtractedData), &extractedData); err == nil {
				codes := llm.GetStringSlice(extractedData, "verification_codes")
				amounts := llm.GetStringSlice(extractedData, "amounts")
				dueDates := llm.GetStringSlice(extractedData, "due_dates")
				trackingNums := llm.GetStringSlice(extractedData, "tracking_numbers")

				// Verification codes
				if len(codes) > 0 {
					message.WriteString(fmt.Sprintf("<b>🔑 Code:</b> <code>%s</code>\n\n",
						html.EscapeString(codes[0])))
				}

				// Amounts
				if len(amounts) > 0 {
					message.WriteString(fmt.Sprintf("<b>💰 Amount:</b> %s\n",
						html.EscapeString(amounts[0])))
				}

				// Due dates
				if len(dueDates) > 0 {
					message.WriteString(fmt.Sprintf("<b>📅 Due:</b> %s\n",
						html.EscapeString(dueDates[0])))
				}

				// Tracking numbers
				if len(trackingNums) > 0 {
					message.WriteString(fmt.Sprintf("<b>📦 Tracking:</b> <code>%s</code>\n",
						html.EscapeString(trackingNums[0])))
				}

				// Add newline if we showed any extracted data
				if len(codes) > 0 || len(amounts) > 0 || len(dueDates) > 0 || len(trackingNums) > 0 {
					message.WriteString("\n")
				}
			}
		}
	} else {
		// Fallback to preview if no summary
		message.WriteString("<b>Preview:</b>\n")
		preview := f.getEmailPreview(email)
		if preview != "" {
			message.WriteString(preview)
			message.WriteString("\n\n")
		}
	}

	// Attachments
	if email.HasAttachments {
		message.WriteString("📎 Has attachments\n")
	}

	// Inline keyboard
	keyboard := &telebot.ReplyMarkup{}

	btnView := keyboard.Data("🌐 View Full", "view_"+email.ID)
	btnReply := keyboard.Data("↩️ Reply", "reply_"+email.ID)
	btnMarkRead := keyboard.Data("✅ Mark Read", "mark_read_"+email.ID)

	keyboard.Inline(
		keyboard.Row(btnView, btnReply),
		keyboard.Row(btnMarkRead),
	)

	return message.String(), keyboard
}

func (f *Formatter) getEmailPreview(email *models.EmailMessage) string {
	var text string

	// Prefer text body for preview
	if email.TextBody != nil && *email.TextBody != "" {
		text = *email.TextBody
	} else if email.SanitizedHTML != nil && *email.SanitizedHTML != "" {
		// Strip HTML tags for preview
		text = f.stripHTML(*email.SanitizedHTML)
	}

	if text == "" {
		return ""
	}

	// Clean up whitespace
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")

	var preview strings.Builder
	charCount := 0
	maxChars := 200

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if charCount+len(line) > maxChars {
			remaining := maxChars - charCount
			if remaining > 0 {
				preview.WriteString(line[:remaining])
			}
			preview.WriteString("...")
			break
		}

		if preview.Len() > 0 {
			preview.WriteString(" ")
		}
		preview.WriteString(line)
		charCount += len(line)
	}

	return html.EscapeString(preview.String())
}

func (f *Formatter) stripHTML(htmlText string) string {
	// Simple HTML tag removal
	text := htmlText
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n")
	text = strings.ReplaceAll(text, "</div>", "\n")

	// Remove all remaining tags
	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text, ">")
		if start < end {
			text = text[:start] + text[end+1:]
		} else {
			break
		}
	}

	return text
}
