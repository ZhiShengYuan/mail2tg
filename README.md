# Mail-to-Telegram

A production-ready Golang system that fetches emails from Gmail and IMAP providers (like QQmail) and delivers them to Telegram with HTML rendering and reply functionality.

## Features

- **Multi-Provider Support**: Gmail (via OAuth2 + Pub/Sub push) and IMAP (QQmail, etc.)
- **AI-Powered Email Summaries**: LLM-based summarization with structured data extraction (verification codes, amounts, dates)
- **Telegram Integration**: Real-time notifications with inline buttons
- **HTML Email Viewing**: Secure web interface with sanitized HTML rendering
- **Reply Functionality**: Reply to emails directly from Telegram via SMTP
- **Attachment Support**: Download links for email attachments
- **Secure**: AES-256-GCM encryption for credentials, HTML sanitization
- **Scalable**: Redis queue, MariaDB storage, systemd services
- **Production-Ready**: Systemd integration, logging, graceful shutdown

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Email Providers (Gmail API, QQmail IMAP)              │
└─────────────────┬───────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────────┐
│  Part 1: Mail Fetcher Service                          │
│  - Gmail API with Pub/Sub push notifications           │
│  - IMAP polling for QQmail                             │
│  - Email parsing and sanitization                      │
└─────────────────┬───────────────────────────────────────┘
                  │ Redis Queue
┌─────────────────▼───────────────────────────────────────┐
│  Part 2: Telegram Service                              │
│  - Telegram bot with commands                          │
│  - Web server for email viewing                        │
│  - SMTP client for replies                             │
└─────────────────────────────────────────────────────────┘
```

## Prerequisites

- Go 1.21 or later
- MariaDB 10.5+
- Redis 6.0+
- Linux server (for systemd)

## Quick Start

### 1. Clone and Build

```bash
cd /home/kexi/mail-to-tg
make deps
make build
```

### 2. Database Setup

```bash
# Create database and user
export DB_ROOT_PASSWORD=your_root_password
export DB_PASSWORD=your_mail_user_password
make create-db

# Run migrations
make migrate
```

### 3. Configuration

```bash
# Generate encryption key
make gen-key

# Copy and edit configuration
cp configs/.env.example .env
nano .env
```

Fill in:
- `DB_PASSWORD`: MariaDB password
- `REDIS_PASSWORD`: Redis password (if set)
- `TELEGRAM_BOT_TOKEN`: From @BotFather
- `ENCRYPTION_KEY`: From `make gen-key`
- `GMAIL_PROJECT_ID`: GCP project ID (if using Gmail)
- `LLM_API_KEY`: OpenAI API key for email summarization (optional)

### 4. Run Locally (Development)

```bash
# Terminal 1: Mail fetcher
make run-fetcher

# Terminal 2: Telegram service
make run-telegram
```

### 5. Production Deployment

```bash
# Build production binaries
make build-prod

# Install to system
sudo make install

# Setup systemd services
sudo make setup-services

# Start services
sudo make start

# Check status
sudo make status
```

## Telegram Bot Commands

- `/start` - Initialize bot and show welcome message
- `/link` - Link email account (Gmail OAuth or IMAP)
- `/accounts` - List all linked accounts
- `/unlink` - Remove an email account
- `/search <query>` - Search emails (coming soon)
- `/help` - Show help message

## Linking Email Accounts

### Gmail (OAuth2)

1. Set up Google Cloud Project:
   - Enable Gmail API
   - Create OAuth2 credentials
   - Download `credentials.json` to `/etc/mail-to-tg/`
   - Set up Pub/Sub topic and subscription

2. In Telegram, use `/link` → Select "Gmail (OAuth2)"
3. Follow OAuth flow

### IMAP (QQmail, etc.)

1. For QQmail:
   - Enable IMAP in QQ Mail settings
   - Generate app-specific password

2. In Telegram, use `/link` → Select "IMAP"
3. Provide:
   - Email address
   - IMAP server (e.g., `imap.qq.com`)
   - Port (usually `993`)
   - Username (your email)
   - Password (app-specific password)

## Email Notifications

When you receive an email, you'll get a Telegram message with:

- **Subject** and **Sender**
- **AI Summary** (if LLM enabled) with extracted:
  - Verification codes
  - Billing amounts
  - Due dates
  - Tracking numbers
  - Action items
- **Preview** (fallback if LLM disabled, first 200 characters)
- **Buttons**:
  - 🌐 View Full - Open HTML email in browser
  - ↩️ Reply - Start reply mode
  - ✅ Mark Read - Mark as read

### LLM Summarization

To enable AI-powered email summaries, configure an OpenAI-compatible API:

**OpenAI**:
```bash
LLM_API_KEY=sk-your-openai-key
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4o-mini
```

**OpenRouter** (access multiple models):
```bash
LLM_API_KEY=sk-your-openrouter-key
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_MODEL=anthropic/claude-3.5-sonnet
```

**Local LLM** (Ollama, LM Studio, etc.):
```bash
LLM_API_KEY=not-needed
LLM_BASE_URL=http://localhost:11434/v1
LLM_MODEL=llama3.2
```

Set `LLM_ENABLED=false` in config to disable summarization and use preview mode.

## Replying to Emails

1. Click **↩️ Reply** on any email notification
2. Type your reply message
3. Send - Reply will be sent via SMTP

## Security

- **Encryption**: All OAuth tokens and passwords encrypted with AES-256-GCM
- **HTML Sanitization**: Removes scripts, tracking pixels, dangerous elements
- **View Tokens**: 24-hour expiration for email view links
- **TLS**: HTTPS for web server (with Let's Encrypt)
- **User Isolation**: Dedicated system user with limited permissions

## Configuration Files

- `/etc/mail-to-tg/config.yaml` - Main configuration
- `/etc/mail-to-tg/.env` - Environment variables (secrets)
- `/etc/mail-to-tg/credentials.json` - Gmail OAuth credentials

## Storage

- **Database**: MariaDB at `/var/lib/mysql`
- **Attachments**: `/var/lib/mail-to-tg/attachments/`
- **Logs**: `journalctl -u mail-fetcher` / `journalctl -u telegram-service`

## Monitoring

### View Logs

```bash
# Mail fetcher logs
make logs-fetcher

# Telegram service logs
make logs-telegram
```

### Check Status

```bash
make status
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Backup

```bash
# Manual backup
sudo make backup

# Backups stored in /var/backups/mail-to-tg/
```

Set up cron for automated backups:
```bash
0 2 * * * /opt/mail-to-tg/scripts/backup.sh
```

## Troubleshooting

### Service won't start

```bash
# Check logs
journalctl -u mail-fetcher -n 50
journalctl -u telegram-service -n 50

# Check config
cat /etc/mail-to-tg/config.yaml

# Verify database connection
mysql -u mail_user -p mail_to_tg -e "SELECT 1"
```

### Gmail push notifications not working

1. Verify Pub/Sub topic permissions
2. Check watch expiration: expires every 7 days
3. Re-setup watch via API

### IMAP connection fails

1. Verify server and port (usually `imap.qq.com:993`)
2. Check TLS is enabled
3. Use app-specific password, not main password

## Development

### Project Structure

```
mail-to-tg/
├── cmd/                    # Main applications
├── internal/               # Private application code
│   ├── bot/               # Telegram bot
│   ├── fetcher/           # Email fetching
│   ├── notifier/          # Notifications
│   ├── smtp/              # Email sending
│   ├── storage/           # Database/Redis
│   ├── queue/             # Message queue
│   └── web/               # Web server
├── pkg/                   # Public libraries
│   ├── config/            # Configuration
│   ├── crypto/            # Encryption
│   ├── llm/               # LLM client for email summarization
│   ├── logger/            # Logging
│   └── models/            # Data models
├── migrations/            # Database migrations
├── configs/               # Config files
├── scripts/               # Deployment scripts
└── systemd/               # Systemd services
```

### Running Tests

```bash
make test
```

### Adding a New Email Provider

1. Create client in `internal/fetcher/<provider>/`
2. Implement message fetching
3. Add to `internal/fetcher/manager.go`
4. Update bot handlers for account linking

## License

MIT License - See LICENSE file

## Support

For issues and questions:
- GitHub Issues: https://github.com/kexi/mail-to-tg/issues
- Documentation: This README

## Roadmap

- [ ] Email search functionality
- [ ] Multiple recipients for forwarding
- [ ] Email filtering rules
- [ ] Scheduled email sending
- [ ] Advanced attachment handling
- [ ] Email templates for replies
- [ ] Multi-language support
- [ ] Mobile app integration
