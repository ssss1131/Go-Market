package service

import (
	"GoProduct/internal/domain"
	"GoProduct/internal/repo"
	"errors"

	"gorm.io/gorm"
)

var errNotFound = errors.New("product_not_found")

func IsNotFound(err error) bool { return errors.Is(err, errNotFound) }

type CreateProductInput struct {
	Name        string
	Description string
	Price       float64
}

type UpdateProductInput struct {
	ID          uint
	Name        string
	Description string
	Price       float64
}

type ProductService struct {
	products *repo.Products
}

func NewProductService(products *repo.Products) *ProductService {
	return &ProductService{products: products}
}

// CreateProduct — создать новый продукт
func (s *ProductService) CreateProduct(in CreateProductInput) (*domain.Product, error) {
	p := &domain.Product{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
	}

	if err := s.products.Create(p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetProduct — получить продукт по ID
func (s *ProductService) GetProduct(id uint) (*domain.Product, error) {
	p, err := s.products.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errNotFound
		}
		return nil, err
	}
	return p, nil
}

// ListProducts — список всех продуктов
func (s *ProductService) ListProducts() ([]domain.Product, error) {
	return s.products.List()
}

// UpdateProduct — обновить продукт
func (s *ProductService) UpdateProduct(in UpdateProductInput) (*domain.Product, error) {
	p, err := s.products.GetByID(in.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errNotFound
		}
		return nil, err
	}

	p.Name = in.Name
	p.Description = in.Description
	p.Price = in.Price

	if err := s.products.Update(p); err != nil {
		return nil, err
	}

	return p, nil
}

// DeleteProduct — удалить продукт
func (s *ProductService) DeleteProduct(id uint) error {
	// Если хочешь отличать "нет такого" — можно сначала дернуть GetByID
	_, err := s.products.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errNotFound
		}
		return err
	}

	return s.products.Delete(id)
}
