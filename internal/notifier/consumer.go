package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kexi/mail-to-tg/internal/queue"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/llm"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

type NotificationConsumer struct {
	consumer   *queue.Consumer
	db         *storage.MariaDB
	redis      *storage.Redis
	bot        *telebot.Bot
	formatter  *Formatter
	llmClient  llm.Client
	llmTimeout time.Duration
	cacheTTL   time.Duration
}

func NewNotificationConsumer(
	redis *storage.Redis,
	db *storage.MariaDB,
	bot *telebot.Bot,
	baseURL string,
	llmClient llm.Client,
	llmTimeout time.Duration,
	cacheTTL time.Duration,
) *NotificationConsumer {
	formatter := NewFormatter(baseURL)

	nc := &NotificationConsumer{
		db:         db,
		redis:      redis,
		bot:        bot,
		formatter:  formatter,
		llmClient:  llmClient,
		llmTimeout: llmTimeout,
		cacheTTL:   cacheTTL,
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

	// Generate AI summary if LLM is enabled
	if nc.llmClient != nil {
		nc.generateAISummary(email)
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

func (nc *NotificationConsumer) generateAISummary(email *models.EmailMessage) {
	// Check Redis cache first
	cacheKey := fmt.Sprintf("llm:summary:%s", email.ID)
	cached, err := nc.redis.Get(cacheKey)

	if err == nil && cached != "" {
		// Use cached summary
		var summary llm.SummaryResult
		if err := json.Unmarshal([]byte(cached), &summary); err == nil {
			email.AISummary = &summary.Summary
			extractedJSON, _ := llm.MarshalExtractedData(summary.ExtractedData)
			email.AIExtractedData = &extractedJSON
			email.AISummaryModel = &summary.Model

			log.Debug().
				Str("email_id", email.ID).
				Msg("Using cached LLM summary")
			return
		}
	}

	// Call LLM API with timeout
	ctx, cancel := context.WithTimeout(context.Background(), nc.llmTimeout)
	defer cancel()

	summary, err := nc.llmClient.Summarize(ctx, email)
	if err != nil {
		// Fallback to preview on error
		errStr := err.Error()
		log.Warn().
			Err(err).
			Str("email_id", email.ID).
			Msg("LLM summarization failed, using fallback")

		// Store error in database
		nc.db.UpdateEmailSummary(email.ID, nil, nil, nil, &errStr)
		return
	}

	// Cache and store summary
	summaryJSON, _ := json.Marshal(summary)
	nc.redis.Set(cacheKey, string(summaryJSON), nc.cacheTTL)

	// Update email object for immediate use
	email.AISummary = &summary.Summary
	extractedJSON, _ := llm.MarshalExtractedData(summary.ExtractedData)
	email.AIExtractedData = &extractedJSON
	email.AISummaryModel = &summary.Model

	// Store in database
	nc.db.UpdateEmailSummary(email.ID, &summary.Summary, &extractedJSON, &summary.Model, nil)

	log.Info().
		Str("email_id", email.ID).
		Str("model", summary.Model).
		Int("input_tokens", summary.InputTokens).
		Int("output_tokens", summary.OutputTokens).
		Msg("LLM summarization completed")
}
