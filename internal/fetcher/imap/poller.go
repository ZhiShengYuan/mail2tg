package imap

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kexi/mail-to-tg/internal/parser"
	"github.com/kexi/mail-to-tg/internal/queue"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
)

type Poller struct {
	account   *models.EmailAccount
	db        *storage.MariaDB
	publisher *queue.Publisher
	parser    *parser.Parser
	interval  time.Duration
	stopped   bool
}

func NewPoller(
	account *models.EmailAccount,
	db *storage.MariaDB,
	publisher *queue.Publisher,
	emailParser *parser.Parser,
	interval time.Duration,
) *Poller {
	return &Poller{
		account:   account,
		db:        db,
		publisher: publisher,
		parser:    emailParser,
		interval:  interval,
	}
}

func (p *Poller) Start() error {
	log.Info().
		Str("account_id", p.account.ID).
		Str("email", p.account.EmailAddress).
		Dur("interval", p.interval).
		Msg("Starting IMAP poller")

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Fetch immediately on start
	if err := p.fetchOnce(); err != nil {
		log.Error().Err(err).Msg("Initial fetch failed")
	}

	for !p.stopped {
		select {
		case <-ticker.C:
			if err := p.fetchOnce(); err != nil {
				log.Error().Err(err).Msg("Fetch failed")
			}
		}
	}

	return nil
}

func (p *Poller) Stop() {
	log.Info().
		Str("account_id", p.account.ID).
		Msg("Stopping IMAP poller")
	p.stopped = true
}

func (p *Poller) fetchOnce() error {
	if p.account.IMAPServer == nil || p.account.IMAPPort == nil ||
		p.account.IMAPUsername == nil || p.account.IMAPPasswordEncrypted == nil {
		return fmt.Errorf("IMAP credentials not configured")
	}

	// Decrypt password
	password, err := p.parser.DecryptPassword(*p.account.IMAPPasswordEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt IMAP password: %w", err)
	}

	client := NewClient(*p.account.IMAPServer, *p.account.IMAPPort, *p.account.IMAPUsername, password)

	messages, err := client.FetchUnread()
	if err != nil {
		// Update account with error
		p.account.LastError = new(string)
		*p.account.LastError = err.Error()
		p.db.UpdateEmailAccount(p.account)
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	log.Debug().
		Str("account_id", p.account.ID).
		Int("count", len(messages)).
		Msg("Fetched messages from IMAP")

	for _, msg := range messages {
		if err := p.processMessage(msg); err != nil {
			log.Error().
				Err(err).
				Uint32("uid", msg.UID).
				Str("message_id", msg.MessageID).
				Msg("Failed to process message")
			continue
		}
	}

	// Update last fetch time
	now := time.Now()
	p.account.LastFetchAt = &now
	p.account.LastError = nil
	if err := p.db.UpdateEmailAccount(p.account); err != nil {
		log.Error().Err(err).Msg("Failed to update account fetch time")
	}

	return nil
}

func (p *Poller) processMessage(msg *Message) error {
	// Check if message already exists
	existing, err := p.db.GetEmailMessageByAccountAndMessageID(p.account.ID, msg.MessageID)
	if err != nil {
		return fmt.Errorf("failed to check existing message: %w", err)
	}
	if existing != nil {
		log.Debug().
			Str("message_id", msg.MessageID).
			Msg("Message already exists, skipping")
		return nil
	}

	// Parse email
	parsed, err := p.parser.ParseRaw(msg.RawMessage)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Create email message record
	email := &models.EmailMessage{
		ID:            uuid.New().String(),
		AccountID:     p.account.ID,
		MessageID:     msg.MessageID,
		IMAPUID:       new(int64),
		FromAddress:   parsed.FromAddress,
		FromName:      parsed.FromName,
		ToAddresses:   parsed.ToAddresses,
		Subject:       parsed.Subject,
		Date:          parsed.Date,
		TextBody:      parsed.TextBody,
		HTMLBody:      parsed.HTMLBody,
		SanitizedHTML: parsed.SanitizedHTML,
		InReplyTo:     parsed.InReplyTo,
		References:    parsed.References,
		IsRead:        false,
		IsNotified:    false,
	}
	*email.IMAPUID = int64(msg.UID)

	// Handle attachments
	if len(parsed.Attachments) > 0 {
		email.HasAttachments = true
		email.Attachments = parsed.AttachmentsJSON
	}

	// Save to database
	if err := p.db.CreateEmailMessage(email); err != nil {
		return fmt.Errorf("failed to save email: %w", err)
	}

	log.Info().
		Str("email_id", email.ID).
		Str("message_id", msg.MessageID).
		Str("subject", *email.Subject).
		Msg("Saved new email message")

	// Publish to queue for notification
	event := &queue.EmailEvent{
		EmailID:   email.ID,
		AccountID: p.account.ID,
		UserID:    p.account.UserID,
	}

	if err := p.publisher.PublishEmailEvent(event); err != nil {
		log.Error().Err(err).Msg("Failed to publish email event")
		// Don't return error, email is already saved
	}

	return nil
}
