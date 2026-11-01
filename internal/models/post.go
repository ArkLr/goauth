package models

import (
	"time"
)

type Post struct {
	Id        int64
	UserId    int64
	Headline  string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time

	AuthorEmail string // filled by List only
}
