# Configuration Files

Mail-to-Telegram uses **a single JSON configuration file** for simplicity:

## `config.json` - All Configuration 🎯

**Everything in one place**:
- Application settings (ports, workers, timeouts)
- Secrets (passwords, API keys, tokens)
- Feature flags
- All configuration

**Example**:
```json
{
  "environment": "production",
  "database": {
    "host": "localhost",
    "port": 3306,
    "name": "mail_to_tg",
    "user": "mail_user",
    "password": "your_password_here",
    "max_open_conns": 25,
    "max_idle_conns": 5
  },
  "telegram": {
    "bot_token": "123456:ABC-DEF...",
    "webhook_url": ""
  },
  "security": {
    "encryption_key": "base64_key...",
    "jwt_secret": "random_secret"
  },
  "llm": {
    "enabled": true,
    "api_key": "sk-...",
    "base_url": "https://api.openai.com/v1",
    "model": "gpt-4o-mini",
    "timeout_seconds": 10,
    "max_tokens": 300,
    "fallback_on_error": true,
    "cache_ttl_hours": 24
  },
  "web": {
    "host": "0.0.0.0",
    "port": 8080,
    "base_url": "http://localhost:8080"
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

## Setup Instructions

### Development

1. Copy example file:
   ```bash
   cp configs/config.json.example configs/config.json
   ```

2. Edit with your credentials:
   ```bash
   nano configs/config.json
   ```

3. Run services:
   ```bash
   make run-fetcher
   # or
   make run-telegram
   ```

### Production

1. Install the system:
   ```bash
   make install
   ```

2. Edit configuration:
   ```bash
   sudo nano /etc/mail-to-tg/config.json
   ```

3. Start services:
   ```bash
   sudo systemctl start mail-fetcher telegram-service
   ```

## Security Notes

- `config.json` is automatically ignored by git (`.gitignore`)
- File permissions should be `600` (read/write owner only)
- Never commit `config.json` to version control
- Use `config.json.example` as a template

## Benefits

✅ **Simple** - One file, not two
✅ **Clear** - JSON structure with validation
✅ **Easy** - No need to manage separate secret files
✅ **Auto-migrate** - Database tables created automatically on startup
