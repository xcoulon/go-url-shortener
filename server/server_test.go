package server_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xcoulon/go-url-shortener/configuration"
	"github.com/xcoulon/go-url-shortener/connection"
	"github.com/xcoulon/go-url-shortener/server"
	"github.com/xcoulon/go-url-shortener/storage"
)

func TestServer(t *testing.T) {
	config := configuration.New()
	db, err := connection.New(config)
	require.Nil(t, err)
	repository := storage.New(db)
	s := server.New(repository)

	t.Run("status", func(t *testing.T) {
		// given
		req := httptest.NewRequest(echo.GET, "/status", nil)
		rec := httptest.NewRecorder()
		// when
		s.ServeHTTP(rec, req)
		// then
		assert.Equal(t, 200, rec.Code)
	})

	t.Run("POST and GET", func(t *testing.T) {
		// given
		req1 := httptest.NewRequest(echo.POST, "http://localhost:8080/", strings.NewReader("full_url=http://redhat.com"))
		req1.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec1 := httptest.NewRecorder()
		// when
		s.ServeHTTP(rec1, req1)
		// then
		require.Equal(t, 201, rec1.Code)
		require.NotNil(t, rec1.Header()[echo.HeaderLocation])
		location := rec1.Header()[echo.HeaderLocation][0]
		// given

		req2 := httptest.NewRequest(echo.GET, location, nil)
		rec2 := httptest.NewRecorder()
		// when
		s.ServeHTTP(rec2, req2)
		// then
		require.Equal(t, 307, rec2.Code)
		require.NotNil(t, rec2.Header()[echo.HeaderLocation])
		assert.Equal(t, "http://redhat.com", rec2.Header()[echo.HeaderLocation][0])
	})

	t.Run("GET unknown", func(t *testing.T) {
		// given
		req := httptest.NewRequest(echo.GET, "http://localhost:8080/foo", nil)
		rec := httptest.NewRecorder()
		// when
		s.ServeHTTP(rec, req)
		// then
		require.Equal(t, 404, rec.Code)
		require.Nil(t, rec.Header()[echo.HeaderLocation])
	})
}
