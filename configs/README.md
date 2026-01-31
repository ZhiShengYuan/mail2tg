# Configuration Files

Mail-to-Telegram uses **two configuration files** for clarity and security:

## 1. `config.yaml` - Application Settings ⚙️

**Non-sensitive configuration** (can be version controlled):
- Server ports and hosts
- Worker counts and timeouts
- Feature toggles
- File paths
- Logging settings

**Example**:
```yaml
environment: production

database:
  host: localhost
  port: 3306
  name: mail_to_tg
  user: mail_user
  password: ""  # Set in secrets.json

web:
  host: 0.0.0.0
  port: 8080

mail_fetcher:
  workers: 3
  imap_poll_interval: 60

llm:
  enabled: true
  timeout_seconds: 10
  max_tokens: 300
```

## 2. `secrets.json` - Secrets Only 🔐

**Sensitive data** (NOT version controlled, in `.gitignore`):
- Database passwords
- API keys
- Bot tokens
- Encryption keys

**Example**:
```json
{
  "database": {
    "password": "your_db_password"
  },
  "telegram": {
    "bot_token": "123456:ABC-DEF..."
  },
  "security": {
    "encryption_key": "base64_key...",
    "jwt_secret": "random_secret"
  },
  "llm": {
    "api_key": "sk-...",
    "base_url": "https://api.openai.com/v1",
    "model": "gpt-4o-mini"
  }
}
```

## Setup Instructions

### Development

1. Copy example files:
   ```bash
   cp configs/config.yaml.example configs/config.yaml  # if needed
   cp configs/secrets.json.example configs/secrets.json
   ```

2. Edit `secrets.json` with your credentials:
   ```bash
   nano configs/secrets.json
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

2. Edit secrets:
   ```bash
   sudo nano /etc/mail-to-tg/secrets.json
   ```

3. (Optional) Customize settings:
   ```bash
   sudo nano /etc/mail-to-tg/config.yaml
   ```

4. Start services:
   ```bash
   sudo systemctl start mail-fetcher telegram-service
   ```

## How It Works

The application loads configuration in this order:

1. **Read `config.yaml`** - Loads all application settings
2. **Read `secrets.json`** - Loads secrets from same directory
3. **Merge** - Secrets override empty values in config
4. **Start** - Application runs with merged configuration

## Security Notes

- `secrets.json` is automatically ignored by git (`.gitignore`)
- `secrets.json` permissions should be `600` (read/write owner only)
- Never commit `secrets.json` to version control
- Use `secrets.json.example` as a template only

## Migration from .env

If you're upgrading from the old `.env` system:

1. Create `secrets.json`:
   ```bash
   cp configs/secrets.json.example configs/secrets.json
   ```

2. Copy values from `.env` to `secrets.json`:
   - `DB_PASSWORD` → `database.password`
   - `TELEGRAM_BOT_TOKEN` → `telegram.bot_token`
   - `ENCRYPTION_KEY` → `security.encryption_key`
   - `LLM_API_KEY` → `llm.api_key`
   - etc.

3. Delete `.env` file:
   ```bash
   rm .env
   ```

4. Restart services - they'll use the new configuration automatically
