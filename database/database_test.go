package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDB(t *testing.T) {
	t.Parallel()
	db, err := NewDB("test-new-db.db?mode=memory")
	assert.NotNil(t, db)
	assert.NoError(t, err)
}

func TestSetDB(t *testing.T) {
	t.Parallel()
	testdb, err := NewDB("test-set-db.db?mode=memory")
	assert.NoError(t, err)
	SetDB(testdb)
	assert.Equal(t, testdb, db)
}

func TestGetDB(t *testing.T) {
	// should return default db
	defaultDB, err := GetDB()
	assert.NoError(t, err)
	assert.Equal(t, defaultDB, db)
	db.Close()
	db = nil

	// set db then get it using GetDB
	testdb, err := NewDB("test-get-db.db?mode=memory")
	assert.NoError(t, err)
	SetDB(testdb)
	gotdb, err := GetDB()
	assert.NoError(t, err)
	assert.Equal(t, gotdb, testdb)
}
