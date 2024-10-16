package database

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setUpApiKeyTest(t *testing.T, dbName string) (*sql.DB, *User) {
	testdb, err := NewDB(fmt.Sprintf("%s?mode=memory", dbName))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, ApplyMigrations(testdb)) {
		t.FailNow()
	}
	user, err := CreateUser(testdb, "test-user", "test-user@example.com", "npt a password")
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return testdb, user
}

func tearDownApiKeyTest(t *testing.T, testdb *sql.DB, user *User, apiKeys []*ApiKey) {
	for _, apiKey := range apiKeys {
		_, err := testdb.Exec("DELETE FROM api_keys WHERE id=?", apiKey.Id)
		if !assert.NoError(t, err) {
			t.Fail()
		}
	}
	_, err := testdb.Exec("DELETE FROM users WHERE id=?", user.Id)
	if !assert.NoError(t, err) {
		t.Fail()
	}
}

func TestCreateApiKey(t *testing.T) {
	t.Parallel()

	testdb, user := setUpApiKeyTest(t, "test-create-api-key.db")

	// create API key
	key, err := generateKey(ApiKeyLength)
	assert.NoError(t, err)
	apiKey, err := CreateApiKey(testdb, user.Id, "test-name", key)
	assert.NoError(t, err)
	// check hash
	assert.NoError(t, ValidateHash(key, apiKey.Hash))

	// try to create an API key with a duplicate name
	key, err = generateKey(ApiKeyLength)
	assert.NoError(t, err)
	nilApiKey, err := CreateApiKey(testdb, user.Id, "test-name", key)
	assert.ErrorIs(t, err, ErrApiKeyExists)
	assert.Nil(t, nilApiKey)

	tearDownApiKeyTest(t, testdb, user, []*ApiKey{apiKey})
}

func TestDeleteApiKey(t *testing.T) {
	t.Parallel()

	testdb, user := setUpApiKeyTest(t, "test-delete-api-key.db")
	defer tearDownApiKeyTest(t, testdb, user, nil)
	key, err := generateKey(ApiKeyLength)
	assert.NoError(t, err)
	apiKey, err := CreateApiKey(testdb, user.Id, "test-name", key)
	assert.NoError(t, err)
	assert.NoError(t, DeleteApiKey(testdb, user.Id, apiKey.Name))

	// make sure the api key was actually deleted
	var count int
	row := testdb.QueryRow("SELECT COUNT(*) FROM api_keys WHERE id=?", apiKey.Id)
	assert.NoError(t, row.Scan(&count))
	assert.Equal(t, 0, count)

	// make sure that deleting an API key that doesn't exist doesn't throw an
	// error unexpectedly
	assert.NoError(t, DeleteApiKey(testdb, user.Id, "not a name in db"))
}

func TestSetApiKeyActivation(t *testing.T) {
	t.Parallel()

	testdb, user := setUpApiKeyTest(t, "test-set-api-key-activation.db")
	key, err := generateKey(ApiKeyLength)
	assert.NoError(t, err)
	apiKey, err := CreateApiKey(testdb, user.Id, "test name", key)
	assert.NoError(t, err)

	// deactivate API key
	assert.NoError(t, SetApiKeyActivation(testdb, user.Id, apiKey.Name, false))
	var count int
	row := testdb.QueryRow(
		"SELECT COUNT(*) FROM api_keys WHERE user_id=? AND name=? AND active=FALSE",
		user.Id,
		apiKey.Name,
	)
	assert.NoError(t, row.Scan(&count))
	assert.Equal(t, 1, count)

	// reactivate it
	assert.NoError(t, SetApiKeyActivation(testdb, user.Id, apiKey.Name, true))
	row = testdb.QueryRow(
		"SELECT COUNT(*) FROM api_keys WHERE user_id=? AND name=? AND active=TRUE",
		user.Id,
		apiKey.Name,
	)
	assert.NoError(t, row.Scan(&count))
	assert.Equal(t, 1, count)

	tearDownApiKeyTest(t, testdb, user, []*ApiKey{apiKey})
}
