package storage

import (
	"github.com/dchest/uniuri"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Repository the repository structure to create and retrieve Shortened URLs
type Repository struct {
	db *gorm.DB
}

// New returns a new repository configured with the given 'db'
func New(db *gorm.DB) *Repository {
	db.AutoMigrate(&ShortenedURL{})
	return &Repository{db: db}
}

//Create creates a new entry
func (r *Repository) Create(fullURL string) (*string, error) {
	shortURL := uniuri.NewLen(7)
	s := ShortenedURL{
		LongURL:  fullURL,
		ShortURL: shortURL,
	}
	err := r.db.Create(&s).Error
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create new shortened URL record in DB")
	}
	logrus.WithField("long_url", s.LongURL).WithField("short_url", s.ShortURL).Infof("Stored URL in DB")
	return &shortURL, nil
}

//Lookup looks-up an entry in the DB
func (r *Repository) Lookup(shortURL string) (*string, error) {
	logrus.WithField("short_url", shortURL).Debug("Looking-up by short_url in DB")
	var record ShortenedURL
	result := r.db.Where("short_url = ?", shortURL).First(&record)
	if result.RecordNotFound() {
		logrus.WithField("short_url", shortURL).Warn("No entry for short_url in DB")
		return nil, nil
	} else if result.Error != nil {
		return nil, errors.Wrapf(result.Error, "failed to look-up shortened URL record in DB")
	}
	return &record.LongURL, nil
}
