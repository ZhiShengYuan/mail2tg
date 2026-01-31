package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment string            `yaml:"environment"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	MailFetcher MailFetcherConfig `yaml:"mail_fetcher"`
	Telegram    TelegramConfig    `yaml:"telegram"`
	Web         WebConfig         `yaml:"web"`
	Security    SecurityConfig    `yaml:"security"`
	Storage     StorageConfig     `yaml:"storage"`
	Logging     LoggingConfig     `yaml:"logging"`
	LLM         LLMConfig         `yaml:"llm"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Name         string `yaml:"name"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type MailFetcherConfig struct {
	Workers          int         `yaml:"workers"`
	IMAPPollInterval int         `yaml:"imap_poll_interval"`
	Gmail            GmailConfig `yaml:"gmail"`
}

type GmailConfig struct {
	ProjectID          string `yaml:"project_id"`
	PubSubTopic        string `yaml:"pubsub_topic"`
	PubSubSubscription string `yaml:"pubsub_subscription"`
	CredentialsPath    string `yaml:"credentials_path"`
}

type TelegramConfig struct {
	BotToken   string `yaml:"bot_token"`
	WebhookURL string `yaml:"webhook_url"`
}

type WebConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	BaseURL    string `yaml:"base_url"`
	TLSEnabled bool   `yaml:"tls_enabled"`
	TLSCert    string `yaml:"tls_cert"`
	TLSKey     string `yaml:"tls_key"`
}

type SecurityConfig struct {
	EncryptionKey string `yaml:"encryption_key"`
	JWTSecret     string `yaml:"jwt_secret"`
}

type StorageConfig struct {
	AttachmentsPath string `yaml:"attachments_path"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type LLMConfig struct {
	Enabled         bool   `yaml:"enabled"`
	BaseURL         string `yaml:"base_url"`
	APIKey          string `yaml:"api_key"`
	Model           string `yaml:"model"`
	TimeoutSeconds  int    `yaml:"timeout_seconds"`
	MaxTokens       int    `yaml:"max_tokens"`
	FallbackOnError bool   `yaml:"fallback_on_error"`
	CacheTTLHours   int    `yaml:"cache_ttl_hours"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Read YAML config file
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

	// Load secrets from JSON file (secrets.json in same directory as config)
	secretsPath := filepath.Join(filepath.Dir(configPath), "secrets.json")
	secrets, err := LoadSecrets(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load secrets: %w", err)
	}

	// Apply secrets to config
	cfg.ApplySecrets(secrets)

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
