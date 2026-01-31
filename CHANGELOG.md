# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [2.0.0] - 2026-01-31

### 🎯 Major Changes

#### Single JSON Configuration
- **BREAKING CHANGE**: Replaced YAML+JSON dual-config with single `config.json`
- All settings and secrets now in one file
- Simpler setup and maintenance
- No more YAML dependency

#### Automatic Database Migrations
- Database tables now created automatically on startup
- No manual migration commands needed
- Migration history tracked in `schema_migrations` table
- Services auto-apply pending migrations

#### LLM Email Summarization
- AI-powered email summaries in Telegram notifications
- OpenAI-compatible API support (OpenAI, OpenRouter, local LLMs)
- Extracts structured data: verification codes, amounts, dates, tracking numbers
- Redis caching (24h TTL) to minimize API costs
- Graceful fallback to preview mode on errors

### ✨ Added

- **LLM Integration** (`pkg/llm/`)
  - OpenAI-compatible client
  - Email summarization with structured data extraction
  - Support for multiple providers (OpenAI, OpenRouter, Ollama, etc.)
  - Token usage tracking for cost monitoring
  - Redis caching layer

- **Auto-Migration System** (`internal/storage/migrations.go`)
  - Automatic SQL migration execution
  - Migration tracking and history
  - No manual migration steps required

- **New Configuration Files**
  - `configs/config.json.example` - Development template
  - `configs/config.production.json.example` - Production template
  - `configs/README.md` - Configuration documentation

- **Database Schema**
  - `migrations/003_add_ai_summary.sql` - AI summary fields
  - AI summary storage: `ai_summary`, `ai_extracted_data`, etc.

### 🔧 Changed

- **Configuration System**
  - Switched from `config.yaml` + `secrets.json` to single `config.json`
  - Removed YAML dependency (`gopkg.in/yaml.v3`)
  - All struct tags changed from `yaml:` to `json:`
  - Config loading simplified (single JSON decoder)

- **Service Initialization**
  - Both services now run migrations on startup
  - Added `-migrations` flag for migration directory path
  - Default config path changed to `config.json`

- **Systemd Services**
  - Removed `EnvironmentFile` directive
  - Updated `ExecStart` to use `config.json`
  - Added `-migrations` flag to service commands

- **Installation Scripts**
  - `scripts/install.sh` now installs `config.json`
  - Copies migrations to `/opt/mail-to-tg/migrations`
  - Updated instructions for JSON config

- **Makefile**
  - `make run-*` commands use `config.json`
  - Deprecated `make migrate` (auto-migration now)
  - Updated all config paths

- **Documentation**
  - README.md - Updated for single JSON config
  - SETUP.md - Simplified setup instructions
  - docs/LLM_INTEGRATION.md - Comprehensive LLM guide
  - configs/README.md - Configuration documentation

### 🗑️ Removed

- **YAML Configuration**
  - Removed `config.yaml` support
  - Removed `secrets.json` (merged into `config.json`)
  - Removed `pkg/config/secrets.go`
  - Removed YAML dependency

- **Environment Variables**
  - No longer using `.env` files
  - Removed `envconfig` dependency
  - Removed `EnvironmentFile` from systemd

- **Manual Migrations**
  - `make migrate` deprecated (migrations automatic)
  - No need to run SQL files manually

### 🔒 Security

- `config.json` contains all secrets (permissions: 600)
- Automatic `.gitignore` for `config.json`
- AES-256-GCM encryption for stored credentials
- JWT secrets configurable per environment

### 📈 Performance

- LLM summaries cached in Redis (24h TTL)
- ~70% reduction in API calls via caching
- Graceful fallback prevents notification delays

### 💰 Cost Optimization

- Redis caching minimizes LLM API costs
- ~$0.27/day for 1000 emails with gpt-4o-mini
- Configurable model selection
- Support for free/local LLMs (Ollama)

## [1.0.0] - 2026-01-31

### Initial Release

- Gmail and IMAP email fetching
- Telegram bot integration
- HTML email viewing in web interface
- Email reply functionality via SMTP
- Attachment support
- Secure credential encryption
- Redis message queue
- MariaDB storage
- Systemd service integration

---

## Migration Guide

### From 1.x to 2.x

#### Configuration Migration

**Old system** (config.yaml + secrets.json):
```yaml
# config.yaml
database:
  host: localhost
  port: 3306
  password: ""  # In secrets.json
```

```json
// secrets.json
{
  "database": {
    "password": "secret123"
  }
}
```

**New system** (single config.json):
```json
{
  "database": {
    "host": "localhost",
    "port": 3306,
    "password": "secret123"
  }
}
```

#### Migration Steps

1. **Create new config**:
   ```bash
   cp configs/config.json.example config.json
   ```

2. **Merge your old configs**:
   - Copy settings from `config.yaml`
   - Copy secrets from `secrets.json` (or `.env`)
   - Put everything into `config.json`

3. **Update paths**:
   ```bash
   # If using systemd
   sudo cp config.json /etc/mail-to-tg/config.json
   sudo chown mail-to-tg:mail-to-tg /etc/mail-to-tg/config.json
   sudo chmod 600 /etc/mail-to-tg/config.json
   ```

4. **Remove old files**:
   ```bash
   rm config.yaml secrets.json .env
   ```

5. **Restart services**:
   ```bash
   sudo systemctl restart mail-fetcher telegram-service
   ```

#### Database Migration

No manual steps required! Migrations run automatically on service startup.

The services will:
- Create `schema_migrations` table
- Apply `003_add_ai_summary.sql` if not already applied
- Track migration history

### Breaking Changes Summary

- Configuration file changed from `config.yaml` to `config.json`
- No more separate `secrets.json` file
- No more `.env` file support
- Default config path changed in CLI flags
- YAML dependency removed
- Manual migrations replaced with auto-migration

---

## Version History

- **2.0.0** (2026-01-31) - Single JSON config, auto-migrations, LLM summarization
- **1.0.0** (2026-01-31) - Initial release

---

## Links

- [Documentation](README.md)
- [Setup Guide](SETUP.md)
- [LLM Integration](docs/LLM_INTEGRATION.md)
- [GitHub Repository](https://github.com/ZhiShengYuan/mail2tg)
