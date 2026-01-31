package config

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment string          `yaml:"environment" envconfig:"ENVIRONMENT"`
	Database    DatabaseConfig  `yaml:"database"`
	Redis       RedisConfig     `yaml:"redis"`
	MailFetcher MailFetcherConfig `yaml:"mail_fetcher"`
	Telegram    TelegramConfig  `yaml:"telegram"`
	Web         WebConfig       `yaml:"web"`
	Security    SecurityConfig  `yaml:"security"`
	Storage     StorageConfig   `yaml:"storage"`
	Logging     LoggingConfig   `yaml:"logging"`
	LLM         LLMConfig       `yaml:"llm"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host" envconfig:"DB_HOST"`
	Port         int    `yaml:"port" envconfig:"DB_PORT"`
	Name         string `yaml:"name" envconfig:"DB_NAME"`
	User         string `yaml:"user" envconfig:"DB_USER"`
	Password     string `yaml:"password" envconfig:"DB_PASSWORD"`
	MaxOpenConns int    `yaml:"max_open_conns" envconfig:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int    `yaml:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS"`
}

type RedisConfig struct {
	Host     string `yaml:"host" envconfig:"REDIS_HOST"`
	Port     int    `yaml:"port" envconfig:"REDIS_PORT"`
	Password string `yaml:"password" envconfig:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" envconfig:"REDIS_DB"`
}

type MailFetcherConfig struct {
	Workers          int         `yaml:"workers" envconfig:"FETCHER_WORKERS"`
	IMAPPollInterval int         `yaml:"imap_poll_interval" envconfig:"IMAP_POLL_INTERVAL"`
	Gmail            GmailConfig `yaml:"gmail"`
}

type GmailConfig struct {
	ProjectID           string `yaml:"project_id" envconfig:"GMAIL_PROJECT_ID"`
	PubSubTopic         string `yaml:"pubsub_topic" envconfig:"GMAIL_PUBSUB_TOPIC"`
	PubSubSubscription  string `yaml:"pubsub_subscription" envconfig:"GMAIL_PUBSUB_SUBSCRIPTION"`
	CredentialsPath     string `yaml:"credentials_path" envconfig:"GMAIL_CREDENTIALS_PATH"`
}

type TelegramConfig struct {
	BotToken   string `yaml:"bot_token" envconfig:"TELEGRAM_BOT_TOKEN"`
	WebhookURL string `yaml:"webhook_url" envconfig:"TELEGRAM_WEBHOOK_URL"`
}

type WebConfig struct {
	Host       string `yaml:"host" envconfig:"WEB_HOST"`
	Port       int    `yaml:"port" envconfig:"WEB_PORT"`
	BaseURL    string `yaml:"base_url" envconfig:"WEB_BASE_URL"`
	TLSEnabled bool   `yaml:"tls_enabled" envconfig:"WEB_TLS_ENABLED"`
	TLSCert    string `yaml:"tls_cert" envconfig:"WEB_TLS_CERT"`
	TLSKey     string `yaml:"tls_key" envconfig:"WEB_TLS_KEY"`
}

type SecurityConfig struct {
	EncryptionKey string `yaml:"encryption_key" envconfig:"ENCRYPTION_KEY"`
	JWTSecret     string `yaml:"jwt_secret" envconfig:"JWT_SECRET"`
}

type StorageConfig struct {
	AttachmentsPath string `yaml:"attachments_path" envconfig:"ATTACHMENTS_PATH"`
}

type LoggingConfig struct {
	Level  string `yaml:"level" envconfig:"LOG_LEVEL"`
	Format string `yaml:"format" envconfig:"LOG_FORMAT"`
}

type LLMConfig struct {
	Enabled         bool   `yaml:"enabled" envconfig:"LLM_ENABLED"`
	BaseURL         string `yaml:"base_url" envconfig:"LLM_BASE_URL"`
	APIKey          string `yaml:"api_key" envconfig:"LLM_API_KEY"`
	Model           string `yaml:"model" envconfig:"LLM_MODEL"`
	TimeoutSeconds  int    `yaml:"timeout_seconds" envconfig:"LLM_TIMEOUT_SECONDS"`
	MaxTokens       int    `yaml:"max_tokens" envconfig:"LLM_MAX_TOKENS"`
	FallbackOnError bool   `yaml:"fallback_on_error" envconfig:"LLM_FALLBACK_ON_ERROR"`
	CacheTTLHours   int    `yaml:"cache_ttl_hours" envconfig:"LLM_CACHE_TTL_HOURS"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Read YAML file
	if configPath != "" {
		f, err := os.Open(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer f.Close()

		decoder := yaml.NewDecoder(f)
		if err := decoder.Decode(cfg); err != nil {
			return nil, fmt.Errorf("failed to decode config: %w", err)
		}
	}

	// Override with environment variables
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}

	// Set defaults
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.MailFetcher.Workers == 0 {
		cfg.MailFetcher.Workers = 3
	}
	if cfg.MailFetcher.IMAPPollInterval == 0 {
		cfg.MailFetcher.IMAPPollInterval = 60
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Web.Host == "" {
		cfg.Web.Host = "0.0.0.0"
	}
	if cfg.Web.Port == 0 {
		cfg.Web.Port = 8080
	}
	if cfg.LLM.TimeoutSeconds == 0 {
		cfg.LLM.TimeoutSeconds = 10
	}
	if cfg.LLM.MaxTokens == 0 {
		cfg.LLM.MaxTokens = 300
	}
	if cfg.LLM.CacheTTLHours == 0 {
		cfg.LLM.CacheTTLHours = 24
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "gpt-4o-mini"
	}
	if cfg.LLM.BaseURL == "" {
		cfg.LLM.BaseURL = "https://api.openai.com/v1"
	}

	return cfg, nil
}
