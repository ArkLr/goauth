package models

import (
	"time"
)

type Post struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"userId"`
	Headline  string    `json:"headline"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	AuthorEmail string `json:"authorEmail,omitempty"` // filled by List only
}
