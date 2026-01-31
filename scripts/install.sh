#!/bin/bash
set -e

echo "Mail-to-Telegram Installation Script"
echo "====================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

# Create user
echo "Creating mail-to-tg user..."
if ! id -u mail-to-tg > /dev/null 2>&1; then
    useradd -r -s /bin/false mail-to-tg
    echo "User created"
else
    echo "User already exists"
fi

# Create directories
echo "Creating directories..."
mkdir -p /opt/mail-to-tg/bin
mkdir -p /etc/mail-to-tg
mkdir -p /etc/mail-to-tg/ssl
mkdir -p /var/lib/mail-to-tg/attachments
mkdir -p /var/log/mail-to-tg

# Set permissions
echo "Setting permissions..."
chown -R mail-to-tg:mail-to-tg /opt/mail-to-tg
chown -R mail-to-tg:mail-to-tg /etc/mail-to-tg
chown -R mail-to-tg:mail-to-tg /var/lib/mail-to-tg
chown -R mail-to-tg:mail-to-tg /var/log/mail-to-tg

chmod 755 /opt/mail-to-tg
chmod 750 /etc/mail-to-tg
chmod 750 /var/lib/mail-to-tg
chmod 750 /var/log/mail-to-tg

# Copy binaries if they exist
if [ -f "bin/mail-fetcher" ]; then
    echo "Installing mail-fetcher binary..."
    cp bin/mail-fetcher /opt/mail-to-tg/bin/
    chmod 755 /opt/mail-to-tg/bin/mail-fetcher
fi

if [ -f "bin/telegram-service" ]; then
    echo "Installing telegram-service binary..."
    cp bin/telegram-service /opt/mail-to-tg/bin/
    chmod 755 /opt/mail-to-tg/bin/telegram-service
fi

# Copy config if it doesn't exist
if [ ! -f "/etc/mail-to-tg/config.yaml" ]; then
    echo "Installing config..."
    cp configs/config.production.yaml /etc/mail-to-tg/config.yaml
    chown mail-to-tg:mail-to-tg /etc/mail-to-tg/config.yaml
    chmod 640 /etc/mail-to-tg/config.yaml
fi

# Copy .env.example if .env doesn't exist
if [ ! -f "/etc/mail-to-tg/.env" ]; then
    echo "Installing .env template..."
    cp configs/.env.example /etc/mail-to-tg/.env
    chown mail-to-tg:mail-to-tg /etc/mail-to-tg/.env
    chmod 600 /etc/mail-to-tg/.env
    echo "IMPORTANT: Edit /etc/mail-to-tg/.env with your credentials!"
fi

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit /etc/mail-to-tg/.env with your credentials"
echo "2. Generate encryption key: openssl rand -base64 32"
echo "3. Set up MariaDB database (run migrations)"
echo "4. Install systemd services: ./scripts/setup-services.sh"
echo "5. Start services: systemctl start mail-fetcher telegram-service"
