package queue

import (
	"encoding/json"
	"fmt"

	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/rs/zerolog/log"
)

const (
	EmailQueueKey = "mail-to-tg:queue:emails"
)

type EmailEvent struct {
	EmailID   string `json:"email_id"`
	AccountID string `json:"account_id"`
	UserID    string `json:"user_id"`
}

type Publisher struct {
	redis *storage.Redis
}

func NewPublisher(redis *storage.Redis) *Publisher {
	return &Publisher{redis: redis}
}

func (p *Publisher) PublishEmailEvent(event *EmailEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal email event: %w", err)
	}

	if err := p.redis.RPush(EmailQueueKey, data); err != nil {
		return fmt.Errorf("failed to push to queue: %w", err)
	}

	log.Debug().
		Str("email_id", event.EmailID).
		Str("account_id", event.AccountID).
		Str("user_id", event.UserID).
		Msg("Published email event to queue")

	return nil
}

func (p *Publisher) GetQueueLength() (int64, error) {
	return p.redis.LLen(EmailQueueKey)
}
