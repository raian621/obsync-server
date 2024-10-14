package database

import (
	"database/sql"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("test-create-user.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))

	testCases := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "Create user with good parameters",
			username: "test-user",
			email:    "test-user@localhost.com",
			password: "this is not a secure password",
			wantErr:  nil,
		},
		{
			name:     "Username too long",
			username: strings.Repeat("a", 101),
			email:    "test-user@localhost.com",
			password: "this is not a secure password",
			wantErr:  ErrUsernameFormat,
		},
		{
			name:     "Username too short",
			username: "",
			email:    "test-user@localhost.com",
			password: "this is not a secure password",
			wantErr:  ErrUsernameFormat,
		},
		{
			name:     "Email invalid",
			username: "test-user",
			email:    "test-userlocalhost.com",
			password: "this is not a secure password",
			wantErr:  ErrEmailFormat,
		},
		{
			name:     "Email too short",
			username: "test-user",
			email:    "",
			password: "this is not a secure password",
			wantErr:  ErrEmailFormat,
		},
		{
			name:     "Email too long",
			username: "test-user",
			email:    strings.Repeat("a", 200) + "@email.com",
			password: "this is not a secure password",
			wantErr:  ErrEmailFormat,
		},
		{
			name:     "Password too short",
			username: "test-user",
			email:    "test-user@email.com",
			password: "shorts!",
			wantErr:  ErrPasswordLength,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			user, err := CreateUser(testdb, tc.username, tc.email, tc.password)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, tc.wantErr, err)
			}
			if err == nil {
				assert.Equal(t, tc.username, user.Username)
				assert.Equal(t, tc.email, user.Email)
				assert.NoError(t, ValidatePassword(tc.password, user.Passhash))
				_, err = testdb.Exec(
					"DELETE FROM users WHERE id=:id",
					sql.Named("id", user.Id),
				)
				assert.NoError(t, err)
			}
		})
	}

	var (
		username = "test-user-1"
		email    = "test-user1@email.com"
		password = "not a secure password"
	)
	// all usernames must be unique
	user1, err := CreateUser(testdb, username, email, password)
	assert.NoError(t, err)
	user2, err := CreateUser(
		testdb, user1.Username, "test-user-2@example.com", password)
	assert.Error(t, err)
	assert.Nil(t, user2)

	// all emails must be unique
	user2, err = CreateUser(
		testdb,
		"test-user2@email.com",
		user1.Email,
		password,
	)
	assert.Error(t, err)
	assert.Nil(t, user2)
}

func TestGetUserBy(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("test-get-user-by.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))
	users := []*User{
		{Username: "alice", Email: "alice@example.com"},
		{Username: "bob", Email: "bob@example.com"},
		{Username: "charlie", Email: "charlie@example.com"},
		{Username: "david", Email: "david@example.com"},
		{Username: "emily", Email: "emily@example.com"},
		{Username: "frank", Email: "frank@example.com"},
		{Username: "grace", Email: "grace@example.com"},
	}
	for i, user := range users {
		dbUser, err := CreateUser(testdb, user.Username, user.Email, "shared password lol")
		assert.NoError(t, err)
		users[i] = dbUser
	}

	testCases := []struct {
		name       string
		username   string
		email      string
		id         uint64
		wantUserId uint64
		wantErr    error
	}{
		{
			name:       "user found by username",
			username:   users[2].Username,
			wantUserId: users[2].Id,
		},
		{
			name:     "user not found by username",
			username: "jeff",
			wantErr:  sql.ErrNoRows,
		},
		{
			name:     "username too long",
			username: strings.Repeat("j", 101),
			wantErr:  ErrUsernameFormat,
		},
		{
			name:       "user found by email",
			email:      users[2].Email,
			wantUserId: users[2].Id,
		},
		{
			name:    "user not found by email",
			email:   "jeff@example.com",
			wantErr: sql.ErrNoRows,
		},
		{
			name:    "email too long",
			email:   strings.Repeat("j", 200) + "@example.com",
			wantErr: ErrEmailFormat,
		},
		{
			name:       "user found by id",
			id:         users[2].Id,
			wantUserId: users[2].Id,
		},
		{
			name:    "user not found by id",
			id:      20,
			wantErr: sql.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			var (
				err  error
				user *User
			)
			if len(tc.username) > 0 {
				user, err = GetUserByUsername(testdb, tc.username)
			}
			if len(tc.email) > 0 {
				user, err = GetUserByEmail(testdb, tc.email)
			}
			if tc.id != 0 {
				user, err = GetUserById(testdb, tc.id)
			}
			assert.ErrorIs(t, err, tc.wantErr)
			if err == nil {
				assert.Equal(t, tc.wantUserId, user.Id)
			}
		})
	}
}

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
	assert.Equal(t, session.Id, user.Id)
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

func TestGetSyncFileBy(t *testing.T) {
	t.Parallel()

	testdb, err := NewDB("get-sync-file-by.db?mode=memory")
	assert.NoError(t, err)
	assert.NoError(t, ApplyMigrations(testdb))

	user, err := CreateUser(testdb, "test-user", "test-user@example.com", "not a secure password")
	assert.NoError(t, err)
	syncfiles := []*SyncFile{
		{Filepath: "/folder/file1.md", Etag: "f0f9ef0cbb7e0d836aea4a4c6fe6420a"},
		{Filepath: "/folder/file2.md", Etag: "8c42bf48c4b5d8553ad3ab5b30b484df"},
		{Filepath: "/folder/file3.md", Etag: "37e904b58a2a5e61babc827ded3a828d"},
	}
	for i, syncfile := range syncfiles {
		syncfiles[i], err = CreateSyncFile(testdb, syncfile.Filepath, syncfile.Etag, user.Id)
		assert.NoError(t, err)
	}

	testCases := []struct {
		name         string
		id           int64
		userId       int64
		filepath     string
		wantSyncfile *SyncFile
		wantErr      error
	}{
		{
			name:         "found by id",
			id:           int64(syncfiles[1].Id),
			wantSyncfile: syncfiles[1],
		},
		{
			name:    "not found by id",
			id:      int64(syncfiles[2].Id + 1),
			wantErr: ErrNoResults,
		},
		{
			name:         "found by filepath",
			filepath:     "/folder/file1.md",
			wantSyncfile: syncfiles[0],
		},
		{
			name:     "not found by filepath",
			filepath: "/folder/file4.md",
			wantErr:  ErrNoResults,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			var (
				syncfile *SyncFile
				err      error
			)

			if tc.id > 0 {
				syncfile, err = GetSyncFileById(testdb, uint64(tc.id))
			}
			if len(tc.filepath) > 0 {
				syncfile, err = GetSyncFileByFilepath(testdb, tc.filepath)
			}

			assert.ErrorIs(t, err, tc.wantErr)
			if tc.wantSyncfile != nil && tc.wantErr == nil {
				assert.Equal(t, *tc.wantSyncfile, *syncfile)
			}
		})
	}

	// sync files for user are found
	dbSyncfiles, err := GetSyncFilesByUserId(testdb, user.Id)
	assert.NoError(t, err)
	slices.SortFunc(dbSyncfiles, func(a *SyncFile, b *SyncFile) int {
		return int(a.Id - b.Id)
	})
	assert.ElementsMatch(t, syncfiles, dbSyncfiles)

	// sync files for user are not found
	dbSyncfiles, err = GetSyncFilesByUserId(testdb, user.Id+1)
	assert.ErrorIs(t, err, ErrNoResults)
	assert.Empty(t, dbSyncfiles)

	for _, syncfile := range syncfiles {
		_, err := testdb.Exec("DELETE FROM file_syncs WHERE id=?", syncfile.Id)
		assert.NoError(t, err)
	}
	_, err = testdb.Exec("DELETE FROM users WHERE id=?", user.Id)
	assert.NoError(t, err)
}
