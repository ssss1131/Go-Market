package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword_Success(t *testing.T) {
	pw := "mysecret123"
	hash, err := HashPassword(pw)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, pw, hash)
}

func TestCheckPassword_Success(t *testing.T) {
	pw := "mypassword"
	hash, _ := HashPassword(pw)

	err := CheckPassword(hash, pw)
	assert.NoError(t, err)
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	pw := "correct"
	wrong := "incorrect"
	hash, _ := HashPassword(pw)

	err := CheckPassword(hash, wrong)
	assert.Error(t, err)
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	pw := "pass"
	invalidHash := "not-a-valid-bcrypt-hash"

	err := CheckPassword(invalidHash, pw)
	assert.Error(t, err)
}
