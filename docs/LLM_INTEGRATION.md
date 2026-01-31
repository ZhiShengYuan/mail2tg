# LLM Integration Guide

This guide explains how to configure and use the LLM-based email summarization feature in Mail-to-Telegram.

## Overview

The LLM integration adds AI-powered email summarization that:

- Generates concise 2-3 sentence summaries of emails
- Extracts structured data (verification codes, amounts, dates, tracking numbers)
- Displays summaries in Telegram notifications instead of plain text previews
- Caches results in Redis to minimize API costs
- Stores summaries in the database for future reference
- Falls back to traditional preview mode if LLM fails

## Supported Providers

The system uses OpenAI-compatible APIs, supporting:

1. **OpenAI** - Official OpenAI API
2. **OpenRouter** - Access to multiple LLM providers
3. **Local LLMs** - Ollama, LM Studio, etc.
4. **Custom endpoints** - Any OpenAI-compatible API

## Configuration

### 1. OpenAI

**Cost**: ~$0.27/day for 1000 emails with gpt-4o-mini (~$8/month)

```yaml
# config.yaml
llm:
  enabled: true
  base_url: https://api.openai.com/v1
  api_key: ${LLM_API_KEY}
  model: gpt-4o-mini
  timeout_seconds: 10
  max_tokens: 300
  fallback_on_error: true
  cache_ttl_hours: 24
```

```bash
# .env
LLM_API_KEY=sk-your-openai-api-key
```

**Recommended models**:
- `gpt-4o-mini` - Fastest and cheapest ($0.15/1M input, $0.60/1M output)
- `gpt-4o` - Higher quality but more expensive
- `gpt-3.5-turbo` - Legacy, not recommended

### 2. OpenRouter

**Benefits**: Access to multiple models, pay-as-you-go pricing

```yaml
llm:
  enabled: true
  base_url: https://openrouter.ai/api/v1
  api_key: ${LLM_API_KEY}
  model: anthropic/claude-3.5-sonnet
  timeout_seconds: 10
  max_tokens: 300
  fallback_on_error: true
  cache_ttl_hours: 24
```

```bash
# .env
LLM_API_KEY=sk-or-your-openrouter-api-key
```

**Popular models on OpenRouter**:
- `anthropic/claude-3.5-sonnet` - Excellent quality
- `google/gemini-2.0-flash-exp:free` - Free tier available
- `meta-llama/llama-3.3-70b-instruct` - Good balance

### 3. Local LLM (Ollama)

**Benefits**: Free, privacy-focused, no internet required

1. Install Ollama:
   ```bash
   curl -fsSL https://ollama.com/install.sh | sh
   ```

2. Download a model:
   ```bash
   ollama pull llama3.2
   # or
   ollama pull mistral
   ```

3. Configure:
   ```yaml
   llm:
     enabled: true
     base_url: http://localhost:11434/v1
     api_key: not-needed
     model: llama3.2
     timeout_seconds: 30  # Local models may be slower
     max_tokens: 300
     fallback_on_error: true
     cache_ttl_hours: 24
   ```

**Note**: Local models may be slower. Increase `timeout_seconds` to 30-60.

### 4. LM Studio

1. Download LM Studio from [lmstudio.ai](https://lmstudio.ai/)
2. Download a model (e.g., Llama 3.2)
3. Start the local server in LM Studio
4. Configure:
   ```yaml
   llm:
     enabled: true
     base_url: http://localhost:1234/v1
     api_key: not-needed
     model: llama-3.2-3b-instruct
     timeout_seconds: 30
     max_tokens: 300
     fallback_on_error: true
     cache_ttl_hours: 24
   ```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `enabled` | `false` | Enable/disable LLM summarization |
| `base_url` | `https://api.openai.com/v1` | API endpoint URL |
| `api_key` | - | API key (required for cloud providers) |
| `model` | `gpt-4o-mini` | Model to use for summarization |
| `timeout_seconds` | `10` | Request timeout (increase for local LLMs) |
| `max_tokens` | `300` | Maximum tokens in summary |
| `fallback_on_error` | `true` | Use preview if LLM fails |
| `cache_ttl_hours` | `24` | Redis cache duration |

## How It Works

### 1. Email Processing Flow

```
New Email Received
    ↓
Saved to Database
    ↓
Redis Queue → Email ID
    ↓
Notification Consumer
    ↓
Check Redis Cache ──┐
    ↓               │ Cache Hit
    ├── Cache Miss  │
    ↓               ↓
Call LLM API    Use Cached
    ↓               ↓
Parse Response  ────┘
    ↓
Extract Data (codes, amounts, etc.)
    ↓
Store in Database
    ↓
Cache in Redis (24h)
    ↓
Format Telegram Message
    ↓
Send to User
```

### 2. Prompt Engineering

The system uses a structured prompt to extract:

- **Summary**: 2-3 sentence overview
- **Verification Codes**: Numeric/alphanumeric codes
- **Amounts**: Billing amounts with currency
- **Due Dates**: Payment or action deadlines (ISO format)
- **Action Items**: What the user needs to do
- **Tracking Numbers**: Shipping/order tracking

Example response:
```json
{
  "summary": "Your order #12345 has shipped via UPS. Tracking number provided. Expected delivery by Dec 25.",
  "extracted_data": {
    "verification_codes": [],
    "amounts": ["$49.99"],
    "due_dates": ["2024-12-25"],
    "action_items": ["Track your package"],
    "tracking_numbers": ["1Z999AA10123456784"]
  }
}
```

### 3. Telegram Notification Format

**With LLM enabled**:
```
📧 New Email

From: Amazon <noreply@amazon.com>
Subject: Your order has shipped

🤖 Summary:
Your order #12345 has shipped via UPS. Tracking number provided. Expected delivery by Dec 25.

💰 Amount: $49.99
📅 Due: 2024-12-25
📦 Tracking: 1Z999AA10123456784

[🌐 View Full] [↩️ Reply] [✅ Mark Read]
```

**Without LLM** (fallback):
```
📧 New Email

From: Amazon <noreply@amazon.com>
Subject: Your order has shipped

Preview:
Hello, your order #12345 containing Widget XYZ has been shipped via UPS Ground. You can track your package using the tracking number 1Z999AA10...

[🌐 View Full] [↩️ Reply] [✅ Mark Read]
```

## Database Schema

The migration adds these fields to `email_messages`:

```sql
ALTER TABLE email_messages
ADD COLUMN ai_summary TEXT NULL,
ADD COLUMN ai_extracted_data JSON NULL,
ADD COLUMN ai_summary_model VARCHAR(50) NULL,
ADD COLUMN ai_summary_at TIMESTAMP NULL,
ADD COLUMN ai_summary_error TEXT NULL,
ADD INDEX idx_ai_summary_at (ai_summary_at);
```

## Cost Optimization

### 1. Redis Caching

- Summaries cached for 24 hours by default
- Viewing same email multiple times = 1 API call
- Adjust `cache_ttl_hours` to increase cache duration

### 2. Model Selection

Choose cheaper models for cost savings:

| Model | Input ($/1M) | Output ($/1M) | Quality |
|-------|-------------|---------------|---------|
| gpt-4o-mini | $0.15 | $0.60 | Good |
| gpt-3.5-turbo | $0.50 | $1.50 | Fair |
| gpt-4o | $2.50 | $10.00 | Excellent |
| claude-3.5-sonnet | $3.00 | $15.00 | Excellent |
| llama3.2 (local) | Free | Free | Good |

### 3. Token Limits

- Email body truncated to 4000 chars before sending to LLM
- `max_tokens: 300` limits summary length
- Average cost: $0.0003 per email with gpt-4o-mini

### 4. Disable for Low-Value Emails

Consider disabling LLM for specific accounts:
```yaml
# Future enhancement - not yet implemented
llm:
  exclude_accounts:
    - spam@example.com
    - newsletter@example.com
```

## Monitoring and Logging

### Token Usage Tracking

The system logs token usage for cost monitoring:

```
INFO LLM summarization completed email_id=abc123 model=gpt-4o-mini input_tokens=450 output_tokens=120
```

### Error Handling

If LLM fails, the system:
1. Logs the error with `ai_summary_error` in database
2. Falls back to traditional preview
3. Still sends the notification to user
4. Does NOT retry (to avoid duplicate API calls)

Example error log:
```
WARN LLM summarization failed, using fallback email_id=abc123 error="timeout exceeded"
```

### Health Monitoring

Check LLM status:
```bash
# Check recent summarizations
mysql -u mail_user -p mail_to_tg -e "
SELECT
  COUNT(*) as total,
  SUM(CASE WHEN ai_summary IS NOT NULL THEN 1 ELSE 0 END) as summarized,
  SUM(CASE WHEN ai_summary_error IS NOT NULL THEN 1 ELSE 0 END) as errors
FROM email_messages
WHERE created_at > DATE_SUB(NOW(), INTERVAL 24 HOUR);"

# Check cache hit rate
redis-cli INFO stats | grep keyspace_hits
```

## Troubleshooting

### LLM Calls Failing

**Symptoms**: All emails show preview instead of summary

**Solutions**:
1. Check API key is valid:
   ```bash
   grep LLM_API_KEY .env
   ```

2. Test API endpoint:
   ```bash
   curl -H "Authorization: Bearer $LLM_API_KEY" \
        https://api.openai.com/v1/models
   ```

3. Check logs:
   ```bash
   journalctl -u telegram-service | grep "LLM"
   ```

4. Verify config loaded:
   ```bash
   # In telegram-service logs, you should see:
   # "LLM client initialized" model=gpt-4o-mini base_url=...
   ```

### Timeouts

**Symptoms**: Errors like "context deadline exceeded"

**Solutions**:
1. Increase timeout:
   ```yaml
   llm:
     timeout_seconds: 30  # Increased from 10
   ```

2. Use faster model (gpt-4o-mini instead of gpt-4o)

3. For local LLMs, ensure adequate GPU/RAM

### High API Costs

**Solutions**:
1. Switch to cheaper model (gpt-4o-mini)
2. Increase cache TTL:
   ```yaml
   llm:
     cache_ttl_hours: 72  # 3 days instead of 24h
   ```
3. Disable for high-volume, low-value accounts

### Local LLM Too Slow

**Solutions**:
1. Use smaller model:
   ```bash
   ollama pull llama3.2:1b  # Smaller than default
   ```

2. Increase timeout:
   ```yaml
   llm:
     timeout_seconds: 60
   ```

3. Enable GPU acceleration (if available)

4. Use quantized models (faster but slightly lower quality)

## Privacy Considerations

### Data Handling

- **Cloud LLMs** (OpenAI, OpenRouter): Email content sent to third-party servers
- **Local LLMs** (Ollama): All processing on your server, no external calls
- **Caching**: Summaries stored in Redis (encrypted at rest) and MariaDB

### Compliance

For sensitive emails:

1. **Use local LLM** (Ollama) to keep data on-premise
2. **Disable LLM** for accounts handling PII:
   ```yaml
   llm:
     enabled: false
   ```
3. **Review vendor DPAs** (Data Processing Agreements) for OpenAI/OpenRouter

## Future Enhancements

Planned features:

1. **Per-account LLM settings** - Enable/disable for specific accounts
2. **Custom prompts** - User-defined extraction rules
3. **Multi-language support** - Summaries in user's language
4. **Priority scoring** - Highlight urgent emails
5. **Smart categorization** - Auto-tag emails (bills, shipping, security)
6. **Thread summarization** - Summarize entire email threads
7. **Regenerate summary** - Button to re-run LLM with different prompt

## API Costs Estimation

Based on 1000 emails/day:

| Provider | Model | Daily Cost | Monthly Cost |
|----------|-------|------------|--------------|
| OpenAI | gpt-4o-mini | $0.27 | $8 |
| OpenAI | gpt-4o | $3.60 | $108 |
| OpenRouter | claude-3.5-sonnet | $5.40 | $162 |
| OpenRouter | gemini-flash (free) | $0 | $0 |
| Ollama | llama3.2 | $0 | $0 (hardware cost only) |

**Note**: Costs assume 1000 input tokens and 200 output tokens per email.

## Support

For issues or questions:
- Check logs: `journalctl -u telegram-service | grep LLM`
- Review configuration: `cat /etc/mail-to-tg/config.yaml`
- GitHub Issues: https://github.com/kexi/mail-to-tg/issues
