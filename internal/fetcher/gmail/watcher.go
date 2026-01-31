package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
)

type PubSubMessage struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    uint64 `json:"historyId"`
}

type Watcher struct {
	projectID    string
	subscription string
	client       *pubsub.Client
	handler      func(emailAddress string, historyID uint64) error
}

func NewWatcher(projectID, subscription string, handler func(string, uint64) error) (*Watcher, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
	}

	return &Watcher{
		projectID:    projectID,
		subscription: subscription,
		client:       client,
		handler:      handler,
	}, nil
}

func (w *Watcher) Start() error {
	ctx := context.Background()
	sub := w.client.Subscription(w.subscription)

	log.Info().
		Str("subscription", w.subscription).
		Msg("Starting Gmail Pub/Sub watcher")

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// Decode message
		var notification PubSubMessage

		// The data is base64 encoded
		decoded, err := base64.StdEncoding.DecodeString(string(msg.Data))
		if err != nil {
			log.Error().Err(err).Msg("Failed to decode Pub/Sub message")
			msg.Nack()
			return
		}

		if err := json.Unmarshal(decoded, &notification); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal Pub/Sub message")
			msg.Nack()
			return
		}

		log.Debug().
			Str("email", notification.EmailAddress).
			Uint64("history_id", notification.HistoryID).
			Msg("Received Gmail push notification")

		// Handle notification
		if err := w.handler(notification.EmailAddress, notification.HistoryID); err != nil {
			log.Error().
				Err(err).
				Str("email", notification.EmailAddress).
				Msg("Failed to handle push notification")
			msg.Nack()
			return
		}

		msg.Ack()
	})

	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	return nil
}

func (w *Watcher) Stop() {
	if w.client != nil {
		w.client.Close()
	}
}
