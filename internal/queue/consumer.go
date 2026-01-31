package queue

import (
	"encoding/json"
	"time"

	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	redis   *storage.Redis
	handler func(*EmailEvent) error
	stopped bool
}

func NewConsumer(redis *storage.Redis, handler func(*EmailEvent) error) *Consumer {
	return &Consumer{
		redis:   redis,
		handler: handler,
	}
}

func (c *Consumer) Start() error {
	log.Info().Msg("Starting queue consumer")

	for !c.stopped {
		result, err := c.redis.BRPop(5*time.Second, EmailQueueKey)
		if err != nil {
			log.Error().Err(err).Msg("Failed to pop from queue")
			time.Sleep(time.Second)
			continue
		}

		// BRPop returns nil when timeout
		if result == nil {
			continue
		}

		// result is [queueKey, value]
		if len(result) < 2 {
			continue
		}

		data := result[1]
		var event EmailEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			log.Error().Err(err).Str("data", data).Msg("Failed to unmarshal email event")
			continue
		}

		log.Debug().
			Str("email_id", event.EmailID).
			Str("account_id", event.AccountID).
			Str("user_id", event.UserID).
			Msg("Processing email event from queue")

		if err := c.handler(&event); err != nil {
			log.Error().
				Err(err).
				Str("email_id", event.EmailID).
				Msg("Failed to handle email event")
			// In production, consider re-queuing or dead-letter queue
		}
	}

	log.Info().Msg("Queue consumer stopped")
	return nil
}

func (c *Consumer) Stop() {
	log.Info().Msg("Stopping queue consumer")
	c.stopped = true
}
