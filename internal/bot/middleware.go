package bot

import (
	"github.com/google/uuid"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

func (b *Bot) authMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if c.Sender() == nil {
			return nil
		}

		telegramID := c.Sender().ID

		// Check if user exists
		user, err := b.db.GetUserByTelegramID(telegramID)
		if err != nil {
			log.Error().Err(err).Int64("telegram_id", telegramID).Msg("Failed to get user")
			return c.Send("An error occurred. Please try again later.")
		}

		// Create user if not exists
		if user == nil {
			username := c.Sender().Username
			firstName := c.Sender().FirstName
			lastName := c.Sender().LastName

			user = &models.User{
				ID:         uuid.New().String(),
				TelegramID: telegramID,
				IsActive:   true,
			}

			if username != "" {
				user.Username = &username
			}
			if firstName != "" {
				user.FirstName = &firstName
			}
			if lastName != "" {
				user.LastName = &lastName
			}

			if err := b.db.CreateUser(user); err != nil {
				log.Error().Err(err).Msg("Failed to create user")
				return c.Send("Failed to initialize user. Please try again.")
			}

			log.Info().
				Str("user_id", user.ID).
				Int64("telegram_id", telegramID).
				Msg("Created new user")
		}

		// Store user in context
		c.Set("user", user)

		return next(c)
	}
}

func (b *Bot) loggingMiddleware(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if c.Sender() != nil {
			log.Debug().
				Int64("telegram_id", c.Sender().ID).
				Str("username", c.Sender().Username).
				Str("text", c.Text()).
				Msg("Received message")
		}

		return next(c)
	}
}
