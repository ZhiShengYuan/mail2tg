package bot

import (
	"fmt"
	"time"

	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

type Bot struct {
	bot    *telebot.Bot
	db     *storage.MariaDB
	redis  *storage.Redis
	cfg    *config.Config
}

func NewBot(cfg *config.Config, db *storage.MariaDB, redis *storage.Redis) (*Bot, error) {
	pref := telebot.Settings{
		Token:  cfg.Telegram.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot := &Bot{
		bot:   b,
		db:    db,
		redis: redis,
		cfg:   cfg,
	}

	bot.setupHandlers()

	return bot, nil
}

func (b *Bot) setupHandlers() {
	// Middleware
	b.bot.Use(b.authMiddleware)
	b.bot.Use(b.loggingMiddleware)

	// Commands
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/link", b.handleLink)
	b.bot.Handle("/unlink", b.handleUnlink)
	b.bot.Handle("/accounts", b.handleAccounts)
	b.bot.Handle("/search", b.handleSearch)

	// Callback queries (for inline buttons)
	b.bot.Handle(telebot.OnCallback, b.handleCallback)

	// Text messages (for reply mode)
	b.bot.Handle(telebot.OnText, b.handleText)
}

func (b *Bot) Start() error {
	log.Info().Msg("Starting Telegram bot")
	b.bot.Start()
	return nil
}

func (b *Bot) Stop() {
	log.Info().Msg("Stopping Telegram bot")
	b.bot.Stop()
}

func (b *Bot) GetBot() *telebot.Bot {
	return b.bot
}
