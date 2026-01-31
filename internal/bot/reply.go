package bot

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kexi/mail-to-tg/internal/smtp"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

func (b *Bot) startReplyMode(c telebot.Context, emailID string) error {
	user := c.Get("user").(*models.User)

	// Store reply state in Redis
	replyKey := fmt.Sprintf("reply:%d", user.TelegramID)
	b.redis.Set(replyKey, emailID, 10*time.Minute)

	return c.Respond(&telebot.CallbackResponse{
		Text:      "Reply mode activated. Send your reply message.",
		ShowAlert: true,
	})
}

func (b *Bot) handleReplyText(c telebot.Context, user *models.User, emailID, text string) error {
	// Get original email
	email, err := b.db.GetEmailMessageByID(emailID)
	if err != nil || email == nil {
		b.redis.Del(fmt.Sprintf("reply:%d", user.TelegramID))
		return c.Send("Original email not found. Reply cancelled.")
	}

	// Get email account
	account, err := b.db.GetEmailAccountByID(email.AccountID)
	if err != nil || account == nil {
		b.redis.Del(fmt.Sprintf("reply:%d", user.TelegramID))
		return c.Send("Email account not found. Reply cancelled.")
	}

	// Send reply via SMTP
	smtpClient := smtp.NewClient(b.cfg, b.db)

	subject := "Re: "
	if email.Subject != nil {
		subject += *email.Subject
	}

	err = smtpClient.SendReply(account, email, subject, text)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send reply")

		// Log failed attempt
		reply := &models.SentReply{
			ID:              uuid.New().String(),
			UserID:          user.ID,
			OriginalEmailID: &emailID,
			AccountID:       account.ID,
			ToAddress:       email.FromAddress,
			Subject:         &subject,
			Body:            &text,
			Error:           new(string),
		}
		*reply.Error = err.Error()
		b.db.CreateSentReply(reply)

		b.redis.Del(fmt.Sprintf("reply:%d", user.TelegramID))
		return c.Send(fmt.Sprintf("Failed to send reply: %v", err))
	}

	// Log successful reply
	reply := &models.SentReply{
		ID:              uuid.New().String(),
		UserID:          user.ID,
		OriginalEmailID: &emailID,
		AccountID:       account.ID,
		ToAddress:       email.FromAddress,
		Subject:         &subject,
		Body:            &text,
	}
	b.db.CreateSentReply(reply)

	// Clear reply state
	b.redis.Del(fmt.Sprintf("reply:%d", user.TelegramID))

	log.Info().
		Str("user_id", user.ID).
		Str("email_id", emailID).
		Str("to", email.FromAddress).
		Msg("Sent email reply")

	return c.Send("Reply sent successfully!")
}
