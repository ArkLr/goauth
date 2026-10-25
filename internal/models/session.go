package models

import "time"

type Session struct {
	Token     string
	UserId    int64
	CSRFToken string
	ExpiresAt time.Time
}
