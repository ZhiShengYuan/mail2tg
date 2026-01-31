# Upgrading to v2.0

This guide helps you upgrade from v1.x to v2.0.

## What's New in v2.0

### 🎯 Single JSON Configuration
- **One file**: `config.json` (was: `config.yaml` + `secrets.json`)
- Simpler setup and maintenance
- All settings and secrets in one place

### 🔄 Automatic Migrations
- Database tables created automatically on startup
- No more manual `make migrate` commands
- Migration history tracked automatically

### 🤖 LLM Email Summarization
- AI-powered summaries in Telegram
- Extracts codes, amounts, dates automatically
- Works with OpenAI, OpenRouter, local LLMs

## Quick Upgrade (5 minutes)

### Step 1: Stop Services

```bash
sudo systemctl stop mail-fetcher telegram-service
```

### Step 2: Backup Current Config

```bash
sudo cp /etc/mail-to-tg/config.yaml /etc/mail-to-tg/config.yaml.backup
sudo cp /etc/mail-to-tg/secrets.json /etc/mail-to-tg/secrets.json.backup 2>/dev/null || true
sudo cp /etc/mail-to-tg/.env /etc/mail-to-tg/.env.backup 2>/dev/null || true
```

### Step 3: Create New Config

```bash
# Pull latest code
cd /path/to/mail-to-tg
git pull origin main

# Copy example config
sudo cp configs/config.production.json.example /etc/mail-to-tg/config.json
```

### Step 4: Migrate Your Settings

Edit `/etc/mail-to-tg/config.json` and copy your values:

```bash
sudo nano /etc/mail-to-tg/config.json
```

**From your old files, copy**:

| Old Location | New Location in config.json |
|--------------|---------------------------|
| `config.yaml` → `database.host` | `"database": {"host": "..."}` |
| `secrets.json` → `database.password` | `"database": {"password": "..."}` |
| `.env` → `DB_PASSWORD` | `"database": {"password": "..."}` |
| `secrets.json` → `telegram.bot_token` | `"telegram": {"bot_token": "..."}` |
| `secrets.json` → `security.encryption_key` | `"security": {"encryption_key": "..."}` |
| `secrets.json` → `llm.api_key` | `"llm": {"api_key": "..."}` |

### Step 5: Set Permissions

```bash
sudo chown mail-to-tg:mail-to-tg /etc/mail-to-tg/config.json
sudo chmod 600 /etc/mail-to-tg/config.json
```

### Step 6: Rebuild and Reinstall

```bash
cd /path/to/mail-to-tg
make build
sudo make install
sudo ./scripts/setup-services.sh
```

### Step 7: Start Services

```bash
sudo systemctl start mail-fetcher telegram-service
sudo systemctl status mail-fetcher telegram-service
```

### Step 8: Verify

Check logs to ensure migrations ran:

```bash
sudo journalctl -u telegram-service -n 50 | grep -i migration
```

You should see:
```
INFO Starting auto-migration
INFO Migration applied successfully file=001_initial_schema.sql
INFO Migration applied successfully file=002_add_indexes.sql
INFO Migration applied successfully file=003_add_ai_summary.sql
INFO Migrations completed count=3
```

### Step 9: Clean Up Old Files

Once everything works:

```bash
sudo rm /etc/mail-to-tg/config.yaml.backup
sudo rm /etc/mail-to-tg/secrets.json.backup
sudo rm /etc/mail-to-tg/.env.backup
```

## Configuration Example

**Complete `config.json` example**:

```json
{
  "environment": "production",
  "database": {
    "host": "localhost",
    "port": 3306,
    "name": "mail_to_tg",
    "user": "mail_user",
    "password": "YOUR_DB_PASSWORD",
    "max_open_conns": 50,
    "max_idle_conns": 10
  },
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "YOUR_REDIS_PASSWORD",
    "db": 0
  },
  "mail_fetcher": {
    "workers": 5,
    "imap_poll_interval": 60,
    "gmail": {
      "project_id": "YOUR_GCP_PROJECT_ID",
      "pubsub_topic": "gmail-notifications",
      "pubsub_subscription": "gmail-sub",
      "credentials_path": "/etc/mail-to-tg/credentials.json"
    }
  },
  "telegram": {
    "bot_token": "YOUR_BOT_TOKEN",
    "webhook_url": ""
  },
  "web": {
    "host": "0.0.0.0",
    "port": 8080,
    "base_url": "https://your-domain.com",
    "tls_enabled": true,
    "tls_cert": "/etc/mail-to-tg/ssl/cert.pem",
    "tls_key": "/etc/mail-to-tg/ssl/key.pem"
  },
  "security": {
    "encryption_key": "YOUR_ENCRYPTION_KEY",
    "jwt_secret": "YOUR_JWT_SECRET"
  },
  "storage": {
    "attachments_path": "/var/lib/mail-to-tg/attachments"
  },
  "logging": {
    "level": "info",
    "format": "json"
  },
  "llm": {
    "enabled": true,
    "base_url": "https://api.openai.com/v1",
    "api_key": "YOUR_OPENAI_API_KEY",
    "model": "gpt-4o-mini",
    "timeout_seconds": 10,
    "max_tokens": 300,
    "fallback_on_error": true,
    "cache_ttl_hours": 24
  }
}
```

## Troubleshooting

### Services won't start

**Check config syntax**:
```bash
# Validate JSON
python3 -m json.tool /etc/mail-to-tg/config.json
```

**Check permissions**:
```bash
ls -la /etc/mail-to-tg/config.json
# Should be: -rw------- mail-to-tg mail-to-tg
```

**Check logs**:
```bash
sudo journalctl -u telegram-service -n 100
```

### Config not found

Make sure path is correct in systemd:
```bash
# Check service file
cat /etc/systemd/system/telegram-service.service | grep config

# Should show:
# ExecStart=/opt/mail-to-tg/bin/telegram-service -config /etc/mail-to-tg/config.json
```

If wrong, reinstall:
```bash
sudo ./scripts/setup-services.sh
sudo systemctl daemon-reload
```

### Migrations not running

Check migration directory exists:
```bash
ls -la /opt/mail-to-tg/migrations/
```

Should show:
```
001_initial_schema.sql
002_add_indexes.sql
003_add_ai_summary.sql
```

If missing:
```bash
sudo mkdir -p /opt/mail-to-tg/migrations
sudo cp migrations/*.sql /opt/mail-to-tg/migrations/
sudo chown -R mail-to-tg:mail-to-tg /opt/mail-to-tg/migrations
```

### LLM features not working

1. Check LLM is enabled in config:
   ```json
   "llm": {
     "enabled": true,
     ...
   }
   ```

2. Verify API key is set:
   ```bash
   sudo cat /etc/mail-to-tg/config.json | grep -A 8 '"llm"'
   ```

3. Test API connectivity:
   ```bash
   curl -H "Authorization: Bearer YOUR_API_KEY" \
        https://api.openai.com/v1/models
   ```

4. Check logs for LLM errors:
   ```bash
   sudo journalctl -u telegram-service | grep -i llm
   ```

## Breaking Changes

### Removed Features

- ❌ No more `.env` file support
- ❌ No more separate `secrets.json`
- ❌ No more `config.yaml`
- ❌ Manual `make migrate` command (automatic now)

### Changed Behavior

- ⚠️ Config file path changed: `config.yaml` → `config.json`
- ⚠️ Default paths updated in all services
- ⚠️ Systemd services no longer use `EnvironmentFile`

## Rollback Plan

If you need to rollback to v1.x:

```bash
# Stop services
sudo systemctl stop mail-fetcher telegram-service

# Restore backups
sudo cp /etc/mail-to-tg/config.yaml.backup /etc/mail-to-tg/config.yaml
sudo cp /etc/mail-to-tg/secrets.json.backup /etc/mail-to-tg/secrets.json

# Checkout old version
cd /path/to/mail-to-tg
git checkout v1.0.0  # or previous commit

# Rebuild
make build
sudo make install

# Start services
sudo systemctl start mail-fetcher telegram-service
```

## Need Help?

- 📖 [Documentation](README.md)
- 🔧 [Setup Guide](SETUP.md)
- 🤖 [LLM Integration Guide](docs/LLM_INTEGRATION.md)
- 📝 [Changelog](CHANGELOG.md)
- 🐛 [GitHub Issues](https://github.com/ZhiShengYuan/mail2tg/issues)
