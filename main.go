package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/xcoulon/go-url-shortener/configuration"
	"github.com/xcoulon/go-url-shortener/connection"
	"github.com/xcoulon/go-url-shortener/server"
	"github.com/xcoulon/go-url-shortener/storage"

	"github.com/jinzhu/gorm"
)

func main() {
	// load the logging configuration and init the logger
	// logrus.SetFormatter(&logrus.JSONFormatter{})

	config := configuration.New()
	// load the configuration and init the storage
	err := connection.SetupUUIDExtension(config)
	if err != nil {
		logrus.Fatalf("failed to start: %s", err.Error())
	}
	// load the configuration and init the storage
	db, err := connection.NewUserConnection(config)
	if err != nil {
		logrus.Fatalf("failed to start: %s", err.Error())
	}
	repository := storage.New(db)
	// handle shutdown
	go handleShutdown(db)
	s := server.New(repository)
	// listen and serve on 0.0.0.0:8080
	s.Start(":8080")
}

func handleShutdown(db *gorm.DB) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	// handle ctrl+c event here
	// for example, close database
	logrus.Warn("Closing DB connection before complete shutdown")
	err := db.Close()
	if err != nil {
		logrus.Errorf("error while closing the connection to the database: %v", err)
	}
	os.Exit(0)
}
