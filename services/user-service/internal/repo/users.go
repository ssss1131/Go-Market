package repo

import (
	"GoMarket/internal/domain"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Users struct {
	db *gorm.DB
}

func NewUsers(db *gorm.DB) *Users { return &Users{db: db} }

func (u *Users) Create(user *domain.User) error {
	return u.db.Create(user).Error
}

func IsUniqueEmail(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
	}
	return false
}

func (u *Users) ByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := u.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
