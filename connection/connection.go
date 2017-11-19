package connection

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xcoulon/go-url-shortener/configuration"
)

// New returns a new database connection.
func New(config *configuration.Configuration) (*gorm.DB, error) {
	logrus.Infof("Connecting to Postgres database using: host=`%s:%d` dbname=`%s` username=`%s`",
		config.GetPostgresHost(), config.GetPostgresPort(), config.GetPostgresDatabase(), config.GetPostgresUser())
	db, err := gorm.Open("postgres", config.GetPostgresConfig())
	if err != nil {
		return nil, errors.Wrap(err, "failed to open connection to database")
	}
	return db, nil
}
