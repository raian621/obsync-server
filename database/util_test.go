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
	assert.NoError(t, ValidateHash(password, passhash))

	incorrect := "this is not the password"
	assert.Error(t, ErrInvalidPassword, ValidateHash(incorrect, passhash))
}

func TestValidEmail(t *testing.T) {
	t.Parallel()
	assert.True(t, validEmail("test@localhost"))
	assert.False(t, validEmail("test@localhost."))
	assert.False(t, validEmail(""))
}
