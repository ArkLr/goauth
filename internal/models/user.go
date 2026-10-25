package models

import "time"

type User struct {
	Id           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
