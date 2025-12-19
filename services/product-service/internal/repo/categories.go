package repo

import (
	"GoProduct/internal/domain"

	"gorm.io/gorm"
)

type Categories struct {
	db *gorm.DB
}

func NewCategories(db *gorm.DB) *Categories {
	return &Categories{db: db}
}

func (r *Categories) Create(c *domain.Category) error {
	return r.db.Create(c).Error
}

func (r *Categories) GetByID(id uint) (*domain.Category, error) {
	var c domain.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Categories) GetBySlug(slug string) (*domain.Category, error) {
	var c domain.Category
	if err := r.db.Where("slug = ?", slug).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Categories) List() ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *Categories) Update(c *domain.Category) error {
	return r.db.Save(c).Error
}

func (r *Categories) Delete(id uint) error {
	return r.db.Delete(&domain.Category{}, id).Error
}
