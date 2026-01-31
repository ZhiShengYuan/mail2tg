package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/models"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient implements the Client interface using OpenAI-compatible APIs
type OpenAIClient struct {
	client   *openai.Client
	model    string
	maxTokens int
}

// NewOpenAIClient creates a new OpenAI-compatible LLM client
func NewOpenAIClient(cfg *config.LLMConfig) (*OpenAIClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("LLM API key is required")
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)

	// Support custom base URLs for OpenRouter, local LLMs, etc.
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}

	client := openai.NewClientWithConfig(clientConfig)

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini" // Default model
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 300 // Default max tokens for summary
	}

	return &OpenAIClient{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
	}, nil
}

// Summarize generates a summary and extracts structured data from an email
func (c *OpenAIClient) Summarize(ctx context.Context, email *models.EmailMessage) (*SummaryResult, error) {
	// Build the prompt
	prompt, err := BuildEmailPrompt(email)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Create the chat completion request
	req := openai.ChatCompletionRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful AI assistant that summarizes emails and extracts structured information. Always respond with valid JSON.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower temperature for more consistent output
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	// Call the API
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices returned", ErrInvalidResponse)
	}

	// Parse the response
	content := resp.Choices[0].Message.Content

	var llmResponse struct {
		Summary       string                 `json:"summary"`
		ExtractedData map[string]interface{} `json:"extracted_data"`
	}

	if err := json.Unmarshal([]byte(content), &llmResponse); err != nil {
		return nil, fmt.Errorf("%w: failed to parse JSON response: %v", ErrInvalidResponse, err)
	}

	// Ensure extracted_data is not nil
	if llmResponse.ExtractedData == nil {
		llmResponse.ExtractedData = make(map[string]interface{})
	}

	result := &SummaryResult{
		Summary:       llmResponse.Summary,
		ExtractedData: llmResponse.ExtractedData,
		Model:         c.model,
		InputTokens:   resp.Usage.PromptTokens,
		OutputTokens:  resp.Usage.CompletionTokens,
	}

	return result, nil
}
