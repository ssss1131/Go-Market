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
	UserID      uint
	Name        string
	Description string
	Price       float64
	CategoryIDs []uint
}

type UpdateProductInput struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	CategoryIDs []uint
}

type ProductService struct {
	products *repo.Products
}

type ListProductsFilter struct {
	Q        string
	Category string
	MinPrice *float64
	MaxPrice *float64
	Sort     string
	Limit    int
	Offset   int
}

func NewProductService(products *repo.Products) *ProductService {
	return &ProductService{products: products}
}

func (s *ProductService) CreateProduct(in CreateProductInput) (*domain.Product, error) {
	p := &domain.Product{
		UserID:      in.UserID,
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
	}

	if err := s.products.Create(p); err != nil {
		return nil, err
	}

	if err := s.products.SetCategories(p.ID, in.CategoryIDs); err != nil {
		return nil, err
	}

	p, err := s.products.GetByID(p.ID)
	if err != nil {
		return nil, err
	}

	return p, nil

}

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

func (s *ProductService) ListProducts() ([]domain.Product, error) {
	return s.products.List()
}

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

	if err := s.products.SetCategories(p.ID, in.CategoryIDs); err != nil {
		return nil, err
	}

	p, err = s.products.GetByID(p.ID)
	if err != nil {
		return nil, err
	}

	return p, nil

}

func (s *ProductService) DeleteProduct(id uint) error {
	_, err := s.products.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errNotFound
		}
		return err
	}

	return s.products.Delete(id)
}

func (s *ProductService) ListProductsFiltered(f ListProductsFilter) ([]domain.Product, error) {
	return s.products.ListFiltered(repo.ProductFilter{
		Q:        f.Q,
		Category: f.Category,
		MinPrice: f.MinPrice,
		MaxPrice: f.MaxPrice,
		Sort:     f.Sort,
		Limit:    f.Limit,
		Offset:   f.Offset,
	})
}
