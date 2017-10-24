package storage

import (
	"time"

	uuid "github.com/satori/uuid.go"
)

// ShortenedURL the structure for shortened URLs
type ShortenedURL struct {
	ID        uuid.UUID `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
	CreatedAt time.Time
	LongURL   string
	ShortURL  string
}

//TableName set ShortenedURL's table name to be `urls`
func (ShortenedURL) TableName() string {
	return "urls"
}
