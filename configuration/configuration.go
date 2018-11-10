package configuration

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// BuildCommit lastest build commit (set by build script)
	BuildCommit = "unknown"
	// BuildTime set by build script
	BuildTime = "unknown"
)

const (
	// Constants for viper variable names. Will be used to set
	// default values as well as to get each value
	varPostgresHost                 = "postgres.host"
	varPostgresPort                 = "postgres.port"
	varPostgresDatabase             = "postgres.database"
	varPostgresUser                 = "postgres.user"
	varPostgresPassword             = "postgres.password"
	varPostgresSuperUser            = "postgres.superuser"
	varPostgresAdminPassword        = "postgres.admin.password"
	varPostgresSSLMode              = "postgres.sslmode"
	varPostgresConnectionTimeout    = "postgres.connection.timeout"
	varPostgresTransactionTimeout   = "postgres.transaction.timeout"
	varPostgresConnectionRetrySleep = "postgres.connection.retrysleep"
	varPostgresConnectionMaxIdle    = "postgres.connection.maxidle"
	varPostgresConnectionMaxOpen    = "postgres.connection.maxopen"
	varPathToConfig                 = "config.file"
	varLogLevel                     = "log.level"
)

// Configuration the application Configuration, based on ENV variables
type Configuration struct {
	v *viper.Viper
}

// New initializes a new Configuration from the ENV variables
func New() *Configuration {
	c := Configuration{
		v: viper.New(),
	}
	c.v.SetDefault(varPostgresHost, "localhost")
	c.v.SetDefault(varPostgresPort, 5432)
	c.v.SetDefault(varPostgresDatabase, "postgres")
	c.v.SetDefault(varPostgresUser, "postgres")
	c.v.SetDefault(varPostgresSuperUser, "postgres")
	c.v.SetDefault(varPostgresSSLMode, "disable")
	c.v.SetDefault(varPostgresConnectionTimeout, 5)
	c.v.SetDefault(varPostgresConnectionMaxIdle, -1)
	c.v.SetDefault(varPostgresConnectionMaxOpen, -1)
	c.v.SetDefault(varPostgresConnectionRetrySleep, time.Duration(time.Second))
	c.v.SetDefault(varPostgresTransactionTimeout, time.Duration(5*time.Minute))
	c.v.SetDefault(varPathToConfig, "./config.yaml")
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.v.SetDefault(varLogLevel, "info")
	c.v.SetConfigFile(c.GetPathToConfig())
	c.v.SetTypeByDefaultValue(true)
	err := c.v.ReadInConfig() // Find and read the config file
	logrus.WithField("path", c.GetPathToConfig()).Warn("loading config")
	// just use the default value(s) if the config file was not found
	if _, ok := err.(*os.PathError); ok {
		logrus.Warnf("no config file '%s' not found. Using default values", c.GetPathToConfig())
	} else if err != nil { // Handle other errors that occurred while reading the config file
		panic(fmt.Errorf("fatal error while reading the config file: %s", err))
	}
	setLogLevel(c.GetLogLevel())
	// monitor the changes in the config file
	c.v.WatchConfig()
	c.v.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithField("file", e.Name).Warn("Config file changed")
		setLogLevel(c.GetLogLevel())
	})
	return &c
}

func setLogLevel(logLevel string) {
	logrus.WithField("level", logLevel).Warn("setting log level")
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithField("level", logLevel).Fatalf("failed to start: %s", err.Error())
	}
	logrus.SetLevel(level)

}

// GetPostgresHost returns the postgres host as set via default, config file, or environment variable
func (c *Configuration) GetPostgresHost() string {
	return c.v.GetString(varPostgresHost)
}

// GetPostgresPort returns the postgres port as set via default, config file, or environment variable
func (c *Configuration) GetPostgresPort() int64 {
	return c.v.GetInt64(varPostgresPort)
}

// GetPostgresDatabase returns the postgres database as set via default, config file, or environment variable
func (c *Configuration) GetPostgresDatabase() string {
	return c.v.GetString(varPostgresDatabase)
}

// GetPostgresUser returns the postgres user as set via default, config file, or environment variable
func (c *Configuration) GetPostgresUser() string {
	return c.v.GetString(varPostgresUser)
}

// GetPostgresPassword returns the postgres password as set via default, config file, or environment variable
func (c *Configuration) GetPostgresPassword() string {
	return c.v.GetString(varPostgresPassword)
}

// GetPostgresSuperUser returns the postgres superuser as set via default, config file, or environment variable
func (c *Configuration) GetPostgresSuperUser() string {
	return c.v.GetString(varPostgresSuperUser)
}

// GetPostgresAdminPassword returns the postgres password as set via default, config file, or environment variable
func (c *Configuration) GetPostgresAdminPassword() string {
	return c.v.GetString(varPostgresAdminPassword)
}

// GetPostgresSSLMode returns the postgres sslmode as set via default, config file, or environment variable
func (c *Configuration) GetPostgresSSLMode() string {
	return c.v.GetString(varPostgresSSLMode)
}

// GetPostgresConnectionTimeout returns the postgres connection timeout as set via default, config file, or environment variable
func (c *Configuration) GetPostgresConnectionTimeout() int64 {
	return c.v.GetInt64(varPostgresConnectionTimeout)
}

// GetPostgresConnectionRetrySleep returns the number of seconds (as set via default, config file, or environment variable)
// to wait before trying to connect again
func (c *Configuration) GetPostgresConnectionRetrySleep() time.Duration {
	return c.v.GetDuration(varPostgresConnectionRetrySleep)
}

// GetPostgresTransactionTimeout returns the number of minutes to timeout a transaction
func (c *Configuration) GetPostgresTransactionTimeout() time.Duration {
	return c.v.GetDuration(varPostgresTransactionTimeout)
}

// GetPostgresConnectionMaxIdle returns the number of connections that should be keept alive in the database connection pool at
// any given time. -1 represents no restrictions/default behavior
func (c *Configuration) GetPostgresConnectionMaxIdle() int {
	return c.v.GetInt(varPostgresConnectionMaxIdle)
}

// GetPostgresConnectionMaxOpen returns the max number of open connections that should be open in the database connection pool.
// -1 represents no restrictions/default behavior
func (c *Configuration) GetPostgresConnectionMaxOpen() int {
	return c.v.GetInt(varPostgresConnectionMaxOpen)
}

//GetPostgresConfig returns the settings for opening a new connection on a PostgreSQL server
func (c *Configuration) GetPostgresConfig() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresUser(),
		c.GetPostgresPassword(),
		c.GetPostgresDatabase(),
		c.GetPostgresSSLMode(),
		c.GetPostgresConnectionTimeout())
}

//GetPostgresAdminConfig returns the settings for opening a new connection on a PostgreSQL server
func (c *Configuration) GetPostgresAdminConfig() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresSuperUser(),
		c.GetPostgresAdminPassword(),
		c.GetPostgresDatabase(),
		c.GetPostgresSSLMode(),
		c.GetPostgresConnectionTimeout())
}

// GetPathToConfig returns the path to the config file
func (c *Configuration) GetPathToConfig() string {
	return c.v.GetString(varPathToConfig)
}

// GetLogLevel returns the log level
func (c *Configuration) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}
