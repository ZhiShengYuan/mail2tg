#!/bin/bash
set -e

BACKUP_DIR="/var/backups/mail-to-tg"
DATE=$(date +%Y%m%d_%H%M%S)

echo "Mail-to-Telegram Backup Script"
echo "==============================="

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
echo "Backing up database..."
mysqldump -u mail_user -p mail_to_tg > $BACKUP_DIR/mail_to_tg_$DATE.sql
gzip $BACKUP_DIR/mail_to_tg_$DATE.sql

# Backup attachments
echo "Backing up attachments..."
tar -czf $BACKUP_DIR/attachments_$DATE.tar.gz /var/lib/mail-to-tg/attachments/

# Backup config
echo "Backing up config..."
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /etc/mail-to-tg/

# Clean up old backups (keep last 7 days)
echo "Cleaning up old backups..."
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo ""
echo "Backup complete!"
echo "Backup location: $BACKUP_DIR"
ls -lh $BACKUP_DIR/*$DATE*
