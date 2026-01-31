package llm

import (
	"context"

	"github.com/kexi/mail-to-tg/pkg/models"
)

// Client is the interface for LLM-based email summarization
type Client interface {
	// Summarize generates a summary and extracts structured data from an email
	Summarize(ctx context.Context, email *models.EmailMessage) (*SummaryResult, error)
}

// SummaryResult contains the LLM-generated summary and extracted data
type SummaryResult struct {
	Summary       string                 `json:"summary"`         // Human-readable summary
	ExtractedData map[string]interface{} `json:"extracted_data"`  // Structured data (codes, amounts, dates, etc.)
	Model         string                 `json:"model"`           // Model used for generation
	InputTokens   int                    `json:"input_tokens"`    // Input tokens for cost tracking
	OutputTokens  int                    `json:"output_tokens"`   // Output tokens for cost tracking
}

// ExtractedDataFields contains structured data extracted from emails
type ExtractedDataFields struct {
	VerificationCodes []string `json:"verification_codes,omitempty"`
	Amounts           []string `json:"amounts,omitempty"`
	DueDates          []string `json:"due_dates,omitempty"`
	ActionItems       []string `json:"action_items,omitempty"`
	TrackingNumbers   []string `json:"tracking_numbers,omitempty"`
}
