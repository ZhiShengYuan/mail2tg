# Mail-to-Telegram Setup Guide

## Prerequisites Installation

### 1. Install MariaDB

```bash
sudo apt update
sudo apt install mariadb-server
sudo mysql_secure_installation
```

### 2. Install Redis

```bash
sudo apt install redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
```

## Database Setup

### 1. Create Database and User

```bash
# Login to MariaDB as root
sudo mysql -u root -p

# Create database
CREATE DATABASE mail_to_tg CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# Create user (change password!)
CREATE USER 'mail_user'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON mail_to_tg.* TO 'mail_user'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 2. Run Migrations

```bash
cd /home/kexi/mail-to-tg

# Method 1: Using Makefile
export DB_PASSWORD=your_secure_password
make migrate

# Method 2: Manual
mysql -u mail_user -p mail_to_tg < migrations/001_initial_schema.sql
mysql -u mail_user -p mail_to_tg < migrations/002_add_indexes.sql
```

## Configuration

### 1. Generate Encryption Key

```bash
# Generate 32-byte encryption key
openssl rand -base64 32
# Save this output - you'll need it for .env file
```

### 2. Create Environment File

```bash
cp configs/.env.example .env
nano .env
```

Fill in the following:

```env
# Database
DB_PASSWORD=your_secure_password

# Redis (if password is set)
REDIS_PASSWORD=your_redis_password

# Telegram Bot
# Create bot via @BotFather on Telegram
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11

# Security
# Use the key generated above
ENCRYPTION_KEY=your_32_byte_base64_key_here
JWT_SECRET=your_random_jwt_secret

# Gmail API (optional, only if using Gmail)
GMAIL_PROJECT_ID=your-gcp-project-id
```

### 3. Telegram Bot Setup

1. Open Telegram and find @BotFather
2. Send `/newbot` command
3. Follow prompts to create your bot
4. Copy the bot token to your `.env` file

### 4. Gmail Setup (Optional)

If you want to support Gmail:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable Gmail API
4. Create OAuth 2.0 credentials
5. Download credentials as `credentials.json`
6. Set up Pub/Sub:
   ```bash
   # Create topic
   gcloud pubsub topics create gmail-notifications

   # Create subscription
   gcloud pubsub subscriptions create gmail-sub --topic=gmail-notifications

   # Grant Gmail API permission
   gcloud pubsub topics add-iam-policy-binding gmail-notifications \
     --member=serviceAccount:gmail-api-push@system.gserviceaccount.com \
     --role=roles/pubsub.publisher
   ```
7. Copy `credentials.json` to `/etc/mail-to-tg/credentials.json` (after installation)

## Local Development

### Build and Run

```bash
# Build binaries
make build

# Run mail-fetcher (Terminal 1)
make run-fetcher

# Run telegram-service (Terminal 2)
make run-telegram
```

### Testing

1. Open Telegram and find your bot
2. Send `/start` command
3. Send `/link` to add an email account
4. For QQmail:
   - Select "IMAP"
   - Server: `imap.qq.com`
   - Port: `993`
   - Username: Your QQ email
   - Password: App-specific password (enable in QQ Mail settings)

## Production Deployment

### 1. Build Production Binaries

```bash
make build-prod
```

### 2. Install to System

```bash
# Install binaries and directories (requires root)
sudo make install

# Edit production config
sudo nano /etc/mail-to-tg/config.yaml

# Edit environment file
sudo nano /etc/mail-to-tg/.env
```

### 3. Setup Systemd Services

```bash
# Install and enable services
sudo make setup-services

# Start services
sudo make start

# Check status
sudo make status
```

### 4. View Logs

```bash
# Mail fetcher logs
sudo make logs-fetcher

# Telegram service logs
sudo make logs-telegram

# Or directly with journalctl
sudo journalctl -u mail-fetcher -f
sudo journalctl -u telegram-service -f
```

## Web Server Setup (For Email Viewing)

### Option 1: Direct Access (Development)

The web server runs on port 8080 by default. Access at `http://your-server-ip:8080`

### Option 2: Reverse Proxy with Nginx (Production)

```bash
sudo apt install nginx certbot python3-certbot-nginx

# Create nginx config
sudo nano /etc/nginx/sites-available/mail-to-tg
```

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/mail-to-tg /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Get SSL certificate
sudo certbot --nginx -d your-domain.com

# Update config.yaml
sudo nano /etc/mail-to-tg/config.yaml
# Set web.base_url to https://your-domain.com
# Set web.tls_enabled to false (nginx handles TLS)

# Restart services
sudo make restart
```

## Backup

### Manual Backup

```bash
sudo /opt/mail-to-tg/scripts/backup.sh
```

### Automated Backups with Cron

```bash
sudo crontab -e

# Add line for daily backup at 2 AM
0 2 * * * /opt/mail-to-tg/scripts/backup.sh
```

Backups are stored in `/var/backups/mail-to-tg/`

## Maintenance

### Update Code

```bash
cd /home/kexi/mail-to-tg
git pull  # If using git
make build-prod
sudo make stop
sudo make install
sudo make start
```

### Database Cleanup

```bash
# Clean expired view tokens
mysql -u mail_user -p mail_to_tg -e "DELETE FROM email_view_tokens WHERE expires_at < NOW()"

# Clean old emails (older than 90 days)
mysql -u mail_user -p mail_to_tg -e "DELETE FROM email_messages WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY)"
```

### Monitor Disk Space

```bash
# Check attachment storage
du -sh /var/lib/mail-to-tg/attachments

# Check database size
sudo mysql -u root -p -e "SELECT table_name, ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)' FROM information_schema.TABLES WHERE table_schema = 'mail_to_tg' ORDER BY (data_length + index_length) DESC;"
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u mail-fetcher -n 100
sudo journalctl -u telegram-service -n 100

# Verify config
sudo cat /etc/mail-to-tg/config.yaml
sudo cat /etc/mail-to-tg/.env

# Test database connection
mysql -u mail_user -p mail_to_tg -e "SELECT 1"

# Test Redis connection
redis-cli ping
```

### No Emails Being Fetched

```bash
# Check mail-fetcher logs
sudo journalctl -u mail-fetcher -f

# Verify account is active in database
mysql -u mail_user -p mail_to_tg -e "SELECT id, email_address, provider, is_active, last_fetch_at, last_error FROM email_accounts"

# Test IMAP connection manually
openssl s_client -connect imap.qq.com:993 -crlf
```

### Telegram Bot Not Responding

```bash
# Check telegram-service logs
sudo journalctl -u telegram-service -f

# Verify bot token
curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getMe

# Check Redis connection
redis-cli ping

# Restart service
sudo systemctl restart telegram-service
```

### Gmail Push Not Working

1. Check Pub/Sub subscription is receiving messages
2. Verify watch hasn't expired (renews every 7 days)
3. Check service account permissions
4. Review gmail client logs

## Security Best Practices

1. **Strong Passwords**: Use strong, unique passwords for database and Redis
2. **Encryption Key**: Keep encryption key secure, never commit to git
3. **File Permissions**:
   ```bash
   sudo chmod 600 /etc/mail-to-tg/.env
   sudo chmod 640 /etc/mail-to-tg/config.yaml
   ```
4. **Firewall**: Only expose necessary ports
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```
5. **Updates**: Keep system and dependencies updated
6. **SSL/TLS**: Always use HTTPS in production

## Quick Reference

### Makefile Commands

```bash
make build          # Build binaries
make build-prod     # Build production binaries
make test           # Run tests
make install        # Install to system
make setup-services # Setup systemd services
make start          # Start services
make stop           # Stop services
make restart        # Restart services
make status         # Show service status
make logs-fetcher   # View mail-fetcher logs
make logs-telegram  # View telegram-service logs
make backup         # Run backup
make gen-key        # Generate encryption key
make help           # Show all commands
```

### Important Paths

- **Binaries**: `/opt/mail-to-tg/bin/`
- **Config**: `/etc/mail-to-tg/config.yaml`
- **Environment**: `/etc/mail-to-tg/.env`
- **Attachments**: `/var/lib/mail-to-tg/attachments/`
- **Logs**: `journalctl -u mail-fetcher` / `journalctl -u telegram-service`
- **Backups**: `/var/backups/mail-to-tg/`

## Support

For issues and questions:
- Check logs first
- Review this setup guide
- Check README.md for architecture details
- Open an issue on GitHub (if applicable)
