package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/xcoulon/go-url-shortener/configuration"
	"github.com/xcoulon/go-url-shortener/storage"
)

// New instanciates a new Echo server
func New(repository *storage.Repository) *echo.Echo {
	// starts the HTTP engine to handle requests
	e := echo.New()
	// graceful handle of errors, i.e., just logging with the same logger as everywhere else in the app.
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok {
			logrus.WithField("code", he.Code).Error(he.Message)
			if msg, ok := he.Message.(string); ok {
				c.String(he.Code, msg)
			} else {
				c.NoContent(he.Code)
			}
		}
	}
	e.POST("/", CreateURL(repository))
	e.GET("/:shortURL", RetrieveURL(repository))
	e.GET("/status", Status())
	return e
}

// Status returns a basic `ping/pong` handler
func Status() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("build.time: %s - build.commit: %s 👷‍♂️", configuration.BuildTime, configuration.BuildCommit))
	}
}

// CreateURL returns a handler to create an db record from the `full_url` form param of the request.
func CreateURL(repository *storage.Repository) echo.HandlerFunc {
	return func(c echo.Context) error {
		scheme := c.Scheme()
		host := c.Request().Host
		logrus.Debugf("Processing incoming request on %s://%s%s", scheme, host, c.Request().URL)
		fullURL := c.FormValue("full_url")
		if fullURL == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing `full_url` form param in request")
		}
		shortURL, err := repository.Create(fullURL)
		if err != nil {
			logrus.Errorf("failed to store url: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to store URL")
		}
		location := fmt.Sprintf("%s://%s/%s", scheme, host, *shortURL)
		logrus.Infof("Generated location for further usage: %s", location)
		c.Response().Header().Set(echo.HeaderLocation, location)
		c.String(http.StatusCreated, location)
		return nil
	}
}

// RetrieveURL returns a handler that retrieves the full URL from the `shortURL` request param.
// Returns a `Temporary Redirect` response with the result or `Not Found` if no match was found.
func RetrieveURL(repository *storage.Repository) echo.HandlerFunc {
	return func(c echo.Context) error {
		shortURL := c.Param("shortURL")
		fullURL, err := repository.Lookup(shortURL)
		if err != nil {
			logrus.Errorf("failed to retrieve url: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve URL")
		} else if fullURL == nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("No record found for '%s'", shortURL))
		}
		var result string
		if !strings.HasPrefix(*fullURL, "http://") && !strings.HasPrefix(*fullURL, "https://") {
			result = "http://" + *fullURL
		} else {
			result = *fullURL
		}
		c.Response().Header().Set(echo.HeaderLocation, result)
		c.Response().WriteHeader(http.StatusTemporaryRedirect)
		return nil
	}
}
