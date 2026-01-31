-- Additional indexes for performance optimization

-- Index for finding unnotified emails
CREATE INDEX IF NOT EXISTS idx_unnotified ON email_messages(is_notified, created_at)
WHERE is_notified = FALSE;

-- Index for email search by sender
CREATE INDEX IF NOT EXISTS idx_from_address ON email_messages(from_address);

-- Index for email search by date range
CREATE INDEX IF NOT EXISTS idx_date ON email_messages(date DESC);

-- Index for active accounts lookup
CREATE INDEX IF NOT EXISTS idx_active_accounts ON email_accounts(is_active, last_fetch_at);

-- Index for expired tokens cleanup
CREATE INDEX IF NOT EXISTS idx_expired_tokens ON email_view_tokens(expires_at);
