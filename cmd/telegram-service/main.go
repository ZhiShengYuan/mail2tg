package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kexi/mail-to-tg/internal/bot"
	"github.com/kexi/mail-to-tg/internal/notifier"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/internal/web"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/llm"
	"github.com/kexi/mail-to-tg/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	configPath := flag.String("config", "/etc/mail-to-tg/config.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)

	log.Info().
		Str("environment", cfg.Environment).
		Msg("Starting telegram-service")

	// Connect to database
	db, err := storage.NewMariaDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Connected to MariaDB")

	// Connect to Redis
	redis, err := storage.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	log.Info().Msg("Connected to Redis")

	// Create Telegram bot
	telegramBot, err := bot.NewBot(cfg, db, redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Telegram bot")
	}

	log.Info().Msg("Telegram bot initialized")

	// Initialize LLM client if enabled
	var llmClient llm.Client
	if cfg.LLM.Enabled {
		llmClient, err = llm.NewOpenAIClient(&cfg.LLM)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize LLM client, will use fallback")
			llmClient = nil
		} else {
			log.Info().
				Str("model", cfg.LLM.Model).
				Str("base_url", cfg.LLM.BaseURL).
				Msg("LLM client initialized")
		}
	} else {
		log.Info().Msg("LLM summarization disabled")
	}

	// Create notification consumer
	llmTimeout := time.Duration(cfg.LLM.TimeoutSeconds) * time.Second
	cacheTTL := time.Duration(cfg.LLM.CacheTTLHours) * time.Hour
	consumer := notifier.NewNotificationConsumer(
		redis,
		db,
		telegramBot.GetBot(),
		cfg.Web.BaseURL,
		llmClient,
		llmTimeout,
		cacheTTL,
	)

	// Start consumer in goroutine
	go func() {
		if err := consumer.Start(); err != nil {
			log.Error().Err(err).Msg("Notification consumer stopped")
		}
	}()

	log.Info().Msg("Notification consumer started")

	// Start web server in goroutine
	webServer := web.NewServer(&cfg.Web, db)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Error().Err(err).Msg("Web server stopped")
		}
	}()

	log.Info().Msg("Web server started")

	// Start Telegram bot (blocking)
	go func() {
		if err := telegramBot.Start(); err != nil {
			log.Error().Err(err).Msg("Telegram bot stopped")
		}
	}()

	log.Info().Msg("Telegram service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Info().Msg("Shutdown signal received, stopping service...")

	// Graceful shutdown
	consumer.Stop()
	telegramBot.Stop()

	log.Info().Msg("Telegram service stopped")
}
