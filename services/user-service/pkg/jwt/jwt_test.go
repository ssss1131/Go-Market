package jwtutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSigner_NewAccess_And_Verify(t *testing.T) {
	s := NewSigner("test-secret", time.Minute)

	token, jti, err := s.NewAccess(42, "john@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, jti)

	claims, err := s.Verify(token)
	assert.NoError(t, err)

	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "john@example.com", claims.Email)
	assert.Equal(t, claims.ID, jti)
}

func TestSigner_Verify_InvalidToken(t *testing.T) {
	s := NewSigner("secret", time.Minute)

	_, err := s.Verify("not-a-real-token")
	assert.Error(t, err)
}

func TestSigner_Verify_WrongSecret(t *testing.T) {
	s1 := NewSigner("secret1", time.Minute)
	s2 := NewSigner("secret2", time.Minute)

	token, _, _ := s1.NewAccess(1, "a@b.com")

	_, err := s2.Verify(token)
	assert.Error(t, err)
}

func TestSigner_TokenExpired(t *testing.T) {
	s := NewSigner("secret", time.Millisecond*10)

	token, _, err := s.NewAccess(5, "user@mail.com")
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 20)

	_, err = s.Verify(token)
	assert.Error(t, err)
}
