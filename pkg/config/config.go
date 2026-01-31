package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Environment string            `json:"environment"`
	Database    DatabaseConfig    `json:"database"`
	Redis       RedisConfig       `json:"redis"`
	MailFetcher MailFetcherConfig `json:"mail_fetcher"`
	Telegram    TelegramConfig    `json:"telegram"`
	Web         WebConfig         `json:"web"`
	Security    SecurityConfig    `json:"security"`
	Storage     StorageConfig     `json:"storage"`
	Logging     LoggingConfig     `json:"logging"`
	LLM         LLMConfig         `json:"llm"`
}

type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Name         string `json:"name"`
	User         string `json:"user"`
	Password     string `json:"password"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type MailFetcherConfig struct {
	Workers          int         `json:"workers"`
	IMAPPollInterval int         `json:"imap_poll_interval"`
	Gmail            GmailConfig `json:"gmail"`
}

type GmailConfig struct {
	ProjectID          string `json:"project_id"`
	PubSubTopic        string `json:"pubsub_topic"`
	PubSubSubscription string `json:"pubsub_subscription"`
	CredentialsPath    string `json:"credentials_path"`
}

type TelegramConfig struct {
	BotToken   string `json:"bot_token"`
	WebhookURL string `json:"webhook_url"`
}

type WebConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	BaseURL    string `json:"base_url"`
	TLSEnabled bool   `json:"tls_enabled"`
	TLSCert    string `json:"tls_cert"`
	TLSKey     string `json:"tls_key"`
}

type SecurityConfig struct {
	EncryptionKey string `json:"encryption_key"`
	JWTSecret     string `json:"jwt_secret"`
}

type StorageConfig struct {
	AttachmentsPath string `json:"attachments_path"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type LLMConfig struct {
	Enabled         bool   `json:"enabled"`
	BaseURL         string `json:"base_url"`
	APIKey          string `json:"api_key"`
	Model           string `json:"model"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	MaxTokens       int    `json:"max_tokens"`
	FallbackOnError bool   `json:"fallback_on_error"`
	CacheTTLHours   int    `json:"cache_ttl_hours"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Read JSON config file
	if configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
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
