package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/models"
)

type MariaDB struct {
	db *sqlx.DB
}

func NewMariaDB(cfg *config.DatabaseConfig) (*MariaDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	return &MariaDB{db: db}, nil
}

func (m *MariaDB) Close() error {
	return m.db.Close()
}

func (m *MariaDB) Ping() error {
	return m.db.Ping()
}

// User operations
func (m *MariaDB) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, telegram_id, username, first_name, last_name, is_active)
		VALUES (:id, :telegram_id, :username, :first_name, :last_name, :is_active)`
	_, err := m.db.NamedExec(query, user)
	return err
}

func (m *MariaDB) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE telegram_id = ?`
	err := m.db.Get(&user, query, telegramID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (m *MariaDB) GetUserByID(id string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = ?`
	err := m.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (m *MariaDB) UpdateUser(user *models.User) error {
	query := `UPDATE users SET username = :username, first_name = :first_name,
		last_name = :last_name, is_active = :is_active, updated_at = NOW()
		WHERE id = :id`
	_, err := m.db.NamedExec(query, user)
	return err
}

// Email account operations
func (m *MariaDB) CreateEmailAccount(account *models.EmailAccount) error {
	query := `INSERT INTO email_accounts (
		id, user_id, provider, email_address, oauth_token_encrypted,
		oauth_refresh_token_encrypted, oauth_expiry, imap_server, imap_port,
		imap_username, imap_password_encrypted, smtp_server, smtp_port,
		smtp_username, smtp_password_encrypted, gmail_history_id,
		gmail_watch_expiration, is_active
	) VALUES (
		:id, :user_id, :provider, :email_address, :oauth_token_encrypted,
		:oauth_refresh_token_encrypted, :oauth_expiry, :imap_server, :imap_port,
		:imap_username, :imap_password_encrypted, :smtp_server, :smtp_port,
		:smtp_username, :smtp_password_encrypted, :gmail_history_id,
		:gmail_watch_expiration, :is_active
	)`
	_, err := m.db.NamedExec(query, account)
	return err
}

func (m *MariaDB) GetEmailAccountByID(id string) (*models.EmailAccount, error) {
	var account models.EmailAccount
	query := `SELECT * FROM email_accounts WHERE id = ?`
	err := m.db.Get(&account, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &account, err
}

func (m *MariaDB) GetEmailAccountsByUserID(userID string) ([]*models.EmailAccount, error) {
	var accounts []*models.EmailAccount
	query := `SELECT * FROM email_accounts WHERE user_id = ? ORDER BY created_at DESC`
	err := m.db.Select(&accounts, query, userID)
	return accounts, err
}

func (m *MariaDB) GetActiveEmailAccounts() ([]*models.EmailAccount, error) {
	var accounts []*models.EmailAccount
	query := `SELECT * FROM email_accounts WHERE is_active = TRUE`
	err := m.db.Select(&accounts, query)
	return accounts, err
}

func (m *MariaDB) UpdateEmailAccount(account *models.EmailAccount) error {
	query := `UPDATE email_accounts SET
		provider = :provider, email_address = :email_address,
		oauth_token_encrypted = :oauth_token_encrypted,
		oauth_refresh_token_encrypted = :oauth_refresh_token_encrypted,
		oauth_expiry = :oauth_expiry, imap_server = :imap_server,
		imap_port = :imap_port, imap_username = :imap_username,
		imap_password_encrypted = :imap_password_encrypted,
		smtp_server = :smtp_server, smtp_port = :smtp_port,
		smtp_username = :smtp_username, smtp_password_encrypted = :smtp_password_encrypted,
		gmail_history_id = :gmail_history_id, gmail_watch_expiration = :gmail_watch_expiration,
		is_active = :is_active, last_fetch_at = :last_fetch_at,
		last_error = :last_error, updated_at = NOW()
		WHERE id = :id`
	_, err := m.db.NamedExec(query, account)
	return err
}

func (m *MariaDB) DeleteEmailAccount(id string) error {
	query := `DELETE FROM email_accounts WHERE id = ?`
	_, err := m.db.Exec(query, id)
	return err
}

// Email message operations
func (m *MariaDB) CreateEmailMessage(email *models.EmailMessage) error {
	query := `INSERT INTO email_messages (
		id, account_id, message_id, thread_id, gmail_id, imap_uid,
		from_address, from_name, to_addresses, subject, date,
		text_body, html_body, sanitized_html, has_attachments, attachments,
		in_reply_to, ` + "`references`" + `, is_read, is_notified
	) VALUES (
		:id, :account_id, :message_id, :thread_id, :gmail_id, :imap_uid,
		:from_address, :from_name, :to_addresses, :subject, :date,
		:text_body, :html_body, :sanitized_html, :has_attachments, :attachments,
		:in_reply_to, :references, :is_read, :is_notified
	)`
	_, err := m.db.NamedExec(query, email)
	return err
}

func (m *MariaDB) GetEmailMessageByID(id string) (*models.EmailMessage, error) {
	var email models.EmailMessage
	query := `SELECT * FROM email_messages WHERE id = ?`
	err := m.db.Get(&email, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &email, err
}

func (m *MariaDB) GetEmailMessageByAccountAndMessageID(accountID, messageID string) (*models.EmailMessage, error) {
	var email models.EmailMessage
	query := `SELECT * FROM email_messages WHERE account_id = ? AND message_id = ?`
	err := m.db.Get(&email, query, accountID, messageID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &email, err
}

func (m *MariaDB) GetUnnotifiedEmails(limit int) ([]*models.EmailMessage, error) {
	var emails []*models.EmailMessage
	query := `SELECT * FROM email_messages
		WHERE is_notified = FALSE
		ORDER BY date ASC
		LIMIT ?`
	err := m.db.Select(&emails, query, limit)
	return emails, err
}

func (m *MariaDB) UpdateEmailMessage(email *models.EmailMessage) error {
	query := `UPDATE email_messages SET
		is_read = :is_read, is_notified = :is_notified,
		notified_at = :notified_at, updated_at = NOW()
		WHERE id = :id`
	_, err := m.db.NamedExec(query, email)
	return err
}

func (m *MariaDB) MarkEmailAsNotified(id string) error {
	query := `UPDATE email_messages SET is_notified = TRUE, notified_at = NOW() WHERE id = ?`
	_, err := m.db.Exec(query, id)
	return err
}

func (m *MariaDB) UpdateEmailSummary(emailID string, summary *string, extractedData *string, model *string, summaryError *string) error {
	query := `UPDATE email_messages SET
		ai_summary = ?,
		ai_extracted_data = ?,
		ai_summary_model = ?,
		ai_summary_at = IF(? IS NULL, NULL, NOW()),
		ai_summary_error = ?,
		updated_at = NOW()
		WHERE id = ?`
	_, err := m.db.Exec(query, summary, extractedData, model, summary, summaryError, emailID)
	return err
}

// Email view token operations
func (m *MariaDB) CreateEmailViewToken(token *models.EmailViewToken) error {
	query := `INSERT INTO email_view_tokens (id, email_id, token, expires_at)
		VALUES (:id, :email_id, :token, :expires_at)`
	_, err := m.db.NamedExec(query, token)
	return err
}

func (m *MariaDB) GetEmailViewToken(token string) (*models.EmailViewToken, error) {
	var viewToken models.EmailViewToken
	query := `SELECT * FROM email_view_tokens WHERE token = ? AND expires_at > NOW()`
	err := m.db.Get(&viewToken, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &viewToken, err
}

func (m *MariaDB) IncrementTokenViewCount(token string) error {
	query := `UPDATE email_view_tokens SET view_count = view_count + 1 WHERE token = ?`
	_, err := m.db.Exec(query, token)
	return err
}

func (m *MariaDB) DeleteExpiredTokens() error {
	query := `DELETE FROM email_view_tokens WHERE expires_at < NOW()`
	_, err := m.db.Exec(query)
	return err
}

// Sent reply operations
func (m *MariaDB) CreateSentReply(reply *models.SentReply) error {
	query := `INSERT INTO sent_replies (
		id, user_id, original_email_id, account_id, to_address,
		subject, body, smtp_message_id, error
	) VALUES (
		:id, :user_id, :original_email_id, :account_id, :to_address,
		:subject, :body, :smtp_message_id, :error
	)`
	_, err := m.db.NamedExec(query, reply)
	return err
}

func (m *MariaDB) GetSentRepliesByUserID(userID string, limit int) ([]*models.SentReply, error) {
	var replies []*models.SentReply
	query := `SELECT * FROM sent_replies WHERE user_id = ? ORDER BY sent_at DESC LIMIT ?`
	err := m.db.Select(&replies, query, userID, limit)
	return replies, err
}
