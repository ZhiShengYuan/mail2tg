-- Add AI summary fields to email_messages table
-- Migration: 003_add_ai_summary

ALTER TABLE email_messages
ADD COLUMN ai_summary TEXT NULL COMMENT 'LLM-generated summary',
ADD COLUMN ai_extracted_data JSON NULL COMMENT 'Structured data: codes, amounts, dates',
ADD COLUMN ai_summary_model VARCHAR(50) NULL COMMENT 'Model used (e.g., gpt-4o-mini)',
ADD COLUMN ai_summary_at TIMESTAMP NULL COMMENT 'Timestamp of summary generation',
ADD COLUMN ai_summary_error TEXT NULL COMMENT 'Error if summarization failed',
ADD INDEX idx_ai_summary_at (ai_summary_at);
