package repo

import (
	"errors"
	"testing"

	"GoMarket/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// FakeUsersRepo simulates a Users repository for testing without GORM
type FakeUsersRepo struct {
	store map[string]domain.User
	err   error
}

func NewFakeUsersRepo() *FakeUsersRepo {
	return &FakeUsersRepo{store: make(map[string]domain.User)}
}

func (f *FakeUsersRepo) Create(user *domain.User) error {
	if f.err != nil {
		return f.err
	}
	f.store[user.Email] = *user
	return nil
}

func (f *FakeUsersRepo) ByEmail(email string) (*domain.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	u, ok := f.store[email]
	if !ok {
		return nil, errors.New("record not found")
	}
	return &u, nil
}

func TestUsers_Create(t *testing.T) {
	fake := NewFakeUsersRepo()

	user := &domain.User{Email: "a@b.com"}
	// Use fake directly
	err := fake.Create(user)
	assert.NoError(t, err)
}

func TestIsUniqueEmail(t *testing.T) {
	assert.False(t, IsUniqueEmail(nil))
	assert.False(t, IsUniqueEmail(errors.New("x")))
	pgErr := &pgconn.PgError{Code: "99999"}
	assert.False(t, IsUniqueEmail(pgErr))
	pgErr2 := &pgconn.PgError{Code: "23505"}
	assert.True(t, IsUniqueEmail(pgErr2))
}

func TestUsers_ByEmail(t *testing.T) {
	fake := NewFakeUsersRepo()
	user := domain.User{Email: "john@doe.com"}
	_ = fake.Create(&user)

	res, err := fake.ByEmail("john@doe.com")
	assert.NoError(t, err)
	assert.Equal(t, "john@doe.com", res.Email)

	res, err = fake.ByEmail("notfound@doe.com")
	assert.Nil(t, res)
	assert.Error(t, err)
}
