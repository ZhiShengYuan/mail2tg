package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Secrets contains sensitive configuration loaded from JSON
type Secrets struct {
	Database DatabaseSecrets `json:"database"`
	Redis    RedisSecrets    `json:"redis"`
	Telegram TelegramSecrets `json:"telegram"`
	Gmail    GmailSecrets    `json:"gmail"`
	Security SecuritySecrets `json:"security"`
	LLM      LLMSecrets      `json:"llm"`
}

type DatabaseSecrets struct {
	Password string `json:"password"`
}

type RedisSecrets struct {
	Password string `json:"password"`
}

type TelegramSecrets struct {
	BotToken   string `json:"bot_token"`
	WebhookURL string `json:"webhook_url"`
}

type GmailSecrets struct {
	ProjectID       string `json:"project_id"`
	CredentialsPath string `json:"credentials_path"`
}

type SecuritySecrets struct {
	EncryptionKey string `json:"encryption_key"`
	JWTSecret     string `json:"jwt_secret"`
}

type LLMSecrets struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

// LoadSecrets reads secrets from a JSON file
func LoadSecrets(secretsPath string) (*Secrets, error) {
	secrets := &Secrets{}

	// If secrets file doesn't exist, return empty secrets
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		return secrets, nil
	}

	f, err := os.Open(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open secrets file: %w", err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	if err := decoder.Decode(secrets); err != nil {
		return nil, fmt.Errorf("failed to decode secrets JSON: %w", err)
	}

	return secrets, nil
}

// ApplySecrets merges secrets into the main config
func (c *Config) ApplySecrets(secrets *Secrets) {
	if secrets.Database.Password != "" {
		c.Database.Password = secrets.Database.Password
	}
	if secrets.Redis.Password != "" {
		c.Redis.Password = secrets.Redis.Password
	}
	if secrets.Telegram.BotToken != "" {
		c.Telegram.BotToken = secrets.Telegram.BotToken
	}
	if secrets.Telegram.WebhookURL != "" {
		c.Telegram.WebhookURL = secrets.Telegram.WebhookURL
	}
	if secrets.Gmail.ProjectID != "" {
		c.MailFetcher.Gmail.ProjectID = secrets.Gmail.ProjectID
	}
	if secrets.Gmail.CredentialsPath != "" {
		c.MailFetcher.Gmail.CredentialsPath = secrets.Gmail.CredentialsPath
	}
	if secrets.Security.EncryptionKey != "" {
		c.Security.EncryptionKey = secrets.Security.EncryptionKey
	}
	if secrets.Security.JWTSecret != "" {
		c.Security.JWTSecret = secrets.Security.JWTSecret
	}
	if secrets.LLM.APIKey != "" {
		c.LLM.APIKey = secrets.LLM.APIKey
	}
	if secrets.LLM.BaseURL != "" {
		c.LLM.BaseURL = secrets.LLM.BaseURL
	}
	if secrets.LLM.Model != "" {
		c.LLM.Model = secrets.LLM.Model
	}
}
