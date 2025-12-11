package repo

import (
	"GoCart/internal/domain"

	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetCart(userID uint) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

func (r *CartRepository) GetItem(userID, productID uint) (*domain.CartItem, error) {
	var item domain.CartItem
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepository) Create(item *domain.CartItem) error {
	return r.db.Create(item).Error
}

func (r *CartRepository) Update(item *domain.CartItem) error {
	return r.db.Save(item).Error
}

func (r *CartRepository) Delete(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&domain.CartItem{}).Error
}
