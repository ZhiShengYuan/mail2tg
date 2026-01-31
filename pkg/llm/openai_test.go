package llm

import (
	"testing"

	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/models"
)

func TestNewOpenAIClient(t *testing.T) {
	cfg := &config.LLMConfig{
		APIKey:    "test-key",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4o-mini",
		MaxTokens: 300,
	}

	client, err := NewOpenAIClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.model != "gpt-4o-mini" {
		t.Errorf("Expected model gpt-4o-mini, got %s", client.model)
	}

	if client.maxTokens != 300 {
		t.Errorf("Expected maxTokens 300, got %d", client.maxTokens)
	}
}

func TestNewOpenAIClient_MissingAPIKey(t *testing.T) {
	cfg := &config.LLMConfig{
		APIKey: "",
	}

	_, err := NewOpenAIClient(cfg)
	if err == nil {
		t.Fatal("Expected error when API key is missing")
	}
}

func TestBuildEmailPrompt(t *testing.T) {
	fromName := "John Doe"
	subject := "Test Email"
	textBody := "This is a test email with a verification code: 123456"

	email := &models.EmailMessage{
		FromAddress: "john@example.com",
		FromName:    &fromName,
		Subject:     &subject,
		TextBody:    &textBody,
	}

	prompt, err := BuildEmailPrompt(email)
	if err != nil {
		t.Fatalf("Failed to build prompt: %v", err)
	}

	if prompt == "" {
		t.Fatal("Prompt should not be empty")
	}

	// Check that prompt contains email details
	if !contains(prompt, "john@example.com") {
		t.Error("Prompt should contain sender email")
	}

	if !contains(prompt, "Test Email") {
		t.Error("Prompt should contain subject")
	}

	if !contains(prompt, "verification code: 123456") {
		t.Error("Prompt should contain email body")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
