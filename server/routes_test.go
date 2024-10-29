package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/raian621/obsync-server/database"
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

func createTestDB(t *testing.T) *sql.DB {
	if database.ExistingDB() {
		if db, err := database.GetDB(); err != nil {
			t.FailNow()
		} else if db != nil {
			return db
		}
	}

	db, err := database.NewDB("test-routes.db?mode=memory")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.NoError(t, database.ApplyMigrations(db)) {
		t.FailNow()
	}

	return db
}
