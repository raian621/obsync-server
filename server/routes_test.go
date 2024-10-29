package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestUserLoginAndLogout(t *testing.T) {
	e := echo.New()
	db := createTestDB(t)

	// create test user
	user, err := database.CreateUser(db, "test-user", "test-user@example.com", "not a password")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	credentials := map[string]string{
		"username": user.Username,
		"password": "not a password",
	}

	// login as user with correct credentials
	body, err := json.Marshal(credentials)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	srv, err := NewServer(db, t.TempDir())
	if !assert.NoError(t, err) {
		t.Log(err)
		t.FailNow()
	}
	if assert.NoError(t, srv.PostUserLogin(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
	var cookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "OBSYNC_SESSION_ID" {
			cookie = c
		}
	}
	assert.NotNil(t, cookie)

	// login as user with incorrect password
	credentials["password"] = "wrong password"
	body, err = json.Marshal(credentials)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.PostUserLogin(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}

	// login as user with incorrect username
	credentials["username"] = "wrong username"
	credentials["password"] = "not a password"
	body, err = json.Marshal(credentials)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewBuffer(body))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.PostUserLogin(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}

	// logout as user
	req = httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
	req.AddCookie(cookie)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.PostUserLogout(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// delete user
	if err := database.DeleteUser(db, user.Id); err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestPostUser(t *testing.T) {
	testCases := []struct {
		name     string
		user     map[string]string
		wantCode int
	}{
		{
			name: "valid user",
			user: map[string]string{
				"username": "test-user",
				"email":    "test-user@example.com",
				"password": "not a password",
			},
			wantCode: http.StatusOK,
		},
		{
			name: "username too long",
			user: map[string]string{
				"username": strings.Repeat("a", 101),
				"email":    "test-user@example.com",
				"password": "not a password",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "username too short",
			user: map[string]string{
				"username": "",
				"email":    "test-user@example.com",
				"password": "not a password",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "email too long",
			user: map[string]string{
				"username": "test-user",
				"email":    strings.Repeat("a", 200) + "@example.com",
				"password": "not a password",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "email too short",
			user: map[string]string{
				"username": "test-user",
				"email":    "",
				"password": "not a password",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "email invalid",
			user: map[string]string{
				"username": "test-user",
				"email":    "invalid-email",
				"password": "not a password",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "password too short",
			user: map[string]string{
				"username": "test-user",
				"email":    "test-user@example.com",
				"password": "",
			},
			wantCode: http.StatusBadRequest,
		},
	}

	e := echo.New()
	user := map[string]string{
		"username": "test-user",
		"email":    "test-user@example.com",
		"password": "not a password",
	}
	body, err := json.Marshal(user)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	srv, err := NewServer(createTestDB(t), t.TempDir())
	if !assert.NoError(t, err) {
		t.Log(err)
		t.FailNow()
	}
	if assert.NoError(t, srv.PostUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.user)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewBuffer(body))
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			srv, err := NewServer(createTestDB(t), t.TempDir())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if assert.NoError(t, srv.PostUser(c)) {
				assert.Equal(t, tc.wantCode, rec.Code)
			}
		})
	}
}

func TestUserUpdateRoutes(t *testing.T) {
	db := createTestDB(t)
	user, err := database.CreateUser(db, "test-user-update", "test-user-update@example.com", "not a password")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	e := echo.New()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// unauthenticated user
	req := httptest.NewRequest(http.MethodPut, "/api/v1/user/username", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	srv, err := NewServer(db, t.TempDir())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if assert.NoError(t, srv.PutUserUsername(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/user/email", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.PutUserEmail(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/user/password", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	if assert.NoError(t, srv.PutUserPassword(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}

	// authenticated user
	session, err := database.CreateSession(db, user.Id)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	cookie := http.Cookie{
		Name:     "OBSYNC_SESSION_ID",
		Value:    session.SessionKey,
		Expires:  session.Expires,
		HttpOnly: true,
		Path:     "/",
	}

	testCases := []struct {
		name     string
		username string
		email    string
		password string
		wantCode int
	}{
		{
			name:     "update username",
			username: "new-username",
			wantCode: http.StatusOK,
		},
		{
			name:     "update username too long",
			username: strings.Repeat("a", 101),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "update email",
			email:    "new-email@example.com",
			wantCode: http.StatusOK,
		},
		{
			name:     "update email too long",
			email:    strings.Repeat("a", 201),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "update password",
			password: "new-password",
			wantCode: http.StatusOK,
		},
		{
			name:     "update password too short",
			password: "short",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var (
				req *http.Request
				err error
			)

			if len(tc.username) > 0 {
				req = httptest.NewRequest(http.MethodPut, "/api/v1/user/username", bytes.NewBufferString(tc.username))
			} else if len(tc.email) > 0 {
				req = httptest.NewRequest(http.MethodPut, "/api/v1/user/email", bytes.NewBufferString(tc.email))
			} else if len(tc.password) > 0 {
				req = httptest.NewRequest(http.MethodPut, "/api/v1/user/password", bytes.NewBufferString(tc.password))
			}

			req.AddCookie(&cookie)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if len(tc.username) > 0 {
				err = srv.PutUserUsername(c)
			} else if len(tc.email) > 0 {
				err = srv.PutUserEmail(c)
			} else if len(tc.password) > 0 {
				err = srv.PutUserPassword(c)
			}

			if assert.NoError(t, err) {
				assert.Equal(t, tc.wantCode, rec.Code)
			}

			user, err := database.GetUserById(db, user.Id)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			if rec.Code == 200 {
				if len(tc.username) > 0 {
					assert.Equal(t, tc.username, user.Username)
				} else if len(tc.email) > 0 {
					assert.Equal(t, tc.email, user.Email)
				} else if len(tc.password) > 0 {
					assert.NotEqual(t, tc.password, user.Passhash)
				}
			}
		})
	}

	assert.NoError(t, database.DeleteSession(db, session.SessionKey))
	assert.NoError(t, database.DeleteUser(db, user.Id))
}
