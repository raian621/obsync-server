package database

import (
	"database/sql"
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
      name: "Username too long",
      username: strings.Repeat("a", 101),
      email: "test-user@localhost.com",
      password: "this is not a secure password",
      wantErr: ErrUsernameFormat,
    },
    {
      name: "Username too short",
      username: "",
      email: "test-user@localhost.com",
      password: "this is not a secure password",
      wantErr: ErrUsernameFormat,
    },
		{
			name:     "Email invalid",
			username: "test-user",
			email:    "test-userlocalhost.com",
			password: "this is not a secure password",
			wantErr: ErrEmailFormat,
		},
		{
			name:     "Email too short",
			username: "test-user",
			email:    "",
			password: "this is not a secure password",
			wantErr: ErrEmailFormat,
		},
		{
			name:     "Email too long",
			username: "test-user",
			email:    strings.Repeat("a", 200) + "@email.com",
			password: "this is not a secure password",
			wantErr: ErrEmailFormat,
		},
    {
      name: "Password too short",
      username: "test-user",
      email: "test-user@email.com",
      password: "shorts!",
      wantErr: ErrPasswordLength,
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
    email = "test-user1@email.com"
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
  users := []*User {
    {Username: "alice",  Email: "alice@example.com"},
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

  testCases := []struct{
    name       string
    username   string
    email      string
    id         uint64
    wantUserId uint64
    wantErr    error
  }{
    {
      name: "user found by username",
      username: users[2].Username,
      wantUserId: users[2].Id,
    },
    {
      name: "user not found by username",
      username: "jeff",
      wantErr: sql.ErrNoRows,
    },
    {
      name: "username too long",
      username: strings.Repeat("j", 101),
      wantErr: ErrUsernameFormat,
    },
    {
      name: "user found by email",
      email: users[2].Email,
      wantUserId: users[2].Id,
    },
    {
      name: "user not found by email",
      email: "jeff@example.com",
      wantErr: sql.ErrNoRows,
    },
    {
      name: "email too long",
      email: strings.Repeat("j", 200)+"@example.com",
      wantErr: ErrEmailFormat,
    },
    {
      name: "user found by id",
      id: users[2].Id,
      wantUserId: users[2].Id,
    },
    {
      name: "user not found by id",
      id: 20,
      wantErr: sql.ErrNoRows,
    },
  }

  for _, tc := range testCases {
    tc := tc

    t.Run(tc.name, func(t *testing.T) {
      var (
        err error
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

  // incorrect username
  session, err = LoginUser(testdb, "not the username", "password123")
  assert.ErrorIs(t, ErrIncorrectCredentials, err)
}
