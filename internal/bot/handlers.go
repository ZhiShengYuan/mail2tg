package bot

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kexi/mail-to-tg/pkg/crypto"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

func (b *Bot) handleStart(c telebot.Context) error {
	message := `Welcome to Mail-to-Telegram Bot!

This bot forwards your emails to Telegram with HTML rendering and allows you to reply directly.

Commands:
/link - Link an email account (Gmail or IMAP)
/accounts - List your linked accounts
/unlink - Unlink an email account
/search <query> - Search your emails
/help - Show this help message

Get started by linking an email account with /link`

	return c.Send(message)
}

func (b *Bot) handleHelp(c telebot.Context) error {
	message := `Mail-to-Telegram Bot Commands:

/start - Show welcome message
/link - Link an email account
  • Gmail: OAuth2 authentication
  • IMAP: For QQmail and other providers

/accounts - List all linked email accounts
/unlink - Remove an email account
/search <query> - Search emails by subject or sender

When you receive an email, you'll get a notification with:
• Subject and sender
• Preview of the content
• Buttons to view full email or reply

To reply to an email, click the [Reply] button and send your message.`

	return c.Send(message)
}

func (b *Bot) handleLink(c telebot.Context) error {
	selector := &telebot.ReplyMarkup{}
	btnGmail := selector.Data("Gmail (OAuth2)", "link_gmail")
	btnIMAP := selector.Data("IMAP (QQmail, etc.)", "link_imap")
	btnCancel := selector.Data("Cancel", "cancel")

	selector.Inline(
		selector.Row(btnGmail),
		selector.Row(btnIMAP),
		selector.Row(btnCancel),
	)

	return c.Send("Choose how to link your email account:", selector)
}

func (b *Bot) handleUnlink(c telebot.Context) error {
	user := c.Get("user").(*models.User)

	accounts, err := b.db.GetEmailAccountsByUserID(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get accounts")
		return c.Send("Failed to load accounts. Please try again.")
	}

	if len(accounts) == 0 {
		return c.Send("You don't have any linked email accounts.")
	}

	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for _, account := range accounts {
		btnText := fmt.Sprintf("%s (%s)", account.EmailAddress, account.Provider)
		btn := selector.Data(btnText, "unlink_"+account.ID)
		rows = append(rows, selector.Row(btn))
	}

	rows = append(rows, selector.Row(selector.Data("Cancel", "cancel")))
	selector.Inline(rows...)

	return c.Send("Select an account to unlink:", selector)
}

func (b *Bot) handleAccounts(c telebot.Context) error {
	user := c.Get("user").(*models.User)

	accounts, err := b.db.GetEmailAccountsByUserID(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get accounts")
		return c.Send("Failed to load accounts. Please try again.")
	}

	if len(accounts) == 0 {
		return c.Send("You don't have any linked email accounts.\n\nUse /link to add one.")
	}

	var message strings.Builder
	message.WriteString("Your linked email accounts:\n\n")

	for i, account := range accounts {
		status := "Active"
		if !account.IsActive {
			status = "Inactive"
		}

		message.WriteString(fmt.Sprintf("%d. %s (%s) - %s\n",
			i+1, account.EmailAddress, account.Provider, status))

		if account.LastFetchAt != nil {
			message.WriteString(fmt.Sprintf("   Last fetched: %s\n",
				account.LastFetchAt.Format("2006-01-02 15:04:05")))
		}

		if account.LastError != nil && *account.LastError != "" {
			message.WriteString(fmt.Sprintf("   Error: %s\n", *account.LastError))
		}

		message.WriteString("\n")
	}

	return c.Send(message.String())
}

func (b *Bot) handleSearch(c telebot.Context) error {
	query := c.Text()
	if query == "/search" || strings.TrimSpace(strings.TrimPrefix(query, "/search")) == "" {
		return c.Send("Usage: /search <query>\n\nExample: /search important meeting")
	}

	query = strings.TrimSpace(strings.TrimPrefix(query, "/search"))

	return c.Send(fmt.Sprintf("Searching for: %s\n\nSearch functionality coming soon!", query))
}

func (b *Bot) handleCallback(c telebot.Context) error {
	data := c.Callback().Data

	switch {
	case data == "cancel":
		return c.Edit("Cancelled.")

	case data == "link_gmail":
		return b.handleLinkGmail(c)

	case data == "link_imap":
		return b.handleLinkIMAP(c)

	case strings.HasPrefix(data, "unlink_"):
		accountID := strings.TrimPrefix(data, "unlink_")
		return b.handleUnlinkAccount(c, accountID)

	case strings.HasPrefix(data, "view_"):
		emailID := strings.TrimPrefix(data, "view_")
		return b.handleViewEmail(c, emailID)

	case strings.HasPrefix(data, "reply_"):
		emailID := strings.TrimPrefix(data, "reply_")
		return b.handleReplyButton(c, emailID)

	case strings.HasPrefix(data, "mark_read_"):
		emailID := strings.TrimPrefix(data, "mark_read_")
		return b.handleMarkRead(c, emailID)
	}

	return c.Respond(&telebot.CallbackResponse{Text: "Unknown action"})
}

func (b *Bot) handleLinkGmail(c telebot.Context) error {
	return c.Edit("Gmail OAuth2 linking coming soon!\n\nFor now, please use IMAP linking.")
}

func (b *Bot) handleLinkIMAP(c telebot.Context) error {
	user := c.Get("user").(*models.User)

	stateKey := fmt.Sprintf("link_imap:%d", user.TelegramID)
	b.redis.HSet(stateKey, "step", "email")
	b.redis.Expire(stateKey, 600) // 10 minutes

	return c.Edit("Let's link your IMAP account.\n\nPlease send your email address:")
}

func (b *Bot) handleUnlinkAccount(c telebot.Context, accountID string) error {
	account, err := b.db.GetEmailAccountByID(accountID)
	if err != nil || account == nil {
		return c.Edit("Account not found.")
	}

	if err := b.db.DeleteEmailAccount(accountID); err != nil {
		log.Error().Err(err).Str("account_id", accountID).Msg("Failed to delete account")
		return c.Edit("Failed to unlink account. Please try again.")
	}

	log.Info().
		Str("account_id", accountID).
		Str("email", account.EmailAddress).
		Msg("Unlinked email account")

	return c.Edit(fmt.Sprintf("Successfully unlinked %s", account.EmailAddress))
}

func (b *Bot) handleViewEmail(c telebot.Context, emailID string) error {
	// Generate view token and send URL
	token := uuid.New().String()

	viewToken := &models.EmailViewToken{
		ID:        uuid.New().String(),
		EmailID:   emailID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := b.db.CreateEmailViewToken(viewToken); err != nil {
		log.Error().Err(err).Msg("Failed to create view token")
		return c.Respond(&telebot.CallbackResponse{Text: "Failed to generate link"})
	}

	url := fmt.Sprintf("%s/email/%s", b.cfg.Web.BaseURL, token)

	return c.Respond(&telebot.CallbackResponse{
		Text: fmt.Sprintf("View email: %s", url),
		ShowAlert: true,
	})
}

func (b *Bot) handleReplyButton(c telebot.Context, emailID string) error {
	return b.startReplyMode(c, emailID)
}

func (b *Bot) handleMarkRead(c telebot.Context, emailID string) error {
	if err := b.db.MarkEmailAsNotified(emailID); err != nil {
		log.Error().Err(err).Msg("Failed to mark email as read")
		return c.Respond(&telebot.CallbackResponse{Text: "Failed to mark as read"})
	}

	return c.Respond(&telebot.CallbackResponse{Text: "Marked as read"})
}

func (b *Bot) handleText(c telebot.Context) error {
	user := c.Get("user").(*models.User)
	text := c.Text()

	// Check if in linking flow
	stateKey := fmt.Sprintf("link_imap:%d", user.TelegramID)
	step, err := b.redis.HGet(stateKey, "step")
	if err == nil && step != "" {
		return b.handleLinkIMAPFlow(c, user, stateKey, step, text)
	}

	// Check if in reply mode
	replyKey := fmt.Sprintf("reply:%d", user.TelegramID)
	emailID, err := b.redis.Get(replyKey)
	if err == nil && emailID != "" {
		return b.handleReplyText(c, user, emailID, text)
	}

	return c.Send("I don't understand. Use /help to see available commands.")
}

func (b *Bot) handleLinkIMAPFlow(c telebot.Context, user *models.User, stateKey, step, text string) error {
	switch step {
	case "email":
		// Save email and ask for IMAP server
		b.redis.HSet(stateKey, "email", text)
		b.redis.HSet(stateKey, "step", "imap_server")
		return c.Send("Email address saved.\n\nNow enter your IMAP server (e.g., imap.qq.com):")

	case "imap_server":
		b.redis.HSet(stateKey, "imap_server", text)
		b.redis.HSet(stateKey, "step", "imap_port")
		return c.Send("IMAP server saved.\n\nEnter IMAP port (usually 993):")

	case "imap_port":
		b.redis.HSet(stateKey, "imap_port", text)
		b.redis.HSet(stateKey, "step", "imap_username")
		return c.Send("Port saved.\n\nEnter your IMAP username (usually your email address):")

	case "imap_username":
		b.redis.HSet(stateKey, "imap_username", text)
		b.redis.HSet(stateKey, "step", "imap_password")
		return c.Send("Username saved.\n\nEnter your IMAP password:")

	case "imap_password":
		return b.completeLinkIMAP(c, user, stateKey, text)
	}

	return nil
}

func (b *Bot) completeLinkIMAP(c telebot.Context, user *models.User, stateKey, password string) error {
	// Get all saved data
	data, err := b.redis.HGetAll(stateKey)
	if err != nil {
		return c.Send("Session expired. Please start over with /link")
	}

	// Encrypt password
	encKey, _ := base64.StdEncoding.DecodeString(b.cfg.Security.EncryptionKey)
	encPassword, err := crypto.Encrypt(password, encKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt password")
		return c.Send("Failed to save account. Please try again.")
	}

	// Create account
	imapServer := data["imap_server"]
	imapUsername := data["imap_username"]

	account := &models.EmailAccount{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		Provider:     "imap",
		EmailAddress: data["email"],
		IMAPServer:   &imapServer,
		IMAPUsername: &imapUsername,
		SMTPServer:   &imapServer, // Default to same as IMAP
		SMTPUsername: &imapUsername,
		IsActive:     true,
	}

	// Parse port
	var port int
	fmt.Sscanf(data["imap_port"], "%d", &port)
	account.IMAPPort = &port
	account.SMTPPort = &port

	account.IMAPPasswordEncrypted = &encPassword
	account.SMTPPasswordEncrypted = &encPassword

	if err := b.db.CreateEmailAccount(account); err != nil {
		log.Error().Err(err).Msg("Failed to create account")
		return c.Send("Failed to save account. Please try again.")
	}

	// Clean up state
	b.redis.Del(stateKey)

	log.Info().
		Str("account_id", account.ID).
		Str("email", account.EmailAddress).
		Msg("Linked IMAP account")

	return c.Send(fmt.Sprintf("Successfully linked %s!\n\nYou'll start receiving email notifications shortly.", account.EmailAddress))
}
