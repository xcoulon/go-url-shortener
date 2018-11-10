package connection

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xcoulon/go-url-shortener/configuration"
)

// NewUserConnection returns a new database connection.
func NewUserConnection(config *configuration.Configuration) (*gorm.DB, error) {
	logrus.Infof("Connecting to Postgres database using: host=`%s:%d` dbname=`%s` username=`%s`",
		config.GetPostgresHost(), config.GetPostgresPort(), config.GetPostgresDatabase(), config.GetPostgresUser())
	db, err := gorm.Open("postgres", config.GetPostgresConfig())
	if err != nil {
		return nil, errors.Wrap(err, "failed to open connection to database")
	}
	return db, nil
}

// SetupUUIDExtension setup the extension to use UUID, which require superuse privileges
func SetupUUIDExtension(config *configuration.Configuration) error {
	logrus.Infof("Connecting to Postgres database using: host=`%s:%d` dbname=`%s` username=`%s`",
		config.GetPostgresHost(), config.GetPostgresPort(), config.GetPostgresDatabase(), config.GetPostgresSuperUser())
	db, err := gorm.Open("postgres", config.GetPostgresAdminConfig())
	if err != nil {
		return errors.Wrap(err, "failed to open connection to database")
	}
	// ensure that the Postgres DB has the "uuid-ossp" extension to generate UUIDs as the primary keys for the ShortenedURL records
	logrus.Info(`Adding the 'uuid-ossp' extension...`)
	err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error
	if err != nil {
		return errors.Wrap(err, "failed to setup the database")
	}
	return nil
}
