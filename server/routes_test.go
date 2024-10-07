package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDocRoutes(t *testing.T) {
	t.Parallel()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	srv := &ObsyncServer{}
	if assert.NoError(t, srv.GetOpenapiYaml(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.GetDocs(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.GetRedocStandaloneJs(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
