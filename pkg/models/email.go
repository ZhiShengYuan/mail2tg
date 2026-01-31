package models

import "time"

type EmailMessage struct {
	ID             string     `db:"id" json:"id"`
	AccountID      string     `db:"account_id" json:"account_id"`
	MessageID      string     `db:"message_id" json:"message_id"`
	ThreadID       *string    `db:"thread_id" json:"thread_id,omitempty"`
	GmailID        *string    `db:"gmail_id" json:"gmail_id,omitempty"`
	IMAPUID        *int64     `db:"imap_uid" json:"imap_uid,omitempty"`
	FromAddress    string     `db:"from_address" json:"from_address"`
	FromName       *string    `db:"from_name" json:"from_name,omitempty"`
	ToAddresses    *string    `db:"to_addresses" json:"to_addresses,omitempty"`
	Subject        *string    `db:"subject" json:"subject,omitempty"`
	Date           time.Time  `db:"date" json:"date"`
	TextBody       *string    `db:"text_body" json:"text_body,omitempty"`
	HTMLBody       *string    `db:"html_body" json:"html_body,omitempty"`
	SanitizedHTML  *string    `db:"sanitized_html" json:"sanitized_html,omitempty"`
	HasAttachments bool       `db:"has_attachments" json:"has_attachments"`
	Attachments    *string    `db:"attachments" json:"attachments,omitempty"` // JSON array
	InReplyTo      *string    `db:"in_reply_to" json:"in_reply_to,omitempty"`
	References     *string    `db:"references" json:"references,omitempty"`
	IsRead           bool       `db:"is_read" json:"is_read"`
	IsNotified       bool       `db:"is_notified" json:"is_notified"`
	NotifiedAt       *time.Time `db:"notified_at" json:"notified_at,omitempty"`
	AISummary        *string    `db:"ai_summary" json:"ai_summary,omitempty"`
	AIExtractedData  *string    `db:"ai_extracted_data" json:"ai_extracted_data,omitempty"` // JSON object
	AISummaryModel   *string    `db:"ai_summary_model" json:"ai_summary_model,omitempty"`
	AISummaryAt      *time.Time `db:"ai_summary_at" json:"ai_summary_at,omitempty"`
	AISummaryError   *string    `db:"ai_summary_error" json:"ai_summary_error,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

type EmailViewToken struct {
	ID        string    `db:"id" json:"id"`
	EmailID   string    `db:"email_id" json:"email_id"`
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	ViewCount int       `db:"view_count" json:"view_count"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type SentReply struct {
	ID              string     `db:"id" json:"id"`
	UserID          string     `db:"user_id" json:"user_id"`
	OriginalEmailID *string    `db:"original_email_id" json:"original_email_id,omitempty"`
	AccountID       string     `db:"account_id" json:"account_id"`
	ToAddress       string     `db:"to_address" json:"to_address"`
	Subject         *string    `db:"subject" json:"subject,omitempty"`
	Body            *string    `db:"body" json:"body,omitempty"`
	SentAt          time.Time  `db:"sent_at" json:"sent_at"`
	SMTPMessageID   *string    `db:"smtp_message_id" json:"smtp_message_id,omitempty"`
	Error           *string    `db:"error" json:"error,omitempty"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Path        string `json:"path"`
}
