#!/bin/bash
set -e

echo "Setting up systemd services..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

# Copy service files
echo "Installing service files..."
cp systemd/mail-fetcher.service /etc/systemd/system/
cp systemd/telegram-service.service /etc/systemd/system/

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable services
echo "Enabling services..."
systemctl enable mail-fetcher.service
systemctl enable telegram-service.service

echo ""
echo "Services installed and enabled!"
echo ""
echo "To start services:"
echo "  systemctl start mail-fetcher"
echo "  systemctl start telegram-service"
echo ""
echo "To check status:"
echo "  systemctl status mail-fetcher"
echo "  systemctl status telegram-service"
echo ""
echo "To view logs:"
echo "  journalctl -u mail-fetcher -f"
echo "  journalctl -u telegram-service -f"
