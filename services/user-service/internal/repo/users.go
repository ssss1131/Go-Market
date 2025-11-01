package repo

import (
	"GoMarket/internal/domain"
	"errors"
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

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	// TODO: не забыть убрать если хватит верхней проверки
	//var pgErr *pgconn.PgError
	//if errors.As(err, &pgErr) {
	//	if pgErr.Code == "23505" {
	//		return true
	//	}
	//}
	return false
}
