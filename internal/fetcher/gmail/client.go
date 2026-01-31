package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kexi/mail-to-tg/internal/parser"
	"github.com/kexi/mail-to-tg/internal/queue"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Client struct {
	account   *models.EmailAccount
	db        *storage.MariaDB
	publisher *queue.Publisher
	parser    *parser.Parser
	cfg       *config.GmailConfig
	srv       *gmail.Service
	stopped   bool
}

func NewClient(
	account *models.EmailAccount,
	db *storage.MariaDB,
	publisher *queue.Publisher,
	emailParser *parser.Parser,
	cfg *config.GmailConfig,
) (*Client, error) {
	if account.OAuthTokenEncrypted == nil || account.OAuthRefreshTokenEncrypted == nil {
		return nil, fmt.Errorf("OAuth tokens not configured")
	}

	// Decrypt tokens
	accessToken, err := emailParser.DecryptPassword(*account.OAuthTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	refreshToken, err := emailParser.DecryptPassword(*account.OAuthRefreshTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       *account.OAuthExpiry,
	}

	// Create Gmail service
	ctx := context.Background()

	oauthMgr, err := NewOAuthManager(cfg.CredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth manager: %w", err)
	}

	tokenSource := *oauthMgr.GetClient(token)

	srv, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &Client{
		account:   account,
		db:        db,
		publisher: publisher,
		parser:    emailParser,
		cfg:       cfg,
		srv:       srv,
	}, nil
}

func (c *Client) Start() error {
	log.Info().
		Str("account_id", c.account.ID).
		Str("email", c.account.EmailAddress).
		Msg("Starting Gmail client")

	// Set up watch if not already done or expired
	if c.account.GmailWatchExpiration == nil || time.Now().After(*c.account.GmailWatchExpiration) {
		if err := c.SetupWatch(); err != nil {
			log.Error().Err(err).Msg("Failed to setup Gmail watch")
		}
	}

	// Fetch existing unread messages on start
	if err := c.FetchUnreadMessages(); err != nil {
		log.Error().Err(err).Msg("Initial fetch failed")
	}

	// Periodically renew watch (every 6 days)
	ticker := time.NewTicker(6 * 24 * time.Hour)
	defer ticker.Stop()

	for !c.stopped {
		<-ticker.C
		if err := c.SetupWatch(); err != nil {
			log.Error().Err(err).Msg("Failed to renew Gmail watch")
		}
	}

	return nil
}

func (c *Client) Stop() {
	log.Info().
		Str("account_id", c.account.ID).
		Msg("Stopping Gmail client")
	c.stopped = true
}

func (c *Client) SetupWatch() error {
	topicName := fmt.Sprintf("projects/%s/topics/%s", c.cfg.ProjectID, c.cfg.PubSubTopic)

	watchReq := &gmail.WatchRequest{
		TopicName: topicName,
		LabelIds:  []string{"INBOX"},
	}

	resp, err := c.srv.Users.Watch("me", watchReq).Do()
	if err != nil {
		return fmt.Errorf("failed to setup watch: %w", err)
	}

	// Save history ID and expiration
	c.account.GmailHistoryID = new(int64)
	*c.account.GmailHistoryID = int64(resp.HistoryId)

	expiryMillis := resp.Expiration
	expiryTime := time.Unix(0, expiryMillis*int64(time.Millisecond))
	c.account.GmailWatchExpiration = &expiryTime

	if err := c.db.UpdateEmailAccount(c.account); err != nil {
		return fmt.Errorf("failed to save watch state: %w", err)
	}

	log.Info().
		Str("account_id", c.account.ID).
		Int64("history_id", *c.account.GmailHistoryID).
		Time("expiry", expiryTime).
		Msg("Gmail watch setup successful")

	return nil
}

func (c *Client) FetchUnreadMessages() error {
	query := "is:unread in:inbox"

	req := c.srv.Users.Messages.List("me").Q(query).MaxResults(50)
	resp, err := req.Do()
	if err != nil {
		return fmt.Errorf("failed to list messages: %w", err)
	}

	log.Debug().
		Str("account_id", c.account.ID).
		Int("count", len(resp.Messages)).
		Msg("Found unread messages")

	for _, msg := range resp.Messages {
		if err := c.FetchAndProcessMessage(msg.Id); err != nil {
			log.Error().
				Err(err).
				Str("message_id", msg.Id).
				Msg("Failed to process message")
			continue
		}
	}

	return nil
}

func (c *Client) FetchAndProcessMessage(gmailID string) error {
	msg, err := c.srv.Users.Messages.Get("me", gmailID).Format("raw").Do()
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	// Decode raw message
	rawEmail, err := base64.URLEncoding.DecodeString(msg.Raw)
	if err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	// Parse email
	parsed, err := c.parser.ParseRaw(rawEmail)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Extract Message-ID from headers
	messageID := ""
	for _, header := range msg.Payload.Headers {
		if header.Name == "Message-ID" {
			messageID = header.Value
			break
		}
	}
	if messageID == "" {
		messageID = msg.Id
	}

	// Check if already exists
	existing, err := c.db.GetEmailMessageByAccountAndMessageID(c.account.ID, messageID)
	if err != nil {
		return fmt.Errorf("failed to check existing message: %w", err)
	}
	if existing != nil {
		log.Debug().Str("message_id", messageID).Msg("Message already exists")
		return nil
	}

	// Create email message record
	email := &models.EmailMessage{
		ID:            uuid.New().String(),
		AccountID:     c.account.ID,
		MessageID:     messageID,
		ThreadID:      &msg.ThreadId,
		GmailID:       &msg.Id,
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

	// Handle attachments
	if len(parsed.Attachments) > 0 {
		email.HasAttachments = true
		email.Attachments = parsed.AttachmentsJSON
	}

	// Save to database
	if err := c.db.CreateEmailMessage(email); err != nil {
		return fmt.Errorf("failed to save email: %w", err)
	}

	log.Info().
		Str("email_id", email.ID).
		Str("message_id", messageID).
		Str("subject", *email.Subject).
		Msg("Saved new Gmail message")

	// Publish to queue
	event := &queue.EmailEvent{
		EmailID:   email.ID,
		AccountID: c.account.ID,
		UserID:    c.account.UserID,
	}

	if err := c.publisher.PublishEmailEvent(event); err != nil {
		log.Error().Err(err).Msg("Failed to publish email event")
	}

	return nil
}

func (c *Client) HandlePushNotification(historyID uint64) error {
	if c.account.GmailHistoryID == nil {
		return c.FetchUnreadMessages()
	}

	startHistoryID := uint64(*c.account.GmailHistoryID)

	req := c.srv.Users.History.List("me").StartHistoryId(startHistoryID)
	resp, err := req.Do()
	if err != nil {
		return fmt.Errorf("failed to list history: %w", err)
	}

	for _, history := range resp.History {
		for _, msg := range history.MessagesAdded {
			if err := c.FetchAndProcessMessage(msg.Message.Id); err != nil {
				log.Error().
					Err(err).
					Str("message_id", msg.Message.Id).
					Msg("Failed to process message from history")
			}
		}
	}

	// Update history ID
	c.account.GmailHistoryID = new(int64)
	*c.account.GmailHistoryID = int64(historyID)
	c.db.UpdateEmailAccount(c.account)

	return nil
}
