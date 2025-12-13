package service

import (
	"GoProduct/internal/domain"
	"GoProduct/internal/repo"
	"errors"

	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New("product not found")

type CartService struct {
	cartRepo       *repo.Cart
	productService *ProductService // Используем напрямую, а не HTTP!
}

func NewCartService(cartRepo *repo.Cart, productService *ProductService) *CartService {
	return &CartService{
		cartRepo:       cartRepo,
		productService: productService,
	}
}

func (s *CartService) AddItem(userID, productID uint, quantity int) error {
	// Проверяем, что продукт существует (через ProductService напрямую!)
	_, err := s.productService.GetProduct(productID)
	if err != nil {
		if IsNotFound(err) {
			return ErrProductNotFound
		}
		return err
	}

	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// Проверяем, есть ли уже этот товар в корзине
	existing, err := s.cartRepo.GetByUserAndProduct(userID, productID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing.ID != 0 {
		// Обновляем количество
		existing.Quantity += quantity
		return s.cartRepo.Update(existing)
	}

	// Создаем новую запись
	item := &domain.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}
	return s.cartRepo.Create(item)
}

func (s *CartService) UpdateItem(userID, productID uint, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	item, err := s.cartRepo.GetByUserAndProduct(userID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("item not in cart")
		}
		return err
	}

	item.Quantity = quantity
	return s.cartRepo.Update(item)
}

func (s *CartService) DeleteItem(userID, productID uint) error {
	return s.cartRepo.Delete(userID, productID)
}

func (s *CartService) GetCart(userID uint) ([]domain.CartItem, error) {
	return s.cartRepo.GetByUserID(userID)
}
