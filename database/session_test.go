package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginUser(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("test-login-user.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))

	user, err := CreateUser(
		testdb,
		"test-user",
		"test-user@gmail.com",
		"password123",
	)
	assert.NoError(t, err)

	// happy path, successful login
	session, err := LoginUser(testdb, user.Username, "password123")
	assert.NoError(t, err)
	assert.Equal(t, session.UserId, user.Id)
	err = LogoutUser(testdb, session.SessionKey)
	assert.NoError(t, err)

	// incorrect password
	session, err = LoginUser(testdb, user.Username, "not the password")
	assert.ErrorIs(t, ErrIncorrectCredentials, err)
	assert.Nil(t, session)

	// incorrect username
	session, err = LoginUser(testdb, "not the username", "password123")
	assert.ErrorIs(t, ErrIncorrectCredentials, err)
	assert.Nil(t, session)
}

func TestCreateSyncFile(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("test-create-sync-file.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))

	user, err := CreateUser(testdb, "test-user-1", "test-user@gmail.com", "password123")
	assert.NoError(t, err)
	filepath := "/cool/filepath"
	etag := "ff3e4b618b07a1f9b2ab04c201ae6613"

	syncFile, err := CreateSyncFile(testdb, filepath, etag, user.Id)
	assert.NoError(t, err)
	assert.Equal(t, syncFile.UserId, user.Id)
	assert.Equal(t, syncFile.Filepath, filepath)
	assert.Equal(t, syncFile.Etag, etag)

	assert.NoError(t, testdb.Close())
}
