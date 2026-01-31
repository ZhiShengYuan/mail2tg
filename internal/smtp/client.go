package smtp

import (
	"encoding/base64"
	"fmt"

	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/crypto"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/wneessen/go-mail"
)

type Client struct {
	cfg *config.Config
	db  *storage.MariaDB
}

func NewClient(cfg *config.Config, db *storage.MariaDB) *Client {
	return &Client{
		cfg: cfg,
		db:  db,
	}
}

func (c *Client) SendReply(account *models.EmailAccount, originalEmail *models.EmailMessage, subject, body string) error {
	if account.SMTPServer == nil || account.SMTPPort == nil ||
		account.SMTPUsername == nil || account.SMTPPasswordEncrypted == nil {
		return fmt.Errorf("SMTP credentials not configured")
	}

	// Decrypt password
	encKey, err := base64.StdEncoding.DecodeString(c.cfg.Security.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}

	password, err := crypto.Decrypt(*account.SMTPPasswordEncrypted, encKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt SMTP password: %w", err)
	}

	// Create message
	m := mail.NewMsg()

	if err := m.From(account.EmailAddress); err != nil {
		return fmt.Errorf("failed to set from: %w", err)
	}

	if err := m.To(originalEmail.FromAddress); err != nil {
		return fmt.Errorf("failed to set to: %w", err)
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, body)

	// Set threading headers
	if originalEmail.MessageID != "" {
		m.SetMessageID()
		m.SetHeader("In-Reply-To", originalEmail.MessageID)

		// Build references
		references := originalEmail.MessageID
		if originalEmail.References != nil && *originalEmail.References != "" {
			references = *originalEmail.References + " " + originalEmail.MessageID
		}
		m.SetHeader("References", references)
	}

	// Create SMTP client
	smtpClient, err := mail.NewClient(*account.SMTPServer,
		mail.WithPort(*account.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(*account.SMTPUsername),
		mail.WithPassword(password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}

	// Send message
	if err := smtpClient.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (c *Client) SendEmail(account *models.EmailAccount, to, subject, body string) error {
	if account.SMTPServer == nil || account.SMTPPort == nil ||
		account.SMTPUsername == nil || account.SMTPPasswordEncrypted == nil {
		return fmt.Errorf("SMTP credentials not configured")
	}

	// Decrypt password
	encKey, err := base64.StdEncoding.DecodeString(c.cfg.Security.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decode encryption key: %w", err)
	}

	password, err := crypto.Decrypt(*account.SMTPPasswordEncrypted, encKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt SMTP password: %w", err)
	}

	// Create message
	m := mail.NewMsg()

	if err := m.From(account.EmailAddress); err != nil {
		return fmt.Errorf("failed to set from: %w", err)
	}

	if err := m.To(to); err != nil {
		return fmt.Errorf("failed to set to: %w", err)
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, body)
	m.SetMessageID()

	// Create SMTP client
	smtpClient, err := mail.NewClient(*account.SMTPServer,
		mail.WithPort(*account.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(*account.SMTPUsername),
		mail.WithPassword(password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}

	// Send message
	if err := smtpClient.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
