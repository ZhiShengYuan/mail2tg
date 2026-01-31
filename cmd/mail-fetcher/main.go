package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/kexi/mail-to-tg/internal/fetcher"
	"github.com/kexi/mail-to-tg/internal/queue"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	configPath := flag.String("config", "/etc/mail-to-tg/config.json", "Path to config file")
	migrationsDir := flag.String("migrations", "./migrations", "Path to migrations directory")
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
		Msg("Starting mail-fetcher service")

	// Connect to database
	db, err := storage.NewMariaDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	log.Info().Msg("Connected to MariaDB")

	// Run database migrations automatically
	if err := db.RunMigrations(*migrationsDir); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Connect to Redis
	redis, err := storage.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	log.Info().Msg("Connected to Redis")

	// Create publisher
	publisher := queue.NewPublisher(redis)

	// Create fetch manager
	manager, err := fetcher.NewManager(db, publisher, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create fetch manager")
	}

	// Start fetcher
	if err := manager.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start fetch manager")
	}

	log.Info().Msg("Mail fetcher service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Info().Msg("Shutdown signal received, stopping service...")

	// Graceful shutdown
	manager.Stop()

	log.Info().Msg("Mail fetcher service stopped")
}
