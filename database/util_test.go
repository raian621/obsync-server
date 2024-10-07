package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()
	password := "thisisnotasecurepassword"
	passhash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.Equal(t, len(passhash), 97)
}

func TestValidateHash(t *testing.T) {
	t.Parallel()
	password := "thisisnotasecurepassword"
	passhash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.Equal(t, len(passhash), 97)
	assert.NoError(t, ValidatePassword(password, passhash))

	incorrect := "this is not the password"
	assert.Error(t, ErrInvalidPassword, ValidatePassword(incorrect, passhash))
}
