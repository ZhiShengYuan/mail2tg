package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
	"github.com/kexi/mail-to-tg/pkg/crypto"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
)

type Parser struct {
	sanitizer       *Sanitizer
	encryptionKey   []byte
	attachmentsPath string
}

type ParsedEmail struct {
	FromAddress     string
	FromName        *string
	ToAddresses     *string
	Subject         *string
	Date            time.Time
	TextBody        *string
	HTMLBody        *string
	SanitizedHTML   *string
	InReplyTo       *string
	References      *string
	Attachments     []*models.Attachment
	AttachmentsJSON *string
}

func NewParser(encryptionKey []byte, attachmentsPath string) *Parser {
	return &Parser{
		sanitizer:       NewSanitizer(),
		encryptionKey:   encryptionKey,
		attachmentsPath: attachmentsPath,
	}
}

func (p *Parser) ParseRaw(rawEmail []byte) (*ParsedEmail, error) {
	envelope, err := enmime.ReadEnvelope(bytes.NewReader(rawEmail))
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	emailDate, _ := envelope.Date()
	parsed := &ParsedEmail{
		Date: emailDate,
	}

	// From address
	if from := envelope.GetHeader("From"); from != "" {
		parsed.FromAddress = from
		// Try to extract name
		if strings.Contains(from, "<") {
			parts := strings.Split(from, "<")
			if len(parts) > 0 {
				name := strings.TrimSpace(parts[0])
				name = strings.Trim(name, `"`)
				if name != "" {
					parsed.FromName = &name
				}
			}
		}
	}

	// To addresses
	if to := envelope.GetHeader("To"); to != "" {
		parsed.ToAddresses = &to
	}

	// Subject
	if subject := envelope.GetHeader("Subject"); subject != "" {
		parsed.Subject = &subject
	}

	// Text body
	if text := envelope.Text; text != "" {
		parsed.TextBody = &text
	}

	// HTML body
	if html := envelope.HTML; html != "" {
		parsed.HTMLBody = &html
		// Sanitize HTML
		sanitized := p.sanitizer.Sanitize(html)
		parsed.SanitizedHTML = &sanitized
	}

	// In-Reply-To
	if inReplyTo := envelope.GetHeader("In-Reply-To"); inReplyTo != "" {
		parsed.InReplyTo = &inReplyTo
	}

	// References
	if references := envelope.GetHeader("References"); references != "" {
		parsed.References = &references
	}

	// Attachments
	if len(envelope.Attachments) > 0 {
		attachments, err := p.saveAttachments(envelope.Attachments)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save attachments")
		} else {
			parsed.Attachments = attachments
			// Convert to JSON
			jsonData, err := json.Marshal(attachments)
			if err == nil {
				jsonStr := string(jsonData)
				parsed.AttachmentsJSON = &jsonStr
			}
		}
	}

	return parsed, nil
}

func (p *Parser) saveAttachments(attachments []*enmime.Part) ([]*models.Attachment, error) {
	emailID := uuid.New().String()
	emailDir := filepath.Join(p.attachmentsPath, emailID)

	if err := os.MkdirAll(emailDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create attachments directory: %w", err)
	}

	var result []*models.Attachment

	for _, part := range attachments {
		filename := part.FileName
		if filename == "" {
			filename = fmt.Sprintf("attachment_%d", len(result)+1)
		}

		// Sanitize filename
		filename = filepath.Base(filename)
		filePath := filepath.Join(emailDir, filename)

		content := part.Content
		if content == nil {
			log.Error().Str("filename", filename).Msg("Attachment content is nil")
			continue
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			log.Error().Err(err).Str("filename", filename).Msg("Failed to write attachment")
			continue
		}

		attachment := &models.Attachment{
			Filename:    filename,
			ContentType: part.ContentType,
			Size:        int64(len(content)),
			Path:        filePath,
		}

		result = append(result, attachment)

		log.Debug().
			Str("filename", filename).
			Int64("size", attachment.Size).
			Msg("Saved attachment")
	}

	return result, nil
}

func (p *Parser) DecryptPassword(encrypted string) (string, error) {
	return crypto.Decrypt(encrypted, p.encryptionKey)
}

func (p *Parser) EncryptPassword(password string) (string, error) {
	return crypto.Encrypt(password, p.encryptionKey)
}
