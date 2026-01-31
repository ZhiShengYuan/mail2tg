package models

import "time"

type User struct {
	ID         string    `db:"id" json:"id"`
	TelegramID int64     `db:"telegram_id" json:"telegram_id"`
	Username   *string   `db:"username" json:"username,omitempty"`
	FirstName  *string   `db:"first_name" json:"first_name,omitempty"`
	LastName   *string   `db:"last_name" json:"last_name,omitempty"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
