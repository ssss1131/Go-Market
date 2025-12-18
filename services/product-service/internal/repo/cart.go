package repo

import (
	"GoProduct/internal/domain"

	"gorm.io/gorm"
)

type Cart struct {
	db *gorm.DB
}

func NewCart(db *gorm.DB) *Cart {
	return &Cart{db: db}
}

func (r *Cart) GetByUserID(userID uint) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *Cart) GetByUserAndProduct(userID, productID uint) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	return &item, err
}

func (r *Cart) Create(item *domain.CartItem) error {
	return r.db.Create(item).Error
}

func (r *Cart) Update(item *domain.CartItem) error {
	return r.db.Save(item).Error
}

func (r *Cart) Delete(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&domain.CartItem{}).Error
}
