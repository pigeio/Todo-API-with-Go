package models

import "time"

type RefreshSession struct {
	JTI       string    `db:"jti"`
	UserID    int       `db:"user_id"`
	UserAgent string    `db:"user_agent"`
	IP        string    `db:"ip_address"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
