package models

import "time"

type EmailAccount struct {
	ID                         string     `db:"id" json:"id"`
	UserID                     string     `db:"user_id" json:"user_id"`
	Provider                   string     `db:"provider" json:"provider"` // "gmail", "imap"
	EmailAddress               string     `db:"email_address" json:"email_address"`
	OAuthTokenEncrypted        *string    `db:"oauth_token_encrypted" json:"-"`
	OAuthRefreshTokenEncrypted *string    `db:"oauth_refresh_token_encrypted" json:"-"`
	OAuthExpiry                *time.Time `db:"oauth_expiry" json:"oauth_expiry,omitempty"`
	IMAPServer                 *string    `db:"imap_server" json:"imap_server,omitempty"`
	IMAPPort                   *int       `db:"imap_port" json:"imap_port,omitempty"`
	IMAPUsername               *string    `db:"imap_username" json:"imap_username,omitempty"`
	IMAPPasswordEncrypted      *string    `db:"imap_password_encrypted" json:"-"`
	SMTPServer                 *string    `db:"smtp_server" json:"smtp_server,omitempty"`
	SMTPPort                   *int       `db:"smtp_port" json:"smtp_port,omitempty"`
	SMTPUsername               *string    `db:"smtp_username" json:"smtp_username,omitempty"`
	SMTPPasswordEncrypted      *string    `db:"smtp_password_encrypted" json:"-"`
	GmailHistoryID             *int64     `db:"gmail_history_id" json:"gmail_history_id,omitempty"`
	GmailWatchExpiration       *time.Time `db:"gmail_watch_expiration" json:"gmail_watch_expiration,omitempty"`
	IsActive                   bool       `db:"is_active" json:"is_active"`
	LastFetchAt                *time.Time `db:"last_fetch_at" json:"last_fetch_at,omitempty"`
	LastError                  *string    `db:"last_error" json:"last_error,omitempty"`
	CreatedAt                  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                  time.Time  `db:"updated_at" json:"updated_at"`
}
