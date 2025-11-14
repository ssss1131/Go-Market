package repo

import (
	"GoProduct/internal/domain"

	"gorm.io/gorm"
)

type Products struct {
	db *gorm.DB
}

func NewProducts(db *gorm.DB) *Products {
	return &Products{db: db}
}

func (r *Products) Create(p *domain.Product) error {
	return r.db.Create(p).Error
}

func (r *Products) GetByID(id uint) (*domain.Product, error) {
	var p domain.Product
	if err := r.db.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Products) List() ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *Products) Update(p *domain.Product) error {
	return r.db.Save(p).Error
}

func (r *Products) Delete(id uint) error {
	return r.db.Delete(&domain.Product{}, id).Error
}
