package notifier

import (
	"github.com/kexi/mail-to-tg/internal/queue"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

type NotificationConsumer struct {
	consumer  *queue.Consumer
	db        *storage.MariaDB
	bot       *telebot.Bot
	formatter *Formatter
}

func NewNotificationConsumer(
	redis *storage.Redis,
	db *storage.MariaDB,
	bot *telebot.Bot,
	baseURL string,
) *NotificationConsumer {
	formatter := NewFormatter(baseURL)

	nc := &NotificationConsumer{
		db:        db,
		bot:       bot,
		formatter: formatter,
	}

	consumer := queue.NewConsumer(redis, nc.handleEmailEvent)
	nc.consumer = consumer

	return nc
}

func (nc *NotificationConsumer) Start() error {
	return nc.consumer.Start()
}

func (nc *NotificationConsumer) Stop() {
	nc.consumer.Stop()
}

func (nc *NotificationConsumer) handleEmailEvent(event *queue.EmailEvent) error {
	log.Debug().
		Str("email_id", event.EmailID).
		Str("user_id", event.UserID).
		Msg("Handling email notification event")

	// Get email
	email, err := nc.db.GetEmailMessageByID(event.EmailID)
	if err != nil || email == nil {
		log.Error().Err(err).Str("email_id", event.EmailID).Msg("Failed to get email")
		return err
	}

	// Get user
	user, err := nc.db.GetUserByID(event.UserID)
	if err != nil || user == nil {
		log.Error().Err(err).Str("user_id", event.UserID).Msg("Failed to get user")
		return err
	}

	// Format notification message
	message, keyboard := nc.formatter.FormatEmailNotification(email)

	// Send to Telegram
	recipient := &telebot.User{ID: user.TelegramID}
	_, err = nc.bot.Send(recipient, message, &telebot.SendOptions{
		ParseMode:   telebot.ModeHTML,
		ReplyMarkup: keyboard,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("email_id", email.ID).
			Int64("telegram_id", user.TelegramID).
			Msg("Failed to send Telegram notification")
		return err
	}

	// Mark as notified
	if err := nc.db.MarkEmailAsNotified(email.ID); err != nil {
		log.Error().Err(err).Msg("Failed to mark email as notified")
	}

	log.Info().
		Str("email_id", email.ID).
		Int64("telegram_id", user.TelegramID).
		Msg("Sent email notification to Telegram")

	return nil
}
