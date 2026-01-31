-- Mail to Telegram Database Schema

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_telegram_id (telegram_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Email accounts table
CREATE TABLE IF NOT EXISTS email_accounts (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    email_address VARCHAR(255) NOT NULL,

    -- Gmail OAuth2
    oauth_token_encrypted TEXT,
    oauth_refresh_token_encrypted TEXT,
    oauth_expiry TIMESTAMP NULL,

    -- IMAP/SMTP credentials
    imap_server VARCHAR(255),
    imap_port INT,
    imap_username VARCHAR(255),
    imap_password_encrypted TEXT,
    smtp_server VARCHAR(255),
    smtp_port INT,
    smtp_username VARCHAR(255),
    smtp_password_encrypted TEXT,

    -- Gmail push state
    gmail_history_id BIGINT,
    gmail_watch_expiration TIMESTAMP NULL,

    is_active BOOLEAN DEFAULT TRUE,
    last_fetch_at TIMESTAMP NULL,
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_email (user_id, email_address),
    INDEX idx_user_id (user_id),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Email messages table
CREATE TABLE IF NOT EXISTS email_messages (
    id CHAR(36) PRIMARY KEY,
    account_id CHAR(36) NOT NULL,
    message_id VARCHAR(255) NOT NULL,
    thread_id VARCHAR(255),
    gmail_id VARCHAR(255),
    imap_uid BIGINT,

    from_address VARCHAR(500) NOT NULL,
    from_name VARCHAR(500),
    to_addresses TEXT,
    subject TEXT,
    date TIMESTAMP NOT NULL,

    text_body MEDIUMTEXT,
    html_body MEDIUMTEXT,
    sanitized_html MEDIUMTEXT,

    has_attachments BOOLEAN DEFAULT FALSE,
    attachments JSON,

    in_reply_to VARCHAR(255),
    `references` TEXT,

    is_read BOOLEAN DEFAULT FALSE,
    is_notified BOOLEAN DEFAULT FALSE,
    notified_at TIMESTAMP NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (account_id) REFERENCES email_accounts(id) ON DELETE CASCADE,
    UNIQUE KEY unique_account_message (account_id, message_id),
    INDEX idx_account_date (account_id, date DESC),
    INDEX idx_thread (thread_id),
    INDEX idx_gmail_id (gmail_id),
    INDEX idx_is_read (is_read),
    INDEX idx_is_notified (is_notified)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Email view tokens table
CREATE TABLE IF NOT EXISTS email_view_tokens (
    id CHAR(36) PRIMARY KEY,
    email_id CHAR(36) NOT NULL,
    token CHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    view_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (email_id) REFERENCES email_messages(id) ON DELETE CASCADE,
    INDEX idx_token (token),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Sent replies table
CREATE TABLE IF NOT EXISTS sent_replies (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    original_email_id CHAR(36),
    account_id CHAR(36) NOT NULL,
    to_address VARCHAR(500) NOT NULL,
    subject TEXT,
    body TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    smtp_message_id VARCHAR(255),
    error TEXT,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (account_id) REFERENCES email_accounts(id) ON DELETE CASCADE,
    INDEX idx_user_sent (user_id, sent_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
