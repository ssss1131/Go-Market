package repo

import (
	"GoProduct/internal/domain"

	"gorm.io/gorm"
)

type Products struct {
	db *gorm.DB
}
type ProductFilter struct {
	Q        string
	Category string
	MinPrice *float64
	MaxPrice *float64
	Sort     string
	Limit    int
	Offset   int
}

func NewProducts(db *gorm.DB) *Products {
	return &Products{db: db}
}

func (r *Products) Create(p *domain.Product) error {
	return r.db.Create(p).Error
}

func (r *Products) GetByID(id uint) (*domain.Product, error) {
	var p domain.Product
	if err := r.db.Preload("Categories").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Products) List() ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Preload("Categories").Find(&products).Error; err != nil {
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

func (r *Products) ListFiltered(f ProductFilter) ([]domain.Product, error) {
	db := r.db.Model(&domain.Product{}).Preload("Categories")

	if f.Q != "" {
		like := "%" + f.Q + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	if f.MinPrice != nil {
		db = db.Where("price >= ?", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		db = db.Where("price <= ?", *f.MaxPrice)
	}

	if f.Category != "" {
		db = db.Joins("JOIN product_categories pc ON pc.product_id = products.id").
			Joins("JOIN categories c ON c.id = pc.category_id").
			Where("c.slug = ?", f.Category).
			Group("products.id")
	}

	switch f.Sort {
	case "price_asc":
		db = db.Order("price ASC")
	case "price_desc":
		db = db.Order("price DESC")
	case "newest":
		db = db.Order("created_at DESC")
	default:
		db = db.Order("id DESC")
	}

	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	var products []domain.Product
	if err := db.Limit(f.Limit).Offset(f.Offset).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *Products) SetCategories(productID uint, categoryIDs []uint) error {
	var categories []domain.Category
	if len(categoryIDs) > 0 {
		if err := r.db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			return err
		}
	}
	return r.db.Model(&domain.Product{ID: productID}).Association("Categories").Replace(categories)
}
